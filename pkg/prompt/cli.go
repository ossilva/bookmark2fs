// package prompt

// import (
// 	"fmt"
// 	"os"

// 	"github.com/manifoldco/promptui"
// 	"github.com/ossilva/bookmark2fs/pkg/configuration"
// 	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
// 	"github.com/ossilva/bookmark2fs/pkg/conversion/htmlconv"
// 	"github.com/ossilva/bookmark2fs/pkg/db"
// 	"github.com/ossilva/bookmark2fs/pkg/fstree"
// 	"github.com/ossilva/bookmark2fs/pkg/util"
// )

// //Init prompts user for input
// func Init(config *configuration.Bm2fsConfig, bmRoots []*base.BookmarkNodeBase, tracker *util.BookmarkTracker) {
// 	fmt.Println("******************************************")

// 	// var tmpRootPaths []string
// 	var exportRoots []*base.BookmarkNodeBase
// 	var tmpDirPath string

// 	for {
// 		prompt := promptui.Select{
// 			Label: fmt.Sprintf(
// 				"Select operation [temporary dir: %s]",
// 				config.TmpRoot,
// 			),
// 			Items: []string{
// 				"populate filesystem tree",
// 				"export to browser HTML",
// 				"change root directory",
// 				"create/save sqlite",
// 				// "show changes",
// 				"EXIT",
// 			},
// 		}
// 		_, result, _ := prompt.Run()

// 		if result == "change root directory" {
// 			var tryPath string
// 			for {
// 				finfo, err := os.Stat(tryPath)
// 				if err == nil {
// 					if finfo.IsDir() {
// 						break
// 					}
// 				}

// 				if tryPath != "" {
// 					fmt.Println("Error: specify existing directory")
// 				}

// 				prompt := promptui.Prompt{
// 					Label: "Specify temporary directory",
// 				}
// 				tryPath, _ = prompt.Run()
// 			}
// 			config.TmpRoot = tryPath
// 		} else if result == "populate filesystem tree" {
// 			tmpDirPath, _ = fstree.PopulateTmpDir(bmRoots, tracker, config.TmpRoot)
// 			defer os.RemoveAll(tmpDirPath)
// 		} else if result == "export to browser HTML" {
// 			exportRoots = fstree.CollectFSTrees(tmpDirPath, tracker)
// 			htmlconv.BuildTreeHTML(exportRoots, config.OutputFile)
// 		} else if result == "show changes" {
// 			// TODO varies according to bookmark array source
// 		} else if result == "create/save sqlite" {
// 			var rootsToSave []*base.BookmarkNodeBase
// 			if exportRoots != nil {
// 				rootsToSave = exportRoots
// 			} else if bmRoots != nil {
// 				rootsToSave = bmRoots
// 			}
// 			cache := db.OpenDBCache()
// 			for _, root := range rootsToSave {
// 				for _, node := range base.GetNodesBFS(root) {
// 					rec := node.ToRecordable()
// 					db.BmInsertToTable(rec, cache)
// 				}
// 			}
// 		} else if result == "EXIT" {
// 			return
// 		}
// 	}

// }
