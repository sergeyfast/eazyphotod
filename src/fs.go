package main

import (
	"github.com/rwcarlsen/goexif/exif"
    "github.com/disintegration/imaging"
    "io/ioutil"
	"model"
	"os"
	"strings"
    "image"
    "log"
    "path/filepath"
    "time"
)

// exists returns whether the given file or directory exists or not
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func ReadExif(filename string) (x *exif.Exif, json string, err error) {
	var f *os.File
	f, err = os.Open(filename)
	if err != nil {
		return
	}
    defer f.Close()

	x, err = exif.Decode(f)
	if err != nil {
		return
	}

	jsonb, _ := x.MarshalJSON()
	json = string(jsonb)

	return
}

func WatchDir( dir string ) error {
    if err := Watcher.Watch(dir); err != nil {
        return err
    }

    return nil
}


func ReadPhoto( filename string  ) (*model.Photo, error) {
    f, err := os.Stat( filename )
    if err != nil {
        return nil, err
    }

    p := &model.Photo{
        OriginalName: f.Name(),
        FileSize: int(f.Size()),
    }

    return p, nil
}


// Get All .jpg Files from Directory
func ReadPhotos(dir string) (result model.PhotoList, err error) {
	var files []os.FileInfo
	if files, err = ioutil.ReadDir(dir); err != nil {
		return
	}

	for _, f := range files {
		if !checkJpgExt( f.Name() ) {
			continue
		}

		p := model.Photo{
			OriginalName: f.Name(),
			FileSize: int(f.Size()),
		}

		result = append(result, &p)
	}

	return
}

// Rewrite Image to disk
func rewriteImage( dst *image.NRGBA, filename string ) (int, error) {
    if r, _ := Exists( filename ); r  {
        os.Remove( filename )
    }

    // save the image to file
    if err := imaging.Save(dst, filename); err != nil {
        return 0, err
    }

    f, err := os.Stat( filename )
    if err != nil {
        return 0, err
    }

    return int(f.Size()), err
}


func WatcherLoop() {
    log.Println( "Starting watcher loop.")
    for {
        select {
        case ev := <-Watcher.Event:
            if ev.IsCreate() || ev.IsRename() {
                time.Sleep( time.Second ) // we need this sleep because will be another events after CREATE
                si := &SyncItem{
                    Filename: ev.Name,
                }

                si.GoSync()
            }
        case err := <-Watcher.Error:
            log.Println( err )
        }
    }
}


// Checks if filename has .jpg or .JPG extension
func checkJpgExt( filename string ) bool {
    return strings.ToLower( filepath.Ext( filename ) ) == ".jpg"
}

// Replace \ to / and add trailing /
func baseDir( filename string ) string {
    return filepath.Clean( filepath.Dir( filename ) )
}
