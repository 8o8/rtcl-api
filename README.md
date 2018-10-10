# RTCL API

## Overview

RTCL (Article) is an app that indexes medical articles from Pubmed
in particular areas of interest.

Users are able to save keyword searches on the indexed articles and
be notified when new articles are available. Users can also log their
reading time and generate a report for their CPD system.

This repo contains the a simple API for the frontend application.

It is written in Go and uses a Mongo database.

## Configuration

The service requires the following env vars (example values):

```
API_URL="https://api.rtcl.io"
APP_URL="http://rtcl.io"
MONGODB_URI="mongodb://localhost"
MONGODB_NAME="rtcldev"
MONGODB_DESC="dev db"
SENDGRID_API_KEY="123123sfsdfsdf"
TOKEN_ISSUER="RTCL system"
TOKEN_SIGNINGKEY="ABigRandomString1234$%^&"
TOKEN_HOURS_TTL=48
```

These can be set in three ways, in order of precedence:

Firstly, by specifying a config file with the `-c` flag, eg:

```
# go run main -c "./env_example.txt"
```

Secondly, in the absence of a specified config file it will look for
the default `.env` file.

Finally, if the deployment environment allows for env vars to be set via
a control panel or similar (eg Heroku)then no configuration file needs
to be specified.

**Port Number**

If a `PORT` env var is present in the deployment environment then the
server will listen on that port. This is the case for Heroku and similar.

Otherwise, port number can be specified with an optional `-p` flag, or
left to the default of 8080.

## Testing

Most of the integration tests are run against real databases with a
small set of data.

To run all tests from root dir:

```
# go test ./...
```

Test files can be run individually as each `*_test.go` file sets up its
own test database and then runs a group  of test in parallel.

To run an individual set of tests:

```
# go test -v article_test.go
```

Ref: https://blog.golang.org/subtests


## End Points


## Journal Selection

Highest ranking journals, in each category, from here: https://www.scimagojr.com/

Note that the issn number(s) from the above site will look like this:

```
Journal of the American Academy of Dermatology
ISSN: 10976787, 01909622
```

When searching for the journal on the Medline list (below) the issn numbers have a dash in the middle. So the above 
numbers will be `1097-6787` and `0190-9622`.

Then, need journal codes from here: https://www.ncbi.nlm.nih.gov/books/NBK3827/table/pubmedhelp.T.journal_lists/

Or, use this link to load them into a browser: ftp://ftp.ncbi.nih.gov/pubmed/J_Medline.txt

Then search the page for the ISSN number (or Journal Title) and copy the `MedAbbr` to the Pubmed query.

```
JrId: 4671
JournalTitle: Journal of the American Academy of Dermatology
MedAbbr: J Am Acad Dermatol <-- #### This is the one we need ###
ISSN (Print): 0190-9622
ISSN (Online): 1097-6787
IsoAbbr: J. Am. Acad. Dermatol.
NlmId: 7907132
```

Additional journals are added with an `OR` clause and the Pubmed query looks like this:

```
loattrfree full text[Filter]
AND (
"J Am Acad Dermatol"[jour] OR
"J Invest Dermatol"[jour] OR
"JAMA Dermatol"[jour]
)
```

## User

The user record is simple, and looks like this:

```
{
	"_id" : ObjectId("5b5ff8c7a9fb6e53bb474af1"),
	"firstName" : "Barry",
	"lastName" : "Smith",
	"email" : "barry@smith.net",
	"password" : "b1db70b4fa849105af...",
	"locked" : false,
	"notification": ISODate("2018-11-03T00:00:00Z"),
	"categories" : ["cardilogy", "physiotherapy"],
	"searches" : [
		{
			"created" : ISODate("2018-09-07T01:21:15.758Z"),
			"query" : "atherosclerosis"
		},
		{
			"created" : ISODate("2018-09-23T00:13:21.413Z"),
			"query" : "calcium score"
		},
	]
}
```

## Notifications

A cron job runs and checks the `notification` field in each user doc. If it is in the past then a notification is due 
for that user.

The notification job is run and the user's `notification` field value is pushed forward by a specified amount of time.

The notification schedule should be configurable by the user, however, at this stage it is hard-coded to 7 days.

  





 


