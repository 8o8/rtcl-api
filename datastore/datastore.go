// Package datastore provides a common structure for accessing data across multiple database connections
package datastore

import (
	"errors"
	"github.com/mikedonnici/rtcl-api/datastore/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const usersCollection = "users"
const logsCollection = "logs"

// Datastore contains connections to the data sources required for the service
type Datastore struct {
	Mongo *mongo.Connection
}

// New returns a pointer to a Datastore
func New() *Datastore {
	return &Datastore{}
}

// NewUser returns a pointer to a new User value with the datastore attached.
// Important to note that this User value is empty and needs to be populated before its methods
// will be of much use.
func (ds *Datastore) NewUser() *User {
	return &User{ds: ds}
}

// NewLog returns a pointer to a Log with the datastore attached.
func (ds *Datastore) NewLog() *Log {
	return &Log{ds: ds}
}

// UserByID queries user by id and returns a pointer to a User with fields populated from the database
func (ds *Datastore) UserByID(id string) (*User, error) {
	u := &User{}
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("object id is not valid")
	}
	q := bson.M{"_id": bson.ObjectIdHex(id)}
	err := ds.usersCollection().Find(q).One(&u)
	if err != nil {
		return nil, err
	}

	u.ds = ds // attach datastore!
	return u, nil
}

// UserUpdate updates a user doc
//func (ds *Datastore) UserUpdate()

// UserByEmail queries user by email and returns a pointer to a User with fields populated from the database
func (ds *Datastore) UserByEmail(email string) (*User, error) {
	u := &User{}
	q := bson.M{"email": email}
	err := ds.usersCollection().Find(q).One(&u)
	if err != nil {
		return nil, err
	}

	u.ds = ds // attach datastore!
	return u, nil
}

// UserByIDOrEmail queries user by id first, and then email
func (ds *Datastore) UserByIDOrEmail(idOrEmail string) (*User, error) {
	u := &User{}
	u, err := ds.UserByID(idOrEmail)
	if err == nil {
		return u, err
	}
	return ds.UserByEmail(idOrEmail)
}

// UserEmailExists returns true if a user record exists with the specified email
func (ds *Datastore) UserEmailExists(email string) bool {
	_, err := ds.UserByEmail(email)
	return err == nil
}

// UserAuth authenticates the user and return a populated User on success
func (ds *Datastore) UserAuth(email, password string) (*User, error) {
	u, err := ds.UserByEmail(email)
	if err != nil {
		return nil, err
	}
	q := bson.M{"email": u.Email, "password": u.hashPass(password)}
	err = ds.usersCollection().Find(q).One(&u)
	return u, err
}

// FreshUserToken returns a fresh JWT for the user identified by ID
func (ds *Datastore) FreshUserToken(userID, issuer, signingkey string, hoursttl int) (string, error) {

	u := ds.NewUser()
	err := u.ByID(userID)
	if err != nil {
		return "", errors.New("could not fetch user to re-issue token")
	}

	t, err := u.Token(issuer, signingkey, hoursttl)
	if err != nil {
		return "", errors.New("could not issue token")
	}

	return t.String(), nil
}

// LogByID returns a pointer to a Log value with fields populated from the database
func (ds *Datastore) LogByID(id string) (*Log, error) {
	l := &Log{}
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("object id is not valid")
	}
	q := bson.M{"_id": bson.ObjectIdHex(id)}
	err := ds.logsCollection().Find(q).One(&l)
	if err != nil {
		return nil, err
	}

	l.ds = ds // attach datastore!
	return l, nil
}

// LogsByUserID fetches all log records for the specified user id
func (ds *Datastore) LogsByUserID(userID string) ([]Log, error) {
	var xl []Log
	q := bson.M{"user_id": bson.ObjectIdHex(userID)}
	err := ds.logsCollection().Find(q).All(&xl)
	if err != nil {
		return nil, err
	}
	return xl, nil
}

// UsersDueNotification returns Users with notifications due - that is, with a notification field value in the past.
// Note that this needs to exclude dates that are zero, null or missing, hence the check for values greater than epoch.
func (ds *Datastore) UsersDueNotification() ([]User, error) {
	var xu []User
	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	q := bson.M{
		"notification": bson.M{
			"$gt": epoch,
			"$lt": time.Now(),
		},
	}
	err := ds.usersCollection().Find(q).All(&xu)
	if err != nil {
		return nil, err
	}
	return xu, err
}

// returns the users collection
func (ds *Datastore) usersCollection() *mgo.Collection {
	return ds.Mongo.Session.DB(ds.Mongo.DBName).C(usersCollection)
}

// returns the logs collection
func (ds *Datastore) logsCollection() *mgo.Collection {
	return ds.Mongo.Session.DB(ds.Mongo.DBName).C(logsCollection)
}
