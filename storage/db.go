// MPC DB Storage
// Saves Videos Information in a boltdb database
//
package mpcstorage

import (
	"github.com/jempe/mpc/library"
	"github.com/jempe/mpc/users"
	"github.com/jempe/mpc/utils"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Storage struct {
	Db         *bolt.DB
	Path       string
	Videos     map[string]Video
	Categories map[int]mpclibrary.Category
	Actors     map[int]mpclibrary.Actor
}

type Video struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	ThumbURL     string    `json:"thumbURL"`
	ImgURL       string    `json:"imgURL"`
	VideoURL     string    `json:"videoURL"`
	Description  string    `json:"description"`
	PubDate      time.Time `json:"pubDate"`
	SubtitlesURL string    `json:"subtitlesURL"`
	Categories   []int     `json:"categories"`
	Actors       []int     `json:"actors"`
	Extension    string    `json:"extension"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Duration     int       `json:"duration"`
	Step         int       `json:"step"`
	File         string    `json:"file"`
	OrigFile     string    `json:"orig_file"`
	Path         string    `json:"path"`
	Md5Sum       string    `json:"md5sum"`
	Encrypted    bool      `json:"encrypted"`
}

type VideoResults struct {
	Videos []mpclibrary.Video `json:"videos"`
	Total  int                `json:"total"`
	Offset int                `json:"offset"`
	View   int                `json:"view"`
}

type VideoFilter struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Actor    string `json:"actor"`
	Quality  string `json:"quality"`
	Duration [2]int `json:"duration"`
}

// Initialize DB
//
// Creates folder to save DB, Creates DB file and DB buckets, it also loads all the DB data
//
func (storage *Storage) InitDb() error {
	var err error

	if !mpcutils.Exists(storage.Path) {
		fmt.Println("Configuration folder doesn't exist. Creating folder ", storage.Path)
		err = os.MkdirAll(storage.Path, 0700)

		if err != nil {
			return err
		}
	}

	dbPath := storage.Path + "/mpc.db"

	if !mpcutils.Exists(dbPath) {
		fmt.Println("Creating DB")
	}

	storage.Db, err = bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return err
	}

	err = storage.createBucket("videos")
	if err != nil {
		return err
	}

	err = storage.createBucket("categories")
	if err != nil {
		return err
	}

	err = storage.createBucket("actors")
	if err != nil {
		return err
	}

	err = storage.createBucket("settings")
	if err != nil {
		return err
	}

	err = storage.createBucket("users")
	if err != nil {
		return err
	}

	err = storage.getAllActors()
	if err != nil {
		return err
	}

	err = storage.getAllCategories()
	if err != nil {
		return err
	}

	err = storage.GetAllVideos()
	if err != nil {
		return err
	}

	return err
}

// Get Videos
//
// Get videos that meets the search criteria
//
func (storage *Storage) GetVideos(offset int, view int, sortBy string, filter VideoFilter, seed int64) VideoResults {
	var videos mpclibrary.Videos
	var dbVideo Video
	var err error

	totalActiveFilters := 0

	filterActor := false

	var actorReg *regexp.Regexp

	if filter.Actor != "" {
		actorReg, err = regexp.Compile(filter.Actor)
		if err == nil {
			totalActiveFilters++
			filterActor = true
		}
	}

	filterCategory := false

	var categoryReg *regexp.Regexp

	if filter.Category != "" {
		categoryReg, err = regexp.Compile(filter.Category)
		if err == nil {
			totalActiveFilters++
			filterCategory = true
		}
	}

	filterTitle := false

	var titleReg *regexp.Regexp

	if filter.Title != "" {
		titleReg, err = regexp.Compile(filter.Title)
		if err == nil {
			totalActiveFilters++
			filterTitle = true
		}
	}

	_ = storage.Db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("videos"))

		c := b.Cursor()

		var videoIndex int64 = 0

		for k, v := c.First(); k != nil; k, v = c.Next() {
			json.Unmarshal(v, &dbVideo)
			if dbVideo.File != "" {
				libVideo := storage.videoToLibraryVideo(dbVideo)
				rand.Seed(seed + videoIndex)
				videoIndex++
				libVideo.Order = rand.Intn(1000)

				activeFilters := 0

				actorPassed := false

				if filterActor {
					for _, actor := range libVideo.Actors {
						if actorReg.Match([]byte(strings.ToLower(actor.Name))) {
							actorPassed = true
						}
					}
				}

				if actorPassed {
					activeFilters++
				}

				categoryPassed := false

				if filterCategory {
					for _, category := range libVideo.Categories {
						if categoryReg.Match([]byte(strings.ToLower(category.Name))) {
							categoryPassed = true
						}
					}
				}

				if categoryPassed {
					activeFilters++
				}

				if filterTitle {
					if titleReg.Match([]byte(strings.ToLower(libVideo.Title))) {
						activeFilters++
					}
				}

				if activeFilters == totalActiveFilters {
					videos = append(videos, libVideo)
				}
			} else {
				return errors.New("error getting video " + dbVideo.ID)
			}
		}

		return nil
	})

	if sortBy == "title" {
		sort.Sort(mpclibrary.ByTitle{videos})
	} else if sortBy == "titleDesc" {
		sort.Sort(mpclibrary.ByTitleDesc{videos})
	} else if sortBy == "duration" {
		sort.Sort(mpclibrary.ByDuration{videos})
	} else if sortBy == "durationDesc" {
		sort.Sort(mpclibrary.ByTitleDesc{videos})
	} else {
		sort.Sort(mpclibrary.ByRandom{videos})
	}

	var videoResults mpclibrary.Videos

	lastResult := offset + view

	if lastResult > len(videos) {
		lastResult = len(videos)
	}

	if offset < len(videos) {
		videoResults = videos[offset:lastResult]
	}

	results := VideoResults{Videos: videoResults, Offset: offset, View: view, Total: len(videos)}

	return results
}

// GetAllVideos gets all the videos from the DB
//
func (storage *Storage) GetAllVideos() error {
	storage.Videos = make(map[string]Video)
	var dbVideo Video
	err := storage.Db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("videos"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			json.Unmarshal(v, &dbVideo)
			if dbVideo.File != "" {
				storage.Videos[string(k)] = dbVideo
			} else {
				return errors.New("error getting video " + dbVideo.ID)
			}
		}

		return nil
	})

	return err
}

// getAllActors gets all the actors from the DB
//
func (storage *Storage) getAllActors() error {
	allActors := make(map[int]mpclibrary.Actor)
	var dbActor mpclibrary.Actor
	err := storage.Db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("actors"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			json.Unmarshal(v, &dbActor)

			if dbActor.Name != "" {
				allActors[btoi(k)] = dbActor
			} else {
				return errors.New("error getting actors")
			}
		}

		return nil
	})

	storage.Actors = allActors
	return err
}

// getAllCategories gets all the categories from the DB
//
func (storage *Storage) getAllCategories() error {
	allCategories := make(map[int]mpclibrary.Category)
	var dbCategory mpclibrary.Category
	err := storage.Db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("categories"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			json.Unmarshal(v, &dbCategory)

			if dbCategory.Name != "" {
				allCategories[btoi(k)] = dbCategory
			} else {
				return errors.New("error getting categories")
			}
		}

		return nil
	})

	storage.Categories = allCategories

	return err
}

// createBucket creates a new bucket in the boltdb Database only if doesn't exist
//
func (storage *Storage) createBucket(bucketName string) error {
	err := storage.Db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return err
}

// InsertVideos inserts new videos in the DB.
// A list of videos should be provided
//
func (storage *Storage) InsertVideos(videos []mpclibrary.Video) error {
	tx, err := storage.Db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	videosBucket := tx.Bucket([]byte("videos"))
	categoriesBucket := tx.Bucket([]byte("categories"))
	actorsBucket := tx.Bucket([]byte("actors"))

	for _, video := range videos {
		if video.File != "" {
			videoMd5 := mpcutils.MD5SumString(video.File)
			if video.Md5Sum != "" {
				video.ID = video.Md5Sum
			} else {
				video.ID = videoMd5
			}

			stVideo, err := storage.GetVideoByFileName(video.File)

			if err == nil && stVideo.ID == "" && video.Md5Sum != "" {
				stVideo, err = storage.GetVideoByID(video.Md5Sum)
			}

			if err == nil && stVideo.ID == "" {
				for _, actor := range video.Actors {
					dbActor, err := storage.GetActorByName(actor.Name)
					if err != nil {
						return err
					}
					if dbActor.Name == "" {
						fmt.Println("Inserting actor", actor.Name)
						storage.insertActor(actorsBucket, actor)
					}
				}

				for _, category := range video.Categories {
					dbCategory, err := storage.GetCategoryByName(category.Name)
					if err != nil {
						return err
					}
					if dbCategory.Name == "" {
						storage.insertCategory(categoriesBucket, category)
					}
				}

				dbVideo := storage.storageVideoToVideo(video)

				if !dbVideo.Encrypted {
					videoInfo, err := mpcutils.FFProbe(dbVideo.Path + dbVideo.File)
					if err == nil {
						dbVideo.Width = videoInfo.Width
						dbVideo.Height = videoInfo.Height
						dbVideo.Duration = videoInfo.Duration
						dbVideo.Step = videoInfo.Step
					}
				}

				jsonVideo, err := json.Marshal(dbVideo)
				if err != nil {
					return err
				}

				err = videosBucket.Put([]byte(video.ID), jsonVideo)

				if err != nil {
					return err
				}
			}
		}
	}

	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// insertActor inserts a new actor in the DB
//
func (storage *Storage) insertActor(bucket *bolt.Bucket, actor mpclibrary.Actor) error {
	if actor.Name != "" {
		id, _ := bucket.NextSequence()

		actor.ID = int(id)

		jsonActor, err := json.Marshal(actor)
		if err != nil {
			return err
		}

		err = bucket.Put(itob(actor.ID), jsonActor)
		if err != nil {
			return err
		}

		storage.Actors[int(id)] = actor
	}
	return nil
}

// insertCategory inserts a new category in the DB
//
func (storage *Storage) insertCategory(bucket *bolt.Bucket, category mpclibrary.Category) error {
	if category.Name != "" {
		id, _ := bucket.NextSequence()

		category.ID = int(id)

		jsonCategory, err := json.Marshal(category)
		if err != nil {
			return err
		}

		err = bucket.Put(itob(category.ID), jsonCategory)
		if err != nil {
			return err
		}

		storage.Categories[int(id)] = category
	}
	return nil
}

// GetVideoByFileName searchs a video in the DB, it Gets the video ID by using the MD5 sum of its name
//
func (storage *Storage) GetVideoByFileName(fileName string) (video mpclibrary.Video, err error) {
	videoMd5 := mpcutils.MD5SumString(fileName)

	return storage.GetVideoByID(videoMd5)
}

// GetVideoByID searchs a video in the DB using the video ID.
// The Video ID can be the MD5 sum of its name for normal videos or the MD5 sum of the video file for encrypted files
//
func (storage *Storage) GetVideoByID(videoMd5 string) (video mpclibrary.Video, err error) {
	for key, thisVideo := range storage.Videos {
		if key == videoMd5 {
			return storage.videoToLibraryVideo(thisVideo), nil
		}
	}

	return
}

// GetVideoByOriginalName searchs a video in the DB using the file name when it was imported.
//
func (storage *Storage) GetVideoByOriginalName(name string) (video mpclibrary.Video, err error) {
	for _, thisVideo := range storage.Videos {
		if thisVideo.OrigFile == name {
			return storage.videoToLibraryVideo(thisVideo), nil
		}
	}

	return
}

// videoToLibraryVideo converts a DB Video object to a MPC library video object
//
func (storage *Storage) videoToLibraryVideo(dbVideo Video) (video mpclibrary.Video) {
	video.ID = dbVideo.ID
	video.Title = dbVideo.Title
	video.ThumbURL = dbVideo.ThumbURL
	video.ImgURL = dbVideo.ImgURL
	video.VideoURL = dbVideo.VideoURL
	video.Description = dbVideo.Description
	video.PubDate = dbVideo.PubDate
	video.SubtitlesURL = dbVideo.SubtitlesURL
	video.Path = dbVideo.Path
	//Categories   []Category `json:"categories"`
	//Actors       []Actor    `json:"actors"`
	video.Extension = dbVideo.Extension
	video.Width = dbVideo.Width
	video.Height = dbVideo.Height
	video.Duration = dbVideo.Duration
	video.Step = dbVideo.Step
	video.File = dbVideo.File
	video.OrigFile = dbVideo.OrigFile
	video.Md5Sum = dbVideo.Md5Sum
	video.Encrypted = dbVideo.Encrypted

	var categories []mpclibrary.Category

	for _, category := range dbVideo.Categories {
		videoCategory, err := storage.GetCategoryByID(category)
		if err == nil {
			categories = append(categories, videoCategory)
		}
	}

	video.Categories = categories

	var actors []mpclibrary.Actor

	for _, actor := range dbVideo.Actors {
		videoActor, err := storage.GetActorByID(actor)
		if err == nil {
			actors = append(actors, videoActor)
		}
	}

	video.Actors = actors

	return
}

// storageVideoToVideo transforms a library Video object to a DB Video object
//
func (storage *Storage) storageVideoToVideo(video mpclibrary.Video) (dbVideo Video) {
	dbVideo.ID = video.ID
	dbVideo.Title = video.Title
	dbVideo.ThumbURL = video.ThumbURL
	dbVideo.ImgURL = video.ImgURL
	dbVideo.VideoURL = video.VideoURL
	dbVideo.Description = video.Description
	dbVideo.PubDate = video.PubDate
	dbVideo.SubtitlesURL = video.SubtitlesURL
	dbVideo.Path = video.Path
	//Actors       []Actor    `json:"actors"`
	dbVideo.Extension = video.Extension
	dbVideo.Width = video.Width
	dbVideo.Height = video.Height
	dbVideo.Duration = video.Duration
	dbVideo.Step = video.Step
	dbVideo.File = video.File
	dbVideo.OrigFile = video.OrigFile
	dbVideo.Md5Sum = video.Md5Sum
	dbVideo.Encrypted = video.Encrypted

	var videoCategories []int

	for _, category := range video.Categories {
		videoCategory, err := storage.GetCategoryByName(category.Name)
		if err == nil {
			videoCategories = append(videoCategories, videoCategory.ID)
		}
	}

	dbVideo.Categories = videoCategories

	var videoActors []int

	for _, actor := range video.Actors {
		videoActor, err := storage.GetActorByName(actor.Name)
		if err == nil {
			videoActors = append(videoActors, videoActor.ID)
		}
	}

	dbVideo.Actors = videoActors
	return
}

// GetActorByName gets an actor from the list by name
//
func (storage *Storage) GetActorByName(actorName string) (actor mpclibrary.Actor, err error) {
	for _, thisActor := range storage.Actors {
		if actorName == thisActor.Name {
			return thisActor, nil
		}
	}
	return
}

// GetActorByID gets an actor from the list by ID
//
func (storage *Storage) GetActorByID(actorID int) (actor mpclibrary.Actor, err error) {
	for id, thisActor := range storage.Actors {
		if actorID == id {
			return thisActor, nil
		}
	}
	return
}

// GetCategoryByName gets category from the list by Name
//
func (storage *Storage) GetCategoryByName(categoryName string) (category mpclibrary.Category, err error) {
	for _, thisCategory := range storage.Categories {
		if categoryName == thisCategory.Name {
			return thisCategory, nil
		}
	}
	return
}

// GetCategoryByName gets category from the list by ID
//
func (storage *Storage) GetCategoryByID(categoryID int) (category mpclibrary.Category, err error) {
	for id, thisCategory := range storage.Categories {
		if categoryID == id {
			return thisCategory, nil
		}
	}
	return
}

// GetSettings gets settings from the DB
//
func (storage *Storage) GetSettings() (settings mpclibrary.Settings, err error) {
	err = storage.Db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("settings"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			key := string(k)
			if key == "library" {
				settings.LibraryPath = string(v)
			} else if key == "cluster" {
				settings.ClusterID = string(v)
			} else if key == "instance" {
				settings.InstanceID = string(v)
			} else if key == "HMACkey" {
				settings.HMACkey = v
			}
		}

		return nil
	})
	return
}

// Save Settings
//
func (storage *Storage) SaveSettings(settings mpclibrary.Settings) (err error) {
	err = storage.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("settings"))

		if settings.LibraryPath != "" {
			err = bucket.Put([]byte("library"), []byte(settings.LibraryPath))

			if err != nil {
				return err
			}
		}

		err = bucket.Put([]byte("HMACkey"), []byte(settings.HMACkey))

		if err != nil {
			return err
		}
		return err
	})

	return err
}

// Get Users get user list from the DB
//
func (storage *Storage) GetUsers() (users []mpcusers.User, err error) {
	err = storage.Db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("users"))

		c := b.Cursor()

		var user mpcusers.User

		for k, v := c.First(); k != nil; k, v = c.Next() {

			json.Unmarshal(v, &user)

			users = append(users, user)
		}

		return nil
	})
	return
}

// InsertUser inserts new user in the DB.
//
func (storage *Storage) InsertUser(user mpcusers.User) (id string, err error) {

	tx, err := storage.Db.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	usersBucket := tx.Bucket([]byte("users"))

	if user.Email == "" {
		return id, errors.New("user_email_empty")
	}

	if !govalidator.IsEmail(user.Email) {
		return id, errors.New("user_email_invalid")
	}

	if user.Name == "" {
		return id, errors.New("user_name_empty")
	}

	if len(user.Password) < 6 {
		return id, errors.New("user_short_password")
	}

	users, _ := storage.GetUsers()

	for _, dbUser := range users {
		if dbUser.Email == user.Email {
			return id, errors.New("user_email_exists")
		}
	}

	userID := uuid.New()

	user.UUID = userID.String()

	user.Password = HashPassword(user.Password)

	jsonUser, err := json.Marshal(user)
	if err != nil {
		return id, err
	}

	err = usersBucket.Put([]byte(user.UUID), jsonUser)

	if err != nil {
		return id, err
	}

	if err := tx.Commit(); err != nil {
		return user.UUID, err
	}

	return
}

//HashPassword generates sha256 has of password
func HashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	sum := h.Sum(nil)
	return fmt.Sprintf("%x", sum)
}

// itob converts integer to byte
//
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// btoi converts byte to int
//
func btoi(v []byte) int {
	i := int(binary.BigEndian.Uint64(v[:]))
	return i
}
