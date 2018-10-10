package notifier

import (
	"github.com/mikedonnici/rtcl-api/datastore"
)

// notificationsDue fetches returns a set of User that have the notification field value in the past
func notificationsDue(ds *datastore.Datastore) ([]datastore.User, error) {
	return ds.UsersDueNotification()
}
