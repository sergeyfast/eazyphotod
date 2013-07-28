package main

import (
	"errors"
	"github.com/disintegration/imaging"
	"github.com/howeyc/fsnotify"
	"image"
	"log"
	"model"
	"time"
    "strconv"
)

var (
	SyncQueue  = make(chan *SyncItem)      // Main Sync Queue Chan
	AlbumQueue = make(chan *AlbumItem, 50) // Status or Meta update for Album
	Watcher    *fsnotify.Watcher
)

// Album Item
type AlbumItem struct {
	AlbumId      int
	StatusUpdate bool
	MetaUpdate   bool
}

// Main sync item for Dispatch Loop. Unified access to shared resources
type SyncItem struct {
	Album    *model.Album
	FsPhotos model.PhotoList
	FullSync bool
	Filename string
}

// Send self to SyncQueue
func (si *SyncItem) GoSync() {
	SyncQueue <- si
}

// Creates New Fs Sync Item
func NewFsSyncItem(a *model.Album) (si *SyncItem, err error) {
	var photos model.PhotoList
	if photos, err = ReadPhotos(a.PathSource()); err != nil {
		return
	}

	return &SyncItem{
		Album:    a,
		FsPhotos: photos,
		FullSync: true,
	}, nil
}

// Run Full sync from filesystem for all albums
func RunFsSync(albums model.AlbumList) {
	log.Println("FsSync started.")
	for _, a := range albums {
		if r, err := a.CreateDirs(); !r || err != nil {
            log.Println( "a.CreateDirs() failed")
            logFatal( err )
        }

        if err := WatchDir( a.PathSource() ); err != nil {
            log.Println( err )
        }

		si, _ := NewFsSyncItem(a)
		si.GoSync()
	}
	log.Println("FsSync finished.")
}

// Sync Album with photos.
// Can be partial or full
func syncAlbum(si *SyncItem) {
	if si.FullSync {
		log.Printf("Going full fs sync for album %s\n", si.Album.Alias)
	} else {
		log.Printf("Going partial sync for album %s\n", si.Album.Alias)
	}

	index := make(map[string]*model.Photo)
	for _, p := range si.Album.Photos {
		index[p.OriginalName] = p
	}

	var exists, nextId int
	var newPhotos model.PhotoList

	if si.FullSync {
		nextId = si.Album.Photos.MaxPhotoId() + 1
	} else {
		nextId, _ = si.Album.MaxFilename()
		nextId++
	}

	// critical section???
	for _, p := range si.FsPhotos {
		_, ok := index[p.OriginalName]
		if !ok {
			log.Println("Found new photo " + p.OriginalName)
			FillPhoto(si.Album, p, nextId)
			if err := CreatePhotos(si.Album, p); err != nil {
				log.Println(err)
				continue
			}

			newPhotos = append(newPhotos, p)
			nextId++
		} else {
			exists++
		}
	} //eof

	if len(newPhotos) > 0 {
		tx, err := model.DB().Begin()
		if err != nil {
			log.Println(err)
		}
		if err := model.AddPhotos(newPhotos); err != nil {
			log.Println(err)
			if err = tx.Rollback(); err != nil {
				log.Println(err)
			}
		} else {
			// Update Meta
			if si.Album.MetaInfo, err = model.AlbumMeta(si.Album.AlbumId); err != nil {
				log.Println(err)
			} else if err := model.UpdateMeta(si.Album); err != nil {
				log.Println(err)
			}

			// Commit transaction
			if err = tx.Commit(); err != nil {
				log.Println(err)
			}
		}
	}

	log.Printf("Sync Done. %d new, %d exists\n", len(newPhotos), exists)
}

func NewSyncItemPhoto(filename string) (*SyncItem, error) {
	if !checkJpgExt(filename) {
		return nil, errors.New("Not a jpeg: " + filename)
	}

	a := model.FindAlbumByDir(baseDir(filename), Albums)
	if a == nil {
		return nil, errors.New("Failed to find album for photo " + filename)
	}

	p, err := ReadPhoto(filename)
	if err != nil {
		return nil, err
	}

	p.AlbumId = a.AlbumId
	if exists, err := model.HasPhotoByName(p.OriginalName, p.AlbumId); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("Already exists in DB: " + filename)
	}

	si := &SyncItem{
		Album:    a,
		FsPhotos: model.PhotoList{p},
		FullSync: false,
	}

	return si, nil
}

