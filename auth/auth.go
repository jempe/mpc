package mpcauth

import (
	"github.com/jempe/mpc/storage"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
)

type Auth struct {
	Key            []byte
	Storage        *mpcstorage.Storage
	Authorizations []AuthData
}

type AuthData struct {
	SessionID string
	UUID      string
}

// Authorize checks username and password and generates a token with the session id
//
func (auth *Auth) Authorize(username string, password string) (tokenString string, err error) {
	users, _ := auth.Storage.GetUsers()

	for _, user := range users {
		if username == user.Name || username == user.Email {
			if mpcstorage.HashPassword(password) == user.Password {
				tokenString, err = auth.GenerateTokenString(user.UUID)
				fmt.Println(auth.Authorizations)
				return
			} else {
				return tokenString, errors.New("user_wrong_password")
			}
		}
	}

	err = errors.New("user_not_exists")

	return
}

//GenerateTokenString generates New JWT Token String
//
func (auth *Auth) GenerateTokenString(userID string) (tokenString string, err error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	sessionID, err := auth.NewAuthorization(userID)

	if err != nil {
		return "", errors.New("auth_session_generation_error")
	}

	claims["sessionID"] = sessionID

	tokenString, err = token.SignedString(auth.Key)

	return
}

// NewAuthorization generates a new sessionID
//
func (auth *Auth) NewAuthorization(userID string) (sessionID string, err error) {

	sessionID, err = RandomString(64)

	if err != nil {
		return
	}

	var NewAuthorizations []AuthData

	for _, authorization := range auth.Authorizations {
		if authorization.UUID == userID {
			authorization.SessionID = sessionID
		}

		NewAuthorizations = append(NewAuthorizations, authorization)
	}

	auth.Authorizations = NewAuthorizations

	auth.Authorizations = append(auth.Authorizations, AuthData{UUID: userID, SessionID: sessionID})

	return
}

// ValidateToken validates the token and returns the session id
//
func (auth *Auth) ValidateToken(tokenString string) (sessionId string, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if ok {
			return auth.Key, nil
		} else {
			return nil, errors.New("Wrong singing method")
		}
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["sessionID"].(string), nil
	}

	return
}

// ValidateSessionID validates the sessionID
//
func (auth *Auth) ValidateSessionID(sessionID string) string {
	for _, authorization := range auth.Authorizations {
		if authorization.SessionID == sessionID {
			return authorization.UUID
		}
	}

	return ""
}

//RandomBytes is useful to generate HMAC key
func RandomBytes(length int) (key []byte, err error) {
	randomBytes := make([]byte, length)

	_, err = rand.Read(randomBytes)
	if err == nil {
		key = randomBytes
	}

	return key, err
}

//RandomBytes is useful to generate HMAC key
func RandomString(length int) (key string, err error) {
	randomBytes, err := RandomBytes(length)

	if err == nil {
		key = base64.URLEncoding.EncodeToString(randomBytes)
	}

	return key, err
}
