package datastore_test

import (
	"github.com/matryer/is"
	"gopkg.in/mgo.v2/bson"
	"log"
	"testing"
	"time"

	"github.com/mikedonnici/rtcl-api/datastore"
	"github.com/mikedonnici/rtcl-api/datastore/mongo"
	"github.com/mikedonnici/rtcl-api/testdata"
)

var logTestDB = testdata.New()
var logTestDS = datastore.New()

func TestLog(t *testing.T) {

	var err error

	err = logTestDB.SetupMongoDB()
	if err != nil {
		log.Fatalln(err)
	}
	defer cleanupLogTest()

	logTestDS.Mongo, err = mongo.NewConnection(testdata.MongoDSN, logTestDB.DBName, "test")
	if err != nil {
		log.Fatalln(err)
	}

	t.Run("log", func(t *testing.T) {
		t.Run("testPingDB", testPingDB)
		t.Run("testAddLog", testAddLog)
		t.Run("testUpdateLog", testUpdateLog)
		t.Run("testDeleteLog", testDeleteLog)
		t.Run("testLogByID", testLogByID)
		t.Run("testLogByIDNotFound", testLogByIDNotFound)
		t.Run("testLogsByUserID", testLogsByUserID)
	})
}

func cleanupLogTest() {
	err := logTestDB.TearDownMongoDB()
	if err != nil {
		log.Println(err)
	}
}

func testPingDB(t *testing.T) {
	is := is.New(t)
	err := logTestDS.Mongo.Session.Ping()
	is.NoErr(err) // cannot ping database
}

func testAddLog(t *testing.T) {
	is := is.New(t)
	l := logTestDS.NewLog()
	l.UserID = bson.ObjectIdHex("5b3bcd72463cd6029e04de1a") // valid, from test data
	l.Date = time.Now().Format("2006-02-01")
	l.PMID = "Atherosclerosis 2018-08-27; 277: 53-59"
	l.Minutes = 15
	l.Title = "Direct observation of cargo transfer from HDL particles to the plasma membrane"
	l.Source = "J Cardiovasc Magn Reson 2018-09-03; 20(1): 60"
	l.URL = "https://doi.org/10.1186/s12968-018-0482-7"
	l.Comment = "This was an excellent article."
	err := l.Save()
	is.NoErr(err) // error adding log

	// fetch it and check
	n, err := logTestDS.LogByID(l.ID.Hex())
	is.NoErr(err)                  // error fetching the newly created log
	is.Equal(n.ID, l.ID)           // log id mismatch
	is.Equal(n.Comment, l.Comment) // log comment mismatch
}

func testUpdateLog(t *testing.T) {
	is := is.New(t)
	l, err := logTestDS.LogByID("5b3bcd72463cd6029e04de28")
	is.NoErr(err) // error fetching log
	l.Comment = "This has been updated"
	err = l.Save()
	is.NoErr(err) // error saving log
	n, err := logTestDS.LogByID("5b3bcd72463cd6029e04de28")
	is.NoErr(err)                  // error fetching log
	is.Equal(n.Comment, l.Comment) // comment not updated
}

// testDeleteLog tests the removal of a log record. The record is added before being removed
// so it does not affect other tests.
func testDeleteLog(t *testing.T) {
	is := is.New(t)
	l := logTestDS.NewLog()
	l.UserID = bson.ObjectIdHex("5b3bcd72463cd6029e04de1a") // valid, from test data
	l.Date = time.Now().Format("2006-02-01")
	l.PMID = "12345678"
	l.Title = "This will be added, and then deleted"
	is.NoErr(l.Save())   // error adding log, prior to delete
	is.NoErr(l.Delete()) // error deleting log
}

func testLogByID(t *testing.T) {
	is := is.New(t)
	l, err := logTestDS.LogByID("5b3bcd72463cd6029e04de28")
	is.NoErr(err)
	is.Equal(l.PMID, "30173671") // unexpected pmid value
	is.Equal(l.Minutes, 90)      // unexpected minutes value
}

func testLogByIDNotFound(t *testing.T) {
	is := is.New(t)
	_, err := logTestDS.LogByID("5b3bcd72463cd6029e04de88")
	is.True(err != nil) // expect an error
}

func testLogsByUserID(t *testing.T) {
	is := is.New(t)
	xl, err := logTestDS.LogsByUserID("5b3bcd72463cd6029e04de18") // has 4 logs in test data
	is.NoErr(err)
	is.Equal(len(xl), 4)                                         // expected 4 results
	xl, err = logTestDS.LogsByUserID("5b3bcd72463cd6029e04de88") // has 0 logs in test data
	is.NoErr(err)
	is.Equal(len(xl), 0) // expected 0 results
}
