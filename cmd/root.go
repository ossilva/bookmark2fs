package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/ossilva/bookmark2fs/pkg/configuration"
	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
	"github.com/ossilva/bookmark2fs/pkg/conversion/htmlconv"
	"github.com/ossilva/bookmark2fs/pkg/conversion/jsonconv"
	"github.com/ossilva/bookmark2fs/pkg/fstree"
	"github.com/ossilva/bookmark2fs/pkg/prompt"
	"github.com/ossilva/bookmark2fs/pkg/util"
)

func makeConfig() *configuration.Bm2fsConfig {

	timeString := time.Now().Format("02_01_06")
	defaultFileName := fmt.Sprintf("BM2FS_bookmarks_%s.html",
		timeString,
	)
	wd, err := os.Getwd()
	check(err)
	outFile := wd + "/" + defaultFileName
	var configDir, _ = os.UserConfigDir()
	configDir += "bookmark.sqlite"
	tmpDir, err := getFreeTmpDir()
	check(err)
	var config = &configuration.Bm2fsConfig{
		TmpRoot:    tmpDir,
		UserDB:     configDir,
		OutputFile: outFile,
	}
	return config
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

var (
	//for flags
	inFile       string
	outFile      string
	pop          bool
	depop        bool
	quietp       bool
	convertp     bool
	interactivep bool
	tracker      *util.BookmarkTracker

	rootCmd = &cobra.Command{
		Use:   "bookmark2fs [bookmark file]",
		Short: "Bookmark2fs reads bookmark files as file trees and returns browser-readable html ",
		Long: `A tool for reading JSON/html bookmarks and converting them into simple file trees. Exports
		file trees to browser-compatible bookmark html. When used interactively (with flag -p),
		 the program clears temporary directories, otherwise temporary directories remains`,
		Run: run,
	}
)

func run(cmd *cobra.Command, args []string) {
	config := makeConfig()
	SetupCloseHandler(config)

	if inFile == "" {
		inFile = args[0]
	}
	reader, err := os.Open(inFile)
	check(err)
	if inFile == "Bookmarks" {
		jsonconv.DecodeJSON(reader)
	}

	var bookmarkRoots []*base.BookmarkNodeBase
	switch filepath.Ext(inFile) {
	case ".json":
		bookmarkRoots = jsonconv.DecodeJSON(reader)
	case ".html":
		bookmarkRoots = htmlconv.ParseNetscapeHTML(reader)
	default:
		panic("unrecognized file extension")
	}

	if convertp {
		htmlconv.BuildTreeHTML(bookmarkRoots, config.OutputFile)
		return
	} else if pop {
		fstree.PopulateTmpDir(bookmarkRoots, tracker, config.TmpRoot)
	} else if depop {
		rootPath := path.Join(os.TempDir(), configuration.ProgramName)
		check(err)

		exportRoots := fstree.CollectFSTrees(rootPath, tracker)
		htmlconv.BuildTreeHTML(exportRoots, config.OutputFile)
	}

	if interactivep {
		prompt.Init(config, bookmarkRoots, tracker)
		return
	}
	// displayCLI()

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

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&inFile, "in", "i", "", "input file to parsed of type 'HTML' or 'JSON'")
	rootCmd.PersistentFlags().StringVarP(&outFile, "out", "o", "stdout", "filename to write html to")
	rootCmd.PersistentFlags().BoolVarP(&pop, "populate", "p", false, "populate filesystem tree")
	rootCmd.PersistentFlags().BoolVarP(&depop, "depopulate", "d", false, "depopulate filesystem tree")
	rootCmd.PersistentFlags().BoolVarP(&quietp, "quiet", "q", false, "don't print anything to stdout")
	rootCmd.PersistentFlags().BoolVarP(&convertp, "convert", "c", false, "only perform conversion JSON -> HTML")
	rootCmd.PersistentFlags().BoolVarP(&interactivep, "prompt", "", false, "prompt user for commands")
	tracker = util.NewTracker()
}

func getFreeTmpDir() (name string, err error) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), configuration.ProgramName)
	if tmpDir == "" {
		os.Remove(tmpDir)
	}
	return tmpDir, err
}

func SetupCloseHandler(config *configuration.Bm2fsConfig) {
	cleanTmpDir := func() {
		os.RemoveAll(config.TmpRoot)
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Cleared temporary directory")
		cleanTmpDir()
		os.Exit(0)
	}()
}
