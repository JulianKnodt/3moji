package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) ListPeopleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(404)
			return
		}
		dec := json.NewDecoder(r.Body)
		var req ListPeopleRequest
		if err := dec.Decode(&req); err != nil {
			fmt.Printf("Invalid request: %v\n", err)
			w.WriteHeader(401)
			return
		}
		/*
			    // TODO why is validating tokens not working?
					if err := s.ValidateLoginToken(req.LoginToken); err != nil {
						fmt.Printf("Invalid login token: %v\n", err)
						w.WriteHeader(401)
						return
					}
		*/
		user, exists := s.UserFor(req.LoginToken)
		if !exists {
			fmt.Println("User does not exist")
			w.WriteHeader(401)
			return
		}
		amt := req.Amount
		var resp ListPeopleResponse
		s.mu.Lock()
		defer s.mu.Unlock()
		var cond func(*User) bool
		switch req.Kind {
		case All:
			cond = func(u *User) bool { return true }
		case OnlyFriends:
			cond = func(u *User) bool {
				_, exists := s.Friends[user.Uuid][u.Uuid]
				return exists
			}
		case NotFriends:
			cond = func(u *User) bool {
				_, exists := s.Friends[user.Uuid][u.Uuid]
				return !exists
			}
		default:
			w.WriteHeader(404)
			return
		}
		// TODO this is inefficient since we explicitly iterate over everyone.
		// Probably need to fix later when actually using a database.
		for _, person := range s.Users {
			if person.Uuid == user.Uuid {
				continue
			}
			if !cond(person) {
				continue
			}
			resp.People = append(resp.People, *person)
			amt -= 1
			if amt == 0 {
				break
			}
		}
		enc := json.NewEncoder(w)
		enc.Encode(resp)
		return
	}
}

func (s *Server) AckMsgHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			return
		}
		var req AckMsgRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			fmt.Printf("Error decoding recv message %v", err)
			w.WriteHeader(401)
			return
		}
		token := req.LoginToken
		if err := s.ValidateLoginToken(token); err != nil {
			w.WriteHeader(401)
			return
		}
		user, exists := s.UserFor(token)
		if !exists {
			w.WriteHeader(401)
			return
		}
		originalMessage, exists := s.Messages[req.MsgID]
		if !exists {
			w.WriteHeader(404)
			return
		}
		delete(s.Messages, req.MsgID)

		replyUuid, err := generateUuid()
		if err != nil {
			w.WriteHeader(500)
			return
		}
		s.UserToReplies[user.Uuid] = append(s.UserToReplies[user.Uuid], replyUuid)
		// TODO check for collisions?
		s.Replies[replyUuid] = MessageReply{
			Message:         req.MsgID,
			OriginalContent: originalMessage.Emojis,
			Reply:           req.Reply,
			From:            *user,
		}

		w.WriteHeader(200)
		return
	}
}
