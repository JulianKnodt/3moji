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

type EmojiContent string
type EmojiReply string

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

type Group struct {
	Uuid Uuid `json:"uuid"`
	// Display Name, need not be unique
	Name  string            `json:"name"`
	Users map[Uuid]struct{} `json:"users"`
	// TODO should this have a location attached as well?
}

// Message is a struct that represents an emoji message between two people
type Message struct {
	// Messages Uuid
	Uuid     Uuid         `json:"uuid"`
	Emojis   EmojiContent `json:"emojis"`
	Source   User         `json:"source"`
	Location string       `json:"location"`
	// Unix timestamp for current time.
	SentAt int64 `json:"sentAt"`

	// 0-24 for the hour the message is sent at.
	LocalHour float64 `json:"localHour"`
}

func (m *Message) Expired(now time.Time) bool {
	t := time.Unix(m.SentAt, 0)
	return t.Add(time.Hour).Before(now)
}

type MessageReply struct {
	Message Uuid `json:"message"`

	// This is so the user can see what they originally sent
	OriginalContent EmojiContent `json:"originalContent"`
	Reply           EmojiReply   `json:"reply"`
	From            User         `json:"from"`
	// Unix timestamp
	SentAt int64 `json:"sentAt"`
}

type LoginToken struct {
	// Unix Timestamp
	ValidUntil int64 `json:"validUntil"`
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
