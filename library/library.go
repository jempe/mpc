//MPC Library handles the library videos, scans them process them
//
package mpclibrary

import (
	"github.com/jempe/mpc/utils"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Library struct {
	Path     string
	Videos   Videos
	Settings Settings
}

type Videos []Video

type Video struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	ThumbURL     string     `json:"thumbURL"`
	ImgURL       string     `json:"imgURL"`
	VideoURL     string     `json:"videoURL"`
	Description  string     `json:"description"`
	PubDate      time.Time  `json:"pubDate"`
	SubtitlesURL string     `json:"subtitlesURL"`
	Categories   []Category `json:"categories"`
	Actors       []Actor    `json:"actors"`
	Extension    string     `json:"extension"`
	Width        int        `json:"width"`
	Height       int        `json:"height"`
	Duration     int        `json:"duration"`
	Step         int        `json:"step"`
	File         string     `json:"file"`
	OrigFile     string     `json:"orig_file"`
	Path         string     `json:"path"`
	Md5Sum       string     `json:"md5sum"`
	Encrypted    bool       `json:"encrypted"`
	Order        int        `json:"order"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Actor struct {
	ID     int       `json:"id"`
	Name   string    `json:"name"`
	Gender string    `json:"gender"`
	Birth  time.Time `json:birth`
}

type VideoJSON struct {
	Actors      []string `json:"actors"`
	ReleaseDate string   `json:"releaseDate"`
	Description string   `json:"description"`
	Title       string   `json:"title"`
	Who         []string `json:"who"`
	Categories  []string `json:"categories"`
	URL         string   `json:"url"`
	Image       string   `json:"image"`
	Md5Sum      string   `json:"md5sum"`
}

type Settings struct {
	LibraryPath string
	ClusterID   string
	InstanceID  string
	HMACkey     []byte
}

// ScanDirectory  scans a folder to find videos
//
func (lib *Library) Scan(directory string) (err error) {
	log.Println("scanning ", directory)
	files, err := ioutil.ReadDir(directory)

	if err != nil {
		return err
	}

	lib.Videos = Videos{}

	for _, f := range files {
		isVid, fileName, fileExtension := mpcutils.IsVideo(f.Name())

		if isVid {
			thisVideo := Video{}

			videoJSON := lib.Path + "/" + strings.Replace(fileName, ".mp4", ".json", 1)
			if mpcutils.Exists(videoJSON) {
				thisVideoJSON, err := GetJSONData(videoJSON)

				if err == nil {
					thisVideo = thisVideoJSON
				} else {
					log.Println(err)
				}
			} else {
				thisVideo.Title = f.Name()
			}

			thisVideo.Extension = fileExtension
			thisVideo.File = f.Name()
			thisVideo.Path = lib.Path + "/"

			lib.Videos = append(lib.Videos, thisVideo)
		}
	}

	log.Println(len(lib.Videos), " videos found")

	return err
}

// GetJSONData gets the video data from json file
//
func GetJSONData(videoJSON string) (thisVideo Video, err error) {
	videoData, err := ioutil.ReadFile(videoJSON)
	var videoDataJSON VideoJSON

	if err == nil {
		err = json.Unmarshal(videoData, &videoDataJSON)

		if err == nil {
			var videoActors []Actor
			for _, actor := range videoDataJSON.Actors {
				videoActors = append(videoActors, Actor{Name: actor})
			}

			var videoCategories []Category
			for _, category := range videoDataJSON.Categories {
				videoCategories = append(videoCategories, Category{Name: category})
			}

			thisVideo.VideoURL = videoDataJSON.URL
			thisVideo.ImgURL = videoDataJSON.Image
			thisVideo.Title = videoDataJSON.Title
			thisVideo.Description = videoDataJSON.Description
			thisVideo.Actors = videoActors
			thisVideo.Categories = videoCategories
			thisVideo.Md5Sum = videoDataJSON.Md5Sum

			videoDate, err := mpcutils.ParseDate(videoDataJSON.ReleaseDate)

			if err == nil {
				thisVideo.PubDate = videoDate
			}
		} else {
			return thisVideo, err
		}
	} else {
		return thisVideo, err
	}
	return thisVideo, err
}

// SaveJSONData saves the video data on a json file
//
func (lib *Library) SaveJSONData(videoData Video) (err error) {
	var actors []string

	for _, actor := range videoData.Actors {
		actors = append(actors, actor.Name)
	}

	var categories []string

	for _, category := range videoData.Categories {
		categories = append(categories, category.Name)
	}

	output := VideoJSON{Description: videoData.Description, Title: videoData.Title, URL: videoData.VideoURL, Image: videoData.ImgURL, Actors: actors, Categories: categories, Md5Sum: videoData.Md5Sum, ReleaseDate: videoData.PubDate.Format("2006-01-02")}

	if videoData.Md5Sum != "" {
		json, err := json.MarshalIndent(output, "", "\t")

		if err == nil {
			jsonPath := lib.Path + "/" + videoData.Md5Sum + ".json"

			err = ioutil.WriteFile(jsonPath, json, 0644)
		}
	} else {
		err = errors.New("file doesn't have MD5 checkSum")
	}

	return err
}

// importVideo gets md5sum of video and copies it to the library folder and return video info
//
func (lib *Library) ImportVideo(file string) (video Video, err error) {
	isVideo, fileName, extension := mpcutils.IsVideo(file)

	if mpcutils.Exists(file) && isVideo {
		if mpcutils.Exists(lib.Path) && mpcutils.IsDir(lib.Path) {
			md5Sum, err := mpcutils.FileMD5(file)

			if err == nil {
				targetFile := lib.Path + "/" + md5Sum + "." + extension

				if mpcutils.Exists(targetFile) {
					err = errors.New(fileName + " is already in your library")
					return video, err
				} else {
					err = mpcutils.CopyFile(file, targetFile)

					if err == nil {
						videoJSON := strings.Replace(file, "."+extension, ".json", 1)
						if mpcutils.Exists(videoJSON) {
							thisVideoJSON, err := GetJSONData(videoJSON)

							if err == nil {
								video = thisVideoJSON

								lib.SaveJSONData(video)
							}
						} else {
							video.Title = fileName
						}

						video.Extension = extension
						video.OrigFile = fileName
						video.File = md5Sum + "." + extension
						video.Path = lib.Path + "/"
						video.Md5Sum = md5Sum

						videoInfo, err := mpcutils.FFProbe(file)
						if err == nil {
							video.Width = videoInfo.Width
							video.Height = videoInfo.Height
							video.Duration = videoInfo.Duration
							video.Step = videoInfo.Step
						}

						lib.SaveJSONData(video)
					} else {
						return video, err
					}
				}
			} else {
				return video, err
			}
		} else {
			err = errors.New(lib.Path + "doesn't exist")
		}
	} else {
		err = errors.New(file + "doesn't exist or is not a supported video")
	}
	return
}

// GenerateScreenshots saves video screenshots as jpg
//
func (lib *Library) GenerateScreenshots(videoData Video) (err error) {
	if videoData.Step > 0 {
		screenshotFolder := lib.Path + "/thumbs/" + videoData.Md5Sum

		if !mpcutils.Exists(screenshotFolder) {
			log.Println("Screenshots folder doesn't exist. Creating folder ", screenshotFolder)
			err := os.MkdirAll(screenshotFolder, 0700)

			if err != nil {
				return err
			}
		}

		for screenshotTime := 0; screenshotTime < videoData.Duration; screenshotTime += videoData.Step {
			screenshotPath := screenshotFolder + "/" + strconv.Itoa(screenshotTime) + ".jpg"

			if !mpcutils.Exists(screenshotPath) {
				log.Println("save screenshot", screenshotPath)

				err = mpcutils.SaveScreenshot(lib.Path+"/"+videoData.File, strconv.Itoa(screenshotTime), screenshotPath)

				if err != nil {
					return err
				}
			}
		}
	}

	return
}

func (s Videos) Len() int {
	return len(s)
}
func (s Videos) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type ByTitle struct {
	Videos
}

func (s ByTitle) Less(i, j int) bool {
	return s.Videos[i].Title < s.Videos[j].Title
}

type ByTitleDesc struct {
	Videos
}

func (s ByTitleDesc) Less(i, j int) bool {
	return s.Videos[i].Title > s.Videos[j].Title
}

type ByDuration struct {
	Videos
}

func (s ByDuration) Less(i, j int) bool {
	return s.Videos[i].Duration < s.Videos[j].Duration
}

type ByDuractionDesc struct {
	Videos
}

func (s ByDuractionDesc) Less(i, j int) bool {
	return s.Videos[i].Duration > s.Videos[j].Duration
}

type ByRandom struct {
	Videos
}

func (s ByRandom) Less(i, j int) bool {
	return s.Videos[i].Order < s.Videos[j].Order
}
