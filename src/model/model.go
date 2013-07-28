package model

import (
    "fmt"
    "time"
    "encoding/json"
    "os"
)

const (
    StatusEnabled = 1
    StatusInQueue = 4
)

var (
    AlbumsRoot string
    Source = "source"
    Thumbs = "thubms"
    HD     = "hd"
    IgnoreBTSyncResult = false
)

type AlbumList map[int]*Album

type Album struct {
    AlbumId    int
    Alias      string
    FolderPath string
    ROSecret   string
    ROSecretHD string
    StatusId   int
    StartDate  time.Time
    Photos     PhotoList
    MetaInfo   MetaInfo
}

func (a *Album) Path(t string) string {
    return fmt.Sprintf("%s/%d/%s/%s/", AlbumsRoot, a.StartDate.Year(), a.FolderPath, t)
}

func (a *Album) PathSource() string {
    return a.Path(Source)
}

func (a *Album) PathHD() string {
    return a.Path(HD)
}

func (a *Album) PathThumbs() string {
    return a.Path(Thumbs)
}

// Create Dirs for Source, HD and Thumbs
func ( a *Album ) CreateDirs()  (bool, error) {
    err := os.MkdirAll( a.PathSource(), os.ModePerm )
    if err != nil {
        return false, err
    }

    err = os.MkdirAll( a.PathHD(), os.ModePerm )
    if err != nil {
        return false, err
    }

    err = os.MkdirAll( a.PathThumbs(), os.ModePerm )
    if err != nil {
        return false, err
    }

    return true, nil
}



type PhotoList []*Photo

type Photo struct {
    PhotoId      int
    AlbumId      int
    OriginalName string
    Filename     string
    FileSize     int
    FileSizeHD   int
    OrderNumber  *int
    EXIF         string
    CreatedAt    time.Time
    PhotoDate    time.Time
    StatusId     int
    Photos       PhotoList
}

// Get Max Photo Id from PhotoList
func (photos PhotoList) MaxPhotoId() int {
    id := 0

    for _, p := range photos {
        tId := intFromFilename(p.Filename)
        if tId > id {
            id = tId
        }
    }

    return id
}

// Album MetaInfo (stored in json)
type MetaInfo struct {
    Count    int   `json:"count"`
    Size     int   `json:"size"`
    SizeHD   int   `json:"sizeHd"`
    PhotoIds []string `json:"photoIds"`
}

func (m MetaInfo) Json() string {
    b, _ := json.Marshal(m)
    return string(b)
}


