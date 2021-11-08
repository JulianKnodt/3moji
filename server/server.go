package main

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	emojisSentAt    = expvar.NewMap("emojisSentAt")
	emojisSentCount = expvar.NewMap("emojisSentCount")
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	s := NewServer()
	if err := s.Serve(":" + port); err != nil {
		fmt.Printf("Server exited: %v", err)
	}
}

// Server is a stateful server which represents the current state of the
// universe. For now it should just be a massive struct which contains everything for
// simplicity.
type Server struct {
	// mu guards the map of struct who have signed up
	mu sync.Mutex
	// LoggedIn map[Email]LoginToken

	// In-memory map of recipient to Messages
	UserToMessages map[Uuid]map[Uuid]struct{}

	// TODO this isn't really in use so it's okay that it gets reset
	// List of friends for a given user: user -> their friends
	Friends map[Uuid]map[Uuid]struct{}

	// Replies waiting for a given user
	UserToReplies map[Uuid][]Uuid
	Replies       map[Uuid]*MessageReply

	EmojiSendTime map[EmojiContent]float64

	// A long living redis client for using as a persistent store.
	RedisClient *redis.Client
}

func NewServer() *Server {
	redisURL := os.Getenv("REDIS_URL")
	user := ""
	password := ""
	if redisURL == "" {
		redisURL = ":6379"
	} else {
		u, err := url.Parse(redisURL)
		if err != nil {
			fmt.Println(err)
		}
		redisURL = u.Host
		user = u.User.Username()
		password, _ = u.User.Password()
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     strings.TrimSuffix(redisURL, ":"),
		Username: user,
		// TODO need to set a password through secret.
		Password: password,
		DB:       0,
	})
	// ping the local redis database
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		fmt.Printf("Failed to open ping redis: %v\n", err)
	}
	return &Server{
		// SignedUp:        map[Email]*User{},
		// LoggedIn: map[Email]LoginToken{},

		UserToMessages: map[Uuid]map[Uuid]struct{}{},

		Friends: map[Uuid]map[Uuid]struct{}{},

		UserToReplies: map[Uuid][]Uuid{},
		Replies:       map[Uuid]*MessageReply{},

		EmojiSendTime: map[EmojiContent]float64{},

		RedisClient: rdb,
	}
}

func (srv *Server) Serve(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sign_up/", srv.SignUpHandler())
	mux.HandleFunc("/api/v1/login/", srv.LoginHandler())

	mux.HandleFunc("/api/v1/friend/", srv.FriendHandler())
	mux.HandleFunc("/api/v1/groups/", srv.GroupHandler())

	mux.HandleFunc("/api/v1/list_friends/", srv.ListPeopleHandler())
	mux.HandleFunc("/api/v1/list_groups/", srv.ListGroupHandler())

	mux.HandleFunc("/api/v1/send_msg/", srv.SendMsgHandler())
	mux.HandleFunc("/api/v1/recv_msg/", srv.RecvMsgHandler())
	mux.HandleFunc("/api/v1/ack_msg/", srv.AckMsgHandler())

	mux.HandleFunc("/api/v1/recs/", srv.RecommendationHandler())

	mux.HandleFunc("/api/v1/push_token/", srv.PushNotifTokenHandler())

	mux.HandleFunc("/api/v1/summary/", srv.SummaryHandler())

	mux.Handle("/debug/vars", expvar.Handler())

	s := http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("Listening on", s.Addr, "...")
	return s.ListenAndServe()
}

func (s *Server) AddMessage(ctx context.Context, msg *Message) error {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return nil
	}
	duration := time.Second * time.Duration(msg.TTL)
	return s.RedisClient.Set(ctx, MessageRedisKey(msg.Uuid), msgJSON, duration).Err()
}

func (s *Server) GetMessage(ctx context.Context, uuid Uuid) (*Message, error) {
	msgJSON, err := s.RedisClient.Get(ctx, MessageRedisKey(uuid)).Bytes()
	if err != nil {
		return nil, err
	}
	if len(msgJSON) == 0 {
		return nil, nil
	}
	var message Message
	if err = json.Unmarshal(msgJSON, &message); err != nil {
		return nil, err
	}
	return &message, nil
}

