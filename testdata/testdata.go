package testdata

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/mikedonnici/rtcl-api/datastore"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"time"
)

// Hard coded for local dev and Travis CI
const MongoDSN = "mongodb://localhost"

type TestStore struct {
	DBName         string
	MongoDBSession *mgo.Session
}

// New returns a pointer to a TestStore
func New() *TestStore {

	s, _ := uuid.GenerateUUID()
	n := fmt.Sprintf("%v_test", s[0:7])

	t := TestStore{
		DBName: n,
	}
	return &t
}

// SetupMongoDB connects to the test database and populates a collection
func (t *TestStore) SetupMongoDB() error {

	var err error

	t.MongoDBSession, err = mgo.Dial(MongoDSN)
	if err != nil {
		return errors.Wrap(err, "Error establishing session with Mongo")
	}

	err = t.MongoDBSession.Ping()
	if err != nil {
		return errors.Wrap(err, "Error pinging Mongo")
	}

	err = t.usersData()
	if err != nil {
		return err
	}

	err = t.logData()
	if err != nil {
		return err
	}

	return nil
}

// usersData adds data to the users collection
func (t *TestStore) usersData() error {

	var xu []struct {
		ID           bson.ObjectId      `json:"_id" bson:"_id"`
		FirstName    string             `json:"firstName" bson:"firstName"`
		LastName     string             `json:"lastName" bson:"lastName"`
		Email        string             `json:"email" bson:"email"`
		Password     string             `json:"password" bson:"password"`
		Locked       bool               `json:"locked" bson:"locked"`
		Categories   []string           `json:"categories" bson:"categories"`
		Searches     []datastore.Search `json:"searches" bson:"searches"`
		Notification time.Time          `json:"notification" bson:"notification"`
	}
	err := json.Unmarshal([]byte(MONGO_USERS_DATA), &xu)
	if err != nil {
		return errors.Wrap(err, "Unmarshal error")
	}

	for _, u := range xu {
		salt := hash(u.ID.Hex() + os.Getenv("PASSWORD_SALT")) // empty if not present
		u.Password = hash(u.Password + salt)
		err = t.MongoDBSession.DB(t.DBName).C(MONGO_USERS_COLLECTION).Insert(u)
		if err != nil {
			return errors.Wrap(err, "Error inserting user into mongo")
		}
	}
	return nil
}

// logData adds data to the log collection
func (t *TestStore) logData() error {

	var xl []struct {
		ID      bson.ObjectId `json:"_id" bson:"_id"`
		UserID  bson.ObjectId `json:"user_id" bson:"user_id"`
		Date    string        `json:"date" bson:"date"`
		PMID    string        `json:"pmid" bson:"pmid"`
		Minutes int           `json:"minutes" bson:"minutes"`
		Title   string        `json:"title" bson:"title"`
		Source  string        `json:"source" bson:"source"`
		URL     string        `json:"url" bson:"url"`
		Comment string        `json:"comment"" bson:"comment"`
	}
	err := json.Unmarshal([]byte(MONGO_LOGS_DATA), &xl)
	if err != nil {
		return errors.Wrap(err, "Unmarshal error")
	}

	for _, l := range xl {
		err = t.MongoDBSession.DB(t.DBName).C(MONGO_LOGS_COLLECTION).Insert(l)
		if err != nil {
			return errors.Wrap(err, "Error inserting log into mongo")
		}
	}
	return nil
}

func (t *TestStore) TearDownMongoDB() error {
	err := t.MongoDBSession.DB(t.DBName).DropDatabase()
	if err != nil {
		return errors.Wrap(err, "Error deleting Mongo test database")
	}
	return nil
}

// Note this is a copy of the hash function from the datastore package
func hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
