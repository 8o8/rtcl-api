// Package server brings together various packages to provide a complete http service.
// Func names:
// * funcs that handle a request directly are preceded with 'handle', eg indexHandler()
// * funcs that return a http.HandlerFunc are appended with 'Handler', eg indexHandler()
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mikedonnici/pubmed"
	"github.com/mikedonnici/rtcl-api/datastore"
	"github.com/mikedonnici/rtcl-api/emailer"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"log"
	"net/http"
)

func (s *server) routes() {

	s.router.HandleFunc("/", s.optionsHandler()).Methods("OPTIONS")
	s.router.HandleFunc("/", s.indexHandler()).Methods("GET")
	s.router.HandleFunc("/r/{pmid}", s.redirectHandler()).Methods("GET")
	s.router.HandleFunc("/favicon.ico", s.faviconHandler()).Methods("GET")
	s.router.HandleFunc("/auth", s.authHandler()).Methods("POST")

	// these should all require an app client key
	s.router.HandleFunc("/users", s.addUserHandler()).Methods("POST")
	s.router.HandleFunc("/users/{id}", s.userByIDHandler()).Methods("GET")
	s.router.HandleFunc("/users/{id}/notifications/{notification}", s.userNotificationHandler()).Methods("POST")
	s.router.HandleFunc("/users/{id}/confirm/{key}", s.userConfirmationHandler()).Methods("GET")

	// Auth Middleware
	s.router.HandleFunc("/user", s.requireValidUserToken(s.userByTokenHandler())).Methods("GET")
	s.router.HandleFunc("/user", s.requireValidUserToken(s.updateUserHandler())).Methods("PUT")
	s.router.HandleFunc("/user/search", s.requireValidUserToken(s.saveSearchHandler())).Methods("POST")
	s.router.HandleFunc("/user/search", s.requireValidUserToken(s.deleteSearchHandler())).Methods("DELETE")
	s.router.HandleFunc("/user/log", s.requireValidUserToken(s.saveLogHandler())).Methods("POST")
	s.router.HandleFunc("/user/logs", s.requireValidUserToken(s.userLogsHandler())).Methods("GET")
	s.router.HandleFunc("/user/log/{id}", s.requireValidUserToken(s.deleteLogHandler())).Methods("DELETE")
}

func (s *server) optionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}
}

// indexHandler handles requests for resource
func (s *server) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"hi": "leo"}, nil)
	}
}

// redirectHandler handles requests for resource
func (s *server) redirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		pmid := mux.Vars(r)["pmid"]
		if len(pmid) == 0 {
			respondJSON(w, http.StatusBadRequest, nil, errors.New("no article id"))
		}

		article, err := pubmed.ArticleByPMID(pmid)
		if err != nil {
			respondJSON(w, http.StatusNotFound, nil, err)
		}
		if len(article.URL) == 0 {
			respondJSON(w, http.StatusNotFound, nil, errors.New("no url for this article"))
		}

		http.Redirect(w, r, article.URL, http.StatusFound)
	}
}

func (s *server) faviconHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusNoContent, nil, nil)
	}
}

func (s *server) authHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, nil, err)
			return
		}

		u, err := s.store.UserAuth(body.Email, body.Password)
		if err != nil {
			respondJSON(w, http.StatusUnauthorized, nil, errors.New("could not authorize user"))
			return
		}

		t, err := u.Token(s.config.Token.Issuer, s.config.Token.SigningKey, s.config.Token.HoursTTL)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, nil, err)
			return
		}

		responseBody := map[string]string{
			"token":  t.String(),
			"userId": u.ID.Hex(),
		}
		respondJSON(w, http.StatusOK, responseBody, nil)
		return
	}
}

func (s *server) addUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := s.store.NewUser()
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, nil, err)
			return
		}

		err = u.Save()
		if err != nil {
			respondJSON(w, http.StatusConflict, nil, err)
			return
		}

		u.Password = datastore.PasswordMask
		respondJSON(w, http.StatusCreated, u, nil)
	}
}

func (s *server) userByIDHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if !bson.IsObjectIdHex(id) {
			respondJSON(w, http.StatusBadRequest, nil, errors.New("invalid object id"))
			return
		}
		u, err := s.store.UserByID(id)
		if err != nil {
			respondNotFoundOrBadRequest(w, err)
		}

		u.Password = datastore.PasswordMask
		respondJSON(w, http.StatusOK, u, err)
	}
}

// userByTokenHandler returns the user profile for the user identified by userID value in context
func (s *server) userByTokenHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value("userID")
		u, err := s.store.UserByID(id.(string))
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, nil, err)
		}
		u.Password = datastore.PasswordMask
		respondJSON(w, http.StatusOK, u, err)
	}
}

// updateUserHandler updates the user profile identified by userID value in context
func (s *server) updateUserHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value("userID")

		// get the target record
		u, err := s.store.UserByID(id.(string))
		if err != nil {
			respondJSON(w, http.StatusNotFound, nil, err)
			return
		}

		// decode the body of the request to get fields to update
		var body bson.M
		err = json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, nil, err)
			return
		}

		err = u.SavePartial(body)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, nil, err)
			return
		}

		u.Password = datastore.PasswordMask
		respondJSON(w, http.StatusOK, u, err)
	}
}

