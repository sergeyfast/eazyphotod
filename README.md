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
    go get -fix eazyphotod model
    go build -o eazyphotod src/eazyphotod.go
