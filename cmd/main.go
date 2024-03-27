package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"wget/internal/downloader"
	"wget/internal/types"
	"wget/internal/wget"
)

// Analog wget application

const pathToLogFile = "log.txt"

func main() {

	flags := parseFlags()
	if flags.BgFlag {
		logFile, err := os.OpenFile(pathToLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening log file: %s\n", err.Error())
		}
		cmd := exec.Command(os.Args[0], flag.Args()...)
		cmd.Stdin = logFile
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		err = cmd.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при запуске в фоновом режиме: %s\n", err)
			os.Exit(1)
		}
		return
	}
	app := wget.NewWget()
	switch {
	case flags.MultiFlag != "":
		app.SetDownloader(new(downloader.Multi))
	case flags.MirrorFlag:
		app.SetDownloader(new(downloader.Site))
	default:
		app.SetDownloader(new(downloader.Interactive))
	}
	app.StartDownload(flags)
}

func parseFlags() types.Flags {
	var flags types.Flags

	flag.BoolVar(&flags.BgFlag, "B", false, "run in background")
	flag.StringVar(&flags.RateFlag, "rate-limit", "", "set the rate limit like \"200K or 2M\"")
	flag.StringVar(&flags.PathFlag, "P", "", "set the path to save the file")
	flag.StringVar(&flags.NameFlag, "o", "", "set the name of the file")
	flag.BoolVar(&flags.MirrorFlag, "mirror", false, "mirror the site")
	flag.StringVar(&flags.MultiFlag, "i", "", "download multiple files from file")
	flag.Parse()
	flags.URL = flag.Arg(0)

	return flags
}
