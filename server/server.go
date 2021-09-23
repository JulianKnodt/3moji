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
	SignedUp        map[Email]User
	LoggedIn        map[Email]LoginToken
	HashedPasswords map[Uuid]string

	// In-memory map of recipient to Messages
	UserToMessages map[Uuid][]Uuid
	// TODO add a timer so that these will eventually expire or be cleaned up periodically.
	Messages map[Uuid]Message

	Users map[Uuid]User
	// List of friends for a given user
	Friends       map[Uuid]map[Uuid]struct{}
	MutualFriends map[Uuid]map[Uuid]struct{}

	// Replies waiting for a given user
	UserToReplies map[Uuid][]Uuid
	Replies       map[Uuid]MessageReply
}

func NewServer() *Server {
	return &Server{
		SignedUp:        map[Email]User{},
		LoggedIn:        map[Email]LoginToken{},
		HashedPasswords: map[Uuid]string{},

		UserToMessages: map[Uuid][]Uuid{},
		Messages:       map[Uuid]Message{},

		Users:   map[Uuid]User{},
		Friends: map[Uuid]map[Uuid]struct{}{},

		UserToReplies: map[Uuid][]Uuid{},
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

func (s *Server) SignUp(userEmail Email, userName string, hashedPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.SignedUp[userEmail]; exists {
		return fmt.Errorf("User already exists")
	}

	uuid, err := generateUuid()
	if err != nil {
		return err
	}
	s.SignedUp[userEmail] = User{
		Name:  userName,
		Email: userEmail,
		Uuid:  uuid,
	}
	s.HashedPasswords[uuid] = hashedPassword

	return nil
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
		enc := json.NewEncoder(w)
		if err := s.SignUp(email, sup.Name, sup.HashedPassword); err != nil {
			w.WriteHeader(401)
			enc.Encode(err)
			return
		}
		loginToken, err := s.Login(email, sup.HashedPassword)
		if err != nil {
			w.WriteHeader(401)
			enc.Encode(err)
			return
		}
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
		enc := json.NewEncoder(w)
		loginToken, err := s.Login(email, lp.HashedPassword)
		if err != nil {
			w.WriteHeader(401)
			enc.Encode(err)
			return
		}
		enc.Encode(loginToken)
		return
	}
}

func (s *Server) FriendHandler() http.HandlerFunc {
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

		s.mu.Lock()
		defer s.mu.Unlock()

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

func (s *Server) SendMsgHandler() http.HandlerFunc {
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
		for _, recipUuid := range msg.Recipients {
			s.UserToMessages[recipUuid] = append(s.UserToMessages[recipUuid], msg.Uuid)
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

// s.mu should be held when called.
func (s *Server) MessageForReplyLocked(reply MessageReply) (Message, bool) {
	msg, exists := s.Messages[reply.Message]
	return msg, exists
}

func (s *Server) RecvMsgHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(404)
			return
		}
		var req RecvMsgRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			fmt.Printf("Error decoding recv message %v", err)
			w.WriteHeader(401)
			return
		}
		token := req.LoginToken
		if s.ValidateLoginToken(token) != nil {
			w.WriteHeader(401)
			return
		}
		user, exists := s.UserFor(token)
		if !exists {
			w.WriteHeader(401)
			return
		}

		var out RecvMsgResponse
		s.mu.Lock()
		now := time.Now()
		defer s.mu.Unlock()
		for _, uuid := range s.UserToMessages[user.Uuid] {
			msg, exists := s.Messages[uuid]
			if !exists {
				continue
			} else if msg.Expired(now) {
				delete(s.Messages, uuid)
				continue
			}
			out.NewMessages = append(out.NewMessages, msg)
		}
		for _, uuid := range s.UserToReplies[user.Uuid] {
			reply, replyExists := s.Replies[uuid]
			if !replyExists {
				continue
			}
			msg, exists := s.MessageForReplyLocked(reply)
			if !exists {
				continue
			} else if msg.Expired(now) {
				delete(s.Messages, uuid)
				continue
			}
			out.NewReplies = append(out.NewReplies, reply)
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(&out); err != nil {
			w.WriteHeader(500)
			return
		}
		// if success then empty out the messages
		if req.DeleteOld {
			s.UserToMessages[user.Uuid] = nil
			s.UserToReplies[user.Uuid] = nil
		}
		return
	}
}
