package main

import (
	"codefetcher/codefetcher"
	"context"
	goflag "flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"os"
	"os/signal"
	"time"
)

var (
	helpArg           *bool   = flag.BoolP("help", "h", false, "Show help/usage")
	logLevelArg       *string = flag.String("log-level", log.DebugLevel.String(), "Log level (debug, info, warn, error, fatal, panic)")
	databaseArg       *string = flag.StringP("database", "d", "codes.sqlite3", "SQLite3 database path")
	githubUserArg     *string = flag.String("github-user", "", "Github username")
	githubTokenArg    *string = flag.String("github-token", "", "Github access token")
	queryArg          *string = flag.StringP("query", "q", "*", "Extra search terms for query")
	languageArg       *string = flag.StringP("language", "l", "", fmt.Sprintf("Programming language (%s)", codefetcher.AvailableLanguages))
	maxCodeSizeArg    *int    = flag.Int("max-code-size", 1*1024, "Maximum total code size per language in bytes (default: 1KB)")
	requestTimeoutArg *int    = flag.IntP("timeout", "t", 2000, "Timeout between requests in milliseconds")

	language       codefetcher.Language
	requestTimeout time.Duration = 0
)

func usage() {
	fmt.Println("Usage: ./codefetcher [options]")
	fmt.Println("  also see: https://docs.github.com/en/rest/search?apiVersion=2022-11-28")
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)

	if logLevel, err := log.ParseLevel(*logLevelArg); err == nil {
		log.SetLevel(logLevel)
		log.Infof("Log level set to \"%s\"", logLevel.String())
	}

	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()

	if *helpArg {
		usage()
	}

	if len(*githubUserArg) == 0 {
		log.Error("Missing argument github username")
		usage()
	}

	if len(*githubTokenArg) == 0 {
		log.Error("Missing argument github token")
		usage()
	}

	if len(*languageArg) == 0 {
		log.Error("Missing argument language")
		usage()
	} else {
		var err error
		language, err = codefetcher.ParseLanguage(*languageArg)
		if err != nil {
			log.Error("Invalid argument language")
			usage()
		}
	}

	if *requestTimeoutArg > 0 {
		requestTimeout = time.Duration(*requestTimeoutArg) * time.Millisecond
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Handle Ctrl+C
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // second signal, hard exit
	}()

	s, err := codefetcher.NewSqlite3Storage(*databaseArg)
	if err != nil {
		log.Errorf("Failed to open sqlite database: \"%s\"", err.Error())
		os.Exit(1)
	}
	defer s.Close()

	log.Infof("Connected to database %s", s.Url())
	log.Infof("Fetching code from github.com for language %s with query \"%s\"", language.String(), *queryArg)

	fetcher := codefetcher.NewGithubFetcher(*githubUserArg, *githubTokenArg, s, requestTimeout)
	err = fetcher.FetchCodes(ctx, language, *queryArg, *maxCodeSizeArg)
	if err != nil {
		log.Fatalf("Failed to fetch codes: %s", err.Error())
		os.Exit(2)
	}
}
