package server_test

import (
	"log"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/mikedonnici/rtcl-api/datastore"
	"github.com/mikedonnici/rtcl-api/server"
	"net/http/httptest"
)

// For storing test tokens and their expected validity
type testToken struct {
	Token       string
	ExpectError bool
}

var serverConfig = server.Config{
	Port: "8888",
	Token: server.TokenConfig{
		Issuer:     "Test Server",
		SigningKey: "SuperDuperSecret1234",
		HoursTTL:   1,
	},
}

var tokenClaims = map[string]interface{}{
	"id":   "5b3bcd72463cd6029e04de18",
	"name": "Mike Donnici",
	"role": "user",
}

// TestValidate token tests the middleware helper function that validates a token in a request Authorization header.
// The request does not need to be served to test this function.
func TestValidateToken(t *testing.T) {
	is := is.New(t)
	s := server.NewServer(serverConfig, nil)
	tokens := testTokens(serverConfig.Token, tokenClaims)

	for _, tt := range tokens {
		r := httptest.NewRequest("GET", "/token", nil)
		r.Header.Add("Authorization", "Bearer "+tt.Token)
		_, err := s.AuthHeaderToken(r)
		is.True((err != nil) == tt.ExpectError) // unexpected token validity
	}
}

// TestAuthTokenValidUser tests the middleware helper function when a user id (id) is specified on the url
func TestValidateUserToken(t *testing.T) {
	is := is.New(t)
	s := server.NewServer(serverConfig, nil)
	tokens := testTokens(serverConfig.Token, tokenClaims)

	// test each token with correct user id - validity determined by the token itself
	userID := tokenClaims["id"].(string)
	for _, tt := range tokens {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Add("Authorization", "Bearer "+tt.Token)
		err := s.ValidateAuthHeaderToken(r, userID)
		is.True((err != nil) == tt.ExpectError) // unexpected token validity
	}

	// repeat tests with incorrect user id (switched last 2 chars) - should all be invalid
	for _, tt := range tokens {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Add("Authorization", "Bearer "+tt.Token)
		err := s.ValidateAuthHeaderToken(r, "5b3bcd72463cd6029e04de81")
		is.True(err != nil) // should be invalid for user id mismatch
	}
}

// generates a set of test tokens - key is the token string, value is validity as bool
func testTokens(tokenConfig server.TokenConfig, claims map[string]interface{}) []testToken {
	var xtt []testToken

	// a valid token
	t, err := datastore.NewToken(tokenConfig.Issuer, tokenConfig.SigningKey, 4).CustomClaims(claims).Encode()
	if err != nil {
		log.Fatalln("Could not generate valid token")
	}
	xtt = append(xtt, testToken{t.String(), false})

	// same token with one char missing should be invalid
	t2 := t.String()[:len(t.String())-2]
	xtt = append(xtt, testToken{Token: t2, ExpectError: true})

	// incorrect signing key should be invalid
	t, err = datastore.NewToken(tokenConfig.Issuer, "dodgeyKey", 4).CustomClaims(claims).Encode()
	if err != nil {
		log.Fatalln("Could not generate test token with bad signing key")
	}
	xtt = append(xtt, testToken{Token: t.String(), ExpectError: true})

	// expired token is invalid
	iat := time.Now().Add(-2 * time.Hour)
	t, err = datastore.NewToken(tokenConfig.Issuer, tokenConfig.SigningKey, 1).SetTimes(iat).CustomClaims(claims).Encode()
	if err != nil {
		log.Fatalln("Could not generate expired test token")
	}
	xtt = append(xtt, testToken{Token: t.String(), ExpectError: true})

	// token with iat in future is also invalid
	iat = time.Now().Add(2 * time.Hour)
	t, err = datastore.NewToken(tokenConfig.Issuer, tokenConfig.SigningKey, 1).SetTimes(iat).CustomClaims(claims).Encode()
	if err != nil {
		log.Fatalln("Could not generate expired test token")
	}
	xtt = append(xtt, testToken{Token: t.String(), ExpectError: true})

	return xtt
}
