package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
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
		_, err = s.SignUp(context.Background(), email, sup.Name, sup.HashedPassword)
		if err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Failed when signing up: %v", err)
			return
		}
		loginToken, err := s.Login(context.Background(), email, sup.HashedPassword)
		if err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Failed when logging up: %v", err)
			return
		}
		user, exists := s.UserFor(context.Background(), loginToken)
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
		loginToken, err := s.Login(context.Background(), email, lp.HashedPassword)
		if err != nil {
			w.WriteHeader(401)
			fmt.Fprint(w, "Error logging in, email or password may be incorrect")
			return
		}
		user, exists := s.UserFor(context.Background(), loginToken)
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
		user, exists := s.UserFor(context.Background(), req.LoginToken)
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
				if uuid == user.Uuid {
					continue
				}
				person, err := s.GetUser(context.Background(), uuid)
				if err != nil {
					// TODO log error here
					continue
				} else if person == nil {
					continue
				}
				resp.People = append(resp.People, *person)
				amt -= 1
				if amt == 0 {
					break
				}
			}
		} else {
			users, err := s.GetUsers(context.Background())
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Failed to get users: %v", err)
				return
			}
			for _, person := range users {
				if person.Uuid == user.Uuid {
					continue
				}
				if !cond(&person) {
					continue
				}
				resp.People = append(resp.People, person)
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
			fmt.Fprintf(w, "Error decoding ack message %v", err)
			return
		}
		token := req.LoginToken
		if err := s.ValidateLoginToken(token); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error validating login token %v", err)
			return
		}
		user, exists := s.UserFor(context.Background(), token)
		if !exists {
			w.WriteHeader(401)
			fmt.Fprint(w, "User does not exist")
			return
		}
		// originalMessage, exists := s.Messages[req.MsgID]
		originalMessage, err := s.GetMessage(context.Background(), req.MsgID)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Error retrieving message: %v", err)
			return
		}

		if originalMessage == nil {
			w.WriteHeader(404)
			fmt.Fprint(w, "Message being replied to could not be found!")
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()
		// Do not delete the original message here since other users may need to see it, but now a
		// specific user should not be able to see it anymore.
		delete(s.UserToMessages[user.Uuid], req.MsgID)

		replyUuid, err := generateUuid()
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "Internal server error")
			return
		}
		// TODO check for collisions?
		s.Replies[replyUuid] = &MessageReply{
			Message:         originalMessage,
			OriginalContent: originalMessage.Emojis,
			Reply:           req.Reply,
			From:            *user,
		}
		source := originalMessage.Source
		s.UserToReplies[source.Uuid] = append(s.UserToReplies[source.Uuid], replyUuid)
		go s.sendAckPushNotification(source.Uuid, user.Name, originalMessage.Emojis, req.Reply)
		go s.LogReply(s.Replies[replyUuid])

		w.WriteHeader(200)
		return
	}
}

