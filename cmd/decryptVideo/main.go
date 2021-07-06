package main

import (
	"github.com/jempe/encdec"
	"github.com/jempe/mpc/storage"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var key string

func main() {
	configPath := "/home/mpc/.mpc/"
	target := os.Args[2]

	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	viper.ReadInConfig()

	key = viper.GetString("key")

	source := os.Args[1]

	if !encdec.Exists(source) {
		fmt.Println("source file", source, "doesn't exist")
	} else if !encdec.Exists(target) {
		fmt.Println("target folder", target, "doesn't exist")
	} else if encdec.IsDir(source) {
		fmt.Println("source", source, "can't be a folder")
	} else if !encdec.IsDir(target) {
		fmt.Println("target", target, "is not a folder")
	} else {
		target = strings.TrimSuffix(target, "/") + "/"
		sourceExtension := filepath.Ext(source)

		if strings.ToLower(sourceExtension) == ".enc" {
			data, err := ioutil.ReadFile(source)

			if err != nil {
				fmt.Println(err)
			} else {
				decryptedName := strings.Replace(filepath.Base(source), ".enc", ".mp4", 1)

				targetFile := target + decryptedName

				if encdec.Exists(targetFile) {
					fmt.Println("file", source, "is already decrypted :", targetFile)
				} else {
					decrypted, err := encdec.Decrypt(data, []byte(key))

					err = ioutil.WriteFile(targetFile, decrypted, 0644)

					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(source, "succesfully decrypted :", targetFile)

						decryptThumbnail(strings.Replace(source, ".enc", "_thumb.enc", 1), targetFile+".jpg")

						storage := &mpcstorage.Storage{Path: configPath}

						err = storage.InitDb()
						checkErr(err)

						videoData, err := storage.GetVideoByID(strings.Replace(decryptedName, ".mp4", "", 1))

						if err == nil {
							fmt.Println(videoData)
						}
					}
				}
			}
		} else {
			fmt.Println("Can't decrypt decrypted file", source)
		}
	}
}
func decryptThumbnail(thumbnailFile string, targetFile string) {

	if encdec.Exists(thumbnailFile) {
		data, err := ioutil.ReadFile(thumbnailFile)

		if err == nil {
			decrypted, err := encdec.Decrypt(data, []byte(key))

			err = ioutil.WriteFile(targetFile, decrypted, 0644)

			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(thumbnailFile, "succesfully decrypted :", targetFile)
			}
		} else {
			fmt.Println(err)
		}
	} else {
		fmt.Println("thumbnail ", thumbnailFile, "doesn't exist")
	}
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
