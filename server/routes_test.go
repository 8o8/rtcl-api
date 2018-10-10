package server_test

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"fmt"
	"github.com/matryer/is"
	"github.com/mikedonnici/rtcl-api/datastore"
	"github.com/mikedonnici/rtcl-api/datastore/mongo"
	"github.com/mikedonnici/rtcl-api/server"
	"github.com/mikedonnici/rtcl-api/testdata"
)

var testDB = testdata.New()
var ds = datastore.New()

var srvConfig = server.Config{
	Port: "8888",
	Token: server.TokenConfig{
		Issuer:     "Routes Test",
		SigningKey: "Routes@##!%",
		HoursTTL:   1,
	},
}

// TestRoutes sets up test databases, connects a testDB to the database and starts a server with the datastore.
// It then runs a group of route tests and tears down the test databases.
func TestRoutes(t *testing.T) {

	var err error

	// install databases
	err = setupDatabase()
	if err != nil {
		log.Fatalln(err)
	}
	defer teardownDatabase()

	err = datastoreConnectMongoDB()
	if err != nil {
		log.Fatalln(err)
	}

	// run tests
	t.Run("routes", func(t *testing.T) {
		t.Run("testIndex", testIndex)
		t.Run("testGetUser", testGetUser)
		t.Run("testGetUserBadID", testGetUserBadID)
		t.Run("testAddUser", testAddUser)
		t.Run("testUpdateUser", testUpdateUser)
		t.Run("testAddUserAlreadyExists", testAddUserAlreadyExists)
		t.Run("testAddUserBadBody", testAddUserBadBody)
		t.Run("testAuthUser", testAuthUser)
		t.Run("testMe", testMe)
		t.Run("testSaveSearch", testSaveSearch)
		t.Run("testDeleteSearch", testDeleteSearch)
		t.Run("testRedirect", testRedirect)
		t.Run("testSaveLog", testSaveLog)
		t.Run("testFetchUserLogs", testFetchUserLogs)
		t.Run("testDeleteLog", testDeleteLog)
	})
}

// setUpDatabase creates and populates test database
func setupDatabase() error {
	return testDB.SetupMongoDB()
}

// teardownDatabase cleans up the test databases
func teardownDatabase() {
	testDB.TearDownMongoDB()
}

// datastoreConnectMongoDB connects the datastore to the test Mongo database
func datastoreConnectMongoDB() error {
	var err error
	ds.Mongo, err = mongo.NewConnection(testdata.MongoDSN, testDB.DBName, "test")
	return err
}

func testIndex(t *testing.T) {
	is := is.New(t)
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv := server.NewServer(srvConfig, ds)
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, 200) // response not 200 ok
}

func testGetUser(t *testing.T) {
	is := is.New(t)
	r := httptest.NewRequest("GET", "/users/5b3bcd72463cd6029e04de18", nil)
	w := httptest.NewRecorder()
	srv := server.NewServer(srvConfig, ds)
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, 200) // expected 200 ok
}

func testGetUserBadID(t *testing.T) {
	is := is.New(t)
	r := httptest.NewRequest("GET", "/users/notarealid", nil)
	w := httptest.NewRecorder()
	srv := server.NewServer(srvConfig, ds)
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, 400) // expected response 400 Bad Request
}

func testAddUser(t *testing.T) {
	is := is.New(t)

	b := strings.NewReader(`{"firstName": "Barry", "lastName": "Smith", "email": "bs@rtcl.io"}`)
	r := httptest.NewRequest("POST", "/users", b)
	w := httptest.NewRecorder()
	srv := server.NewServer(srvConfig, ds)
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, 201) // response not 201 created

	uj := datastore.User{}
	err := json.NewDecoder(w.Body).Decode(&uj)
	is.NoErr(err)           // error decoding response
	is.True(len(uj.ID) > 0) // expect ID len > 0

	// Fetch user and check
	u := ds.NewUser()
	err = u.ByID(uj.ID.Hex())
	is.NoErr(err)                   // error fetching newly created user
	is.Equal(u.FirstName, "Barry")  // Wrong first name
	is.Equal(u.LastName, "Smith")   // Wrong last name
	is.Equal(u.Email, "bs@rtcl.io") // Wrong email
}

func testUpdateUser(t *testing.T) {
	is := is.New(t)

	// generate a valid token for a user that is in the test database
	u, err := ds.UserByID("5b3bcd72463cd6029e04de1c")
	is.NoErr(err) // error fetching user record
	tk, err := u.Token(srvConfig.Token.Issuer, srvConfig.Token.SigningKey, 1)
	is.NoErr(err) // error generating token

	// Only change last name
	b := strings.NewReader(`{"lastName": "Hayes-Update"}`)
	r := httptest.NewRequest("PUT", "/user", b)
	r.Header.Set("Authorization", "Bearer "+tk.String())
	w := httptest.NewRecorder()
	srv := server.NewServer(srvConfig, ds)
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, 200) // expected 200 ok
}