func (s *Server) sendAckPushNotification(
	senderUuid Uuid,
	responderName string,
	original EmojiContent,
	reply EmojiReply,
) {
	notifToken, err := s.RedisClient.HGet(context.Background(), "user_notif_tokens", senderUuid.String()).Result()
	if err != nil {
		fmt.Printf("Failed to get user notif token: %v", err)
		return
	} else if notifToken == "" {
		return
	}

	pushBody := fmt.Sprintf("%s: %s ↩️ %s", responderName, reply, original)
	pushMsg := expo.PushMessage{
		To:       []expo.ExponentPushToken{expo.ExponentPushToken(notifToken)},
		Body:     pushBody,
		Sound:    "default",
		Title:    "📨↩️",
		Priority: expo.DefaultPriority,
	}
	client := expo.NewPushClient(nil)
	resp, err := client.Publish(&pushMsg)
	if err != nil {
		fmt.Println(err)
	}
	if resp.ValidateResponse() != nil {
		fmt.Println("Failed to send push notification")
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
		user, exists := s.UserFor(context.Background(), token)
		if !exists {
			w.WriteHeader(401)
			fmt.Fprint(w, "User does not exist")
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		switch req.Kind {
		case JoinGroup:
			group, err := s.GetGroup(context.Background(), req.GroupUuid)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Failed to find group: %v", err)
				return
			} else if group == nil {
				w.WriteHeader(404)
				fmt.Fprint(w, "No Such group")
				return
			}
			group.Users[user.Uuid] = user.Name
			s.AddGroup(context.Background(), group)
			if _, exists := s.UsersToGroups[user.Uuid]; !exists {
				s.UsersToGroups[user.Uuid] = map[Uuid]struct{}{}
			}
			s.UsersToGroups[user.Uuid][req.GroupUuid] = struct{}{}
		case LeaveGroup:
			group, err := s.GetGroup(context.Background(), req.GroupUuid)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Failed to find group: %v", err)
				return
			} else if group == nil {
				w.WriteHeader(404)
				fmt.Fprint(w, "No Such group")
				return
			}
			delete(group.Users, user.Uuid)
			delete(s.UsersToGroups[user.Uuid], req.GroupUuid)
			if len(group.Users) == 0 {
				s.DeleteGroup(context.Background(), req.GroupUuid)
			} else {
				s.AddGroup(context.Background(), group)
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
				Users: map[Uuid]string{
					user.Uuid: user.Name,
				},
			}
			s.AddGroup(context.Background(), &group)
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
		user, exists := s.UserFor(context.Background(), req.LoginToken)
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
		groups, err := s.GetGroups(context.Background())
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Failed to get groups: %v", err)
			return
		}
		// TODO this is inefficient since we explicitly iterate over everyone.
		// Probably need to fix later when actually using a database.
		for _, group := range groups {
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
		var recs map[EmojiContent]struct{}
		switch int(math.Round(req.LocalTime)) % 24 {
		case 6, 7, 8, 9:
			recs = map[EmojiContent]struct{}{
				"🥞🍳🥓":  struct{}{},
				"🫖'🥐🌅": struct{}{},
				"🏃🌄🚲":  struct{}{},
				"💪🤸💪":  struct{}{},
			}
		case 12, 13:
			recs = map[EmojiContent]struct{}{
				"🍕🍔🌯": struct{}{},
				"🥗🥙🍲": struct{}{},
				"🍱🍚🍛": struct{}{},
			}
		case 16, 17:
			recs = map[EmojiContent]struct{}{
				"🏀🎾🏐": struct{}{},
				"🎥🕴🎦": struct{}{},
			}
		case 18, 19:
			recs = map[EmojiContent]struct{}{
				"🍕🍔🌯": struct{}{},
				"🥗🥙🍲": struct{}{},
				"🍱🍚🍛": struct{}{},
			}
		case 21, 22:
			recs = map[EmojiContent]struct{}{
				"🍷🎉🍹": struct{}{},
				"🍰🍦🍡": struct{}{},
			}
		case 23, 0, 1:
			recs = map[EmojiContent]struct{}{
				"🌌🚶🌃": struct{}{},
			}
		}
		for _, v := range s.FindNearRecommendations(5, req.LocalTime) {
			recs[v] = struct{}{}
		}
		var resp RecommendationResponse
		for rec := range recs {
			resp.Recommendations = append(resp.Recommendations, rec)
		}
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
		user, exists := s.UserFor(context.Background(), fp.LoginToken)
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
		user, exists := s.UserFor(context.Background(), req.LoginToken)
		if !exists {
			w.WriteHeader(401)
			fmt.Fprint(w, "Could not find user sending message")
			return
		}

		msg := &req.Message
		msg.Source = *user
		var err error
		if msg.Uuid, err = generateUuid(); err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "Internal server error")
			return
		}

		s.mu.Lock()
		defer s.mu.Unlock()
		var uuids []Uuid
		switch req.RecipientKind {
		case MsgGroup:
			group, err := s.GetGroup(context.Background(), req.To)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Failed to get group: %v", err)
				return
			} else if group == nil {
				w.WriteHeader(404)
				fmt.Fprint(w, "Group does not exist")
				return
			}
			msg.SentTo = group.Name
			if !exists {
				w.WriteHeader(401)
				fmt.Fprint(w, "Group does not exist")
				return
			}
			msg.SentTo = group.Name
			s.AddMessage(context.Background(), msg)
			for userUuid := range group.Users {
				if s.UserToMessages[userUuid] == nil {
					s.UserToMessages[userUuid] = map[Uuid]struct{}{}
				}
				s.UserToMessages[userUuid][msg.Uuid] = struct{}{}
				uuids = append(uuids, userUuid)
			}
		case MsgFriend:
			user, err := s.GetUser(context.Background(), req.To)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Error getting user: %v", err)
				return
			} else if user == nil {
				w.WriteHeader(401)
				fmt.Fprint(w, "User does not exist")
				return
			}
			msg.SentTo = user.Name
			s.AddMessage(context.Background(), msg)
			if s.UserToMessages[req.To] == nil {
				s.UserToMessages[req.To] = map[Uuid]struct{}{}
			}
			s.UserToMessages[req.To][msg.Uuid] = struct{}{}
			uuids = []Uuid{req.To}
		default:
			w.WriteHeader(404)
			fmt.Fprint(w, "Unknown recipient kind")
			return
		}

		go s.LogEmojiContent(msg.Emojis, msg.LocalTime)
		go s.sendMessagePushNotification(uuids, user.Name, msg.Emojis, msg.Location)

		w.WriteHeader(200)
		return
	}
}

