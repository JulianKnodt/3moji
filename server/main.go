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
  if err := s.Serve(); err != nil {
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
	LoggedIn map[Email]LogInToken
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

type LogInToken struct {
	ValidUntil time.Time
	// uuid is some unique way of representing a log in token so that it cannot be forged with
	// just the time.
	Uuid Uuid
	// awful method of protecting a user, TODO eventually replace this with
}

func NewServer() *Server {
	return &Server{}
}

func (srv *Server) Serve() error {
  mux := http.NewServeMux()
  mux.HandleFunc("/api/v1/sign_up/", srv.SignUpHandler())

  s := http.Server {
    Addr:           ":8080",
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
	uuidBytes := [4]byte{}
	if _, err := rand.Read(uuidBytes[:]); err != nil {
		return Uuid(0), err
	}

	uuid := binary.BigEndian.Uint64(uuidBytes[:])
  return Uuid(uuid), nil
}

func (s *Server) LogIn(userEmail Email) (LogInToken, error) {
	if _, exists := s.SignedUp[userEmail]; !exists {
		return LogInToken{}, fmt.Errorf("User does not exist")
	}

	// TODO check collisions of the uuid and retry or crash
  uuid, err := generateUuid()
  if err != nil {
		return LogInToken{}, err
  }

	loginToken := LogInToken{
		ValidUntil: time.Now().Add(72 * time.Hour),
		Uuid:       uuid,
	}
	return loginToken, nil
}

func (s *Server) SignUpHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(400)
			return
		}
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(500)
			return
		}
		email, err := NewEmail(r.Form.Get("email"))
		if err != nil {
			w.WriteHeader(401)
			return
		}
		// TODO validate name
		name := r.Form.Get("name")
		if err = s.SignUp(email, name); err != nil {
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
