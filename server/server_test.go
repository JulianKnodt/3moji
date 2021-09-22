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
	err := mc.SignUp("test", "test@princeton.edu")
	if err != nil {
		t.Fatalf("Failed to sign up: %v", err)
	}
}
