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
	user       User
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
	fmt.Println(resp.Status)
	dec := json.NewDecoder(resp.Body)
	var login LoginResponse
	if err := dec.Decode(&login); err != nil {
		return err
	}
	mc.loginToken = login.LoginToken
	mc.user = login.User
	return nil
}

func (mc *mojiClient) UserID() Uuid {
	return mc.user.Uuid
}

func (mc *mojiClient) ListPeople() error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	req := ListPeopleRequest{LoginToken: mc.loginToken, Amount: 50, Kind: 0}
	if err := enc.Encode(req); err != nil {
		return err
	}
	resp, err := mc.httpc.Post(mc.dst+"/api/v1/list_friends/", "application/json", &buf)
	if err != nil {
		return err
	}
	// TODO check response
	dec := json.NewDecoder(resp.Body)
	var listResp ListPeopleResponse
	if err := dec.Decode(&listResp); err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(listResp)
	return nil
}

func (mc *mojiClient) ListGroups(op ListGroupKind) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	req := ListGroupRequest{LoginToken: mc.loginToken, Amount: 50, Kind: op}
	if err := enc.Encode(req); err != nil {
		return err
	}
	resp, err := mc.httpc.Post(mc.dst+"/api/v1/list_groups/", "application/json", &buf)
	if err != nil {
		return err
	}
	// TODO check response
	dec := json.NewDecoder(resp.Body)
	var listResp ListGroupResponse
	if err := dec.Decode(&listResp); err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(listResp)
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

func (mc *mojiClient) GroupOp(name string, groupUuid Uuid, op GroupOp) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	payload := GroupRequest{
		Kind: op, GroupName: name, GroupUuid: groupUuid, LoginToken: mc.loginToken,
	}
	if err := enc.Encode(payload); err != nil {
		return err
	}
	resp, err := mc.httpc.Post(mc.dst+"/api/v1/groups/", "application/json", &buf)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	return nil
}
