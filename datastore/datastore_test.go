package datastore_test

//import (
//	"log"
//	"testing"
//
//	"github.com/mikedonnici/dev/go/code-organisation/httpservice/testdata"
//)
//
//var store = testdata.New()

// TestMain sets up test databases against which subsequent datastore tests can be run. These are torn
// down once all of the tests have completed.
//func TestMain(m *testing.M) {
//
//	err := store.SetupMySQL()
//	if err != nil {
//		log.Fatalln(err)
//	}
//	defer store.TearDownMySQL()
//
//	err = store.SetupMongoDB()
//	if err != nil {
//		log.Fatalln(err)
//	}
//	defer store.TearDownMongoDB()
//
//	m.Run()
//}

//// Test a direct query
//func TestFetchMySQL(t *testing.T) {
//	is := is.New(t)
//
//	cases := []struct {
//		id        string
//		firstName string
//	}{
//		{id: "1", firstName: "Broderick"},
//		{id: "5", firstName: "Declan"},
//	}
//
//	for _, c := range cases {
//		q := `select firstname from people where id = ?`
//		var name string
//		store.MySQLSession.QueryRow(q, c.id).Scan(&name)
//		is.Equal(name, c.firstName)
//	}
//}
//
//// Test a direct query
//func TestFetchMongoDB(t *testing.T) {
//	is := is.New(t)
//
//	cases := []struct {
//		id        int
//		firstName string
//	}{
//		{id: 1, firstName: "Broderick"},
//		{id: 5, firstName: "Declan"},
//	}
//
//	type person struct {
//		FirstName string `bson:"firstname"`
//	}
//
//	for _, c := range cases {
//		var p person
//		q := bson.M{"id": c.id}
//		s := bson.M{"_id": 0, "firstname": 1}
//		col := store.MongoDBSession.DB(store.DBName).C(testdata.MONGO_COLLECTION)
//		err := col.Find(q).Select(s).One(&p)
//		is.NoErr(err)                      // error fetching from Mongo
//		is.Equal(p.FirstName, c.firstName) // incorrect name
//	}
//}
