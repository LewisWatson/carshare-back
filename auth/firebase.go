package auth

import (
	"crypto/rsa"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
)

// Token firebase secure token
type Token struct {
	kid, token string
}

// Firebase module to verify and extract information from firebase JWT tokens
type Firebase struct {
	publicKeys map[string]*rsa.PublicKey
	projectID  string
}

// NewFirebase loads the firebase keys
func NewFirebase(projectID string) (*Firebase, error) {

	keys, err := updatePublicKeys()
	if err != nil {
		return nil, err
	}

	fb := new(Firebase)
	fb.projectID = projectID
	fb.publicKeys = keys
	return fb, nil
}

func updatePublicKeys() (map[string]*rsa.PublicKey, error) {

	log.Printf("Requesting firebase tokens")
	tokens := make(map[string]interface{})
	err := getFirebaseTokens(tokens)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]*rsa.PublicKey)
	for kid, token := range tokens {
		publicKey, err := crypto.ParseRSAPublicKeyFromPEM([]byte(token.(string)))
		if err != nil {
			log.Printf("Error parsing kid %s, %v", kid, err)
		} else {
			log.Printf("Validated kid %s", kid)
			keys[kid] = publicKey
		}
	}

	return keys, nil
}

// firebase tokens must be signed by one of the keys provided at a certain url.
func getFirebaseTokens(tokens map[string]interface{}) error {
	return getJSON("https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com", &tokens)
}

var myClient = &http.Client{Timeout: 30 * time.Second}

func getJSON(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

// Verify to satisfy the auth.Verify interface
func (fa Firebase) Verify(accessToken string) (jwt.Claims, error) {

	// empty string is clearly invalid
	if accessToken == "" {
		return nil, ErrNilToken
	}

	token, err := jws.ParseJWT([]byte(accessToken))
	if err != nil {
		return nil, err
	}

	// TODO extract kid from header and only verify against that key
	// https://firebase.google.com/docs/auth/admin/verify-id-tokens

	// validate against firebase keys
	for _, key := range fa.publicKeys {
		err = token.Validate(key, crypto.SigningMethodRS256)
		// verification errors indicate that the token isn't valid for this key
		if err == nil || !strings.Contains(err.Error(), "verification error") {
			break
		}
	}

	switch err {
	case jwt.ErrTokenIsExpired:
		err = ErrTokenExpired
		break
	case crypto.ErrECDSAVerification:
		err = ErrECDSAVerification
		break
	case jws.ErrNotCompact:
		err = ErrNotCompact
		break
	}

	aud := token.Claims().Get("aud")
	if aud != fa.projectID {
		log.Printf("Invalid authorisation token audience %v, expecting %v", aud, fa.projectID)
		err = ErrInvalidAud
	}

	iss := token.Claims().Get("iss")
	expectedIss := "https://securetoken.google.com/" + fa.projectID
	if iss != expectedIss {
		log.Printf("Invalid authorisation token issuer %v, expected %v", iss, expectedIss)
		err = ErrInvalidIss
	}

	return token.Claims(), err
}
