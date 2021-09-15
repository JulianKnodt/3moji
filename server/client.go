package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
  "io/ioutil"
)

// mojiClient is a test client for the mojiServer
type mojiClient struct {
	httpc *http.Client
	dst   string
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
	if err := enc.Encode(SignUpPayload{Email: email, Name: name}); err != nil {
    return err
  }
	resp, err := mc.httpc.Post(
		mc.dst + "/api/v1/sign_up/",
		"application/json",
		&buf,
	)
	if err != nil {
		return err
	}
	// TODO check response
  content, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.Status, string(content))
	return nil
}

