package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"
)

type EmojiContent string

func (e EmojiContent) RedisKey() string {
	return fmt.Sprintf("emojis_%s", e)
}

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

	// Whether new users can join this group or not. It will never be displayed
	Locked bool `json:"locked"`
}

// Message is a struct that represents an emoji message between two people
type Message struct {
	// Messages Uuid
	Uuid Uuid `json:"uuid,string"`
	// Name of who this was sent to
	SentTo string `json:"sentTo"`
	// Group which this was sent, or invalid Uuid
	Group Uuid `json:"groupSentTo,omitempty"`

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

func MessageRedisKey(uuid Uuid) string {
	return fmt.Sprintf("message_%d", uuid)
}

type MessageReply struct {
	Message *Message `json:"message"`
	Group   Uuid     `json:"group"`

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

const InvalidUuid = Uuid(0)

func generateUuid() (Uuid, error) {
	uuidBytes := [8]byte{}
	if _, err := rand.Read(uuidBytes[:]); err != nil {
		return Uuid(0), err
	}

	uuid := binary.BigEndian.Uint64(uuidBytes[:])
	return Uuid(uuid), nil
}

func (u Uuid) IsValid() bool {
	return u != InvalidUuid
}

// Convert this Uuid to a string
func (u Uuid) String() string {
	return strconv.FormatUint(uint64(u), 10)
}

func UuidFromString(uuid string) (Uuid, error) {
	u, err := strconv.ParseUint(uuid, 10, 64)
	return Uuid(u), err
}

// What kind of match
type MatchKind uint16

const (
	MatchAll = iota
	MatchPrefix
)

type MatchFilter struct {
	Match string    `json:"match"`
	Kind  MatchKind `json:"kind"`
}

func (mf MatchFilter) MatchFunc() func(string) bool {
	if mf.Match == "" {
		return func(string) bool { return true }
	}
	switch mf.Kind {
	case MatchAll:
		return func(s string) bool { return strings.Contains(s, mf.Match) }
	case MatchPrefix:
		return func(s string) bool { return strings.HasPrefix(s, mf.Match) }
	default:
		// Don't want to panic, so return match nothing if it's unknown
		return func(string) bool { return false }
	}
}
