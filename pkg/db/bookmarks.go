package db

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/Masterminds/structable"
)

type RecordableNode struct {
	structable.Recorder
	builder      squirrel.StatementBuilderType
	ID           string `stbl:"id,PRIMARY_KEY,SERIAL"`
	ParentId     string `stbl:"parent_id"`
	DateCreated  int64  `stbl:"date_created"`
	DateModified int64  `stbl:"date_modified"`
	Name         string `stbl:"name"`
	Type         string `stbl:"type"`
	URL          string `stbl:"url"`
}

func BmInsertToTable(bookmark *RecordableNode, db squirrel.DBProxyBeginner) error {
	bookmark.Recorder = structable.New(db, DbFlavor).Bind(
		BookmarkTable, bookmark,
	)
	return bookmark.Insert()
}

const BookmarkTable = "bookmarks"
const DbFlavor = "sqlite"

func OpenDBCache() squirrel.DBProxyBeginner {
	con, _ := sql.Open(DbFlavor, "dbname=bookmark2fs sslmode=disable")
	return squirrel.NewStmtCacheProxy(con)
}