// saveSearchHandler saves a search for a user
func (s *server) saveSearchHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value("userID")
		u, err := s.store.UserByID(id.(string))
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, nil, err)
			return
		}

		s := datastore.Search{}
		err = json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, nil, err)
			return
		}

		err = u.SaveSearch(s.Query)
		if err != nil {
			respondJSON(w, http.StatusConflict, nil, err)
			return
		}
		u.Password = datastore.PasswordMask
		respondJSON(w, http.StatusCreated, u, err)
	}
}

// deleteSearchHandler returns a hadnler that deletes a saved user search query
func (s *server) deleteSearchHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value("userID")
		u, err := s.store.UserByID(id.(string))
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, nil, err)
			return
		}

		s := datastore.Search{}
		err = json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, nil, err)
			return
		}

		err = u.DeleteSearch(s.Query)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, nil, err)
			return
		}
		u.Password = datastore.PasswordMask
		respondJSON(w, http.StatusOK, u, err)
	}
}

// saveLogHandler saves a log entry for a user
func (s *server) saveLogHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value("userID")
		_, err := s.store.UserByID(id.(string))
		if err != nil {
			respondJSON(w, http.StatusUnauthorized, nil, errors.New("could not get user id from token"))
			return
		}
		defer r.Body.Close()

		l := s.store.NewLog()
		err = json.NewDecoder(r.Body).Decode(&l)
		if err != nil {
			log.Println(err)
			respondJSON(w, http.StatusBadRequest, nil, err)
			return
		}
		l.UserID = bson.ObjectIdHex(id.(string))

		err = l.Save()
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, nil, err)
			return
		}
		respondJSON(w, http.StatusCreated, l, err)
	}
}

// userLogsHandler fetches all of the user logs
func (s *server) userLogsHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID")
		_, err := s.store.UserByID(userID.(string))
		if err != nil {
			respondJSON(w, http.StatusUnauthorized, nil, errors.New("could not get user id from token"))
			return
		}

		xl, err := s.store.LogsByUserID(userID.(string))
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, nil, errors.New("error fetching logs - "+err.Error()))
			return
		}

		respondJSON(w, http.StatusOK, xl, err)
	}
}

func (s *server) deleteLogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if len(id) == 0 {
			respondJSON(w, http.StatusBadRequest, nil, errors.New("no log id"))
			return
		}

		l, err := s.store.LogByID(id)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, nil, errors.New("could not find log with id "+id))
			return
		}

		// check user match here
		userID := r.Context().Value("userID")
		if l.UserID.Hex() != userID {
			respondJSON(w, http.StatusUnauthorized, nil, errors.New("user in token does not match owner of log"))
			return
		}

		err = l.Delete()
		if err != nil {
			respondJSON(w, http.StatusBadRequest, nil, errors.New("error deleting log - "+err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// userNotificationHandler handles notifications for a user. The value of id can be the users ObjectID in the
// database, or the users email address - both are unique.
func (s *server) userNotificationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		idOrEmail := mux.Vars(r)["id"]
		notification := mux.Vars(r)["notification"]

		u, err := s.store.UserByIDOrEmail(idOrEmail)
		if err != nil {
			respondNotFoundOrBadRequest(w, err)
		}

		switch notification {
		case "welcome":
			emailer.WelcomeUser(*u)
			log.Println(fmt.Sprintf("Send %s message to %s (%s)", notification, u.Email, u.ID))
			respondJSON(w, http.StatusAccepted, nil, nil)
		case "reset":
			emailer.ResetPassword(*u)
			log.Println(fmt.Sprintf("Send %s message to %s (%s)", notification, u.Email, u.ID))
			respondJSON(w, http.StatusAccepted, nil, nil)
		}
	}
}

func (s *server) userConfirmationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		key := mux.Vars(r)["key"]

		u, err := s.store.UserByID(id)
		if err == mgo.ErrNotFound {
			respondJSON(w, http.StatusNotFound, nil, errors.New("user not found"))
			return
		}
		if err != nil {
			respondJSON(w, http.StatusBadRequest, nil, err)
			return
		}
		if key != u.KeyGen() {
			respondJSON(w, http.StatusBadRequest, nil, errors.New("invalid key"))
			return
		}

		u.Locked = false
		err = u.Save()
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, nil, err)
		}

		u.Password = datastore.PasswordMask
		respondJSON(w, 200, u, nil)
	}
}

func respondNotFoundOrBadRequest(w http.ResponseWriter, err error) {
	if err == mgo.ErrNotFound {
		respondJSON(w, http.StatusNotFound, nil, errors.New("user not found"))
		return
	}
	if err != nil {
		respondJSON(w, http.StatusBadRequest, nil, err)
		return
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}, err error) {

	var body string
	w.Header().Set("content-type", "application/json")

	// explicitly asked to respond with the error
	if err != nil {
		w.WriteHeader(status)
		body = fmt.Sprintf(`{"error": "%s"}`, err.Error())
		io.WriteString(w, body)
		return
	}

	// an error encoding the response
	xb, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
	}

	w.WriteHeader(status)
	body = string(xb)
	io.WriteString(w, body)
}
