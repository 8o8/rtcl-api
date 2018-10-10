package datastore_test

import (
	"gopkg.in/mgo.v2/bson"
	"log"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/mikedonnici/rtcl-api/datastore"
	"github.com/mikedonnici/rtcl-api/datastore/mongo"
	"github.com/mikedonnici/rtcl-api/testdata"
	"gopkg.in/mgo.v2"
)

const testIssuer = "TestTokenIssuer"
const testSigningKey = "testTokenSigningKey"
const testTTLHours = 4

var userTestDB = testdata.New()
var userTestDS = datastore.New()

func TestUser(t *testing.T) {

	var err error

	err = userTestDB.SetupMongoDB()
	if err != nil {
		log.Fatalln(err)
	}
	defer cleanupUserTest()

	userTestDS.Mongo, err = mongo.NewConnection(testdata.MongoDSN, userTestDB.DBName, "test")
	if err != nil {
		log.Fatalln(err)
	}

	t.Run("user", func(t *testing.T) {
		t.Run("testUserByID", testUserByID)
		t.Run("testUserByIDInvalidHex", testUserByIDInvalidHex)
		t.Run("testUserByEmail", testUserByEmail)
		t.Run("testUserNotFoundByID", testUserNotFoundByID)
		t.Run("testUserNotFoundByEmail", testUserNotFoundByEmail)
		t.Run("testUserExists", testUserExists)
		t.Run("testUserAdd", testUserAdd)
		t.Run("testUserAddMissingFields", testUserAddMissingFields)
		t.Run("testUserAddEmailExists", testUserAddEmailExists)
		t.Run("testUserSave", testUserSave)
		t.Run("testUserSavePartial", testUserSavePartial)
		t.Run("testUserSaveEmailExists", testUserSaveEmailExists)
		t.Run("testUserKeyGen", testUserKeyGen)
		t.Run("testUserAuth", testUserAuth)
		t.Run("testUserToken", testUserToken)
		t.Run("testUserSaveSearch", testUserSaveSearch)
		t.Run("testUserDeleteSearch", testUserDeleteSearch)
		t.Run("testUserSavedSearches", testUserSavedSearches)
		t.Run("testUsersDueNotification", testUsersDueNotification)
		t.Run("testUserIncrementNotification", testUserIncrementNotification)
	})
}

func cleanupUserTest() {
	err := userTestDB.TearDownMongoDB()
	if err != nil {
		log.Println(err)
	}
}

func testUserByID(t *testing.T) {
	is := is.New(t)

	cases := []struct {
		id        string
		firstName string
	}{
		{id: "5b3bcd72463cd6029e04de18", firstName: "Broderick"},
		{id: "5b3bcd72463cd6029e04de1c", firstName: "Dawn"},
	}

	for _, c := range cases {
		u, err := userTestDS.UserByID(c.id)
		is.NoErr(err)                      // error fetching user by object id
		is.Equal(u.FirstName, c.firstName) // incorrect user fetched
	}
}

func testUserByIDInvalidHex(t *testing.T) {
	is := is.New(t)
	_, err := userTestDS.UserByID("InvalidHexValue")
	is.True(err != nil) // expect an error for an invalid hex
}

func testUserByEmail(t *testing.T) {
	is := is.New(t)

	cases := []struct {
		email     string
		firstName string
	}{
		{email: "br@rtcl.io", firstName: "Broderick"},
	}

	for _, c := range cases {
		u, err := userTestDS.UserByEmail(c.email)
		is.NoErr(err)                      // error fetching user by email
		is.Equal(u.FirstName, c.firstName) // incorrect user fetched
	}
}

func testUserNotFoundByID(t *testing.T) {
	is := is.New(t)

	cases := []struct {
		id string
	}{
		{id: "5b3bcd72463cd6029e04abcd"},
	}

	for _, c := range cases {
		_, err := userTestDS.UserByID(c.id)
		is.True(err != nil)            // expected an error
		is.Equal(err, mgo.ErrNotFound) // expected 'not found' error
	}
}

func testUserNotFoundByEmail(t *testing.T) {
	is := is.New(t)

	cases := []struct {
		email string
	}{
		{email: "notexists@rtcl.io"},
	}

	for _, c := range cases {
		_, err := userTestDS.UserByEmail(c.email)
		is.True(err != nil)            // expected an error
		is.Equal(err, mgo.ErrNotFound) // expected 'not found' error
	}
}

