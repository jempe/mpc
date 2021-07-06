//MPC Server takes care of the web server
//
package mpcserver

import (
	"github.com/jempe/encdec"
	"github.com/jempe/mpc/auth"
	"github.com/jempe/mpc/library"
	"github.com/jempe/mpc/storage"
	"github.com/jempe/mpc/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	IP      string
	Storage *mpcstorage.Storage
	Library *mpclibrary.Library
	Auth    *mpcauth.Auth
	Key     string
}

// VideosHandler handles shows JSON videos list
//
func (server *Server) VideosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var filter mpcstorage.VideoFilter

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&filter)
	if err != nil {
		log.Println(err)
	}

	offset := 0
	view := 5
	var seed int64 = 0

	getOffset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err == nil {
		offset = getOffset
	}

	getView, err := strconv.Atoi(r.URL.Query().Get("view"))
	if err == nil {
		view = getView
	}

	getSeed, err := strconv.Atoi(r.URL.Query().Get("seed"))
	if err == nil {
		seed = int64(getSeed)
	}

	sortBy := r.URL.Query().Get("sort")

	results := server.Storage.GetVideos(offset, view, sortBy, filter, seed)

	responseJSON, err := json.Marshal(results)
	if err != nil {
		log.Println(err)
	}

	fmt.Fprintln(w, string(responseJSON))
}

// VideoFileHandler serves video files
//
func (server *Server) VideoFileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	uri := string(r.URL.Path)
	uriSegments := strings.Split(uri, "/")
	thumbFile := uriSegments[2]

	if strings.HasSuffix(thumbFile, ".mp4") {
		videoID := strings.TrimSuffix(thumbFile, ".mp4")

		videoData, err := server.Storage.GetVideoByID(videoID)
		fmt.Println(videoData.File)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		var videoFileData []byte

		if strings.HasSuffix(videoData.File, ".enc") {
			encryptedFilePath := server.Library.Path + "/" + videoData.File

			data, err := ioutil.ReadFile(encryptedFilePath)

			if err != nil {
				fmt.Fprint(w, err)
			} else {
				videoFileData, err = encdec.Decrypt(data, []byte(server.Key))

				if err != nil {
					fmt.Fprint(w, err)
				}

				http.ServeContent(w, r, thumbFile, time.Now(), bytes.NewReader(videoFileData))
			}
		} else {
			videoPath := server.Library.Path + "/" + videoData.File

			file, err := os.Open(videoPath)

			if err != nil {
				fmt.Fprint(w, err)
			}

			defer file.Close()

			http.ServeContent(w, r, thumbFile, time.Now(), file)
		}

	} else {
		fmt.Fprint(w, "Invalid Request")
	}
}

// ThumbsHandler serves thumb files
//
func (server *Server) ThumbsHandler(w http.ResponseWriter, r *http.Request) {
	uri := string(r.URL.Path)
	uriSegments := strings.Split(uri, "/")
	thumbFile := uriSegments[3]

	if strings.HasSuffix(thumbFile, ".jpg") {
		videoID := strings.TrimSuffix(thumbFile, ".jpg")

		videoData, err := server.Storage.GetVideoByID(videoID)

		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		if videoData.Encrypted {
			encryptedThumbPath := server.Library.Path + "/" + strings.Replace(videoData.File, ".enc", "_thumb.enc", 1)

			data, err := ioutil.ReadFile(encryptedThumbPath)

			if err != nil {
				fmt.Fprint(w, err)
				return
			} else {
				thumbData, err := encdec.Decrypt(data, []byte(server.Key))

				if err != nil {
					fmt.Fprint(w, err)
					return
				}
				http.ServeContent(w, r, thumbFile, time.Now(), bytes.NewReader(thumbData))
			}

		} else {

			thumbPath := server.Library.Path + "/" + videoData.File + ".jpg"

			if !mpcutils.Exists(thumbPath) {
				fmt.Println(thumbPath, "doesn't exist")

				if videoData.ImgURL != "" {
					err = mpcutils.DownloadImage(videoData.ImgURL, thumbPath)

					if err != nil {
						fmt.Fprint(w, err)
						return
					}
				} else {
					fmt.Fprint(w, errors.New("This video has no thumbnail to download"))
				}
			}

			file, err := os.Open(thumbPath)

			if err != nil {
				fmt.Fprint(w, err)
			}

			defer file.Close()

			http.ServeContent(w, r, thumbFile, time.Now(), file)
		}
	} else {
		fmt.Fprint(w, "Invalid Request")
	}
}

