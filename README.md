eazyphotod
==========

Daemon for EazyPhoto

Requirements
------
Go v1.1


How-to build
------
	git clone git://github.com/sergeyfast/eazyphotod.git
	cd eazyphotod
    export GOPATH=`pwd`
    go get code.google.com/p/gcfg github.com/disintegration/imaging github.com/go-sql-driver/mysql github.com/howeyc/fsnotify github.com/rwcarlsen/goexif/exif github.com/sergeyfast/btsync-cli/src/btsync
	go build -o eazyphotod src/*.go