func testAddUserAlreadyExists(t *testing.T) {
	is := is.New(t)
	b := strings.NewReader(`{"firstName" :"Broderick", "lastName" : "Reynolds", "email" : "br@rtcl.io"}`)
	r := httptest.NewRequest("POST", "/users", b)
	w := httptest.NewRecorder()
	srv := server.NewServer(srvConfig, ds)
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, 409) // expected response 409 Conflict
}

func testAddUserBadBody(t *testing.T) {
	is := is.New(t)
	// Body is malformed ... missing first "
	b := strings.NewReader(`{firstName" :"Doogie"", "lastName" : "Jangles", "email" : "doogiej@rtcl.io"}`)
	r := httptest.NewRequest("POST", "/users", b)
	w := httptest.NewRecorder()
	srv := server.NewServer(srvConfig, ds)
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, 400) // expected response 400 Bad Request
}

func testAuthUser(t *testing.T) {
	is := is.New(t)
	b := strings.NewReader(`{"email": "br@rtcl.io", "password": "12345abcde"}`)
	r := httptest.NewRequest("POST", "/auth", b)
	w := httptest.NewRecorder()
	srv := server.NewServer(srvConfig, ds)
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, 200) // expected 200 OK

	// should get token in body
	body := struct {
		JWT string `json:"token"`
	}{}
	json.Unmarshal(w.Body.Bytes(), &body)
	is.True(len(strings.Split(body.JWT, ".")) == 3) // doesn't look like a token
}

// testme test the user endpoint that fetches the logged in user profile using a token
func testMe(t *testing.T) {
	is := is.New(t)

	// generate a valid token for a user that is in the test database
	u, err := ds.UserByID("5b3bcd72463cd6029e04de18")
	is.NoErr(err) // error fetching user record
	tk, err := u.Token(srvConfig.Token.Issuer, srvConfig.Token.SigningKey, 1)
	is.NoErr(err) // error generating token

	// call /me with valid token
	r := httptest.NewRequest("GET", "/user", nil)
	r.Header.Set("Authorization", "Bearer "+tk.String())
	w := httptest.NewRecorder()
	srv := server.NewServer(srvConfig, ds)
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, 200) // expected 200 OK

	// again with bad token
	r = httptest.NewRequest("GET", "/user", nil)
	r.Header.Set("Authorization", "Bearer "+"dodgyToken")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, 401) // expected 401 unauthorized
}

// testSaveSearch test the saving of a user search to the user's record
func testSaveSearch(t *testing.T) {
	is := is.New(t)
	srv := server.NewServer(srvConfig, ds)

	// generate a valid token for a user that is in the test database
	u, err := ds.UserByID("5b3bcd72463cd6029e04de18")
	is.NoErr(err) // error fetching user record
	tk, err := u.Token(srvConfig.Token.Issuer, srvConfig.Token.SigningKey, 1)
	is.NoErr(err)                      // error generating token
	currentSearches := len(u.Searches) // current number of searches

	body := strings.NewReader(`{"query": "a new search string to save"}`)
	r := httptest.NewRequest("POST", "/user/search", body)
	r.Header.Set("Authorization", "Bearer "+tk.String())
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusCreated)             // expected 201 Created
	u, err = ds.UserByID("5b3bcd72463cd6029e04de18") // fetch user record again
	is.NoErr(err)                                    // error re-fetching user record
	is.True(len(u.Searches) == currentSearches+1)    // expect number of searches to increase by 1
}

// testDeleteSearch tests the endpoint that removes a query from the user's list of saved searches
func testDeleteSearch(t *testing.T) {
	is := is.New(t)
	srv := server.NewServer(srvConfig, ds)
	uid := "5b3bcd72463cd6029e04de18"
	query := "this search will be added and then deleted"

	// generate auth token
	u, err := ds.UserByID(uid)
	is.NoErr(err) // error fetching user record
	tk, err := u.Token(srvConfig.Token.Issuer, srvConfig.Token.SigningKey, 1)
	is.NoErr(err) // error generating token

	// add the query
	numSearches := len(u.Searches) // current number of searches
	body := strings.NewReader(fmt.Sprintf(`{"query": "%s"}`, query))
	r := httptest.NewRequest("POST", "/user/search", body)
	r.Header.Set("Authorization", "Bearer "+tk.String())
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusCreated)      // expected 201 Created
	u, err = ds.UserByID(uid)                 // fetch user record again
	is.NoErr(err)                             // error re-fetching user record
	is.True(len(u.Searches) == numSearches+1) // expect number of searches to increase by 1

	// delete the same query
	numSearches = len(u.Searches) // current number of searches
	body = strings.NewReader(fmt.Sprintf(`{"query": "%s"}`, query))
	r = httptest.NewRequest("DELETE", "/user/search", body)
	r.Header.Set("Authorization", "Bearer "+tk.String())
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)           // expected 200 Ok
	u, err = ds.UserByID(uid)                 // fetch user record again
	is.NoErr(err)                             // error re-fetching user record
	is.True(len(u.Searches) == numSearches-1) // expect number of searches to decrease by 1
}

