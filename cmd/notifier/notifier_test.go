package notifier

import (
	"log"
	"testing"

	"github.com/matryer/is"
	"github.com/mikedonnici/rtcl-api/datastore"
	"github.com/mikedonnici/rtcl-api/datastore/mongo"
	"github.com/mikedonnici/rtcl-api/testdata"
)

var notificationTestDB = testdata.New()
var notificationTestDS = datastore.New()

func TestNotification(t *testing.T) {

	var err error

	err = notificationTestDB.SetupMongoDB()
	if err != nil {
		log.Fatalln(err)
	}
	defer cleanupUserTest()

	notificationTestDS.Mongo, err = mongo.NewConnection(testdata.MongoDSN, notificationTestDB.DBName, "test")
	if err != nil {
		log.Fatalln(err)
	}

	t.Run("user", func(t *testing.T) {
		t.Run("testUserByID", testNotificationsDue)
	})
}

func cleanupUserTest() {
	err := notificationTestDB.TearDownMongoDB()
	if err != nil {
		log.Println(err)
	}
}

// Tests fetching a list of users that have notifications due. Note this test is really covered by the user tests.
func testNotificationsDue(t *testing.T) {
	is := is.New(t)
	xu, err := notificationsDue(notificationTestDS)
	is.NoErr(err)
	is.Equal(len(xu), 2) // expected 2 users with notifications due
}


