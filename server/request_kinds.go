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
	People []User
}