// Fill Photo with AlbumId, Filename, CreatedAt, Sizes and resize images
func FillPhoto(a *model.Album, p *model.Photo, photoNum int) {
	p.Filename = model.PhotoName(photoNum)
	p.CreatedAt = time.Now()
	p.AlbumId = a.AlbumId

	x, json, _ := ReadExif(a.PathSource() + p.OriginalName)
	p.EXIF = json

	if x != nil {
		if date, err := x.Get("DateTimeOriginal"); err == nil {
			p.PhotoDate, err = time.Parse("2006:01:02 15:04:05", date.StringVal())
		}
	}
}

// Create HD and Thumbs Photos
func CreatePhotos(a *model.Album, p *model.Photo) (err error) {
	src, err := imaging.Open(a.PathSource() + p.OriginalName)
	if err != nil {
		return err
	}

	var dst *image.NRGBA
	filename := a.PathHD() + p.Filename
	log.Printf("Saving HD: %s\n", filename)
	dst = imaging.Fit(src, cfg.Image.MaxWidth, cfg.Image.MaxHeight, imaging.Lanczos)
	size, err := rewriteImage(dst, filename)
	if err != nil {
		return err
	}

	p.FileSizeHD = size
	filename = a.PathThumbs() + p.Filename
	log.Printf("Saving Thumb: %s\n", filename)
	dst = imaging.Thumbnail(dst, cfg.Image.ThumbWidth, cfg.Image.ThumbHeight, imaging.CatmullRom) // resize and crop the image to make a 200x200 thumbnail
	_, err = rewriteImage(dst, filename)

	return err
}

// Main loop for dispatching SyncQueue
// TODO exit
func Dispatch() {
	var err error
	for {
		select {
		case si := <-SyncQueue:
			if !si.FullSync {
				si, err = NewSyncItemPhoto(si.Filename)
			}

			if err != nil {
				log.Println(err)
			} else {
				syncAlbum(si)
			}
		case ai := <-AlbumQueue:
			switch {
            case ai.MetaUpdate:
                log.Println( "Updating Album meta")
				if err = updateMeta(ai.AlbumId); err != nil {
					log.Println(err)
				} else {
                    log.Println( "Metainfo was updated for albumId", ai.AlbumId)
                }
            case ai.StatusUpdate:
                log.Println( "Reloading albums")
                if err = updateAlbums(); err != nil {
                    log.Println(err)
                } else {
                    log.Println( "Albums were reloaded")
                }
			}
		}
	}

}

// Update Albums Meta by Id
func updateMeta(albumId int) error {
	a, ok := Albums[albumId]
	if !ok {
		return errors.New("Unknown AlbumId for meta update: " + strconv.Itoa(albumId) )
	}

	var err error
	if a.MetaInfo, err = model.AlbumMeta(a.AlbumId); err != nil {
		return err
	} else if err := model.UpdateMeta(a); err != nil {
		return err
	}

	return nil
}


// Reload New Albums
func updateAlbums() error {
    albums, err := model.Albums()
    if err != nil {
        return err
    }


    // Filter New albums
    newAlbums := make(model.AlbumList)
    for _, a := range albums {
        if a.StatusId == model.StatusInQueue {
            newAlbums[a.AlbumId] = a
        }
    }

    if len(newAlbums) == 0 {
        return nil
    }

    // Queue Albums
    if wds, err := model.PushToBtSync( newAlbums, bstClient ); err != nil {
        return err
    } else {
        for _, d := range wds {
            if err = WatchDir( d ); err != nil {
                return err
            }
        }
    }

    // Update Albums Storage
    for _, a := range newAlbums {
        if curAl, ok := Albums[a.AlbumId]; ok {
            a.Photos = curAl.Photos
        }

        Albums[a.AlbumId] = a
    }

    return nil
}
