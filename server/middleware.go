package server

import (
	"errors"
	"net/http"
	"strings"

	"context"
	"github.com/mikedonnici/rtcl-api/datastore"
)

func (s *server) requireValidToken(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.ValidateAuthHeaderToken(r, "")
		if err != nil {
			respondJSON(w, http.StatusUnauthorized, nil, err)
			return
		}
		h(w, r)
	}
}

// requireValidUserToken middleware gets the token from the Auth header
func (s *server) requireValidUserToken(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		t, err := s.AuthHeaderToken(r)
		if err != nil {
			respondJSON(w, http.StatusUnauthorized, nil, err)
			return
		}

		err = s.ValidateAuthHeaderToken(r, t.Claims.ID)
		if err != nil {
			respondJSON(w, http.StatusUnauthorized, nil, err)
			return
		}

		freshToken, err := s.store.FreshUserToken(t.Claims.ID, s.config.Token.Issuer, s.config.Token.SigningKey, s.config.Token.HoursTTL)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, nil, err)
			return
		}

		w.Header().Add("fresh-token", freshToken)
		ctx := context.WithValue(r.Context(), "userID", t.Claims.ID)
		r = r.WithContext(ctx)
		h(w, r)
	}
}

// ValidateAuthHeaderToken is a middleware helper that validates the token in a request Authorization header.
// If userID is passed in will check it matches the user id in the token.
func (s *server) ValidateAuthHeaderToken(r *http.Request, userID string) error {

	t, err := s.AuthHeaderToken(r)
	if err != nil {
		return err
	}

	if userID != "" {
		if t.Claims.ID != userID {
			return errors.New("token user id mismatch")
		}
	}

	return nil
}

// AuthHeaderToken is a middleware helper function that extracts the token string from the Authorization header
// and decodes it into a Token value
func (s *server) AuthHeaderToken(r *http.Request) (datastore.Token, error) {
	var t datastore.Token

	xs := strings.Fields(r.Header.Get("Authorization"))
	if len(xs) < 2 || xs[0] != "Bearer" {
		return t, errors.New("authorization header should be: Bearer [jwt]")
	}
	ts := strings.TrimSpace(xs[1])

	return datastore.DecodeToken(ts, s.config.Token.SigningKey)
}
