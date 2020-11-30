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

	for {
		prompt := promptui.Select{
			Label: fmt.Sprintf(
				"Select operation [temporary dir: %s]",
				config.TmpRoot,
			),
			Items: []string{
				"populate filesystem tree",
				"export to browser HTML",
				"change bookmark store",
				"change root directory",
				"create/save sqlite",
				// "show changes",
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
		} else if result == "populate filesystem tree" {
			tmpDirPath, _ = fstree.PopulateTmpDir(bmRoots, tracker, config.TmpRoot)
			defer os.RemoveAll(tmpDirPath)
		} else if result == "export to browser HTML" {
			exportNodeRoots = fstree.CollectFSTrees(tmpDirPath, tracker)
			htmlconv.BuildTreeHTML(exportNodeRoots, config.OutputFile)
		} else if result == "change bookmark store" {
			bmFile := promptBookmarkFile()
			bmRoots = ReadInputFile(bmFile)
		} else if result == "show changes" {
			// TODO varies according to bookmark array source
		} else if result == "create/save sqlite" {
			var rootsToSave []*base.BookmarkNodeBase
			if exportNodeRoots != nil {
				rootsToSave = exportNodeRoots
			} else if bmRoots != nil {
				rootsToSave = bmRoots
			}
			db.BackupNodeRoots(rootsToSave)
		} else if result == "EXIT" {
			return
		}
	}

}

func promptBookmarkFile() string {
	// validate := func(path string) error {
	// 	_, err := os.Stat(path)
	// 	check(err)
	// 	return nil
	// }

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
