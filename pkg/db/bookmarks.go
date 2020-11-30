package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Masterminds/squirrel"
	sq "github.com/Masterminds/squirrel"
	"github.com/Masterminds/structable"
	_ "github.com/mattn/go-sqlite3" // need sqlite3 driver
	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
)

const BookmarkTable = "bookmarks"
const driver = "sqlite3"

func BmInsertToTable(bookmark *base.RecordableNode, db squirrel.DBProxyBeginner) error {
	bookmark.Recorder = structable.New(db, driver).Bind(
		BookmarkTable, bookmark,
	)
	return bookmark.Insert()
}

func OpenDBCache() (squirrel.DBProxyBeginner, *sql.DB) {
	con, err := sql.Open(driver, "bookmark2fs.sqlite")
	// con, err := sql.Open(driver, ":memory:")
	if err != nil {
		fmt.Printf("Couldn't Open database: %s\n", err)
	}

	stmt := `
	CREATE TABLE bookmarks (
	id BIGINT PRIMARY_KEY SERIAL,
	uuid STRING,
	parent_uuid STRING,
	date_created INTEGER,
	date_modified INTEGER,
	name STRING,
	type STRING,
	url STRING
	);
`
	_, err = con.Exec(stmt)
	if err != nil {
		fmt.Printf("Couldn't Exec statement \"%q\": %s\n", err, stmt)
		return nil, nil
	}

	return squirrel.NewStmtCacheProxy(con), con
}

func PrintDuplicates(cache *sq.DBProxyBeginner) {

	bms := sq.Select(
		`
	SELECT
	  name,
		date_created,
		date_modified,
		url
		FROM bookmarks
		GROUP BY 
	  name,
		date_created,
		url
		HAVING COUNT(id) >1; 	
		`,
	).From("bookmarks")
	sql, _, _ := bms.ToSql()

	rows, err := (*cache).Query(sql)
	if err != nil {
		// handle error
	}
	for rows.Next() {
		var name, url string
		var date_created, date_modified int
		if err := rows.Scan(
			&date_created,
			&date_modified, &name, &url); err != nil {
			log.Fatal(err)
		}
		fmt.Println(name, url, date_created, date_modified)
	}
}

// BackupNodeRoots Creates a database table and inserts bookmarks
// returns a database proxy
func BackupNodeRoots(roots []*base.BookmarkNodeBase) *sq.DBProxyBeginner {
	cache, _ := OpenDBCache()
	for _, root := range roots {
		for _, node := range base.GetNodesBFS(root) {
			rec := node.ToRecordable()
			if err := BmInsertToTable(rec, cache); err != nil {
				fmt.Println(err.Error())
			}
		}
	}
	return &cache
}
