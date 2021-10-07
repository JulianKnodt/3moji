package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

func (s *Server) SignUpHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			fmt.Fprint(w, "Not a POST request")
			return
		}
		dec := json.NewDecoder(r.Body)
		var sup SignUpRequest
		if err := dec.Decode(&sup); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Incorrect sign up format: %v", err)
			return
		}
		email, err := NewEmail(sup.Email)
		if err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Failed when parsing email: %v", err)
			return
		}
		enc := json.NewEncoder(w)
		_, err = s.SignUp(email, sup.Name, sup.HashedPassword)
		if err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Failed when signing up: %v", err)
			return
		}
		loginToken, err := s.Login(email, sup.HashedPassword)
		if err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Failed when logging up: %v", err)
			return
		}
		user, exists := s.UserFor(loginToken)
		if !exists {
			w.WriteHeader(500)
			fmt.Fprintf(w, "User does not exist after signing up: %v", err)
			return
		}
		resp := LoginResponse{
			User:       *user,
			LoginToken: loginToken,
		}
		enc.Encode(resp)
		return
	}
}

func (s *Server) LoginHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			fmt.Fprint(w, "Not a POST request")
			return
		}
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		var lp LoginRequest
		if err := dec.Decode(&lp); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error decoding request: %v", err)
			return
		}
		email, err := NewEmail(lp.Email)
		if err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error logging in, email does not appear to be an email: %v", err)
			return
		}
		loginToken, err := s.Login(email, lp.HashedPassword)
		if err != nil {
			w.WriteHeader(401)
			fmt.Fprint(w, "Error logging in, email or password may be incorrect")
			return
		}
		user, exists := s.UserFor(loginToken)
		if !exists {
			w.WriteHeader(500)
			fmt.Fprint(w, "User does not exist")
			return
		}
		resp := LoginResponse{
			User:       *user,
			LoginToken: loginToken,
		}
		enc := json.NewEncoder(w)
		enc.Encode(resp)
		return
	}
}

func (s *Server) ListPeopleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			fmt.Fprint(w, "Not a POST request")
			return
		}
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		var req ListPeopleRequest
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Malformed request: %v", err)
			return
		}
		// Cap the amount manually.
		if req.Amount > 50 {
			req.Amount = 50
		}
		if err := s.ValidateLoginToken(req.LoginToken); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Invalid login token: %v", err)
			return
		}
		user, exists := s.UserFor(req.LoginToken)
		if !exists {
			w.WriteHeader(401)
			fmt.Fprint(w, "User does not exist")
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
			// Omitted due to separate loop below
		case NotFriends:
			cond = func(u *User) bool {
				_, exists := s.Friends[user.Uuid][u.Uuid]
				return !exists
			}
		default:
			w.WriteHeader(404)
			fmt.Fprintf(w, "Unexpected list kind: %v", req.Kind)
			return
		}

		if req.Kind == OnlyFriends {
			for uuid := range s.Friends[user.Uuid] {
				person, exists := s.Users[uuid]
				if !exists || uuid == user.Uuid {
					continue
				}
				resp.People = append(resp.People, *person)
				amt -= 1
				if amt == 0 {
					break
				}
			}
		} else {
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
			fmt.Fprint(w, "Not a POST request")
			return
		}
		var req AckMsgRequest
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error decoding recv message %v", err)
			return
		}
		token := req.LoginToken
		if err := s.ValidateLoginToken(token); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error validating login token %v", err)
			return
		}
		user, exists := s.UserFor(token)
		if !exists {
			w.WriteHeader(401)
			fmt.Fprint(w, "User does not exist")
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		originalMessage, exists := s.Messages[req.MsgID]
		if !exists {
			w.WriteHeader(404)
			fmt.Fprint(w, "Message being replied to could not be found!")
			return
		}
		// Do not delete message here, let it naturally expire but now a user should not be able to
		// see it anymore.
		delete(s.UserToMessages[user.Uuid], req.MsgID)
		// delete(s.Messages, req.MsgID)

		replyUuid, err := generateUuid()
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "Internal server error")
			return
		}
		// TODO check for collisions?
		s.Replies[replyUuid] = MessageReply{
			Message:         req.MsgID,
			OriginalContent: originalMessage.Emojis,
			Reply:           req.Reply,
			From:            *user,
		}
		s.UserToReplies[user.Uuid] = append(s.UserToReplies[user.Uuid], replyUuid)

		w.WriteHeader(200)
		return
	}
}