func (s *Server) DeleteMessage(ctx context.Context, uuid Uuid) error {
	return s.RedisClient.HDel(ctx, "messages", uuid.String()).Err()
}

func (s *Server) AddUser(ctx context.Context, user *User) error {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.RedisClient.HSet(ctx, "users", user.Uuid.String(), userJSON).Err()
}

func (s *Server) GetUser(ctx context.Context, uuid Uuid) (*User, error) {
	userJSON, err := s.RedisClient.HGet(ctx, "users", uuid.String()).Bytes()
	if err != nil {
		return nil, err
	}
	if len(userJSON) == 0 {
		return nil, nil
	}
	var user User
	if err = json.Unmarshal(userJSON, &user); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal user: %v", err)
	}
	return &user, nil
}

func (s *Server) GetUsers(ctx context.Context) ([]User, error) {
	userJSONs, err := s.RedisClient.HVals(ctx, "users").Result()
	if err != nil {
		return nil, err
	}
	if len(userJSONs) == 0 {
		return nil, nil
	}
	out := make([]User, len(userJSONs))
	for i, userJSON := range userJSONs {
		if err = json.Unmarshal([]byte(userJSON), &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Server) AddGroup(ctx context.Context, group *Group) error {
	groupJSON, err := json.Marshal(group)
	if err != nil {
		return err
	}
	return s.RedisClient.HSet(ctx, "groups", group.Uuid.String(), groupJSON).Err()
}

func (s *Server) DeleteGroup(ctx context.Context, uuid Uuid) error {
	err := s.RedisClient.HDel(ctx, "groups", uuid.String()).Err()
	if err != nil {
		return err
	}
	groupUserKey := fmt.Sprintf("%s_group_users", uuid)
	return s.RedisClient.Del(ctx, groupUserKey).Err()
}

func (s *Server) GetGroup(ctx context.Context, uuid Uuid) (*Group, error) {
	groupJSON, err := s.RedisClient.HGet(ctx, "groups", uuid.String()).Bytes()
	if err != nil {
		return nil, err
	}
	if len(groupJSON) == 0 {
		return nil, nil
	}
	var group Group
	if err = json.Unmarshal(groupJSON, &group); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal user: %v", err)
	}
	return &group, nil
}

func (s *Server) GetGroups(ctx context.Context) ([]Group, error) {
	groupJSONs, err := s.RedisClient.HVals(ctx, "groups").Result()
	if err != nil {
		return nil, err
	}
	if len(groupJSONs) == 0 {
		return nil, nil
	}
	out := make([]Group, len(groupJSONs))
	for i, groupJSONs := range groupJSONs {
		if err = json.Unmarshal([]byte(groupJSONs), &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Adds the uuid of a user into a group.
func (s *Server) AddUserToGroup(ctx context.Context, user, group Uuid) error {
	groupUserKey := fmt.Sprintf("%s_group_users", group)
	return s.RedisClient.SAdd(ctx, groupUserKey, user.String()).Err()
}

func (s *Server) DeleteUserFromGroup(ctx context.Context, user, group Uuid) error {
	groupUserKey := fmt.Sprintf("%s_group_users", group)
	return s.RedisClient.SRem(ctx, groupUserKey, user.String()).Err()
}

func (s *Server) UserIsMemberOfGroup(ctx context.Context, user, group Uuid) (bool, error) {
	groupUserKey := fmt.Sprintf("%s_group_users", group)
	return s.RedisClient.SIsMember(ctx, groupUserKey, user.String()).Result()
}

// Finds the uuid of all users in a group.
func (s *Server) UsersInGroup(ctx context.Context, group Uuid) ([]Uuid, error) {
	groupUserKey := fmt.Sprintf("%s_group_users", group)
	uuidStrings, err := s.RedisClient.SMembers(ctx, groupUserKey).Result()
	if err != nil {
		return nil, err
	}
	uuids := make([]Uuid, len(uuidStrings))
	for i, uuidString := range uuidStrings {
		uuids[i], err = UuidFromString(uuidString)
		if err != nil {
			return nil, err
		}
	}
	return uuids, nil
}

func (s *Server) SignUp(ctx context.Context, userEmail Email, userName string, hashedPassword string) (Uuid, error) {
	exists, err := s.RedisClient.HExists(ctx, "signed_up", string(userEmail)).Result()
	if err != nil {
		return Uuid(0), fmt.Errorf("Error signing up: %v", err)
	} else if exists {
		return Uuid(0), fmt.Errorf("User already exists")
	}

	uuid, err := generateUuid()
	if err != nil {
		return Uuid(0), err
	}
	user := &User{
		Name:  userName,
		Email: userEmail,
		Uuid:  uuid,
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		return Uuid(0), fmt.Errorf("Failed to marshal user: %v", err)
	}
	// TODO below needs to be in a transaction with above as well?
	s.RedisClient.HSet(ctx, "signed_up", string(userEmail), userJSON)
	s.RedisClient.HSet(ctx, "hashed_passwords", uuid.String(), hashedPassword)
	s.RedisClient.HSet(ctx, "users", uuid.String(), userJSON)

	return uuid, nil
}

func (s *Server) Login(ctx context.Context, userEmail Email, hashedPassword string) (LoginToken, error) {
	if hashedPassword == "" {
		return LoginToken{}, fmt.Errorf("password must not be empty")
	}

	userJSON, err := s.RedisClient.HGet(ctx, "signed_up", string(userEmail)).Bytes()
	if err != nil {
		return LoginToken{}, err
	} else if len(userJSON) == 0 {
		// Show generic error message, but user does not exist
		return LoginToken{}, fmt.Errorf("Something wrong with login")
	}
	var user User
	if err = json.Unmarshal(userJSON, &user); err != nil {
		return LoginToken{}, fmt.Errorf("Something wrong with login")
	}
	if user.Email != userEmail {
		return LoginToken{}, fmt.Errorf("Something wrong with login")
	}
	existing, err := s.RedisClient.HGet(ctx, "hashed_passwords", user.Uuid.String()).Result()
	if err != nil || existing != hashedPassword {
		// Show generic error message, but password isn't right
		return LoginToken{}, fmt.Errorf("Something wrong with login")
	}

	// TODO check collisions of the uuid and retry or crash
	uuid, err := generateUuid()
	if err != nil {
		return LoginToken{}, err
	}

	loginToken := LoginToken{
		ValidUntil: time.Now().Add(5 * 24 * time.Hour).Unix(),
		Uuid:       uuid,
		UserEmail:  userEmail,
	}
	if err = s.setLoggedIn(ctx, loginToken); err != nil {
		return LoginToken{}, err
	}
	return loginToken, nil
}

func (s *Server) setLoggedIn(ctx context.Context, loginToken LoginToken) error {
	loginTokenJSON, err := json.Marshal(loginToken)
	if err != nil {
		return err
	}
	loginTokenKey := fmt.Sprintf("%s_login_token", loginToken.UserEmail)
	duration := time.Unix(loginToken.ValidUntil, 0).Sub(time.Now())
	return s.RedisClient.Set(ctx, loginTokenKey, loginTokenJSON, duration).Err()
}

// Checks that a login token is correct, and matches the currently existing token kept on the
// token.
func (s *Server) ValidateLoginToken(token LoginToken) error {
	if token.Expired() {
		return fmt.Errorf(
			"Token has expired, was valid until %v & is now %v",
			time.Unix(token.ValidUntil, 0), time.Now(),
		)
	}
	loginTokenKey := fmt.Sprintf("%s_login_token", token.UserEmail)
	tokenJSON, err := s.RedisClient.Get(context.TODO(), loginTokenKey).Bytes()
	if err != nil {
		return err
	}
	var existingToken LoginToken
	if err = json.Unmarshal(tokenJSON, &existingToken); err != nil {
		return err
	}
	if existingToken != token {
		return fmt.Errorf("Tokens do not match want: %v, got: %v", existingToken, token)
	}
	if existingToken.Expired() {
		return fmt.Errorf(
			"Token has expired, was valid until %v & is now %v",
			time.Unix(existingToken.ValidUntil, 0),
			time.Now(),
		)
	}
	return nil
}

// Given a login token, it will return the user who used that login token. mu should not be
// held.
func (s *Server) UserFor(ctx context.Context, token LoginToken) (*User, bool) {
	userJSON, err := s.RedisClient.HGet(ctx, "signed_up", string(token.UserEmail)).Bytes()
	if err != nil || len(userJSON) == 0 {
		return nil, false
	}
	var user User
	if err := json.Unmarshal(userJSON, &user); err != nil {
		return nil, false
	}
	return &user, true
}

func (s *Server) MessageForReply(ctx context.Context, reply *MessageReply) (*Message, error) {
	return s.GetMessage(ctx, reply.Message.Uuid)
}

// time is 0 -> 24, returns two values which represents modular clock position
func to2DTimeModular(t float64) (float64, float64) {
	adj := t * math.Pi / 12.0
	return math.Sincos(adj)
}

func from2DTimeModular(u, v float64) float64 {
	// normalize u and v
	dist := math.Sqrt(u*u + v*v)
	if math.Abs(dist) < 1e-5 {
		dist = 1e-5
	}
	return 12 / math.Pi * math.Asin(u/dist)
}

func weightedAverage(old, newVal, alpha float64) float64 {
	return old*(1-alpha) + newVal*alpha
}

func distance(u1, u2, v1, v2 float64) float64 {
	du := u1 - u2
	dv := v1 - v2
	return du*du + dv*dv
}

func (s *Server) LogEmojiContent(e EmojiContent, localTime float64) {
	emojiString := string(e)
	// increment count
	emojisSentCount.Add(emojiString, 1)
	go s.RedisClient.HIncrBy(context.TODO(), "emojis_sent", emojiString, 1)

	s.mu.Lock()
	defer s.mu.Unlock()
	prevTime, exists := s.EmojiSendTime[e]
	if !exists {
		s.EmojiSendTime[e] = localTime
		return
	}
	oldU, oldV := to2DTimeModular(prevTime)
	newU, newV := to2DTimeModular(localTime)
	u := weightedAverage(oldU, newU, 0.01)
	v := weightedAverage(oldV, newV, 0.01)
	newTime := from2DTimeModular(u, v)
	s.EmojiSendTime[e] = newTime
	old := emojisSentAt.Get(emojiString)
	if old == nil {
		emojisSentAt.AddFloat(emojiString, newTime)
	} else {
		old.(*expvar.Float).Set(newTime)
	}
	go s.RedisClient.HSet(
		context.TODO(), "emoji_sent_at", emojiString, strconv.FormatFloat(newTime, 'E', -1, 64),
	)
}

func (s *Server) LogReply(r *MessageReply) {
	replyString := string(r.Reply)
	go s.RedisClient.HIncrBy(context.TODO(), "emoji_reply", replyString, 1)
	go s.RedisClient.HIncrBy(context.TODO(), r.OriginalContent.RedisKey(), replyString, 1)
}

// TODO weight the recommendations with how frequently they are sent.
func (s *Server) FindNearRecommendations(amt int, localTime float64) []EmojiContent {
	u, v := to2DTimeModular(localTime)
	out := make([]EmojiContent, 0, amt)
	s.mu.Lock()
	for cntnt, t := range s.EmojiSendTime {
		newU, newV := to2DTimeModular(t)
		dist := distance(u, newU, v, newV)
		if dist < 0.05 {
			out = append(out, cntnt)
			if len(out) == amt {
				break
			}
		}
	}
	s.mu.Unlock()
	return out
}
