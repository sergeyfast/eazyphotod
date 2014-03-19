package eazyphotod

import (
	"errors"
	"fmt"
	"github.com/howeyc/fsnotify"
	"github.com/sergeyfast/btsync"
	"model"
	"github.com/golang/glog"
)

func pingModel() error {
	model.ConnectionString = cfg.Database.ConnectionString
	return model.Open()
}

// Log Fatal
func logFatal(err error) {
	if err != nil {
		glog.Fatal(err)
	}
}

func setRootDir() {
	if ok, err := Exists(cfg.Photos.Root); !ok {
		if err != nil {
			glog.Info(err)
		}

		glog.Fatalf("Directory %s doen't exists", cfg.Photos.Root)
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
	if _, err := bstClient.Version(); err != nil {
		return errors.New("Cannot get btsync version.")
	}

	return nil
}

func VarDump(vals interface{}) {
	fmt.Printf("%V\n", vals)
}
