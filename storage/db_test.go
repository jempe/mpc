package mpcstorage

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"testing"
)

var tempDir string
var storage *Storage

func TestInitDB(t *testing.T) {
	storage = &Storage{Path: tempStorageDir()}

	fmt.Println("create database file in ", storage.Path)

	err := storage.InitDb()
	if err != nil {
		t.Fatal("Init DB error")
	}

	fmt.Println("db created in folder", tempStorageDir())
}

func tempStorageDir() string {
	if tempDir == "" {
		randomString, err := RandomString(30)
		if err != nil {
			fmt.Println("error generating temp folder path")
		}

		tempDir = os.TempDir() + "/" + randomString
	} else {
		return tempDir
	}

	return tempDir
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
