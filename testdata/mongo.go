package testdata

const MONGO_USERS_COLLECTION = "users"
const MONGO_USERS_DATA = `[
  {
	"_id": "5b3bcd72463cd6029e04de18", 
	"firstName" : "Broderick", 
	"lastName" : "Reynolds", 
	"email" : "br@rtcl.io", 
	"password": "12345abcde", 
	"locked": false,
	"categories": ["cardiology", "physiotherapy"],
	"notification": "2018-01-01T00:00:00Z"
  },
  {
	"_id": "5b3bcd72463cd6029e04de1a", 
	"firstName": "Osborne",   
	"lastName": "Jast",     
	"email": "oj@rtcl.io", 
	"password": "12345abcde", 
	"locked": true,
	"notification": "2018-01-01T00:00:00Z"
  },
  {
	"_id": "5b3bcd72463cd6029e04de1c", 
	"firstName": "Dawn", 
	"lastName": "Hayes",
	"email": "dh@rtcl.io", 
	"password": "12345abcde", 
	"locked": true,
	"notification": "2118-01-01T00:00:00Z"
  }
]`

const MONGO_LOGS_COLLECTION = "logs"
const MONGO_LOGS_DATA = `[
  {
	"_id": "5b3bcd72463cd6029e04de28", 
	"user_id": "5b3bcd72463cd6029e04de18", 
	"date": "2018-10-02", 
	"pmid": "30173671", 
	"minutes": 90,
	"title": "Assessment of longitudinal distribution of subclinical atherosclerosis...",
	"source": "J Cardiovasc Magn Reson 2018-09-03; 20(1): 60",
	"url": "https://doi.org/10.1186/s12968-018-0482-7",
	"comment": "lorem ipsum..."
  },
  {
	"_id": "5b3bcd72463cd6029e04de30", 
	"user_id": "5b3bcd72463cd6029e04de18", 
	"date": "2018-11-02", 
	"pmid": "30173079", 
	"minutes": 60,
	"title": "Direct observation of cargo transfer from HDL particles to the plasma membrane",
	"source": "Atherosclerosis 2018-08-27; 277: 53-59",
	"url": "https://doi.org/10.1016/j.atherosclerosis.2018.08.032",
	"comment": "lorem ipsum..."
  },
  {
	"_id": "5b3bcd72463cd6029e04de32", 
	"user_id": "5b3bcd72463cd6029e04de18", 
	"date": "2018-12-02", 
	"pmid": "30171974", 
	"minutes": 30,
	"title": "Surviving Refractory Out-of-Hospital Ventricular Fibrillation Cardiac Arrest: Critical Care and Extracorporeal Membrane Oxygenation Management",
	"source": "Resuscitation 2018-08-29",
	"url": "https://doi.org/10.1016/j.resuscitation.2018.08.030",
	"comment": "lorem ipsum..."
  },
  {
	"_id": "5b3bcd72463cd6029e04de34", 
	"user_id": "5b3bcd72463cd6029e04de18", 
	"date": "2018-12-03", 
	"pmid": "30170119", 
	"minutes": 15,
	"title": "Variable cardiac myosin binding protein-C expression in the myofilaments due to MYBPC3 mutations in hypertrophic cardiomyopathy",
	"source": "J. Mol. Cell. Cardiol. 2018-08-28",
	"url": "https://doi.org/10.1016/j.yjmcc.2018.08.023",
	"comment": "lorem ipsum..."
  }
]`