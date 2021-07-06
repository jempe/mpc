package main

import (
	"github.com/jempe/encdec"
	"github.com/jempe/mpc/library"
	"github.com/jempe/mpc/storage"
	"github.com/jempe/mpc/utils"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var key string

func main() {
	configPath := "/home/mpc/.mpc/"
	target := "/home/mpc/Videos/"

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

		if strings.ToLower(sourceExtension) != ".enc" {
			data, err := ioutil.ReadFile(source)

			if err != nil {
				fmt.Println(err)
			} else {
				md5Sum := encdec.Md5Sum(data)

				encryptedName := md5Sum + ".enc"

				targetFile := target + encryptedName

				if encdec.Exists(targetFile) {
					fmt.Println("file", source, "is already encrypted :", targetFile)
				} else {
					encrypted, err := encdec.Encrypt(data, []byte(key))

					err = ioutil.WriteFile(targetFile, encrypted, 0644)

					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(source, "succesfully encrypted :", targetFile)

						encryptThumbnail(source+".jpg", target+md5Sum+"_thumb.enc")

						// check if video has JSON file with data
						videoJSON := strings.Replace(source, sourceExtension, ".json", 1)

						videoJSONData, err := mpclibrary.GetJSONData(videoJSON)

						var videos []mpclibrary.Video

						thisVideo := mpclibrary.Video{}

						if err == nil {
							thisVideo = videoJSONData
						}

						thisVideo.Extension = strings.Replace(sourceExtension, ".", "", 1)
						thisVideo.File = encryptedName
						thisVideo.Path = target
						thisVideo.Md5Sum = md5Sum
						thisVideo.Encrypted = true

						videoInfo, err := mpcutils.FFProbe(source)
						if err == nil {
							thisVideo.Width = videoInfo.Width
							thisVideo.Height = videoInfo.Height
							thisVideo.Duration = videoInfo.Duration
							thisVideo.Step = videoInfo.Step
						}

						videos = append(videos, thisVideo)

						storage := &mpcstorage.Storage{Path: configPath}

						err = storage.InitDb()
						checkErr(err)

						err = storage.InsertVideos(videos)
						checkErr(err)

						// create the screenshots

						if thisVideo.Step > 0 {
							screenshotFolder := target + "thumbs/" + thisVideo.Md5Sum

							if !mpcutils.Exists(screenshotFolder) {
								fmt.Println("Screenshots folder doesn't exist. Creating folder ", screenshotFolder)
								err := os.MkdirAll(screenshotFolder, 0700)

								if err != nil {
									fmt.Println(err)
								}
							}

							for screenshotTime := 0; screenshotTime < thisVideo.Duration; screenshotTime += thisVideo.Step {
								screenshotPath := screenshotFolder + "/" + strconv.Itoa(screenshotTime) + ".jpg"

								err = mpcutils.SaveScreenshot(source, strconv.Itoa(screenshotTime), screenshotPath)

								if err == nil {
									if mpcutils.Exists(screenshotPath) {
										encryptedScreenshotPath := screenshotFolder + "/" + strconv.Itoa(screenshotTime) + ".enc"

										fmt.Println("encrypted screenshot", encryptedScreenshotPath)

										encryptThumbnail(screenshotPath, encryptedScreenshotPath)

										if !encdec.IsDir(screenshotPath) && strings.HasSuffix(screenshotPath, ".jpg") {
											err = os.Remove(screenshotPath)

											if err != nil {
												fmt.Println(err)
											}
										}
									} else {
										fmt.Println(screenshotFolder, "doesn't exist")
									}
								}

							}
						}
					}
				}
			}
		} else {
			fmt.Println("Can't encrypt encrypted file", source)
		}
	}
}
func encryptThumbnail(thumbnailFile string, targetFile string) {

	if encdec.Exists(thumbnailFile) {
		data, err := ioutil.ReadFile(thumbnailFile)

		if err == nil {
			encrypted, err := encdec.Encrypt(data, []byte(key))

			err = ioutil.WriteFile(targetFile, encrypted, 0644)

			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(thumbnailFile, "succesfully encrypted :", targetFile)
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