func (s *Server) GroupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			fmt.Fprint(w, "Not a POST request")
			return
		}
		var req GroupRequest
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error decoding request: %v", err)
			return
		}
		token := req.LoginToken
		if err := s.ValidateLoginToken(token); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error validating request: %v", err)
			return
		}
		user, exists := s.UserFor(token)
		if !exists {
			w.WriteHeader(401)
			fmt.Fprint(w, "User does not exist")
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		switch req.Kind {
		case JoinGroup:
			if _, exists := s.Groups[req.GroupUuid]; !exists {
				w.WriteHeader(404)
				fmt.Fprint(w, "No Such group")
				return
			}
			s.Groups[req.GroupUuid].Users[user.Uuid] = struct{}{}
			s.UsersToGroups[user.Uuid][req.GroupUuid] = struct{}{}
		case LeaveGroup:
			if _, exists := s.Groups[req.GroupUuid]; !exists {
				w.WriteHeader(404)
				fmt.Fprint(w, "No Such group")
				return
			}
			delete(s.Groups[req.GroupUuid].Users, user.Uuid)
			delete(s.UsersToGroups[user.Uuid], req.GroupUuid)
			if len(s.Groups[req.GroupUuid].Users) == 0 {
				delete(s.Groups, req.GroupUuid)
			}
		case CreateGroup:
			if len(req.GroupName) < 3 {
				w.WriteHeader(401)
				fmt.Fprint(w, "Must specify at least 3 characters for group name")
				return
			}
			uuid, err := generateUuid()
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprint(w, "Error while generating Uuid")
				return
			}
			group := Group{
				Uuid: uuid,
				Name: req.GroupName,
				Users: map[Uuid]struct{}{
					user.Uuid: struct{}{},
				},
			}
			s.Groups[uuid] = group
			if s.UsersToGroups[user.Uuid] == nil {
				s.UsersToGroups[user.Uuid] = map[Uuid]struct{}{}
			}
			s.UsersToGroups[user.Uuid][uuid] = struct{}{}
		}

		w.WriteHeader(200)
		return
	}
}

func (s *Server) ListGroupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			fmt.Fprint(w, "Not a POST request")
			return
		}
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		var req ListGroupRequest
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Invalid request: %v\n", err)
			return
		}
		// TODO why is validating tokens not working?
		if err := s.ValidateLoginToken(req.LoginToken); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error validating login token: %v", err)
			return
		}
		user, exists := s.UserFor(req.LoginToken)
		if !exists {
			w.WriteHeader(401)
			fmt.Fprint(w, "User does not exist")
			return
		}
		amt := req.Amount
		var resp ListGroupResponse
		s.mu.Lock()
		defer s.mu.Unlock()
		var cond func(Group) bool
		switch req.Kind {
		case AllGroups:
			cond = func(Group) bool { return true }
		case JoinedGroups:
			cond = func(g Group) bool {
				_, exists := s.UsersToGroups[user.Uuid][g.Uuid]
				return exists
			}
		case NotJoinedGroups:
			cond = func(g Group) bool {
				_, exists := s.UsersToGroups[user.Uuid][g.Uuid]
				return !exists
			}
		default:
			w.WriteHeader(404)
			fmt.Fprint(w, "Invalid op kind")
			return
		}
		// TODO this is inefficient since we explicitly iterate over everyone.
		// Probably need to fix later when actually using a database.
		for _, group := range s.Groups {
			if !cond(group) {
				continue
			}
			resp.Groups = append(resp.Groups, group)
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

func (s *Server) RecommendationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			fmt.Fprint(w, "Not a POST request")
			return
		}
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		var req RecommendationRequest
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error parsing request: %v", err)
			return
		}
		// TODO recommendations should query a data structure which specifies user usage at a given
		// hour.
		var resp RecommendationResponse
		switch int(math.Round(req.LocalTime)) % 24 {
		case 6, 7, 8, 9:
			resp.Recommendations = []EmojiContent{
				"🥞🍳🥓",
				"🫖'🥐🌅",
				"🏃🌄🚲",
				"💪🤸💪",
			}
		case 12, 13:
			resp.Recommendations = []EmojiContent{
				"🍕🍔🌯",
				"🥗🥙🍲",
				"🍱🍚🍛",
			}
		case 16, 17:
			resp.Recommendations = []EmojiContent{
				"🏀🎾🏐",
				"🎥🕴🎦",
			}
		case 21, 22:
			resp.Recommendations = []EmojiContent{
				"🍷🎉🍹",
				"🍰🍦🧋'",
			}
		case 23, 0, 1:
			resp.Recommendations = []EmojiContent{
				"🌌🚶🌃",
			}
		}
		resp.Recommendations = append(
			resp.Recommendations,
			s.FindNearRecommendations(5, req.LocalTime)...,
		)
		enc := json.NewEncoder(w)
		enc.Encode(resp)
		return
	}
}

