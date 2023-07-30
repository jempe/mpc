package main

import (
	"embed"
	"flag"
	"github.com/jempe/mpc/auth"
	"github.com/jempe/mpc/library"
	"github.com/jempe/mpc/remote"
	"github.com/jempe/mpc/server"
	"github.com/jempe/mpc/storage"
	"github.com/jempe/mpc/users"
	"github.com/jempe/mpc/utils"
	"html/template"
	"log"
	"net/http"
)

var libraryPath = flag.String("path", "", "Define the path of your videos folder")
var mpcConfigPath = flag.String("config", "", "Define the path of config folder")
var storage *mpcstorage.Storage
var port = "3000"
var libPath string
var library *mpclibrary.Library
var indexTemplate *template.Template
var key string = "ThisisAT3stKey123"

//go:embed tmpl/index.html
//go:embed html/js/* html/fonts/* html/css/* html/images/*

var content embed.FS

type IndexPage struct {
	AuthMethod string
}

func main() {
	flag.Parse()

	var configPath string
	if *mpcConfigPath == "" {
		configPath = mpcutils.ConfigFolder()
	} else {
		configPath = *mpcConfigPath
	}

	// Initialize BoltDB
	storage = &mpcstorage.Storage{Path: configPath}

	err := storage.InitDb()
	mpcutils.CheckErr(err)

	// Load all videos
	err = storage.GetAllVideos()
	mpcutils.CheckErr(err)
	settings, err := storage.GetSettings()
	mpcutils.CheckErr(err)

	if settings.LibraryPath == "" && *libraryPath != "" {
		settings.LibraryPath = *libraryPath
		err = storage.SaveSettings(settings)
		mpcutils.CheckErr(err)

		settings.LibraryPath = *libraryPath
	}

	if settings.HMACkey == nil {
		authkey, err := mpcauth.RandomBytes(256)

		if err == nil {
			settings.HMACkey = authkey
		} else {
			mpcutils.CheckErr(err)

		}

		err = storage.SaveSettings(settings)
		mpcutils.CheckErr(err)
	}

	library = &mpclibrary.Library{Path: settings.LibraryPath, Settings: settings}

	//Init auth library
	auth := &mpcauth.Auth{Key: settings.HMACkey, Storage: storage}

	id, err := storage.InsertUser(mpcusers.User{Name: "Admin", Email: "test@jempe.org", Password: "test1234"})
	log.Println("uuid:", id)

	// load and parse index page template
	//paths := []string{"tmpl/index.html"}
	indexTemplate = template.Must(template.ParseFS(content, "tmpl/index.html"))

	localIP := mpcutils.GetLocalIP()

	server := &mpcserver.Server{IP: localIP, Storage: storage, Library: library, Key: key, Auth: auth}

	http.HandleFunc("/", homeHandler)
	http.Handle("/html/", http.FileServer(http.FS(content)))
	//http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.FS(content))))
	//http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.FS(content))))
	//http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.FS(content))))
	http.HandleFunc("/actors.json", server.ActorsHandler)
	http.HandleFunc("/videos.json", server.VideosHandler)
	http.HandleFunc("/videos/", server.VideoFileHandler)
	http.HandleFunc("/scan/", server.ScanHandler)
	http.HandleFunc("/videos/thumbs/", server.ThumbsHandler)
	http.HandleFunc("/videos/screenshots/", server.ScreenshotsHandler)

	http.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(http.Dir("html/admin"))))
	http.HandleFunc("/login", server.LoginHandler)
	http.HandleFunc("/isloggedin", server.IsLoggedIn)

	remote := mpcremote.NewRemote()

	http.Handle("/remote", remote)

	go remote.Run()

	log.Println("MPC server running on", "http://"+localIP+":"+port)
	panic(http.ListenAndServe(":"+port, nil))
}

// Handle index page server requests
//
func homeHandler(w http.ResponseWriter, r *http.Request) {
	indexTemplate.Execute(w, nil)
}