// ScreenshotsHandler serve screenshots
//
func (server *Server) ScreenshotsHandler(w http.ResponseWriter, r *http.Request) {
	uri := string(r.URL.Path)
	uriSegments := strings.Split(uri, "/")
	videoID := uriSegments[3]
	frameFile := uriSegments[4]

	videoData, err := server.Storage.GetVideoByID(videoID)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	if videoData.ID != "" && strings.HasSuffix(frameFile, ".jpg") {
		screenshotFolder := server.Library.Path + "/thumbs/" + videoID

		if !mpcutils.Exists(screenshotFolder) {
			fmt.Println("Screenshots folder doesn't exist. Creating folder ", screenshotFolder)
			err := os.MkdirAll(screenshotFolder, 0700)

			if err != nil {
				fmt.Fprint(w, err)
				return
			}
		}

		screenshotsPath := screenshotFolder + "/" + frameFile

		if videoData.Encrypted {
			screenshotsPath = strings.Replace(screenshotsPath, ".jpg", ".enc", 1)

			if mpcutils.Exists(screenshotsPath) {
				data, err := ioutil.ReadFile(screenshotsPath)

				if err != nil {
					fmt.Fprint(w, err)
					return
				} else {
					thumbData, err := encdec.Decrypt(data, []byte(server.Key))

					if err != nil {
						fmt.Fprint(w, err)
						return
					}

					http.ServeContent(w, r, frameFile, time.Now(), bytes.NewReader(thumbData))
				}
			} else {
				fmt.Fprint(w, errors.New(screenshotsPath+"doesn't exist"))
				return
			}

		} else {
			if !mpcutils.Exists(screenshotsPath) {
				fmt.Println(screenshotsPath, "doesn't exist")

				frameTime := strings.TrimSuffix(frameFile, ".jpg")

				if mpcutils.Exists(server.Library.Path + "/" + videoData.File) {
					err = mpcutils.SaveScreenshot(server.Library.Path+"/"+videoData.File, frameTime, screenshotsPath)

					if err != nil {
						fmt.Fprint(w, err)
						return
					}
				} else {
					fmt.Fprint(w, errors.New(videoData.File+"doesn't exist"))
					return
				}
			}

			file, err := os.Open(screenshotsPath)

			if err != nil {
				fmt.Fprint(w, err)
			}

			defer file.Close()

			http.ServeContent(w, r, frameFile, time.Now(), file)
		}

	} else {
		fmt.Fprint(w, "Invalid Request")
	}
}

// ScanHandler scans the Library
//
func (server *Server) ScanHandler(w http.ResponseWriter, r *http.Request) {
	err := server.Library.Scan(server.Library.Settings.LibraryPath)
	if err != nil {
		fmt.Fprint(w, err)
	}

	err = server.Storage.InsertVideos(server.Library.Videos)
	if err != nil {
		fmt.Fprint(w, err)
	}

	err = server.Storage.GetAllVideos()
	if err != nil {
		fmt.Fprint(w, err)
	}

}

// IsLogggedIn
//
func (server *Server) IsLoggedIn(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sessionID")
	if err != nil {
		fmt.Fprint(w, err)
	}

	sessionID, err := server.Auth.ValidateToken(cookie.Value)
	if err != nil {
		fmt.Fprint(w, err)
	}

	uuid := server.Auth.ValidateSessionID(sessionID)

	fmt.Fprint(w, uuid)
}

// LoginHandler login users
//
func (server *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	token, err := server.Auth.Authorize(r.FormValue("username"), r.FormValue("password"))

	expiration := time.Now().Add(time.Hour)
	cookie := http.Cookie{Name: "sessionID", Value: token, Expires: expiration, HttpOnly: true}
	http.SetCookie(w, &cookie)
	fmt.Fprint(w, token, err)
}
