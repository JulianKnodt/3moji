package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net/mail"
	"strconv"
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
	Uuid  Uuid   `json:"uuid,string"`
	Name  string `json:"name"`
	Email Email  `json:"email"`
	// TODO add other preference fields here
}

type Group struct {
	Uuid Uuid `json:"uuid,string"`
	// Display Name, need not be unique
	Name  string          `json:"name"`
	Users map[Uuid]string `json:"users"`
	// TODO should this have a location attached as well?
}

// Message is a struct that represents an emoji message between two people
type Message struct {
	// Messages Uuid
	Uuid Uuid `json:"uuid,string"`
	// Name of who this was sent to
	SentTo string `json:"sentTo"`

	Emojis   EmojiContent `json:"emojis"`
	Source   User         `json:"source"`
	Location string       `json:"location"`
	// Unix timestamp for current time.
	SentAt int64 `json:"sentAt,string"`
	// number of seconds for this message to live (time to live)
	TTL int64 `json:"ttl,string"`

	// 0-24 for the hour the message is sent at.
	LocalTime float64 `json:"localTime,string"`
}

func (m *Message) Expired(now time.Time) bool {
	t := time.Unix(m.SentAt, 0)
	return t.Add(time.Duration(m.TTL) * time.Second).Before(now)
}

type MessageReply struct {
	Message *Message `json:"message"`

	// This is so the user can see what they originally sent
	OriginalContent EmojiContent `json:"originalContent"`
	Reply           EmojiReply   `json:"reply"`
	From            User         `json:"from"`
	// Unix timestamp
	SentAt int64 `json:"sentAt,string"`
}

type LoginToken struct {
	// Unix Timestamp
	ValidUntil int64 `json:"validUntil,string"`
	// uuid is some unique way of representing a log in token so that it cannot be forged with
	// just the time.
	// XXX this is not the user's uuid.
	Uuid Uuid `json:"uuid,string"`

	UserEmail Email `json:"userEmail"`
}

func (lt *LoginToken) Expired() bool {
	return time.Unix(lt.ValidUntil, 0).Before(time.Now())
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

// Convert this Uuid to a string
func (u Uuid) String() string {
	return strconv.FormatUint(uint64(u), 10)
}