func (s *Server) sendMessagePushNotification(
	uuids []Uuid,
	name string,
	emojis EmojiContent,
	location string,
) {
	var to []expo.ExponentPushToken
	s.mu.Lock()
	for _, uuid := range uuids {
		notifToken, err := s.RedisClient.HGet(context.Background(), "user_notif_tokens", uuid.String()).Result()
		if err != nil {
			fmt.Printf("Failed to get user notif token: %v", err)
			continue
		} else if notifToken == "" {
			continue
		}
		to = append(to, expo.ExponentPushToken(notifToken))
	}
	s.mu.Unlock()
	if len(to) == 0 {
		return
	}
	var pushBody string
	if location == "" {
		pushBody = fmt.Sprintf("%s: %s❓", name, emojis)
	} else {
		pushBody = fmt.Sprintf("%s: %s❓ @ %s", name, emojis, location)
	}
	pushMsg := expo.PushMessage{
		To:       to,
		Body:     pushBody,
		Sound:    "default",
		Title:    "📨‼️",
		Priority: expo.DefaultPriority,
	}
	client := expo.NewPushClient(nil)
	resp, err := client.Publish(&pushMsg)
	if err != nil {
		fmt.Println(err)
	}
	if resp.ValidateResponse() != nil {
		fmt.Println("Failed to send push notification")
	}
}

func (s *Server) PushNotifTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(400)
			fmt.Fprint(w, "Not a post request")
			return
		}
		var req PushNotifTokenRequest
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error decoding request: %v", err)
			return
		}
		if req.Token == "" {
			w.WriteHeader(401)
			fmt.Fprint(w, "Cannot send empty notification token")
			return
		}
		token := req.LoginToken
		if err := s.ValidateLoginToken(token); err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "Error validating login token: %v", err)
			return
		}

		user, exists := s.UserFor(context.Background(), token)
		if !exists {
			w.WriteHeader(401)
			fmt.Fprint(w, "User does not exist")
			return
		}

		switch req.Kind {
		case AddNotifToken:
			expoToken, err := expo.NewExponentPushToken(req.Token)
			if err != nil {
				w.WriteHeader(401)
				fmt.Fprintf(w, "Error parsing expo token: %v", err)
				return
			}
			err = s.RedisClient.HSet(
				context.Background(), "user_notif_tokens", user.Uuid.String(), string(expoToken),
			).Err()
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Failed to save notification setting: %v", err)
				return
			}

			w.WriteHeader(200)
		case RmNotifToken:
			err := s.RedisClient.HDel(
				context.Background(), "user_notif_tokens", user.Uuid.String(),
			).Err()
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Failed to save notification setting: %v", err)
				return
			}

			w.WriteHeader(200)
		default:
			w.WriteHeader(404)
			fmt.Fprintf(w, "Unknown token action %v", req.Kind)
		}
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
		user, exists := s.UserFor(context.Background(), token)
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
			msg, err := s.GetMessage(context.Background(), uuid)
			if err != nil {
				// TODO report error here
				continue
			} else if msg == nil {
				continue
			} else if msg.Expired(now) {
				s.DeleteMessage(context.Background(), uuid)
				continue
			}
			out.NewMessages = append(out.NewMessages, msg)
		}
		for _, uuid := range s.UserToReplies[user.Uuid] {
			reply, replyExists := s.Replies[uuid]
			if !replyExists {
				continue
			}
			msg, err := s.MessageForReply(context.Background(), reply)
			if err != nil {
				// TODO report error here
				continue
			} else if msg == nil {
				continue
			} else if msg.Expired(now) {
				s.DeleteMessage(context.Background(), uuid)
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

// Handler which returns a summary of all the data gathered on the server.
func (s *Server) SummaryHandler() http.HandlerFunc {
	// Buffer this so concurrent requests cannot overload the server/redis.
	n := 5
	buffer := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		buffer <- struct{}{}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		<-buffer
		defer func() { buffer <- struct{}{} }()
		emojisSent, err := s.RedisClient.HGetAll(context.Background(), "emojis_sent").Result()
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Failed to get counts: %v", err)
			return
		}
		emojisSentAt, err := s.RedisClient.HGetAll(context.Background(), "emojis_sent_at").Result()
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Failed to get times: %v", err)
			return
		}
		repliesSent, err := s.RedisClient.HGetAll(context.Background(), "emoji_reply").Result()
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Failed to get replies: %v", err)
			return
		}
		out := SummaryResponse{
			Counts:      make(map[string]int, len(emojisSent)),
			Times:       make(map[string]float64, len(emojisSentAt)),
			ReplyCounts: make(map[string]int, len(repliesSent)),
		}
		for emojis, count := range emojisSent {
			out.Counts[emojis], err = strconv.Atoi(count)
			if err != nil {
				fmt.Printf("Failed to parse count: %v", err)
				continue
			}
		}
		for emojis, sentAt := range emojisSentAt {
			out.Times[emojis], err = strconv.ParseFloat(sentAt, 64)
			if err != nil {
				fmt.Printf("Failed to parse time: %v", err)
				continue
			}
		}
		for replies, count := range repliesSent {
			out.ReplyCounts[replies], err = strconv.Atoi(count)
			if err != nil {
				fmt.Printf("Failed to parse reply count: %v", err)
				continue
			}
		}

		enc := json.NewEncoder(w)
		enc.Encode(out)
		return
	}
}
