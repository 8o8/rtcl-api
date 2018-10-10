package datastore

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// replaces password values in responses
const PasswordMask = "********"

// User holds data for a user and an associated datastore. The datastore is specified in the User value
// so that methods can be hung off the User value. For example User.Save() is more convenient than
// datastore.UserSave(user). Note that Password is unexported so we never return the hashed Password in a JSON response.
// However, it still needs toe json/bson tags because it will be posted in a request body and stored in the database.
type User struct {
	ds           *Datastore
	ID           bson.ObjectId `json:"id" bson:"_id"`
	FirstName    string        `json:"firstName" bson:"firstName"`
	LastName     string        `json:"lastName" bson:"lastName"`
	Email        string        `json:"email" bson:"email"`
	Password     string        `json:"password" bson:"password"`
	Locked       bool          `json:"locked" bson:"locked"`
	Categories   []string      `json:"categories" bson:"categories"`
	Searches     []Search      `json:"searches" bson:"searches"`
	Notification time.Time     `json:"notification" bson:"notification"`
}

// Search represents stored User search
type Search struct {
	Created time.Time `json:"created" bson:"created"`
	Query   string    `json:"query" bson:"query"`
}

// Save adds / updates a user record
func (u *User) Save() error {

	err := u.checkFields()
	if err != nil {
		return err
	}

	checkUsr := User{
		ds: u.ds, // attach the datastore
	}

	// This user should either not exist, or be the same user for which data is being saved
	err = checkUsr.ByEmail(u.Email)
	if err == nil { // found a user record..
		if checkUsr.ID != u.ID { // email exists for a different user
			return errors.New("Email already exists for a different user id")
		}
	}

	// New user
	if !u.ID.Valid() {

		if u.ds.UserEmailExists(u.Email) {
			return errors.New("User with that email already exists")
		}

		u.ID = bson.NewObjectId()
		u.Locked = true
		if u.Password == "" {
			u.Password = impossiblePassword()
		}
		u.Password = u.hashPass(u.Password)
	}

	q := bson.M{"_id": u.ID}
	_, err = u.ds.usersCollection().Upsert(q, u)
	return err
}

// SavePartial updates user fields in update arg
func (u *User) SavePartial(update bson.M) error {

	firstName, ok := update["firstName"]
	if ok {
		u.FirstName = firstName.(string)
	}

	lastName, ok := update["lastName"]
	if ok {
		u.LastName = lastName.(string)
	}

	email, ok := update["email"]
	if ok {
		u.Email = email.(string)
	}

	// password field should not be an empty string
	password, ok := update["password"]
	if ok && len(password.(string)) > 0 {
		u.Password = u.hashPass(password.(string))
	}

	categories, ok := update["categories"]
	if ok {
		u.Categories = []string{}
		cats := categories.([]interface{})
		for _, v := range cats {
			u.Categories = append(u.Categories, v.(string))
		}
	}

	notification, ok := update["notification"]
	if ok {
		t, err := time.Parse("2006-01-02", notification.(string))
		if err != nil {
			return err
		}
		u.Notification = t
	}

	return u.Save()
}

// ByID validates the id string, fetches a user record by id (_id), and populates the User fields.
// Note that the User.ds value is lost when the field values are scanned to u, so it needs to be re-attached
// before the function ends.
func (u *User) ByID(id string) error {
	if !bson.IsObjectIdHex(id) {
		return errors.New("object id is not valid")
	}

	ds := u.ds // detach the datastore
	q := bson.M{"_id": bson.ObjectIdHex(id)}
	err := u.ds.usersCollection().Find(q).One(&u) // datastore is nil now!
	u.ds = ds                                     // re-attach the datastore
	return err
}

// ByEmail looks up a user record by email and populates User fields.
func (u *User) ByEmail(email string) error {
	ds := u.ds
	q := bson.M{"email": email}
	err := u.ds.usersCollection().Find(q).One(&u)
	u.ds = ds // re-attach datastore
	return err
}

// KeyGen generates an access key for a user by hashing a few string values from the user record along
// with a couple of UTC date strings. This allows the same hash hash to be validated until midnight on the same day.
func (u *User) KeyGen() string {
	s := strings.ToLower(u.Email) + u.Password + time.Now().UTC().Month().String() + strconv.Itoa(time.Now().UTC().Day())
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Token returns a valid JWT for the user
func (u *User) Token(issuer, signingKey string, ttl int) (Token, error) {
	c := map[string]interface{}{
		"id":   u.ID.Hex(),
		"name": u.FirstName + " " + u.LastName,
		"role": "user",
	}
	return NewToken(issuer, signingKey, ttl).CustomClaims(c).Encode()
}

// SaveSearch saves a search (one or more search terms) for a user. The mongo $addToSet will not save a
// duplicate value however as the search in an object it will have a new time stamp so need to check for
// duplicate query strings.
func (u *User) SaveSearch(query string) error {

	if u.SearchExists(query) {
		return errors.New("exact search query already exists for this user")
	}

	q := bson.M{"_id": u.ID}
	s := Search{
		Created: time.Now(),
		Query:   query,
	}
	update := bson.M{"$addToSet": bson.M{"searches": s}}
	err := u.ds.usersCollection().Update(q, update)
	if err != nil {
		return err
	}
	u.Searches = append(u.Searches, s)

	return nil
}

// DeleteSearch deletes the query from the user's search list
func (u *User) DeleteSearch(query string) error {

	if !u.SearchExists(query) {
		return errors.New("cannot find the query so unable to delete it")
	}

	q := bson.M{"_id": u.ID}
	// $lt used to match any searches sub doc older than 'now' - ie all, that match the query string
	update := bson.M{"$pull": bson.M{"searches": bson.M{"created": bson.M{"$lt": time.Now()}, "query": query}}}
	err := u.ds.usersCollection().Update(q, update)
	if err != nil {
		return err
	}

	var updatedList []Search
	for _, s := range u.Searches {
		if s.Query != query {
			updatedList = append(updatedList, s)
		}
	}
	u.Searches = updatedList

	return nil
}

// SearchExists returns true if the query string already exists in the user's list of searches
func (u *User) SearchExists(query string) bool {
	for _, s := range u.Searches {
		if matchString(s.Query, query) {
			return true
		}
	}
	return false
}

// SavedSearches retrieves the set of Searches for a user
func (u *User) SavedSearches() ([]Search, error) {
	return u.Searches, nil
}

// IncrementNotification increments the notification date by the specified number of days.
func (u *User) IncrementNotification(days int) error {
	u.Notification = u.Notification.AddDate(0,0, days)
	return u.Save()
}

// hashPass returns a hashed Password from the users email and the clear text Password
func (u *User) hashPass(password string) string {
	salt := hash(u.ID.Hex() + os.Getenv("PASSWORD_SALT")) // empty if not present
	return hash(password + salt)
}

// checkFields checks the fields for a user are valid
func (u *User) checkFields() error {
	if len(u.FirstName) == 0 {
		return errors.New("First name is missing")
	}
	if len(u.LastName) == 0 {
		return errors.New("Last name is missing")
	}
	if len(u.Email) == 0 {
		return errors.New("Email is missing")
	}
	return nil
}

func impossiblePassword() string {
	return hash(time.Now().String())
}

func hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// used to match strings use for search queries in order to prevent duplicates
func matchString(s1, s2 string) bool {
	xs1 := strings.Fields(strings.ToLower(s1))
	xs2 := strings.Fields(strings.ToLower(s2))
	return reflect.DeepEqual(xs1, xs2)
}
