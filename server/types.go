package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

type EmojiContent [3]rune

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
	Uuid  Uuid
	Name  string
	Email Email
	// TODO add other preference fields here
}

// Message is a struct that represents an emoji message between two people
type Message struct {
	Uuid       Uuid
	Emojis     EmojiContent
	Source     User
	Recipients []Email
}

type MessageReply struct {
	OriginalContent EmojiContent `json:"originalContent"`
	Reply           rune         `json:"reply"`
	From            Email        `json:"from"`
}

type LoginToken struct {
	ValidUntil time.Time
	// uuid is some unique way of representing a log in token so that it cannot be forged with
	// just the time.
	Uuid Uuid
	// awful method of protecting a user, TODO eventually replace this

	UserEmail Email
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
