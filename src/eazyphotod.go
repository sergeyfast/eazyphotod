// EazyPhoto Daemon
package main

import (
    "code.google.com/p/gcfg"
    "flag"
    "fmt"
    "github.com/sergeyfast/btsync-cli/src/btsync"
    "log"
    "model"
    "net/http"
    _ "expvar"
)

var (
    cfg        Config
    configFile = flag.String("config", "config.gcfg", "Config File in gcfg format")
    bstClient  *btsync.Client
    p          = fmt.Println
    pf         = fmt.Printf
    Albums     model.AlbumList
)

// Config File Structure
type Config struct {
    Photos struct {
    Root   string // Albums Root Path
    Source string // source folder
    HD     string // hd folder
    Thumbs string // thumbs folder
}
    Server struct {
    Host string
    Port int
}
    BTSync struct {
    Host     string
    Port     string
    User     string
    Password string
}
    Image struct {
    MaxWidth      int
    MaxHeight     int
    ThumbWidth    int
    ThumbHeight   int
    AllowOverride bool
}
    Database struct {
    ConnectionString string
}
}

func main() {
    var err error
    flag.Parse()
    err = gcfg.ReadFileInto(&cfg, *configFile)
    logFatal(err)

    setRootDir()
    logFatal(pingBtSync())

    // open db
    logFatal(pingModel())
    defer model.Close()

    // join all photos
    Albums, err = model.Albums() // TODO Albums as Storage
    logFatal(err)
    photos, err := model.Photos()
    logFatal(err)
    model.JoinPhotos(Albums, photos)


    // Queue Albums
    if wds, err := model.PushToBtSync( Albums, bstClient ); err != nil {
        logFatal( err )
    } else {
        for _, d := range wds {
            if err = WatchDir( d ); err != nil {
                logFatal( err )
            }
        }
    }


    // start goroutines
    go RunFsSync(Albums)
    go WatcherLoop()
    go Dispatch()

    // Start Server
    listenAddress := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
    log.Printf("Starting eazyphotod at %s\n", listenAddress)
    http.HandleFunc("/update/album-meta", updateMetaHandler)
    http.HandleFunc("/update/albums", updateAlbumsHandler)
    err = http.ListenAndServe(listenAddress, nil)
    logFatal(err)
}
