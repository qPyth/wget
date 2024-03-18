package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"wget/cmd/wget"
)

// Analog wget application

const pathToLogFile = "log.txt"

func main() {

	logFile, err := os.OpenFile(pathToLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %s\n", err.Error())
	}
	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}))
	flags := parseFlags()
	app := wget.NewWget(logger)
	err = app.ProcessFlags(flags)
	if err != nil {
		logger.Error("process flags error", "error", err.Error())
		log.Fatal(err.Error())
	}
	app.Start()
}

func parseFlags() wget.Flags {
	var flags wget.Flags

	flag.BoolVar(&flags.BgFlag, "B", false, "run in background")
	flag.StringVar(&flags.RateFlag, "rate-limit", "", "set the rate limit like \"200K or 2M\"")
	flag.StringVar(&flags.PathFlag, "P", "", "set the path to save the file")
	flag.StringVar(&flags.NameFlag, "o", "", "set the name of the file")
	flag.BoolVar(&flags.MirrorFlag, "mirror", false, "mirror the site")
	flag.BoolVar(&flags.MultiFlag, "i", false, "download multiple files")
	flag.Parse()
	flags.URL = flag.Arg(0)

	return flags
}
