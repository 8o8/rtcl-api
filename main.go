package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/34South/envr"
	"github.com/mikedonnici/rtcl-api/datastore"
	"github.com/mikedonnici/rtcl-api/datastore/mongo"
	"github.com/mikedonnici/rtcl-api/server"
)

const defaultPort = "5000"

func main() {

	var err error

	// flags
	portFlag := flag.String("p", "", "Specify port number (optional)")
	cfgFlag := flag.String("c", "", "Specify cfg file (optional - will override env vars)")
	flag.Parse()

	port := setPort(*portFlag)
	setEnv(*cfgFlag)

	d := datastore.New()
	d.Mongo, err = mongo.NewConnection(
		os.Getenv("MONGODB_URI"),
		os.Getenv("MONGODB_NAME"),
		os.Getenv("MONGODB_DESC"),
	)
	if err != nil {
		log.Fatalf("Datastore could not connect to MongoDB")
	}

	if os.Getenv("PASSWORD_SALT") == "" {
		log.Println("**WARNING** server starting without env var: PASSWORD_SALT")
	}

	ttl, err := strconv.Atoi(os.Getenv("TOKEN_HOURS_TTL"))
	if err != nil {
		log.Fatalln("Could not convert TOKEN_HOURS_TTL value to an integer")
	}
	cfg := server.Config{
		Port: port,
		Token: server.TokenConfig{
			Issuer:     os.Getenv("TOKEN_ISSUER"),
			SigningKey: os.Getenv("TOKEN_SIGNINGKEY"),
			HoursTTL:   ttl,
		},
	}
	srv := server.NewServer(cfg, d)
	log.Println("server listening on port " + port)
	log.Fatal(srv.Start())
}

// setPort sets the port number for the server, with the env var taking the highest precedence.
func setPort(port string) string {
	if os.Getenv("PORT") != "" {
		return os.Getenv("PORT")
	}
	if port != "" {
		return port
	}
	return defaultPort
}

func setEnv(cfg string) {

	// declare required env vars
	e := envr.New("rtclEnv", []string{
		"ADMIN_API_KEY",
		"API_URL",
		"APP_URL",
		"ALGOLIA_APP_ID",
		"ALGOLIA_ADMIN_KEY",
		"MONGODB_URI",
		"MONGODB_NAME",
		"MONGODB_DESC",
		"SENDGRID_API_KEY",
		"TOKEN_ISSUER",
		"TOKEN_SIGNINGKEY",
		"TOKEN_HOURS_TTL",
	})

	// override default .env
	if cfg != "" {
		log.Println("Setting env from", cfg)
		e.Files = []string{cfg}
	}
	e.Auto()
}
