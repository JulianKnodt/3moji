package main

type ListPeopleKind int

const (
	OnlyFriends ListPeopleKind = iota
	All
	NotFriends
)

type ListPeopleRequest struct {
	Amount     int            `json:"amount"`
	Kind       ListPeopleKind `json:"friends"`
	LoginToken LoginToken     `json:"loginToken"`
}

type ListPeopleResponse struct {
	People []User `json:"people"`
}

type AckMsgRequest struct {
	// Msg being replied to
	MsgID Uuid `json:"msgID"`
	// Reply is a single emoji reply.
	Reply rune `json:"reply"`
	// LoginToken of the user
	LoginToken LoginToken `json:"loginToken"`
}

// Receives both messages and replies for a given user
type RecvMsgRequest struct {
	LoginToken LoginToken `json:"loginToken"`
}

type RecvMsgResponse struct {
	// New messages the user has not seen
	NewMessages []Message `json:"newMessages"`
	// TODO need to add in who replied here
	NewReplies []MessageReply `json:"newReplies"`
}

type FriendAction int

const (
	Rmfriend FriendAction = iota
	AddFriend
)

type FriendRequest struct {
	Other      Uuid         `json:"other"`
	LoginToken LoginToken   `json:"loginToken"`
	Action     FriendAction `json:"action"`
}

type SignUpRequest struct {
	Email          string `json:"email"`
	Name           string `json:"name"`
	HashedPassword string `json:"hashedPassword"`
}

type LoginRequest struct {
	Email          string `json:"email"`
	HashedPassword string `json:"hashedPassword"`
}

type SendMessageRequest struct {
	LoginToken LoginToken `json:"loginToken"`
	// The uuid for the message is generated on the server side
	Message Message `json:"message"`
}
