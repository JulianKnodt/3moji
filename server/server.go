// package main
package main

import (
	"encoding/json"
	"fmt"
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
	// TODO handle certain functions from below
	s := NewServer()
	// TODO actually serve the sign in handler
	if err := s.Serve(":" + port); err != nil {
		fmt.Printf("Server exited: %v", err)
	}
}

// TODO have in memory data structures with serialization for persistence

// Server is a persistent stateful server which represents the current state of the app
// universe. For now it should just be a massive struct which contains everything for
// simplicity.
type Server struct {
	// mu guards the map of struct who have signed up
	mu       sync.Mutex
	SignedUp map[Email]User
	LoggedIn map[Email]LoginToken

	// In-memory map of recipient to Messages
	UserToMessages map[Email][]Uuid
	// TODO add a timer so that these will eventually expire or be cleaned up periodically.
	Messages map[Uuid]Message

	Users map[Uuid]User
	// List of friends for a given user
	Friends map[Uuid]map[Uuid]struct{}

	// Replies waiting for a given user
	UserToReplies map[Email][]Uuid
	Replies       map[Uuid]MessageReply
}

func NewServer() *Server {
	return &Server{
		SignedUp: map[Email]User{},
		LoggedIn: map[Email]LoginToken{},

		UserToMessages: map[Email][]Uuid{},
		Messages:       map[Uuid]Message{},

		Users:   map[Uuid]User{},
		Friends: map[Uuid]map[Uuid]struct{}{},

		UserToReplies: map[Email][]Uuid{},
		Replies:       map[Uuid]MessageReply{},
	}
}

func (srv *Server) Serve(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is live."))
	})
	mux.HandleFunc("/api/v1/sign_up/", srv.SignUpHandler())
	mux.HandleFunc("/api/v1/login/", srv.LoginHandler())
	// TODO add logout handler
	mux.HandleFunc("/api/v1/friend/", srv.FriendHandler())
	mux.HandleFunc("/api/v1/send_msg/", srv.SendMsgHandler())
	mux.HandleFunc("/api/v1/recv_msg/", srv.RecvMsgHandler())
	mux.HandleFunc("/api/v1/people/", srv.ListPeopleHandler())

	s := http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("Listening on", s.Addr, "...")
	//ticker := time.NewTicker(30 * time.Minute)
	/*
	  go func() {
	    for {
	      <-ticker.C
	      s.Cleanup()
	    }
	  }
	*/
	return s.ListenAndServe()
}

type TimedObject struct {
	uuid      Uuid
	ExpiresAt time.Time
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

func (s *Server) SignUp(userEmail Email, userName string, hashedPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.SignedUp[userEmail]; exists {
		return fmt.Errorf("User already exists")
	}

	s.SignedUp[userEmail] = User{
		Name:  userName,
		Email: userEmail,
	}

	return nil
}

func (s *Server) Login(userEmail Email, hashedPassword string) (LoginToken, error) {
	s.mu.Lock()
	if _, exists := s.SignedUp[userEmail]; !exists {
		s.mu.Unlock()
		return LoginToken{}, fmt.Errorf("User does not exist")
	}
	s.mu.Unlock()

	// TODO check collisions of the uuid and retry or crash
	uuid, err := generateUuid()
	if err != nil {
		return LoginToken{}, err
	}

	loginToken := LoginToken{
		ValidUntil: time.Now().Add(72 * time.Hour),
		Uuid:       uuid,
		UserEmail:  userEmail,
	}
	return loginToken, nil
}

func (s *Server) SignUpHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			return
		}
		dec := json.NewDecoder(r.Body)
		var sup SignUpRequest
		if err := dec.Decode(&sup); err != nil {
			w.WriteHeader(401)
			return
		}
		email, err := NewEmail(sup.Email)
		if err != nil {
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(err)
			return
		}
		if err := s.SignUp(email, sup.Name, sup.HashedPassword); err != nil {
			panic("TODO")
		}
		loginToken, err := s.Login(email, sup.HashedPassword)
		if err != nil {
			panic("TODO")
		}
		enc := json.NewEncoder(w)
		enc.Encode(loginToken)
		return
	}
}