func testUserExists(t *testing.T) {
	is := is.New(t)

	cases := []struct {
		email  string
		exists bool
	}{
		{email: "br@rtcl.io", exists: true},
		{email: "notexists@rtcl.io", exists: false},
	}

	for _, c := range cases {
		is.Equal(userTestDS.UserEmailExists(c.email), c.exists)
	}
}

func testUserAdd(t *testing.T) {
	is := is.New(t)
	u := userTestDS.NewUser()
	u.FirstName = "Barry"
	u.LastName = "Smith"
	u.Email = "bs@rtcl.io"
	u.Categories = []string{"cardiology", "physiotherapy"}
	u.Searches = []datastore.Search{
		{time.Now(), "search one"},
		{time.Now(), "search two"},
	}
	u.Notification = time.Now().AddDate(0, 0, 7)
	err := u.Save()
	is.NoErr(err)
}

func testUserAddMissingFields(t *testing.T) {
	is := is.New(t)

	cases := []struct {
		firstName string
		lastName  string
		email     string
	}{
		{firstName: "", lastName: "MissingFirstName", email: "anything@anywhere.com"},
		{firstName: "MissingLastName", lastName: "", email: "anything@anywhere.com"},
		{firstName: "Missing", lastName: "Email", email: ""},
	}

	for _, c := range cases {
		u := userTestDS.NewUser()
		u.FirstName = c.firstName
		u.LastName = c.lastName
		u.Email = c.email
		err := u.Save()
		is.True(err != nil) // expected an error when a field is missing
	}
}

func testUserAddEmailExists(t *testing.T) {
	is := is.New(t)
	u := userTestDS.NewUser()
	u.FirstName = "Barry"
	u.LastName = "Remington"
	u.Email = "br@rtcl.io"
	err := u.Save()
	is.True(err != nil) // expected an error as email already exists
}

func testUserSave(t *testing.T) {
	is := is.New(t)

	u := userTestDS.NewUser()
	u.FirstName = "Barry"
	u.LastName = "Save"
	u.Email = "bsave@rtcl.io"

	err := u.Save()
	is.NoErr(err) // error adding user

	u.LastName = "WasSaved"
	err = u.Save()
	is.NoErr(err) // error saving user

	// reset and fetch the user to ensure record was updated
	u2, err := userTestDS.UserByID(u.ID.Hex())
	is.NoErr(err)                     // error saving user
	is.Equal(u2.LastName, u.LastName) // last name should have been saved
}

func testUserSavePartial(t *testing.T) {
	is := is.New(t)

	// add a new user
	u := userTestDS.NewUser()
	u.FirstName = "Suzie"
	u.LastName = "Save"
	u.Email = "ssave@rtcl.io"
	err := u.Save()
	is.NoErr(err) // error adding user

	// fields to be updated. Note that the list of categories gets decoded as a []interface{}
	update := bson.M{
		"lastName": "SavedPartially",
		"categories": []interface{}{"one", "two"},
		"notification": "2018-11-02",
	}
	err = u.SavePartial(update)
	is.NoErr(err) // error saving user

	// re-fetch the user to ensure record was updated
	u2, err := userTestDS.UserByID(u.ID.Hex())
	is.NoErr(err)                           // error fetching user
	is.Equal(u2.LastName, "SavedPartially") // last name should have been changed
	is.Equal(len(u2.Categories), 2)         // expected 2 categories after partial save
	newDate, _ := time.Parse("2006-01-02", "2018-11-02")
	is.Equal(u2.Notification.UTC(), newDate.UTC())
}

// Test that a user cannot be saved if the email already exists in the database
func testUserSaveEmailExists(t *testing.T) {
	is := is.New(t)

	// add a new user
	u := userTestDS.NewUser()
	u.FirstName = "Mike"
	u.LastName = "Smith"
	u.Email = "msmith@rtcl.io"
	err := u.Save()
	is.NoErr(err) // error adding user

	// change email to one that already exists from the testdata
	u.FirstName = "Michael"
	u.LastName = "Smithson"
	u.Email = "br@rtcl.io" // should clash
	err = u.Save()
	is.True(err != nil) // expect error saving a duplicate email
}

func testUserKeyGen(t *testing.T) {
	is := is.New(t)
	u, err := userTestDS.UserByEmail("dh@rtcl.io")
	is.NoErr(err) // error fetching user
	k := u.KeyGen()
	is.Equal(len(k), 64) // key should be 64 chars long
}

