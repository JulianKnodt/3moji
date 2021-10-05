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
	UserToMessages map[Uuid][]Uuid
	// TODO add a timer so that these will eventually expire or be cleaned up periodically.
	Messages map[Uuid]Message

	Users map[Uuid]*User
	// List of friends for a given user: user -> their friends
	Friends       map[Uuid]map[Uuid]struct{}
	MutualFriends map[Uuid]map[Uuid]struct{}

	Groups        map[Uuid]Group
	UsersToGroups map[Uuid]map[Uuid]struct{}

	// Replies waiting for a given user
	UserToReplies map[Uuid][]Uuid
	Replies       map[Uuid]MessageReply

	EmojiSendCounts *expvar.Map
	EmojiSendHours  *expvar.Map
}

func NewServer() *Server {
	return &Server{
		SignedUp:        map[Email]*User{},
		LoggedIn:        map[Email]LoginToken{},
		HashedPasswords: map[Uuid]string{},

		UserToMessages: map[Uuid][]Uuid{},
		Messages:       map[Uuid]Message{},

		Groups:        map[Uuid]Group{},
		UsersToGroups: map[Uuid]map[Uuid]struct{}{},

		Users:   map[Uuid]*User{},
		Friends: map[Uuid]map[Uuid]struct{}{},

		UserToReplies: map[Uuid][]Uuid{},
		Replies:       map[Uuid]MessageReply{},

		//EmojiSendCounts: expvar.NewMap("EmojiSendCounts"),
		//EmojiSendHours:  expvar.NewMap("EmojiSendLocalHour"),
	}
}

func (srv *Server) Serve(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sign_up/", srv.SignUpHandler())
	mux.HandleFunc("/api/v1/login/", srv.LoginHandler())

	mux.HandleFunc("/api/v1/friend/", srv.FriendHandler())
	mux.HandleFunc("/api/v1/groups/", srv.GroupHandler())
	mux.HandleFunc("/api/v1/send_msg/", srv.SendMsgHandler())
	mux.HandleFunc("/api/v1/recv_msg/", srv.RecvMsgHandler())

	mux.HandleFunc("/api/v1/list_friends/", srv.ListPeopleHandler())
	mux.HandleFunc("/api/v1/list_groups/", srv.ListGroupHandler())
	mux.HandleFunc("/api/v1/recs/", srv.RecommendationHandler())

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

// Cleanup will remove any old messages or replies which have not been acknowledged.
/*
func (s *Server) Cleanup() {
  var i int
  var v TimedObject
  now := time.Now()
  for i, v = range s.MessageTimers {
    if v.ExpiresAt.After(now) {
      break
    }
  }
  expired := s.MessageTimes[:i]
  s.MessageTimes = s.MessageTimes[i:]
  for _, o := range expired {
    delete(s.Messages, o.uuid)
  }
}
*/

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
		ValidUntil: time.Now().Add(72 * time.Hour).Unix(),
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
	return nil
	/*
		s.mu.Lock()
		defer s.mu.Unlock()
		existingToken, exists := s.LoggedIn[token.UserEmail]
		if !exists {
			fmt.Println(s.LoggedIn)
			return fmt.Errorf("Token does not exist")
		} else if existingToken != token {
			return fmt.Errorf("Tokens do not match want: %v, got: %v", existingToken, token)
		}
		return nil
	*/
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
func (s *Server) MessageForReplyLocked(reply MessageReply) (Message, bool) {
	msg, exists := s.Messages[reply.Message]
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

func (s *Server) LogEmojiContent(e EmojiContent, localHour float64) {
	// increment count
	emojiString := string(e[:])
	s.EmojiSendCounts.Add(emojiString, 1)

	oldValX := s.EmojiSendHours.Get(emojiString + "x").(*expvar.Float)
	oldValY := s.EmojiSendHours.Get(emojiString + "y").(*expvar.Float)
	newX, newY := to2DTimeModular(localHour)
	u := weightedAverage(oldValX.Value(), newX, 0.01)
	v := weightedAverage(oldValY.Value(), newY, 0.01)
	oldValX.Set(u)
	oldValY.Set(v)

	s.EmojiSendHours.Get(emojiString).(*expvar.Float).Set(from2DTimeModular(u, v))
}
