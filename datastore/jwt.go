package datastore

import (
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type Token struct {
	signingKey []byte
	ttlHours   int
	Encoded    string      `json:"token"`
	IssuedAt   time.Time   `json:"issuedAt"`
	ExpiresAt  time.Time   `json:"expiresAt"`
	Claims     TokenClaims `json:"claims"`
}

type TokenClaims struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
	jwt.StandardClaims
}

// New returns a pointer to a Token
func NewToken(issuer, signingKey string, ttlHours int) *Token {

	var t Token

	t.signingKey = []byte(signingKey)
	t.ttlHours = ttlHours

	// Initialise standard claims
	t.Claims.StandardClaims = jwt.StandardClaims{
		Issuer: issuer,
	}

	// Default issue time to now, can override with .ValidFrom()
	t.SetTimes(time.Now())

	return &t
}

// SetTimes sets the issuer, and time-related claims - requires the issue at time to be passed in.
func (t *Token) SetTimes(iat time.Time) *Token {

	t.Claims.StandardClaims.IssuedAt = iat.Unix()
	t.Claims.StandardClaims.ExpiresAt = iat.Add(time.Hour * time.Duration(t.ttlHours)).Unix()

	// Set Unix dates at root of struct for convenience (??)
	t.IssuedAt = time.Unix(int64(t.Claims.StandardClaims.IssuedAt), 0)
	t.ExpiresAt = time.Unix(int64(t.Claims.StandardClaims.ExpiresAt), 0)

	return t
}

// Encode finishes of the Token value and creates the encoded token string or JWS
func (t *Token) Encode() (Token, error) {

	if t.Claims.Issuer == "" {
		return *t, errors.New("Issuer cannot be blank")
	}
	if len(t.signingKey) < 1 {
		return *t, errors.New("Signing key cannot be blank")
	}
	if t.ttlHours < 1 {
		return *t, errors.New("TTL hours must be a positive integer")
	}

	var err error
	t.Encoded, err = jwt.NewWithClaims(jwt.SigningMethodHS256, t.Claims).SignedString([]byte(t.signingKey))
	return *t, err
}

// CustomClaims sets custom claims
func (t *Token) CustomClaims(claims map[string]interface{}) *Token {

	if id, ok := claims["id"]; ok {
		t.Claims.ID = id.(string)
	}
	if name, ok := claims["name"]; ok {
		t.Claims.Name = name.(string)
	}
	if role, ok := claims["role"]; ok {
		t.Claims.Role = role.(string)
	}

	return t
}

// Valid returns true if the Token.Encoded string is a valid JWT
func (t *Token) Valid() bool {
	_, err := jwt.Parse(t.Encoded, func(tok *jwt.Token) (interface{}, error) {
		return []byte(t.signingKey), nil
	})
	if err != nil {
		return false
	}
	return true
}

// String returns the encoded token string (JWS)
func (t *Token) String() string {
	return string(t.Encoded)
}

// Decode attempts to decode token with signingKey and returns a new Token value if everything checks out
func DecodeToken(token, signingKey string) (Token, error) {

	t := Token{
		Encoded:    token,
		signingKey: []byte(signingKey),
	}

	// The jwt library panics if the jwt does not contain 3 '.'s - assume because it splits the string at each period
	// and gets and index out of range if it does not end up with three pieces.
	if len(strings.Split(token, ".")) < 3 {
		return t, errors.New("JWT should have 3 parts in format aaaa.bbbb.cccc")
	}

	// Parse the token which sets the Valid field
	tok, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		return t.signingKey, nil
	})
	if err != nil {
		return t, errors.New("Error parsing token: " + err.Error())
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if ok && tok.Valid {

		//Custom claims
		t.Claims.ID = claims["id"].(string)
		t.Claims.Name = claims["name"].(string)
		t.Claims.Role = claims["role"].(string)

		// Standard claims
		t.Claims.ExpiresAt = int64(claims["exp"].(float64))
		t.Claims.IssuedAt = int64(claims["iat"].(float64))
		t.Claims.Issuer = claims["iss"].(string)

		// reverse engineer ttlHours from iat and exp
		t.ttlHours = (int(t.Claims.ExpiresAt) - int(t.Claims.IssuedAt)) / 3600

		// Set the friendly dates
		issueTime := time.Unix(t.Claims.IssuedAt, 0)
		t.SetTimes(issueTime)

		return t, nil
	}

	// Below here we are in error land
	ve, ok := err.(*jwt.ValidationError)
	if !ok {
		return t, errors.New("Failed to validate token - " + err.Error())
	}

	if ve.Errors&jwt.ValidationErrorMalformed != 0 {
		return t, errors.New("Token malformed")
	}
	if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
		return t, errors.New("Token issue date is in the future - not valid yet?")
	}
	if ve.Errors&jwt.ValidationErrorExpired != 0 {
		return t, errors.New("Token has expired")
	}
	if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
		return t, errors.New("Token signature is invalid")
	}

	return t, errors.New("Token error: " + err.Error())
}

// FromHeader extracts the jwt string from the header Authorization string (a).
// The Header should be in the format: Bearer aaaa.bbbb.cccc
func FromHeader(a string) (string, error) {

	var errMsg error

	t := strings.Fields(a) // splits on any amount of white space
	if len(t) < 2 || t[0] != "Bearer" {
		errMsg = errors.New("Authorization header should be: Bearer [jwt]")
		return "", errMsg
	}

	return strings.TrimSpace(t[1]), nil
}
