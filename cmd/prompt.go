package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/manifoldco/promptui"
	config "github.com/ossilva/bookmark2fs/pkg/configuration"
	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
	"github.com/ossilva/bookmark2fs/pkg/conversion/htmlconv"
	"github.com/ossilva/bookmark2fs/pkg/db"
	"github.com/ossilva/bookmark2fs/pkg/fstree"
	"github.com/ossilva/bookmark2fs/pkg/util"
)

const MaxPromptArgs = 1

var cmdPrompt = &cobra.Command{
	Use:   "prompt",
	Short: "Start in prompt mode",
	Long:  `Access bookmark2fs functionality through user-friendly menus.`,
	Args:  cobra.RangeArgs(0, MaxPromptArgs),
	Run:   Init,
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

func changeFSRootDir(config *config.Bm2fsConfig) *config.Bm2fsConfig {
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
	return config
}

//Init prompts user for input
func Init(cmd *cobra.Command, args []string) {
	fmt.Println("******************************************")

	inFiles := acquireBrowserBMFiles(args)
	var config *config.Bm2fsConfig = makeConfig()

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
	config *config.Bm2fsConfig,
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

	if options[result] == options["change root directory"] {
		config = changeFSRootDir(config)
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

func solveDefaultChromePaths(goOs string, userConfigDir string) (string, string) {
	var chromeBmJSON, chromiumBmJSON string
	switch goOs {
	case "windows":
		chromeBmJSON = path.Join(userConfigDir, "Local", "Google", "Chrome", "User Data", "Default", "Bookmarks")
		chromiumBmJSON = path.Join(userConfigDir, "Local", "Chromium", "User Data", "Default", "Bookmarks")
	default:
		chromeBmJSON = path.Join(userConfigDir, "Chromium", "Default", "Bookmarks")
		chromiumBmJSON = path.Join(userConfigDir, "google-chrome", "Default", "Bookmarks")
	}
	return chromeBmJSON, chromiumBmJSON
}

func getChromeJSONFile() []string {
	userConfigDir, err := os.UserConfigDir()
	check(err)
	chromeBmJSON, chromiumBmJSON := solveDefaultChromePaths(runtime.GOOS, userConfigDir)
	fileAttempts := []string{chromeBmJSON, chromiumBmJSON}
	for _, file := range fileAttempts {
		_, err := os.Stat(file)
		if err == nil {
			return []string{file}
		}
	}
	return []string{}
}

func acquireBrowserBMFiles(args []string) []string {
	sourcePaths := []string{}
	if len(args) == MaxPromptArgs {
		sourcePaths = append(sourcePaths, args[0])
	}
	sourcePaths = append(sourcePaths, getChromeJSONFile()...)
	return sourcePaths
}
