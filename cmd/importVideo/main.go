package main

import (
	"github.com/jempe/mpc/library"
	"github.com/jempe/mpc/storage"
	"github.com/jempe/mpc/utils"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) > 1 {
		source := os.Args[1]
		configPath := mpcutils.ConfigFolder()

		storage := &mpcstorage.Storage{Path: configPath}

		err := storage.InitDb()
		printErr(err)

		settings, err := storage.GetSettings()
		printErr(err)

		fileName := filepath.Base(source)

		fmt.Println("Importing", fileName)

		sameFileName, err := storage.GetVideoByOriginalName(fileName)

		if err == nil && sameFileName.ID == "" {

			library := &mpclibrary.Library{Path: settings.LibraryPath, Settings: settings}

			videoData, err := library.ImportVideo(source)
			printErr(err)

			var videos []mpclibrary.Video
			videos = append(videos, videoData)

			err = storage.InsertVideos(videos)
			printErr(err)

			if videoData.Md5Sum != "" {
				videoScreenShot := source + ".jpg"
				targetScreenshot := library.Path + "/" + videoData.Md5Sum + ".mp4.jpg"

				if !mpcutils.Exists(targetScreenshot) {
					if mpcutils.Exists(videoScreenShot) {
						err = mpcutils.CopyFile(videoScreenShot, targetScreenshot)
						printErr(err)
					} else {
						if videoData.ImgURL != "" {
							err = mpcutils.DownloadImage(videoData.ImgURL, targetScreenshot)
							printErr(err)
						}
					}
				}

				err = library.GenerateScreenshots(videoData)
				printErr(err)
			}
		} else {
			fmt.Println("File with name", fileName, "already imported")
		}
	} else {
		fmt.Println("Please enter the file that you want to import")
	}
}

func printErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
