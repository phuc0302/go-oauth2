package oauth2

import (
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/phuc0302/go-server/expected_format"
	"gopkg.in/mgo.v2/bson"
)

func Test_MongoDBToken(t *testing.T) {
	u := new(TestEnv)
	defer u.Teardown()
	u.Setup()

	mongoStore, _ := Store.(*MongoDBStore)

	token := MongoDBToken{
		ID:      bson.NewObjectId(),
		User:    bson.NewObjectId(),
		Client:  bson.NewObjectId(),
		Created: time.Now(),

		privateKey: mongoStore.privateKey,
	}
	token.Expired = token.Created.Add(Cfg.RefreshTokenDuration)

	// Test token
	tokenString := token.Token()
	if len(tokenString) == 0 {
		t.Error(expectedFormat.NotNil)
	} else {
		jwtToken, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
			}
			return &token.privateKey.PublicKey, nil
		})

		if err != nil || !jwtToken.Valid {
			t.Error("Expected token string should be able to decoded.")
		}
	}
}
