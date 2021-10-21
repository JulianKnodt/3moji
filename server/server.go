package main

import (
	"expvar"
	"fmt"
	"math"
	"net/http"
	"os"
	"sync"
	"time"
	// TODO add boltdb for persistence

	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
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
	mu              sync.Mutex
	SignedUp        map[Email]*User
	LoggedIn        map[Email]LoginToken
	HashedPasswords map[Uuid]string

	// In-memory map of recipient to Messages
	UserToMessages map[Uuid]map[Uuid]struct{}
	Messages       map[Uuid]*Message

	Users map[Uuid]*User
	// List of friends for a given user: user -> their friends
	Friends map[Uuid]map[Uuid]struct{}

	Groups        map[Uuid]Group
	UsersToGroups map[Uuid]map[Uuid]struct{}

	// Replies waiting for a given user
	UserToReplies map[Uuid][]Uuid
	Replies       map[Uuid]MessageReply

	//EmojiSendCounts *expvar.Map
	EmojiSendTime map[EmojiContent]float64

	// ExpoNotificationTokens for sending push notifications
	UserNotificationTokens map[Uuid]expo.ExponentPushToken
}

func NewServer() *Server {
	// TODO only construct expvars if not test environment
	return &Server{
		SignedUp:        map[Email]*User{},
		LoggedIn:        map[Email]LoginToken{},
		HashedPasswords: map[Uuid]string{},

		UserToMessages: map[Uuid]map[Uuid]struct{}{},
		Messages:       map[Uuid]*Message{},

		Groups:        map[Uuid]Group{},
		UsersToGroups: map[Uuid]map[Uuid]struct{}{},

		Users:   map[Uuid]*User{},
		Friends: map[Uuid]map[Uuid]struct{}{},

		UserToReplies: map[Uuid][]Uuid{},
		Replies:       map[Uuid]MessageReply{},

		//EmojiSendCounts: expvar.NewMap("EmojiSendCounts"),
		EmojiSendTime: map[EmojiContent]float64{},

		UserNotificationTokens: map[Uuid]expo.ExponentPushToken{},
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

	mux.Handle("/debug/vars", expvar.Handler())

	go srv.Cleanup()
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

// Cleanup will remove any old messages or replies which have not been acknowledged.
func (s *Server) Cleanup() {
	t := time.NewTicker(5 * time.Minute)
	for range t.C {
		now := time.Now()
		s.mu.Lock()
		for uuid, message := range s.Messages {
			if message.Expired(now) {
				delete(s.Messages, uuid)
			}
		}
		for uuid, reply := range s.Replies {
			if _, exists := s.Messages[reply.Message.Uuid]; !exists {
				delete(s.Replies, uuid)
			}
		}
		s.mu.Unlock()
	}
}

func (s *Server) SignUp(userEmail Email, userName string, hashedPassword string) (Uuid, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.SignedUp[userEmail]; exists {
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
	s.SignedUp[userEmail] = user
	s.Users[uuid] = user
	s.HashedPasswords[uuid] = hashedPassword

	return uuid, nil
}

func (s *Server) Login(userEmail Email, hashedPassword string) (LoginToken, error) {
	if hashedPassword == "" {
		return LoginToken{}, fmt.Errorf("password must not be empty")
	}

	s.mu.Lock()
	user, exists := s.SignedUp[userEmail]
	if !exists || user.Email != userEmail {
		s.mu.Unlock()
		// Show generic error message, but user does not exist
		return LoginToken{}, fmt.Errorf("Something wrong with login")
	}
	if existing := s.HashedPasswords[user.Uuid]; existing != hashedPassword {
		s.mu.Unlock()
		// Show generic error message, but password isn't right
		return LoginToken{}, fmt.Errorf("Something wrong with login")
	}
	s.mu.Unlock()

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
	s.mu.Lock()
	s.LoggedIn[user.Email] = loginToken
	s.mu.Unlock()
	return loginToken, nil
}

// Checks that a login token is correct, and matches the currently existing token kept on the
// token.
func (s *Server) ValidateLoginToken(token LoginToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existingToken, exists := s.LoggedIn[token.UserEmail]
	if !exists {
		return fmt.Errorf("Token does not exist")
	} else if existingToken != token {
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
func (s *Server) UserFor(token LoginToken) (*User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, exists := s.SignedUp[token.UserEmail]
	return user, exists
}

// s.mu should be held when called.
func (s *Server) MessageForReplyLocked(reply MessageReply) (*Message, bool) {
	msg, exists := s.Messages[reply.Message.Uuid]
	return msg, exists
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
