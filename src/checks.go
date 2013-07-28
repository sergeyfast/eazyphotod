package main

import (
    "github.com/sergeyfast/btsync-cli/src/btsync"
    "github.com/howeyc/fsnotify"
    "errors"
    "fmt"
	"log"
	"model"
)

func pingModel() error {
	model.ConnectionString = cfg.Database.ConnectionString
	return model.Open()
}

// Log Fatal
func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func setRootDir() {
	if ok, err := Exists(cfg.Photos.Root); !ok {
		if err != nil {
			log.Print(err)
		}

		log.Fatalf("Directory %s doen't exists", cfg.Photos.Root)
	}

	model.AlbumsRoot = cfg.Photos.Root
	model.Source = cfg.Photos.Source
	model.Thumbs = cfg.Photos.Thumbs
	model.HD = cfg.Photos.HD


    var err error
    Watcher, err = fsnotify.NewWatcher()
    if err != nil {
        logFatal(err)
    }
}

func pingBtSync() error {
	bstClient = btsync.NewClient(cfg.BTSync.Host, cfg.BTSync.Port, cfg.BTSync.User, cfg.BTSync.Password)
	if !bstClient.RequestToken() {
		return errors.New("Invalid Credentials for BTSync")
	}

	return nil
}

func VarDump(vals interface{}) {
	fmt.Printf("%V\n", vals)
}
