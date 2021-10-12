package main

type ListPeopleKind int

const (
	OnlyFriends ListPeopleKind = iota
	All
	NotFriends
)

type ListPeopleRequest struct {
	Amount     int            `json:"amount"`
	Kind       ListPeopleKind `json:"kind"`
	LoginToken LoginToken     `json:"loginToken"`
}

type ListPeopleResponse struct {
	People []User `json:"people"`
}

type AckMsgRequest struct {
	// Msg being replied to
	MsgID Uuid `json:"msgID,string"`
	// Reply is a single emoji reply.
	Reply EmojiReply `json:"reply"`
	// LoginToken of the user
	LoginToken LoginToken `json:"loginToken"`
}

// Receives both messages and replies for a given user
type RecvMsgRequest struct {
	LoginToken LoginToken `json:"loginToken"`
	DeleteOld  bool       `json:"deleteOld"`
}

type RecvMsgResponse struct {
	// New messages the user has not seen
	NewMessages []*Message `json:"newMessages"`
	// TODO need to add in who replied here
	NewReplies []MessageReply `json:"newReplies"`
}

type FriendAction int

const (
	Rmfriend FriendAction = iota
	AddFriend
)

type FriendRequest struct {
	Other      Uuid         `json:"other,string"`
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

// Also doubles as SignupResponse
type LoginResponse struct {
	User User `json:"user"`

	LoginToken LoginToken `json:"loginToken"`
}

type MessageRecipientKind int

const (
	MsgGroup MessageRecipientKind = iota
	MsgFriend
)

type SendMessageRequest struct {
	LoginToken LoginToken `json:"loginToken"`

	// The uuid for the message is generated on the server side
	Message Message `json:"message"`

	// Whether this is a message intended for a group or an individual
	RecipientKind MessageRecipientKind `json:"recipientKind"`
	// Uuid of group or individual being sent to
	To Uuid `json:"to,string"`
}

type GroupOp int

const (
	JoinGroup GroupOp = iota
	// When removing people from a group, if the group is empty it will be deleted.
	LeaveGroup
	CreateGroup
)

type GroupRequest struct {
	Kind GroupOp `json:"kind"`
	// GroupName is empty if not creating a group, but is only used for display and not
	// identification for now.
	GroupName string `json:"groupName"`
	GroupUuid Uuid   `json:"groupUuid,omitempty,string"`

	// User's login token
	LoginToken LoginToken `json:"loginToken"`
}

// TODO also add a list groups operation
type ListGroupKind int

const (
	AllGroups ListGroupKind = iota
	JoinedGroups
	NotJoinedGroups
)

type ListGroupRequest struct {
	Kind       ListGroupKind `json:"kind"`
	Amount     int           `json:"amount"`
	LoginToken LoginToken    `json:"loginToken"`
}

type ListGroupResponse struct {
	Groups []Group `json:"groups"`
}

type RecommendationRequest struct {
	LocalTime float64 `json:"localTime"`
	// TODO add more features here
}

type RecommendationResponse struct {
	Recommendations []EmojiContent `json:"recommendations"`
}

type AddPushNotifTokenRequest struct {
	Token      string     `json:"token"`
	LoginToken LoginToken `json:"loginToken"`
}
