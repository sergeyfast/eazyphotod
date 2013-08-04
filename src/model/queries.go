package model

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
)

var (
	ConnectionString string
	db               *sql.DB
)

func DB() *sql.DB {
	return db
}

// Opens and Ping Database
func Open() (err error) {
	db, err = sql.Open("mysql", ConnectionString)
	if err != nil {
		return err
	}
	return db.Ping()
}

// Close Database
func Close() error {
	return db.Close()
}

// Get All Albums from DB
func Albums() (result AlbumList, err error) {
	rows, err := db.Query(`SELECT albumId, alias, folderPath, roSecret, roSecretHD, startDate, statusId
	    FROM albums where statusId IN (1,4)
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
			log.Println(err)
			continue
		}

		result[a.AlbumId] = &a
	}

	return
}

// Get All Photos from DB
func Photos() (result PhotoList, err error) {
	rows, err := db.Query(`
        SELECT photoId, albumId, originalName, filename, statusId
        FROM photos
    `)

	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var p Photo
		err := rows.Scan(&p.PhotoId, &p.AlbumId, &p.OriginalName, &p.Filename, &p.StatusId)
		if err != nil {
			log.Println(err)
			continue
		}

		result = append(result, &p)
	}

	return
}

// Update Album Meta
func UpdateMeta(a *Album) error {
	sql := "UPDATE `albums` set `metaInfo` = ? WHERE `albumId` = ?;"
	_, err := db.Exec(sql, a.MetaInfo.Json(), a.AlbumId)
	if err != nil {
		return err
	}

	return nil
}

// Update Album Status and ROSecretHD
func UpdateStatus(a *Album) error {
	sql := "UPDATE `albums` set `statusId` = ?, `roSecretHd` = ? WHERE `albumId` = ?;"
	_, err := db.Exec(sql, a.StatusId, a.ROSecretHD, a.AlbumId)
	if err != nil {
		return err
	}

	return nil
}

// Add Batch Photos
func AddPhotos(photos PhotoList) error {
	sql := "INSERT INTO  `photos`" +
		"(  `albumId`,  `originalName`,  `filename`,  `fileSize`,  `fileSizeHd`,  `exif`,  `createdAt`,  `photoDate`,  `statusId`)" +
		"VALUE ( ?, ?, ?, ?, ?, ?, ?, ?, 1 );"

	stmtIns, err := db.Prepare(sql) // ? = placeholder
	if err != nil {
		return err
	}
	defer stmtIns.Close()

	var resultErr error
	for _, p := range photos {
		r, err := stmtIns.Exec(p.AlbumId, p.OriginalName, p.Filename, p.FileSize, p.FileSizeHD, p.EXIF, p.CreatedAt, p.PhotoDate)
		if err != nil {
			resultErr = err
		}
		if id, err := r.LastInsertId(); err == nil {
			p.PhotoId = int(id)
		} else {
			resultErr = err
		}
	}

	return resultErr
}

func AlbumMeta(id int) (MetaInfo, error) {
	var m MetaInfo
	sql := "SELECT sum(`fileSize`) as fs, sum(`fileSizeHd`) as fh, " +
		"  count( case when `statusId` = 1 then `photoId` else null end ) as total, " +
        " GROUP_CONCAT( CASE WHEN statusId = 1 then `photoId` ELSE null END ORDER BY `orderNumber`, `photoDate` ) as ids " +
		" FROM `photos` p " +
		" WHERE `albumId` = ?"

	ds, err := db.Query(sql, id)
	if err != nil {
		return m, err
	}

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
	stmt, err := db.Prepare("SELECT `photoId` FROM `photos` WHERE `originalName` = ? and `albumId` = ? ")
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
	stmt, err := db.Prepare("SELECT max(`filename`) FROM `photos` WHERE `albumId` = ? ")
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
