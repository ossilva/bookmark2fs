package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"github.com/manifoldco/promptui"
	"github.com/ossilva/bookmark2fs/pkg/configuration"
	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
	"github.com/ossilva/bookmark2fs/pkg/conversion/htmlconv"
	"github.com/ossilva/bookmark2fs/pkg/db"
	"github.com/ossilva/bookmark2fs/pkg/fstree"
	"github.com/ossilva/bookmark2fs/pkg/util"
)

var cmdPrompt = &cobra.Command{
	Use:   "prompt",
	Short: "Start in prompt mode",
	Long: `print is for printing anything back to the screen.
For many years people have printed back to the screen.`,
	Run: Init,
}

type optionItem string

var options = map[string]optionItem{
	"populate filesystem tree": "populate filesystem tree",
	"export to browser HTML":   "export to browser HTML",
	"change bookmark store":    "change bookmark store",
	"change root directory":    "change root directory",
	"create/save sqlite":       "create/save sqlite",
	// "show changes": "show changes"// TODO,
	"EXIT": "EXIT",
}

func getOptionSlice() []optionItem {

	var itemSlice []string = []string{
		"populate filesystem tree",
		"export to browser HTML",
		"change bookmark store",
		"change root directory",
		"create/save sqlite",
		"EXIT",
	}
	items := make([]optionItem, len(itemSlice))

	for i, v := range itemSlice {
		items[i] = optionItem(v)
	}
	return items
}

//Init prompts user for input
func Init(cmd *cobra.Command, args []string) {
	fmt.Println("******************************************")

	inFiles := generateSourceList(args)
	var config *configuration.Bm2fsConfig = makeConfig()

	bmStore := inFiles[0]
	var bmRoots []*base.BookmarkNodeBase
	if bmStore == "" {
		for {
			fmt.Println("Could not find bookmark store file")
			fmt.Println("Please provide file path:")
			bmFile := promptBookmarkFile()
			_, err := os.Stat(bmFile)
			if err == nil {
				bmRoots = ReadInputFile(bmFile)
				break
			}
		}
	} else {
		bmRoots = ReadInputFile(bmStore)
	}
	var exportNodeRoots []*base.BookmarkNodeBase
	var tmpDirPath string
	var tracker = util.NewTracker()

	items := getOptionSlice()

	for {
		showMenu(bmRoots, exportNodeRoots, tmpDirPath, config, items, tracker)
	}
}

func showMenu(
	bmRoots []*base.BookmarkNodeBase,
	exportNodeRoots []*base.BookmarkNodeBase,
	tmpDirPath string,
	config *configuration.Bm2fsConfig,
	items []optionItem,
	tracker *util.BookmarkTracker,
) {
	prompt := promptui.Select{
		Label: fmt.Sprintf(
			"Select operation [temporary dir: %s]",
			config.TmpRoot,
		),
		Items: items,
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
	} else if options[result] == options["populate filesystem tree"] {
		tmpDirPath, _ = fstree.PopulateTmpDir(bmRoots, tracker, config.TmpRoot)
		defer os.RemoveAll(tmpDirPath)
	} else if options[result] == options["export to browser HTML"] {
		if _, err := os.Stat(tmpDirPath); err == nil {
			exportNodeRoots = fstree.CollectFSTrees(tmpDirPath, tracker)
			htmlconv.BuildTreeHTML(exportNodeRoots, config.OutputFile)
		}
	} else if options[result] == options["change bookmark store"] {
		bmFile := promptBookmarkFile()
		bmRoots = ReadInputFile(bmFile)
	} else if options[result] == options["show changes"] {
		// TODO varies according to bookmark array source
	} else if options[result] == options["create/save sqlite"] {
		saveSQLite(exportNodeRoots, bmRoots)
	} else if options[result] == options["EXIT"] {
		return
	}
}

func saveSQLite(exportNodeRoots []*base.BookmarkNodeBase, bmRoots []*base.BookmarkNodeBase) {
	var rootsToSave []*base.BookmarkNodeBase
	if exportNodeRoots != nil {
		rootsToSave = exportNodeRoots
	} else if bmRoots != nil {
		rootsToSave = bmRoots
	}
	db.BackupNodeRoots(rootsToSave)
}

func promptBookmarkFile() string {

	defaultPath, err := os.Getwd()
	if err != nil {
		defaultPath, err = os.UserHomeDir()
		check(err)
	}
	prompt := promptui.Prompt{
		Label:   "Bookmark path:",
		Default: defaultPath,
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return ""
	}
	return result
}

func getChromeJSONFile() string {
	var chromeBM, chromiumBM, file string
	switch runtime.GOOS {
	case "windows":
		appdata, err := os.UserConfigDir()
		check(err)
		chromeBM = path.Join(appdata, "Local", "Google", "Chrome", "User Data", "Default", "Bookmarks")
		chromiumBM = path.Join(appdata, "Local", "Chromium", "User Data", "Default", "Bookmarks")
	default:
		configDir, err := os.UserConfigDir()
		check(err)
		chromeBM = path.Join(configDir, "Chromium", "Default", "Bookmarks")
		chromiumBM = path.Join(configDir, "google-chrome", "Default", "Bookmarks")
	}
	fileAttempts := []string{chromeBM, chromiumBM}
	for _, file := range fileAttempts {
		_, err := os.Stat(file)
		if err == nil {
			break
		}
	}
	return file
}

func parseArgs(args []string) string {
	now := time.Now()
	timeString := now.Format("02_01_06")
	defaultFileName := fmt.Sprintf("BM2FS_bookmarks_%s.html",
		timeString,
	)

	candidatesInFile := []string{
		inFile,
		args[len(args)],
		defaultFileName,
	}

	var outname string
	for _, f := range candidatesInFile {
		if !fileExists(f) {
			outname = f
			break
		}
	}

	return outname
}

func generateSourceList(args []string) []string {
	sourcePaths := []string{}
	if len(args) > 0 {
		if args[0] != "" {
			sourcePaths = append(sourcePaths, args[len(args)-1])
		}
	}
	sourcePaths = append(sourcePaths, getChromeJSONFile())
	return sourcePaths
}