// testUserAuth tests the basic user auth function which finds the user by email, and then
// checks for a Password match. Note that there is no hashing of passwords here as we're just checking
// the values that were inserted into the test data.
func testUserAuth(t *testing.T) {
	is := is.New(t)

	// Add a new user first, so Password gets hashed properly
	u := userTestDS.NewUser()
	u.FirstName = "Mike"
	u.LastName = "Donnici"
	u.Email = "mike@rtcl.io"
	u.Password = "ILike2MoveItMoveIt"
	err := u.Save()
	is.NoErr(err) // err adding new user for auth test

	cases := []struct {
		email      string
		password   string
		expectAuth bool
	}{
		{email: "", password: "", expectAuth: false},
		{email: "someDodgyUsername", password: "", expectAuth: false},
		{email: "mike@rtcl.io", password: "", expectAuth: false},
		{email: "mike@rtcl.io", password: "ilike2moveitmoveit", expectAuth: false},
		{email: "mike@rtcl.io", password: "ILike2MoveItMoveIt", expectAuth: true},
	}

	for _, c := range cases {
		_, err := userTestDS.UserAuth(c.email, c.password)
		is.True((err == nil) == c.expectAuth) // auth bool not as expected
	}
}

// testUserToken create a Token value for a user in the test data
func testUserToken(t *testing.T) {
	is := is.New(t)
	u, err := userTestDS.UserByID("5b3bcd72463cd6029e04de18")
	is.NoErr(err) // error fetching user
	tk, err := u.Token(testIssuer, testSigningKey, testTTLHours)
	is.NoErr(err)                 // error generating token
	is.True(len(tk.String()) > 0) // token has no length
	is.True(tk.Claims.ID == "5b3bcd72463cd6029e04de18")
	is.True(tk.Claims.Name == "Broderick Reynolds")
	is.True(tk.Claims.Role == "user")
}

// tests saving a search for a user
func testUserSaveSearch(t *testing.T) {
	is := is.New(t)
	search := "quadricuspid aortic valve"
	u, err := userTestDS.UserByID("5b3bcd72463cd6029e04de18")
	is.NoErr(err) // error fetching user
	err = u.SaveSearch(search)
	is.NoErr(err) // error saving search
}

// tests deleting a search for a user
func testUserDeleteSearch(t *testing.T) {
	is := is.New(t)
	u, err := userTestDS.UserByID("5b3bcd72463cd6029e04de18")
	is.NoErr(err) // error fetching user
	search := "this is a new search"
	err = u.SaveSearch(search)
	is.NoErr(err)                   // error saving search
	is.True(u.SearchExists(search)) // saved search should exist
	err = u.DeleteSearch(search)
	is.True(!u.SearchExists(search)) // saved search should have been removed
}

// tests fetching the searches for a user
func testUserSavedSearches(t *testing.T) {
	is := is.New(t)
	u, err := userTestDS.UserByID("5b3bcd72463cd6029e04de1c")
	is.NoErr(err) // error fetching user
	searches := []string{
		"cardiomyopathy",
		"athletes",
	}
	for _, s := range searches {
		err := u.SaveSearch(s)
		is.NoErr(err) // error saving search
	}
	xs, err := u.SavedSearches()
	is.NoErr(err)        // error fetching saved searches
	is.Equal(len(xs), 2) // expected 3 saved searches

	// Add one more and re-save all... should not get any duplicates
	searches = append(searches, "cath lab")
	for i, s := range searches {
		err := u.SaveSearch(s) // ignoring the error
		if i < 2 {
			is.True(err != nil) // expect errors for the first 3 saves as they already exist
		} else {
			is.NoErr(err) // should not get an error for the new search
		}
	}
	err = u.ByID("5b3bcd72463cd6029e04de1c")
	is.NoErr(err) // error ensuring user record
	xs, err = u.SavedSearches()
	is.NoErr(err)        // error fetching saved searches
	is.Equal(len(xs), 3) // expected 3 saved searches
}

// Tests fetching a list of users that have notifications due
func testUsersDueNotification(t *testing.T) {
	is := is.New(t)
	xu, err := userTestDS.UsersDueNotification()
	is.NoErr(err)
	is.Equal(len(xu), 2) // expected 2 users with notifications due
}

// Tests incrementing the notification date by x days
func testUserIncrementNotification(t *testing.T) {
	is := is.New(t)
	u, err := userTestDS.UserByID("5b3bcd72463cd6029e04de18")
	is.NoErr(err) // error fetching user
	originalDate := u.Notification // assign *before* update
	err = u.IncrementNotification(7)
	is.NoErr(err) // error incrementing notification date

	// refetch user and check date
	u2, err := userTestDS.UserByID("5b3bcd72463cd6029e04de18")
	is.NoErr(err) // error re-fetching user
	is.True(u2.Notification == originalDate.AddDate(0, 0, 7)) // new date incorrect
}


