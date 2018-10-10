# datastore

The `datastore` package provides access to the various data sources for the application.

# User

This is the primary data entity and looks like this:

```json
{
	"_id" : "5b5ff8c7a9fb6e53bb474af1",
	"firstName" : "Mike",
	"lastName" : "Donnici",
	"email" : "michael@mesa.net.au",
	"password" : "b1db70b4fa8491af2595db9ff7b7ef53dabc836da551a99aa33eb7f50007586b",
	"locked" : false,
	"searches": [
	  {
	    "created": "2018-02-03 10:00:00",
	    "query": "quadricuspid aortic valve"
	  },
	  {
      	"created": "2018-02-07 12:00:00",
      	"query": "cardiomyopathy"
      }
	]
}
```

At this stage notifications will be sent weekly.

They will not be configurable, nor will any information be logged about the notifications.

# Articles

To leverage Algolia, and to keep costs down, only articles published in the last 12 months will be indexed, for each
category.

See indexing.md for details about pubmed queries



