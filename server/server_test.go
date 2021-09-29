package main

import (
	"testing"
	"time"
)

func TestSignUp(t *testing.T) {
	s := NewServer()
	addr := "localhost:8080"
	go s.Serve(addr)
	time.Sleep(10 * time.Millisecond)
	mc := NewMojiClient("http://" + addr)
	mc2 := NewMojiClient("http://" + addr)
	if err := mc.SignUp("test", "test@princeton.edu"); err != nil {
		t.Fatalf("Failed to sign up: %v", err)
	}

	if err := mc2.SignUp("test2", "test2@princeton.edu"); err != nil {
		t.Fatalf("Failed to sign up: %v", err)
	}
}

func TestListPeople(t *testing.T) {
	s := NewServer()
	addr := "localhost:8081"
	go s.Serve(addr)
	time.Sleep(10 * time.Millisecond)
	mc := NewMojiClient("http://" + addr)
	mc2 := NewMojiClient("http://" + addr)
	if err := mc.SignUp("test", "test@princeton.edu"); err != nil {
		t.Fatalf("Failed to sign up: %v", err)
	}

	if err := mc2.SignUp("test2", "test2@princeton.edu"); err != nil {
		t.Fatalf("Failed to sign up: %v", err)
	}

	if err := mc.ListPeople(); err != nil {
		t.Fatalf("Failed to list people:  %v", err)
	}

	if err := mc.FriendOp(mc2.UserID(), AddFriend); err != nil {
		t.Fatalf("Failed to add friend:  %v", err)
	}

	if err := mc.ListPeople(); err != nil {
		t.Fatalf("Failed to list people: %v", err)
	}

	if err := mc.GroupOp("example_group", 0, CreateGroup); err != nil {
		t.Fatalf("Failed to create group:  %v", err)
	}

	if err := mc2.ListGroups(AllGroups); err != nil {
		t.Fatalf("Failed to list groups:  %v", err)
	}
}
