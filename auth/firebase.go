package auth

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
)

// Firebase module to verify and extract information from firebase JWT tokens
type Firebase struct {
	projectID          string
	publicKeys         map[string]*rsa.PublicKey
	cacheControlMaxAge int64
	keysLastUpdatesd   int64
	sync.RWMutex
}

// NewFirebase creates a new instance of firebase auth and loads the latest
// keys from the firebase servers
func NewFirebase(projectID string) (*Firebase, error) {
	fb := new(Firebase)
	fb.projectID = projectID
	return fb, fb.UpdatePublicKeys()
}

// UpdatePublicKeys retrieves the latest firebase keys
func (fb *Firebase) UpdatePublicKeys() error {
	log.Printf("Requesting firebase tokens")
	tokens := make(map[string]interface{})
	maxAge, err := getFirebaseTokens(tokens)
	if err != nil {
		return err
	}
	fb.Lock()
	fb.cacheControlMaxAge = maxAge
	fb.publicKeys = make(map[string]*rsa.PublicKey)
	for kid, token := range tokens {
		publicKey, err := crypto.ParseRSAPublicKeyFromPEM([]byte(token.(string)))
		if err != nil {
			log.Printf("Error parsing kid %s, %v", kid, err)
		} else {
			log.Printf("Validated kid %s", kid)
			fb.publicKeys[kid] = publicKey
		}
	}
	fb.Unlock()
	return nil
}

// firebase tokens must be signed by one of the keys provided at a certain url.
// The keys expire after a certain amount of time so we need to track that also.
func getFirebaseTokens(tokens map[string]interface{}) (int64, error) {
	r, err := myClient.Get("https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com")
	if err != nil {
		return 0, err
	}
	maxAge, err := extractMaxAge(r.Header.Get("Cache-Control"))
	if err != nil {
		return maxAge, err
	}
	defer r.Body.Close()
	return maxAge, json.NewDecoder(r.Body).Decode(&tokens)
}

// Extract the max age from the cache control response header value
// The cache control header should look similar to "..., max-age=19008, ..."
func extractMaxAge(cacheControl string) (int64, error) {
	// "..., max-age=19008, ..."" to ["..., max-age="]["19008, ..."]
	tokens := strings.Split(cacheControl, "max-age=")
	if len(tokens) == 1 {
		return 0, fmt.Errorf("cache control header doesn't contain a max age")
	}
	// "19008, ..." to ["19008"][" ..."]
	tokens2 := strings.Split(tokens[1], ",")
	// convert "19008" to int64
	return strconv.ParseInt(tokens2[0], 10, 64)
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

// checks if the current firebase keys are stale and therefore need updating
func (fb *Firebase) keysStale() bool {
	return (time.Now().UnixNano() - fb.keysLastUpdatesd) > fb.cacheControlMaxAge
}

// Verify to satisfy the auth.Verify interface
func (fb *Firebase) Verify(accessToken string) (string, jwt.Claims, error) {

	// empty string is clearly invalid
	if accessToken == "" {
		return "", nil, ErrNilToken
	}

	token, err := jws.ParseJWT([]byte(accessToken))
	if err != nil {
		return "", nil, err
	}

	if fb.keysStale() {
		log.Println("Firebase keys stale")
		fb.UpdatePublicKeys()
	}

	fb.RLock()

	// validate against firebase keys
	for _, key := range fb.publicKeys {
		err = token.Validate(key, crypto.SigningMethodRS256)
		// verification errors indicate that the token isn't valid for this key
		// TODO extract kid from header and only verify against that key
		// https://firebase.google.com/docs/auth/admin/verify-id-tokens
		if err == nil || !strings.Contains(err.Error(), "verification error") {
			break
		}
	}

	fb.RUnlock()

	if err == nil {
		validatior := jwt.Validator{}
		validatior.SetAudience(fb.projectID)
		validatior.SetIssuer("https://securetoken.google.com/" + fb.projectID)
		err = validatior.Validate(token)
	}

	// convert library errors into auth errors

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
	case jwt.ErrInvalidISSClaim:
		err = ErrInvalidIss
		break
	case jwt.ErrInvalidAUDClaim:
		err = ErrInvalidAud
		break
	}

	return token.Claims().Get("sub").(string), token.Claims(), err
}
