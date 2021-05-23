package db_test

import (
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3" // need sqlite3 driver
	"github.com/ossilva/bookmark2fs/cmd"
	"github.com/ossilva/bookmark2fs/pkg/db"
)

func cleanTestDB() {
	err := os.Remove("./bookmark2fs.sqlite")
	if err == nil {
		fmt.Println("removed database from program dir")
	}
}

func TestDB(t *testing.T) {
	cleanTestDB()
	inputFile := "./test/bookmarks_9_25_20.html"
	bookmarkRoots := cmd.ReadInputFile(inputFile)
	cache := db.BackupNodeRoots(bookmarkRoots)
	db.PrintDuplicates(cache)
}
