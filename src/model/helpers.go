package model

import (
	"fmt"
    "path/filepath"
    "github.com/sergeyfast/btsync-cli/src/btsync"
    "strings"
    "strconv"
    "errors"
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

func FindAlbumByDir(  dir string, albums AlbumList ) *Album {
    for _, a := range albums {
        if dir == filepath.Clean( a.PathSource() ) {
            return a
        }
    }

    return nil
}

func PushToBtSync( albums AlbumList, client *btsync.Client ) ([]string,error) {
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
        var s btsync.Secret
        if s, err = client.GenerateSecret(); err != nil {
            return watchDirs, err
        }

        // Add Sources
        if r := client.AddSyncFolder( a.PathSource(), a.ROSecret ); r.Err != nil || r.Error != 0 {
            err := r.Err
            if err == nil {
                err = errors.New( "Failed to add source sync folder. Error code: " + strconv.Itoa( r.Error ) +  ". " + r.Message )
            }

            if !IgnoreBTSyncResult {
                return watchDirs, err
            }
        }

        // Add HD
        a.ROSecretHD = s.ROSecret
        if r := client.AddSyncFolder( a.PathHD(), s.Secret ); r.Err != nil || r.Error != 0 {
            err := r.Err
            if err == nil {
                err = errors.New( "Failed to add hd sync folder. Error code: " + strconv.Itoa( r.Error ) + ". " + r.Message )
            }

            if !IgnoreBTSyncResult {
                return watchDirs, err
            }
        }

        a.StatusId = StatusEnabled
        if err = UpdateStatus( a ); err != nil  {
            return watchDirs, err
        }

        watchDirs = append( watchDirs, a.PathSource() )
    }

    return watchDirs, nil
}


func intFromFilename( filename string ) int {
    sId := strings.Replace(filename, ".jpg", "", 1)
    tId, _ := strconv.Atoi(sId)

    return tId
}
