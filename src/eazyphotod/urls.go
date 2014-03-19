package eazyphotod

import (
	"fmt"
	"net/http"
	"strconv"
)

// Handles /update/album-meta?id=<album-id>
func updateMetaHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	result := "Queued"
	if id == 0 {
		result = "Invalid albumId"
	} else {
		ai := &AlbumItem{
			AlbumId:    id,
			MetaUpdate: true,
		}

		AlbumQueue <- ai
	}

	fmt.Fprintln(w, result)
}

// Handles /update/albums
func updateAlbumsHandler(w http.ResponseWriter, r *http.Request) {
	ai := &AlbumItem{
		StatusUpdate: true,
	}

	AlbumQueue <- ai

	fmt.Fprintln(w, "Queued")
}
