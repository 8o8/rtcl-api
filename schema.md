# Schema


### User

```json
{
    "_id" : ObjectId("59d59b1992d13eeb0512345"),
    "createdAt" : ISODate("2017-06-13T14:43:11Z"),
    "updatedAt" : ISODate("2018-02-13T13:02:54Z"),
    "firstName": "Mike",
    "lastName": "Donnici",
    "email": "michael@mesa.net",
    "password": "12345abcdef"
}
```

### Article

```
{
 	"_id" : ObjectId("59d59b1992d13eeb051ccc74"),
 	"created" : ISODate("2017-06-13T14:43:11Z"),
 	"published" : ISODate("2016-08-02T00:00:00Z"),
 	"title" : "Comparison of Long-Term Mortality in Patients...",
 	"summary" : "Although current guidelines have highlighted the ...",
 	"keywords" : ["Nakamura Y", "Asaumi Y", "Miyagi T"],
    "url" : "https://doi.org/10.1097/MD.0000000000002896",
	"sourceId": "29747859",
    "sourceIssue": "2",
    "sourceName": "The American journal of cardiology",
    "sourceNameAbbrev": "Am J Cardiol",
    "sourcePages": "206-212",
    "sourcePubDate": "2018 Jul 15",
    "sourceVolume": "122",
}
```
