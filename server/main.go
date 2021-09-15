package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

func main() {
	// TODO handle certain functions from below
	s := NewServer()
	// TODO actually serve the sign in handler
	if err := s.Serve("localhost:8080"); err != nil {
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
	Messages map[Email][]Message
}

type Email string

func NewEmail(s string) (Email, error) {
	if s == "" {
		return Email(""), fmt.Errorf("Cannot make empty string an email")
	}
	// XXX temp
	if !strings.HasSuffix(s, "princeton.edu") {
		return Email(""), fmt.Errorf("Currently only accepting princeton emails")
	}

	return Email(s), nil
}

type User struct {
	Name  string
	Email Email
	// TODO add other preference fields here
}

type LoginToken struct {
	ValidUntil time.Time
	// uuid is some unique way of representing a log in token so that it cannot be forged with
	// just the time.
	Uuid Uuid
	// awful method of protecting a user, TODO eventually replace this

	UserEmail Email
}

func NewServer() *Server {
	return &Server{
    SignedUp: map[Email]User{},
    LoggedIn: map[Email]LoginToken{},
    Messages: map[Email][]Message{},
  }
}

func (srv *Server) Serve(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sign_up/", srv.SignUpHandler())
	mux.HandleFunc("/api/v1/send_msg/", srv.SendMsgHandler())
	mux.HandleFunc("/api/v1/recv_msg/", srv.RecvMsgHandler())

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

func (s *Server) SignUp(userEmail Email, userName string) error {
	if _, exists := s.SignedUp[userEmail]; exists {
		return fmt.Errorf("User already exists")
	}

	s.SignedUp[userEmail] = User{
		Name:  userName,
		Email: userEmail,
	}

	return nil
}

type Uuid uint64

func generateUuid() (Uuid, error) {
	uuidBytes := [8]byte{}
	if _, err := rand.Read(uuidBytes[:]); err != nil {
		return Uuid(0), err
	}

	uuid := binary.BigEndian.Uint64(uuidBytes[:])
	return Uuid(uuid), nil
}

func (s *Server) LogIn(userEmail Email) (LoginToken, error) {
	if _, exists := s.SignedUp[userEmail]; !exists {
		return LoginToken{}, fmt.Errorf("User does not exist")
	}

	// TODO check collisions of the uuid and retry or crash
	uuid, err := generateUuid()
	if err != nil {
		return LoginToken{}, err
	}

	loginToken := LoginToken{
		ValidUntil: time.Now().Add(72 * time.Hour),
		Uuid:       uuid,
    UserEmail: userEmail,
	}
	return loginToken, nil
}

type SignUpPayload struct {
	Email string `json:email`
	Name  string `json:name`
}

func (s *Server) SignUpHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			return
		}
    dec := json.NewDecoder(r.Body)
    var sup SignUpPayload
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
		if err := s.SignUp(email, sup.Name); err != nil {
			panic("TODO")
		}
		logInToken, err := s.LogIn(email)
		if err != nil {
			panic("TODO")
		}
		// TODO I forgot how to make this JSON encoded
		enc := json.NewEncoder(w)
		if err = enc.Encode(logInToken); err != nil {
			panic("TODO")
		}
		return
	}
}

type EmojiContent [3]rune

type SendMessagePayload struct {
	LoginToken LoginToken
	Message    Message
}

// Message is a struct that represents an emoji message between two people
type Message struct {
	Emojis     EmojiContent
	Source     User
	Recipients []Email
}

func (s *Server) SendMsgHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			return
		}
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(500)
			return
		}
		var smp SendMessagePayload
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&smp); err != nil {
			fmt.Printf("Error decoding send message %v", err)
			w.WriteHeader(401)
			return
		}
		user := smp.Message.Source
		loginToken := smp.LoginToken
		existingToken, exists := s.LoggedIn[user.Email]
		if !exists || existingToken != loginToken {
			w.WriteHeader(401)
			return
		}
		if loginToken.UserEmail != user.Email {
			// impersonating someone else?
			w.WriteHeader(401)
			return
		}

		// save message for all users
		// TODO delete old messages as well
		msg := smp.Message
		for _, recipEmail := range msg.Recipients {
			s.Messages[recipEmail] = append(s.Messages[recipEmail], msg)
		}

		w.WriteHeader(200)
		return
	}
}

func (s *Server) RecvMsgHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(404)
			return
		}
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(500)
			return
		}
		var loginToken LoginToken
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&loginToken); err != nil {
			fmt.Printf("Error decoding recv message %v", err)
			w.WriteHeader(401)
			return
		}
		// TODO validate token
		enc := json.NewEncoder(w)
		if err := enc.Encode(s.Messages[loginToken.UserEmail]); err != nil {
      w.WriteHeader(401)
      return
    }
		s.Messages[loginToken.UserEmail] = nil
		return
	}
}
