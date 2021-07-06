// MPC Utils has many utilities to process videos and images
//
package mpcutils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type VideoInfo struct {
	Width    int
	Height   int
	Duration int
	Step     int
}

var DefaultSteps int = 30 // Default number of steps to navigate through the video, this changes the total of screenshots that will be taken

// Gets MD5 sum of a string
//
func MD5SumString(name string) string {
	md5sum := md5.Sum([]byte(name))
	return hex.EncodeToString(byte2string(md5sum))
}

// converts bytes to strings
func byte2string(in [16]byte) []byte {
	return in[:16]
}

// FFProbe gets video information using the ffprobe binary
//
func FFProbe(file string) (videoInfo VideoInfo, err error) {
	out, err := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=height,width,duration", file).Output()
	if err != nil {
		return videoInfo, err
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "width=") {

			width, err := strconv.Atoi(strings.Replace(line, "width=", "0", 1))

			if err == nil {
				videoInfo.Width = width
			}
		}
		if strings.HasPrefix(line, "height=") {

			height, err := strconv.Atoi(strings.Replace(line, "height=", "0", 1))

			if err == nil {
				videoInfo.Height = height
			}
		}
		if strings.HasPrefix(line, "duration=") {
			durationRe := regexp.MustCompile("[0-9]+")
			duration, err := strconv.Atoi(durationRe.FindString(line))

			if err == nil {
				videoInfo.Duration = duration
			}
		}
	}

	videoInfo.Step = videoInfo.Duration / DefaultSteps

	return videoInfo, err
}

// SaveScreenshot saves video screenshot as jpg file using ffmpeg
//
func SaveScreenshot(video string, time string, target string) (err error) {
	_, err = exec.Command("ffmpeg", "-ss", time, "-i", video, "-vframes", "1", "-q:v", "2", target).Output()

	fmt.Println(video)

	if err != nil {
		return err
	}

	return err
}

// Exists checks if file exists
//
func Exists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}
	return false
}

// DownloadImage downloads an image and saves it locally
//
func DownloadImage(url string, path string) error {
	client := &http.Client{}

	resp, err := client.Get(url)
	defer resp.Body.Close()

	if err == nil {
		file, err := os.Create(path)

		defer file.Close()

		if err == nil {
			_, err := io.Copy(file, resp.Body)

			if err == nil {
				fmt.Println("save " + path)
			} else {
				return err
			}

		} else {
			return err
		}

	} else {
		return err
	}

	return nil
}

// HomeFolder finds home Folder for every OS
//
func ConfigFolder() (configPath string) {
	folder := "/.mpc"

	if runtime.GOOS == "windows" {
		folder = "/mpc"
	}

	configPath = HomeFolder() + folder

	return
}

// HomeFolder finds home Folder for every OS
//
func HomeFolder() (home string) {
	if runtime.GOOS == "windows" {
		home = os.Getenv("APPDATA")
	} else {
		home = os.Getenv("HOME")
	}
	return home
}

// GetLocalIP gets the IP of LAN
//
func GetLocalIP() string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addresses {
		ip := address.(*net.IPNet)
		if ip.IP.To4() != nil && !ip.IP.IsLoopback() {
			return ip.IP.String()
		}
	}
	return ""
}

// IsDir checks if path is a directory
//
func IsDir(file string) bool {
	if stat, err := os.Stat(file); err == nil && stat.IsDir() {
		return true
	}
	return false
}

// IsVideo checks if file has supported video extension
//
func IsVideo(file string) (is bool, name string, extension string) {
	if IsDir(file) {
		return
	} else {
		extensions := []string{"mp4", "ogv"}

		fileExt := strings.ToLower(filepath.Ext(file))

		for _, extension := range extensions {
			if fileExt == "."+extension {

				is = true
				name = filepath.Base(file)
				return is, name, extension
			}
		}
	}
	return
}

// CopyFile copies files
//
func CopyFile(source string, destination string) (err error) {
	if Exists(source) && !Exists(destination) {
		sourceFile, err := os.Open(source)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		destFile, err := os.Create(destination)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, sourceFile)
		if err != nil {
			return err
		}

		err = destFile.Sync()
		if err != nil {
			return err
		}
	} else {
		err = errors.New("source file doesn't exist or destination file already exists")
	}

	return err
}

// Parse Date parses multiple date formats
//
func ParseDate(date string) (parsedDate time.Time, err error) {
	reg := regexp.MustCompile("[0-9]{2}-[0-9]{2}-[0-9]{4}")

	if reg.Match([]byte(date)) {
		parsedDate, err = time.Parse("01-02-2006", date)
		return
	}

	reg = regexp.MustCompile("[0-9]{4}-[0-9]{2}-[0-9]{2}")

	if reg.Match([]byte(date)) {
		parsedDate, err = time.Parse("2006-01-02", date)
		return
	}

	reg = regexp.MustCompile("[A-Z][a-z]{2} [0-9]+, [0-9]{4}")

	if reg.Match([]byte(date)) {
		parsedDate, err = time.Parse("Jan 2, 2006", date)
		return
	}

	reg = regexp.MustCompile("[A-Z][a-z]+ [0-9]+, [0-9]{4}")

	if reg.Match([]byte(date)) {
		parsedDate, err = time.Parse("January 2, 2006", date)
		return
	}

	return
}

// get MD5 checksum of file
func FileMD5(source string) (md5sum string, err error) {
	if Exists(source) {

		file, err := os.Open(source)
		if err != nil {
			err = errors.New("file_not_exists")
			return md5sum, err
		}
		defer file.Close()
		hash := md5.New()
		_, err = io.Copy(hash, file)
		hashInBytes := hash.Sum(nil)[:16]

		md5sum = hex.EncodeToString(hashInBytes)

	} else {
		err = errors.New("file_not_exists")
	}

	return
}

// checkErr checks and show errors
func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
