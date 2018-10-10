package datastore

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
)

type Log struct {
	ds      *Datastore
	ID      bson.ObjectId `json:"id" bson:"_id"`
	UserID  bson.ObjectId `json:"userId" bson:"user_id"`
	Date    string        `json:"date" bson:"date"`
	PMID    string        `json:"pmid" bson:"pmid"`
	Minutes int           `json:"minutes" bson:"minutes"`
	Title   string        `json:"title" bson:"title"`
	Source  string        `json:"source" bson:"source"`
	URL     string        `json:"url" bson:"url"`
	Comment string        `json:"comment"" bson:"comment"`
}

// Save saves a log record
func (l *Log) Save() error {
	if !l.ID.Valid() {
		l.ID = bson.NewObjectId()
	}
	_, err := l.ds.logsCollection().Upsert(bson.M{"_id": l.ID}, l)
	return err
}

// Delete deletes log from the datastore
func (l *Log) Delete() error {
	return l.ds.logsCollection().RemoveId(l.ID)
}

// checkFields ensures required field values
func (l *Log) checkFields() error {

	// must have a user id
	if bson.IsObjectIdHex(l.UserID.String()) {
		return errors.New("Log has missing or invalid user id")
	}

	if len(l.PMID) == 0 {
		return errors.New("PMID is missing")
	}

	return nil
}
