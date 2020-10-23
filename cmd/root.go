package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ossilva/bookmark2fs/pkg/conversion"
	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
	"github.com/ossilva/bookmark2fs/pkg/fstree"
	"github.com/ossilva/bookmark2fs/pkg/util"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

var (
	//for flags
	inFile   string
	outFile  string
	quietp   bool
	convertp bool
	tracker  util.BookmarkTracker

	rootCmd = &cobra.Command{
		Use:   "bookmark2fs [bookmark file]",
		Short: "Bookmark2fs manages bookmarks as files and returns browser-readable html ",
		Long: `A tool for converting JSON/html bookmarks and simple file trees. Exporting
		to browser-compatible bookmark html`,
		Run: run,
	}
)

func run(cmd *cobra.Command, args []string) {
	var tracker = util.BookmarkTracker{}
	reader, err := os.Open(inFile)
	check(err)
	if inFile == "Bookmarks" {
		conversion.DecodeJSON(reader)
	}

	var bookmarkRoots map[string]*base.BookmarkNodeBase
	switch filepath.Ext(inFile) {
	case "json":
		bookmarkRoots = conversion.DecodeJSON(reader)
	case "html":
		bookmarkRoots = conversion.ParseNetscapeHTML(reader)
	}

	tmpDirPath, tmpRootPathMap := fstree.PopulateTmpDir(bookmarkRoots, tracker)

	var exportRoots = map[string]*base.BookmarkNodeBase{}
	for k, v := range tmpRootPathMap {
		exportRoots[k] = fstree.ConstructFSTree(v, tracker)
	}
	conversion.BuildTreeHTML(exportRoots, outFile)

	// readUserInput()
	defer os.RemoveAll(tmpDirPath)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getChromeJSONFile() ([]byte, error) {
	var chromeBM, chromiumBM string
	switch runtime.GOOS {
	case "windows":
		appdata := os.Getenv("APPDATA")
		chromeBM = path.Join(appdata, "Local", "Google", "Chrome", "User Data", "Default", "Bookmarks")
		chromiumBM = path.Join(appdata, "Local", "Chromium", "User Data", "Default", "Bookmarks")
	default:
		configDir := os.Getenv("XDG_CONFIG")
		chromeBM = path.Join(configDir, "Chromium", "Default", "Bookmarks")
		chromiumBM = path.Join(configDir, "google-chrome", "Default", "Bookmarks")
	}
	fileAttempts := []string{chromeBM, chromiumBM}
	var jsonFile string
	for _, file := range fileAttempts {
		_, err := os.Stat(file)
		if err != nil {
			break
		}
	}
	jsonBytes, err := ioutil.ReadFile(jsonFile)
	return jsonBytes, err
}

func readUserInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("type \"done\" to save new bookmarks")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		if strings.Compare("done", text) == 0 {
			return
		}

	}
}

func parseArgs(args []string) string {
	now := time.Now()
	timeString := now.Format("02_01_06")
	yearDecade := fmt.Sprintf("BM2FS_bookmarks_%s.html",
		timeString,
	)
	candidatesInFile := []string{
		inFile,
		args[len(args)],
		yearDecade,
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

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&inFile, "in", "i", "", "input file to parsed of type 'HTML' or 'JSON'")
	// rootCmd.PersistentFlags().StringVarP(&outFile, "out", "o", "", "filename to write html to")
	rootCmd.PersistentFlags().BoolVarP(&quietp, "quiet", "q", false, "don't print anything to stdout")
	rootCmd.PersistentFlags().BoolVarP(&convertp, "convert", "c", false, "only perform conversion JSON -> HTML")

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	// rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "author name for copyright attribution")
	// rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "name of license for the project")
	// rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	// viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	// viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	// viper.SetDefault("license", "apache")

	// rootCmd.AddCommand(addCmd)
	// rootCmd.AddCommand(initCmd)
}
