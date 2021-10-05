package main

import (
	"testing"
	//"encoding/json"
)

func TestGenerateUuid(t *testing.T) {
	uuid, err := generateUuid()
	if err != nil {
		t.Fatalf("%v %v", uuid, err)
	}
	//  bytes, err := json.Marshal(uuid)
}
