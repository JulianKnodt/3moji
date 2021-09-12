package main

import (
  "net/http"
  "sync"
  "fmt"
  "encoding/binary"
  "time"
)

func main() {
  // TODO handle certain functions from below
  s := NewServer()
  // TODO actually serve the sign in handler
  s.SignInHandler()
}

// TODO have in memory data structures with serialization for persistence

// Server is a persistent stateful server which represents the current state of the app
// universe. For now it should just be a massive struct which contains everything for
// simplicity.
type Server struct {
  // mu guards the map of struct who have signed up
  mu sync.Mutex
  SignedUp map[Email]User

  LoggedIn map[Email]LogInToken
}

type Email string
type User struct {
  Name string
  Email string
  // TODO add other preference fields here
}

type LogInToken struct {
  ValidUntil time.Time
  // uuid is some unique way of representing a log in token so that it cannot be forged with
  // just the time.
  Uuid uint64
  // awful method of protecting a user, TODO eventually replace this with
}

func NewServer() *Server {
  return &Server{}
}

func (s *Server) SignUp(userEmail, userName string) error {
  if _, exists := s.SignedUp[userEmail]; exists {
    return fmt.Errorf("User already exists")
  }

  s.SignedUp[userEmail] = User {
    Name: userName,
    Email: userEmail,
  }

  return nil
}

func (s *Server) LogIn(userEmail string) (LogInToken, error) {
  if _, exists := s.SignedUp[userEmail]; !exists {
    return fmt.Errorf("User does not exist")
  }

  uuidBytes := [4]byte{}
  if _, err := rand.Read(uuidBytes[:]); err != nil {
    return err
  }

  uuid := binary.BigEndian.Uint64(uuidBytes[:])
  // TODO check collisions of the uuid and retry or crash

  loginToken := LogInToken {
    ValidUntil: time.Now().Add(72 * time.Hour),
    Uuid: uuid,
  }
  return loginToken, nil
}

func (s *Server) SignInHandler() {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
      w.WriteHeader(400)
      return
    }
    if err := r.ParseForm(); err != nil {
      w.WriteHeader(500)
      return
    }
    email := r.Form.Get("email")
    if email == "" {
      w.WriteHeader(401)
      return
    } else if !strings.EndsWith(email, "princeton.edu") {
      // temporary only allow princeton email addresses
      w.WriteHeader(401)
      return
    }
    // TODO validate name
    name := r.Form.Get("name")
    user, err := s.SignUp(email, name)
    if err != nil {
      panic("TODO")
    }
    logInToken, err := s.LogIn(email, user)
    if err != nil {
      panic("TODO")
    }
    // TODO I forgot how to make this JSON encoded
    if err = w.Write(logInToken); err != nil {
      panic("TODO")
    }
    return
  }
}
