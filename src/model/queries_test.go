package model

import (
	"testing"
	"time"
)

func init() {
	ConnectionString = "postgres://sergeyfast@localhost/eazyphoto?sslmode=disable"
}


func TestOpen(t *testing.T) {
	if err := Open(); err != nil {
		t.Fatalf("failed to open connection, got error %v", err)
	}
}

func TestAlbums(t *testing.T) {
	r, err := Albums()
	if err != nil {
		t.Error(err)
	}
	t.Logf("albums in db: %v\n", len(r))
}

func TestPhotos(t *testing.T) {
	r, err := Photos()
	if err != nil {
		t.Error(err)
	}
	t.Logf("photos in db: %v\n", len(r))

	if len(r) == 0 {
		return
	}

	tx, err := DB().Begin();
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()


	l := PhotoList{ &Photo{
		AlbumId: r[0].AlbumId,
		OriginalName: r[0].OriginalName,
		Filename: r[0].Filename,
		FileSize: 0,
		FileSizeHD: 0,
		EXIF: "" ,
		CreatedAt: time.Now(),
		PhotoDate: time.Now(),
	}}

	err = AddPhotos(l)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("last inserted photo id is %v\n", l[0].PhotoId)
}