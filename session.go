package main

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

var (
	cookie = "_CHANGE_THIS_"
	store  = *sessions.NewCookieStore([]byte("_CHANGE_THIS_AS_WELL_"))
)

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 4, // 4hrs timeout
		HttpOnly: true,
		Secure:   true, // only work on https
	}
}

// Errors encountered when handling sessions
var (
	ErrSess        = errors.New("invalid session")
	ErrLoggedIn    = errors.New("user is logged in")
	ErrNotLoggedIn = errors.New("user is not logged in")
)

func login(w http.ResponseWriter, r *http.Request, username string) error {
	s, err := store.Get(r, cookie)
	if err != nil {
		return ErrSess
	}
	s.Values["username"] = username
	if err := s.Save(r, w); err != nil {
		return ErrSess
	}
	return nil
}
func logout(w http.ResponseWriter, r *http.Request) error {
	s, err := store.Get(r, cookie)
	if err != nil {
		return ErrSess
	}
	for k := range s.Values {
		delete(s.Values, k)
	}
	if err := s.Save(r, w); err != nil {
		return ErrSess
	}
	return nil
}
func isLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	s, err := store.Get(r, cookie)
	if err != nil {
		return false
	}
	val, ok := s.Values["username"]
	if !ok {
		return false
	}
	if _, ok := val.(string); !ok {
		return false
	}

	_ = s.Save(r, w)
	return true
}

// loggedIn is an indempotent version of the isLoggedIn function
func loggedIn(r *http.Request) bool {
	s, err := store.Get(r, cookie)
	if err != nil {
		return false
	}
	val, ok := s.Values["username"]
	if !ok {
		return false
	}
	_, ok = val.(string)
	return ok
}

func loggedInUser(r *http.Request) string {
	s, err := store.Get(r, cookie)
	if err != nil {
		return ""
	}
	val, ok := s.Values["username"]
	if !ok {
		return ""
	}
	u, ok := val.(string)
	if !ok {
		return ""
	}
	return u
}