func (s *Server) FriendHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			fmt.Fprint(w, "Not a POST request")
			return
		}
		var fp FriendRequest
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		if err := dec.Decode(&fp); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error decoding request: %v", err)
			return
		}
		if err := s.ValidateLoginToken(fp.LoginToken); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error validating login token: %v", err)
			return
		}
		user, exists := s.UserFor(fp.LoginToken)
		if !exists {
			w.WriteHeader(401)
			fmt.Fprint(w, "User does not exist")
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		switch fp.Action {
		case Rmfriend:
			delete(s.Friends[user.Uuid], fp.Other)
		case AddFriend:
			if _, exists := s.Friends[user.Uuid]; !exists {
				s.Friends[user.Uuid] = map[Uuid]struct{}{}
			}
			s.Friends[user.Uuid][fp.Other] = struct{}{}
		default:
			w.WriteHeader(404)
			fmt.Fprint(w, "Unknown friend action")
			return
		}
		w.WriteHeader(200)
		return
	}
}

func (s *Server) SendMsgHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			fmt.Fprint(w, "Not a POST request")
			return
		}
		var req SendMessageRequest
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error decoding request: %v", err)
			return
		}
		if err := s.ValidateLoginToken(req.LoginToken); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error validating login token: %v", err)
			return
		}
		user, exists := s.UserFor(req.LoginToken)
		if !exists {
			w.WriteHeader(401)
			fmt.Fprint(w, "Could not find user sending message")
			return
		}

		msg := req.Message
		msg.Source = *user
		var err error
		if msg.Uuid, err = generateUuid(); err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "Internal server error")
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()
		s.Messages[msg.Uuid] = msg
		switch req.RecipientKind {
		case MsgGroup:
			group, exists := s.Groups[req.To]
			if !exists {
				w.WriteHeader(401)
				fmt.Fprint(w, "Group does not exist")
				return
			}
			msg.SentTo = group.Name
			for userUuid := range group.Users {
				if s.UserToMessages[userUuid] == nil {
					s.UserToMessages[userUuid] = map[Uuid]struct{}{}
				}
				s.UserToMessages[userUuid][msg.Uuid] = struct{}{}
			}
		case MsgFriend:
			user, exists := s.Users[req.To]
			if !exists {
				w.WriteHeader(401)
				fmt.Fprint(w, "User does not exist")
				return
			}
			msg.SentTo = user.Name
			if s.UserToMessages[req.To] == nil {
				s.UserToMessages[req.To] = map[Uuid]struct{}{}
			}
			s.UserToMessages[req.To][msg.Uuid] = struct{}{}
		}

		go s.LogEmojiContent(msg.Emojis, msg.LocalTime)

		w.WriteHeader(200)
		return
	}
}

func (s *Server) RecvMsgHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			fmt.Fprint(w, "Not a post request")
			return
		}
		var req RecvMsgRequest
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error decoding request: %v", err)
			return
		}
		token := req.LoginToken
		if err := s.ValidateLoginToken(token); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error validating login token: %v", err)
			return
		}
		user, exists := s.UserFor(token)
		if !exists {
			w.WriteHeader(401)
			fmt.Fprint(w, "User does not exist")
			return
		}

		var out RecvMsgResponse
		now := time.Now()

		s.mu.Lock()
		defer s.mu.Unlock()
		for uuid := range s.UserToMessages[user.Uuid] {
			msg, exists := s.Messages[uuid]
			if !exists {
				continue
			} else if msg.Expired(now) {
				delete(s.Messages, uuid)
				continue
			}
			out.NewMessages = append(out.NewMessages, msg)
		}
		for _, uuid := range s.UserToReplies[user.Uuid] {
			reply, replyExists := s.Replies[uuid]
			if !replyExists {
				continue
			}
			msg, exists := s.MessageForReplyLocked(reply)
			if !exists {
				continue
			} else if msg.Expired(now) {
				delete(s.Messages, uuid)
				continue
			}
			out.NewReplies = append(out.NewReplies, reply)
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(out); err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "Internal server error")
			return
		}
		// if success then empty out the messages
		if req.DeleteOld {
			s.UserToMessages[user.Uuid] = nil
			s.UserToReplies[user.Uuid] = nil
		}
		return
	}
}