func (s *Server) LoginHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			return
		}
		dec := json.NewDecoder(r.Body)
		var lp LoginRequest
		if err := dec.Decode(&lp); err != nil {
			w.WriteHeader(401)
			return
		}
		email, err := NewEmail(lp.Email)
		if err != nil {
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(err)
			return
		}
		loginToken, err := s.Login(email, lp.HashedPassword)
		if err != nil {
			w.WriteHeader(401)
			return
		}
		enc := json.NewEncoder(w)
		enc.Encode(loginToken)
		return
	}
}

func (s *Server) FriendHandler() http.HandlerFunc {
	// TODO fill this in
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(404)
			return
		}
		var fp FriendRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&fp); err != nil {
			fmt.Printf("Error decoding send message %v", err)
			w.WriteHeader(401)
			return
		}
		if err := s.ValidateLoginToken(fp.LoginToken); err != nil {
			w.WriteHeader(401)
			return
		}
		user, exists := s.UserFor(fp.LoginToken)
		if !exists {
			w.WriteHeader(401)
			return
		}

		switch fp.Action {
		case Rmfriend:
			delete(s.Friends[user.Uuid], fp.Other)
		case AddFriend:
			if _, exists := s.Friends[user.Uuid]; !exists {
				s.Friends[user.Uuid] = map[Uuid]struct{}{}
			}
			s.Friends[user.Uuid][fp.Other] = struct{}{}
		default:
			w.WriteHeader(404)
			return
		}
		w.WriteHeader(200)
		return
	}
}

func (s *Server) SendMsgHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			return
		}
		var smp SendMessageRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&smp); err != nil {
			fmt.Printf("Error decoding send message %v", err)
			w.WriteHeader(401)
			return
		}
		if err := s.ValidateLoginToken(smp.LoginToken); err != nil {
			w.WriteHeader(401)
			return
		}

		// save message for all users
		// TODO delete old messages as well
		msg := smp.Message
		var err error
		if msg.Uuid, err = generateUuid(); err != nil {
			w.WriteHeader(500)
			return
		}
		for _, recipEmail := range msg.Recipients {
			s.UserToMessages[recipEmail] = append(s.UserToMessages[recipEmail], msg.Uuid)
		}
		s.Messages[msg.Uuid] = msg

		w.WriteHeader(200)
		return
	}
}

func (s *Server) ValidateLoginToken(token LoginToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existingToken, exists := s.LoggedIn[token.UserEmail]
	if !exists || existingToken != token {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func (s *Server) UserFor(token LoginToken) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, exists := s.Users[token.Uuid]
	return user, exists
}

func (s *Server) RecvMsgHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(404)
			return
		}
		var recvMsg RecvMsgRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&recvMsg); err != nil {
			fmt.Printf("Error decoding recv message %v", err)
			w.WriteHeader(401)
			return
		}
		token := recvMsg.LoginToken
		var out RecvMsgResponse

		for _, uuid := range s.UserToMessages[token.UserEmail] {
			out.NewMessages = append(out.NewMessages, s.Messages[uuid])
		}
		for _, uuid := range s.UserToReplies[token.UserEmail] {
			out.NewReplies = append(out.NewReplies, s.Replies[uuid])
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(&out); err != nil {
			w.WriteHeader(500)
			return
		}
		s.UserToMessages[token.UserEmail] = nil
		return
	}
}

func (s *Server) AckMsgHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			return
		}
		var ack AckMsgRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&ack); err != nil {
			fmt.Printf("Error decoding recv message %v", err)
			w.WriteHeader(401)
			return
		}
		/*
			if err := s.ValidateLoginToken(&ack.LoginToken); err != nil {
				w.WriteHeader(401)
				return
			}
		*/
		originalMessage, exists := s.Messages[ack.MsgID]
		if !exists {
			w.WriteHeader(404)
			return
		}
		delete(s.Messages, ack.MsgID)

		replyUuid, err := generateUuid()
		if err != nil {
			w.WriteHeader(500)
			return
		}
		email := originalMessage.Source.Email
		s.UserToReplies[email] = append(s.UserToReplies[email], replyUuid)
		// TODO check for collisions?
		s.Replies[replyUuid] = MessageReply{
			OriginalContent: originalMessage.Emojis,
			Reply:           ack.Reply,
			From:            email,
		}

		w.WriteHeader(200)
		return
	}
}
