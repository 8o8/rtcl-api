package mongo_test

import (
	"log"
	"testing"

	"github.com/matryer/is"
	"github.com/mikedonnici/rtcl-api/datastore/mongo"
	"github.com/mikedonnici/rtcl-api/testdata"
)

var data = testdata.New()

func TestMain(m *testing.M) {

	err := data.SetupMongoDB()
	if err != nil {
		log.Fatalln(err)
	}
	defer data.TearDownMongoDB()

	m.Run()
}

func TestNewConnection(t *testing.T) {
	is := is.New(t)
	_, err := mongo.NewConnection(testdata.MongoDSN, "test", "test mongo db")
	is.NoErr(err)
}