// testRedirect tests the endpoint that redirects to an article url. Note that this end point actually fetches a
// pubmed article and returns a pubmed.Article value. So the id needs to be real or need to re-factor and mock the
// pubmed request.
func testRedirect(t *testing.T) {
	const realPubmedID = "30006323"
	is := is.New(t)
	srv := server.NewServer(srvConfig, ds)
	r := httptest.NewRequest("GET", "/r/"+realPubmedID, nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusFound) // expected 302 Found
}

// testSaveLog test the saving of a user reading log entry
func testSaveLog(t *testing.T) {
	is := is.New(t)
	srv := server.NewServer(srvConfig, ds)

	// generate a valid token for a user that is in the test database
	u, err := ds.UserByID("5b3bcd72463cd6029e04de18")
	is.NoErr(err) // error fetching user record
	tk, err := u.Token(srvConfig.Token.Issuer, srvConfig.Token.SigningKey, 1)
	is.NoErr(err) // error generating token

	newLog := `{
		"date": "2018-08-03T15:00:00Z", 
		"pmid": "30048945", 
		"minutes": 30,
		"title": "Circulating Klotho levels can predict long-term macrovascular outcomes in type 2 diabetic patients",
		"source": "Atherosclerosis 2018-09-01; 276: 83-90",
		"url": "https://doi.org/10.1016/j.atherosclerosis.2018.07.006",
		"comment": "An excellent article"
  	}`
	body := strings.NewReader(newLog)
	r := httptest.NewRequest("POST", "/user/log", body)
	r.Header.Set("Authorization", "Bearer "+tk.String())
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusCreated) // expected 201 Created
}

// testFetchUserLogs test the fetching of all user logs
func testFetchUserLogs(t *testing.T) {
	is := is.New(t)
	srv := server.NewServer(srvConfig, ds)

	// generate a valid token for a user that is in the test database
	u, err := ds.UserByID("5b3bcd72463cd6029e04de18")
	is.NoErr(err) // error fetching user record
	tk, err := u.Token(srvConfig.Token.Issuer, srvConfig.Token.SigningKey, 1)
	is.NoErr(err) // error generating token

	r := httptest.NewRequest("GET", "/user/logs", nil)
	r.Header.Set("Authorization", "Bearer "+tk.String())
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK) // expected 200 Created
}

// testDeleteLog tests the endpoint that removes a user log record
func testDeleteLog(t *testing.T) {
	is := is.New(t)
	srv := server.NewServer(srvConfig, ds)

	const ownerUserID = "5b3bcd72463cd6029e04de1a"
	const notOwnerUserID = "5b3bcd72463cd6029e04de18"

	// add a temp log entry
	l := ds.NewLog()
	l.UserID = bson.ObjectIdHex(ownerUserID)
	l.Date = time.Now().Format("2006-02-01")
	l.PMID = "12345678"
	l.Title = "This will be added, and then deleted"
	is.NoErr(l.Save()) // error adding temp log entry, prior to delete

	// test without a token - expect 401 unauthorized
	r := httptest.NewRequest("DELETE", "/user/log/"+l.ID.String(), nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusUnauthorized) // expected 401 Unauthorized for missing token

	// try with INCORRECT user token
	u, err := ds.UserByID(notOwnerUserID)
	is.NoErr(err) // error fetching user record
	tk, err := u.Token(srvConfig.Token.Issuer, srvConfig.Token.SigningKey, 1)
	is.NoErr(err) // error generating token
	r = httptest.NewRequest("DELETE", "/user/log/"+l.ID.Hex(), nil)
	r.Header.Set("Authorization", "Bearer "+tk.String())
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusUnauthorized) // expected 401 Unauthorized for incorrect user token

	// try with correct user token
	u, err = ds.UserByID(ownerUserID)
	is.NoErr(err) // error fetching user record
	tk, err = u.Token(srvConfig.Token.Issuer, srvConfig.Token.SigningKey, 1)
	is.NoErr(err) // error generating token
	r = httptest.NewRequest("DELETE", "/user/log/"+l.ID.Hex(), nil)
	r.Header.Set("Authorization", "Bearer "+tk.String())
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK) // expected 200 OK
}

func logError(e error) {
	if e != nil {
		log.Println("TEST ERROR:", e)
	}
}
