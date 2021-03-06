package model

import (
	"database/sql"
	"github.com/golang/glog"
	_ "github.com/lib/pq"
	"strings"
)

var (
	ConnectionString string
	db               *sql.DB
)

const maxIdleConnections = 5

func DB() *sql.DB {
	return db
}

// Opens and Ping Database
func Open() (err error) {
	db, err = sql.Open("postgres", ConnectionString)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(maxIdleConnections)
	return db.Ping()
}

// Close Database
func Close() error {
	return db.Close()
}

// Get All Albums from DB
func Albums() (result AlbumList, err error) {
	rows, err := db.Query(`SELECT "albumId", "alias", "folderPath", "roSecret", "roSecretHd", "startDate", "statusId"
	    FROM "albums" where "statusId" IN (1,4)
	    `)

	if err != nil {
		return
	}
	defer rows.Close()

	result = make(AlbumList)
	for rows.Next() {
		var a Album
		err := rows.Scan(&a.AlbumId, &a.Alias, &a.FolderPath, &a.ROSecret, &a.ROSecretHD, &a.StartDate, &a.StatusId)
		if err != nil {
			glog.Warningln(err)
			continue
		}

		result[a.AlbumId] = &a
	}

	return
}

// Get All Photos from DB
func Photos() (result PhotoList, err error) {
	rows, err := db.Query(`
        SELECT "photoId", "albumId", "originalName", "filename", "statusId"
        FROM "photos"
    `)

	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var p Photo
		err := rows.Scan(&p.PhotoId, &p.AlbumId, &p.OriginalName, &p.Filename, &p.StatusId)
		if err != nil {
			glog.Warningln(err)
			continue
		}

		result = append(result, &p)
	}

	return
}

// Update Album Meta
func UpdateMeta(a *Album) error {
	sql := `UPDATE "albums" set "metaInfo" = $1 WHERE "albumId" = $2;`
	_, err := db.Exec(sql, a.MetaInfo.Json(), a.AlbumId)
	if err != nil {
		return err
	}

	return nil
}

// Update Album Status and ROSecretHD
func UpdateStatus(a *Album) error {
	sql := `UPDATE "albums" set "statusId" = $1, "roSecretHd" = $2 WHERE "albumId" = $3;`
	_, err := db.Exec(sql, a.StatusId, a.ROSecretHD, a.AlbumId)
	if err != nil {
		return err
	}

	return nil
}

// Add Batch Photos
func AddPhotos(photos PhotoList) error {
	sql := `INSERT INTO  "photos"` +
		`(  "albumId",  "originalName",  "filename",  "fileSize",  "fileSizeHd",  "exif",  "createdAt",  "photoDate",  "statusId")` +
		`VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, 1 ) RETURNING "photoId";`

	stmtIns, err := db.Prepare(sql) // ? = placeholder
	if err != nil {
		return err
	}
	defer stmtIns.Close()

	var resultErr error
	for _, p := range photos {
		var id int
		err := stmtIns.QueryRow(p.AlbumId, p.OriginalName, p.Filename, p.FileSize, p.FileSizeHD, p.EXIF, p.CreatedAt, p.PhotoDate).Scan(&id)
		if err != nil {
			resultErr = err
		} else {
			p.PhotoId = id
		}
	}

	return resultErr
}

func AlbumMeta(id int) (MetaInfo, error) {
	var m MetaInfo
	sql := `SELECT sum("fileSize") as fs, sum("fileSizeHd") as fh, count( case when "statusId" = 1 then "photoId" else null end ) as total,
	string_agg( "photoId"::text, ',' ORDER BY "statusId", "orderNumber", "photoDate" ) as ids
	FROM "photos" p
	WHERE "albumId" = $1`

	ds, err := db.Query(sql, id)
	if err != nil {
		return m, err
	}
	defer ds.Close()

	if ds.Next() {
		var ids string
		err = ds.Scan(&m.Size, &m.SizeHD, &m.Count, &ids)
		if err != nil {
			return m, err
		}

		m.PhotoIds = strings.Split(ids, ",")
		if len(m.PhotoIds) > 6 {
			m.PhotoIds = m.PhotoIds[0:6]
		}
	}

	return m, nil
}

func HasPhotoByName(originalName string, albumId int) (bool, error) {
	stmt, err := db.Prepare(`SELECT "photoId" FROM "photos" WHERE "originalName" = $1 and "albumId" = $2 `)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var r *sql.Rows
	if r, err = stmt.Query(originalName, albumId); err != nil {
		return false, err
	}
	defer r.Close()

	return r.Next(), nil
}

func (a *Album) MaxFilename() (int, error) {
	var id int
	stmt, err := db.Prepare(`SELECT max("filename") FROM "photos" WHERE "albumId" = $1 `)
	if err != nil {
		return id, err
	}
	defer stmt.Close()

	var r *sql.Rows
	if r, err = stmt.Query(a.AlbumId); err != nil {
		return id, err
	}
	defer r.Close()
	if r.Next() {
		var s string
		r.Scan(&s)
		if s != "" {
			id = intFromFilename(s)
		}
	}

	return id, nil
}
