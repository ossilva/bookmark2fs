package prompt

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/ossilva/bookmark2fs/pkg/configuration"
	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
	"github.com/ossilva/bookmark2fs/pkg/conversion/htmlconv"
	"github.com/ossilva/bookmark2fs/pkg/fstree"
	"github.com/ossilva/bookmark2fs/pkg/util"
)

//Init prompts user for input
func Init(config *configuration.Bm2fsConfig, bmRoots map[string]*base.BookmarkNodeBase, tracker *util.BookmarkTracker) {
	fmt.Println("******************************************")

	var tmpRootPathMap map[string]string
	var exportRoots map[string]*base.BookmarkNodeBase
	var tmpDirPath string

	for {
		prompt := promptui.Select{
			Label: fmt.Sprintf(
				"Select operation [temporary dir: %s]",
				config.TmpRoot,
			),
			Items: []string{
				"(re)populate filesystem tree",
				"export to browser HTML",
				"change root directory",
				"create/save to sqlite",
				"show changes",
				"EXIT",
			},
		}
		_, result, _ := prompt.Run()

		if result == "change root directory" {
			var tryPath string
			for {
				finfo, err := os.Stat(tryPath)
				if err == nil {
					if finfo.IsDir() {
						break
					}
				}

				if tryPath != "" {
					fmt.Println("Error: specify existing directory")
				}

				prompt := promptui.Prompt{
					Label: "Specify temporary directory",
				}
				tryPath, _ = prompt.Run()
			}
			config.TmpRoot = tryPath
		}
		if result == "populate filesystem tree" {
			tmpDirPath, tmpRootPathMap = fstree.PopulateTmpDir(bmRoots, tracker, config.TmpRoot)
			defer os.RemoveAll(tmpDirPath)
		}
		if result == "export to browser HTML" {
			exportRoots = fstree.CollectFSTrees(tmpRootPathMap, tracker)
			htmlconv.BuildTreeHTML(exportRoots, config.OutputFile)
			// fmt.Println(exportRoots)
			// fmt.Println(config.OutputFile)
		}
		if result == "show changes" {
		}
		if result == "export to browser HTML" {

		}
		if result == "EXIT" {
			return
		}
	}

}
