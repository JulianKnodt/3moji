package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"net/http"
)

// mojiClient is a test client for the mojiServer
type mojiClient struct {
	httpc *http.Client
	// loginToken will not be zero when the client has logged in
	loginToken LoginToken
	dst        string
}

func NewMojiClient(addr string) *mojiClient {
	return &mojiClient{
		httpc: http.DefaultClient,
		dst:   addr,
	}
}

func (mc *mojiClient) SignUp(name, email string) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	req := SignUpRequest{Email: email, Name: name, HashedPassword: "test"}
	if err := enc.Encode(req); err != nil {
		return err
	}
	resp, err := mc.httpc.Post(mc.dst+"/api/v1/sign_up/", "application/json", &buf)
	if err != nil {
		return err
	}
	// TODO check response
	dec := json.NewDecoder(resp.Body)
	var loginToken LoginToken
	if err := dec.Decode(&loginToken); err != nil {
		return err
	}
	mc.loginToken = loginToken
	fmt.Println(resp.Status)
	return nil
}

func (mc *mojiClient) Login(email string) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	req := LoginRequest{Email: email, HashedPassword: "test"}
	if err := enc.Encode(req); err != nil {
		return err
	}
	resp, err := mc.httpc.Post(mc.dst+"/api/v1/log_in/", "application/json", &buf)
	if err != nil {
		return err
	}
	// TODO check response
	dec := json.NewDecoder(resp.Body)
	var loginToken LoginToken
	if err := dec.Decode(&loginToken); err != nil {
		return err
	}
	mc.loginToken = loginToken
	fmt.Println(resp.Status)
	return nil
}

func (mc *mojiClient) FriendOp(to Uuid, op FriendAction) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	payload := FriendRequest{Other: to, LoginToken: mc.loginToken, Action: op}
	if err := enc.Encode(payload); err != nil {
		return err
	}
	resp, err := mc.httpc.Post(mc.dst+"/api/v1/friend/", "application/json", &buf)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	return nil
}
