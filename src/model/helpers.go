package model

import (
	"errors"
	"fmt"
	"github.com/sergeyfast/btsync"
	"path/filepath"
	"strconv"
	"strings"
)

func JoinPhotos(albums AlbumList, photos PhotoList) {
	for _, p := range photos {
		a, ok := albums[p.AlbumId]
		if !ok {
			continue
		}

		a.Photos = append(a.Photos, p)
	}
}

func PhotoName(number int) string {
	return fmt.Sprintf("%04d.jpg", number)
}

func FindAlbumByDir(dir string, albums AlbumList) *Album {
	for _, a := range albums {
		if dir == filepath.Clean(a.PathSource()) {
			return a
		}
	}

	return nil
}

func PushToBtSync(albums AlbumList, client *btsync.Client) ([]string, error) {
	var watchDirs []string
	var err error

	for _, a := range albums {
		if a.StatusId != StatusInQueue {
			continue
		}

		// Create Folders
		if _, err := a.CreateDirs(); err != nil {
			return watchDirs, err
		}

		// Generate Secret
		var s *btsync.Secrets
		if s, err = client.Secrets("", false); err != nil {
			return watchDirs, err
		}

		// Add Sources
		if r, err := client.AddFolder(a.PathSource(), a.ROSecret, 0); err != nil || r.Error != 0 || r.Result != 0 {
			if err == nil {
				err = errors.New("Failed to add source sync folder. Error code: " + strconv.Itoa(r.Error) + ". " + r.Message)
			}

			if !IgnoreBTSyncResult {
				return watchDirs, err
			}
		}

		// Add HD
		a.ROSecretHD = s.ReadOnly
		if r, err := client.AddFolder(a.PathHD(), s.ReadWrite, 0); err != nil || r.Error != 0 || r.Result != 0 {
			if err == nil {
				err = errors.New("Failed to add hd sync folder. Error code: " + strconv.Itoa(r.Error) + ". " + r.Message)
			}

			if !IgnoreBTSyncResult {
				return watchDirs, err
			}
		}

		a.StatusId = StatusEnabled
		if err = UpdateStatus(a); err != nil {
			return watchDirs, err
		}

		watchDirs = append(watchDirs, a.PathSource())
	}

	return watchDirs, nil
}

func intFromFilename(filename string) int {
	sId := strings.Replace(filename, ".jpg", "", 1)
	tId, _ := strconv.Atoi(sId)

	return tId
}
