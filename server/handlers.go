package main

import (
	"encoding/json"
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
			w.WriteHeader(401)
			return
		}
		if err := s.ValidateLoginToken(req.LoginToken); err != nil {
			w.WriteHeader(401)
			return
		}
		user, exists := s.UserFor(req.LoginToken)
		if !exists {
			w.WriteHeader(401)
			return
		}
		amt := req.Amount
		var resp ListPeopleResponse
		switch req.Kind {
		case All:
			for _, person := range s.Users {
				resp.People = append(resp.People, person)
				amt -= 1
				if amt == 0 {
					break
				}
			}
		case OnlyFriends:
			for _, person := range s.Users {
				if _, exists := s.Friends[user.Uuid]; !exists {
					continue
				}
				resp.People = append(resp.People, person)
				amt -= 1
				if amt == 0 {
					break
				}
			}
		case NotFriends:
			for _, person := range s.Users {
				if _, exists := s.Friends[user.Uuid]; exists {
					continue
				}
				resp.People = append(resp.People, person)
				amt -= 1
				if amt == 0 {
					break
				}
			}
		default:
			w.WriteHeader(404)
			return
		}
		enc := json.NewEncoder(w)
		enc.Encode(resp)
		return
	}
}
