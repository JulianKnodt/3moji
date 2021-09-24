package main

import (
	"crypto/rand"
	//"crypto/aes"
	//"crypto/cipher"
	"encoding/binary"
	"fmt"
	"net/mail"
	"time"
)

/*
// The key is not intended to be secret, obfuscating login tokens so they're not immediately
// visible over the wire.
var (
  block cipher.Block
)

func init() {
  // TODO or get key from environment
  var key = [24]byte{}
  var err error
  if _, err = rand.Read(key[:]); err != nil {
    panic(fmt.Errorf("Failed to read key %v", err))
  }
  if block, err = aes.NewCipher(key[:]); err != nil {
    panic(fmt.Errorf("Failed to create cipher %v", err))
  }
}
*/

type EmojiContent [3]rune

type Email string

func NewEmail(s string) (Email, error) {
	if s == "" {
		return Email(""), fmt.Errorf("Cannot make empty string an email")
	}
	_, err := mail.ParseAddress(s)
	if err != nil {
		return Email(""), err
	}

	return Email(s), nil
}

type User struct {
	Uuid  Uuid   `json:"uuid"`
	Name  string `json:"name"`
	Email Email  `json:"email"`
	// TODO add other preference fields here
}

// Message is a struct that represents an emoji message between two people
type Message struct {
	Uuid       Uuid         `json:"uuid"`
	Emojis     EmojiContent `json:"emojis"`
	Source     User         `json:"source"`
	Recipients []Uuid       `json:"recipients"`
	SentAt     time.Time    `json:"sentAt"`
}

func (m *Message) Expired(now time.Time) bool {
	return m.SentAt.Add(time.Hour).Before(now)
}

type MessageReply struct {
	Message Uuid `json:"message"`

	OriginalContent EmojiContent `json:"originalContent"`
	Reply           rune         `json:"reply"`
	From            User         `json:"from"`
	SentAt          time.Time    `json:"sentAt"`
}

type LoginToken struct {
	ValidUntil time.Time `json:"validUntil"`
	// uuid is some unique way of representing a log in token so that it cannot be forged with
	// just the time.
	Uuid Uuid `json:"uuid"`

	UserEmail Email `json:"userEmail"`
}

// Uuid represents a unique identifier, temporary for now but maybe upgrade to [2]uint64
// at some point.
type Uuid uint64

func generateUuid() (Uuid, error) {
	uuidBytes := [8]byte{}
	if _, err := rand.Read(uuidBytes[:]); err != nil {
		return Uuid(0), err
	}

	uuid := binary.BigEndian.Uint64(uuidBytes[:])
	return Uuid(uuid), nil
}
