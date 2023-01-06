package main

import (
	"codefetcher/codefetcher"
	"context"
	"database/sql"
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
	databaseArg       *string = flag.StringP("database", "d", "codes.db", "SQLite database path")
	githubUserArg     *string = flag.String("github-user", "", "Github username")
	githubTokenArg    *string = flag.String("github-token", "", "Github access token")
	queryArg          *string = flag.StringP("query", "q", "", "Extra search terms for query")
	languageArg       *string = flag.StringP("language", "l", "", fmt.Sprintf("Programming language (%s)", codefetcher.AvailableLanguages))
	maxCodeSizeArg    *int    = flag.Int("max-code-size", 0, "Maximum total code size per language in bytes (0 = unlimited)")
	requestTimeoutArg *int    = flag.IntP("timeout", "t", 2000, "Timeout between requests in milliseconds")

	language       codefetcher.Language
	requestTimeout time.Duration = 0
)

func usage(exitCode int) {
	fmt.Println("Usage: ./codefetcher [options]")
	fmt.Println("  also see: https://docs.github.com/en/rest/search?apiVersion=2022-11-28")
	flag.PrintDefaults()
	os.Exit(exitCode)
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
		usage(1)
	}

	if len(*githubUserArg) == 0 {
		log.Error("Missing argument github username")
		usage(1)
	}

	if len(*githubTokenArg) == 0 {
		log.Error("Missing argument github token")
		usage(1)
	}

	if len(*languageArg) == 0 {
		log.Error("Missing argument language")
		usage(1)
	} else {
		var err error
		language, err = codefetcher.ParseLanguage(*languageArg)
		if err != nil {
			log.Error("Invalid argument language")
			usage(1)
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

	db, err := sql.Open("sqlite", *databaseArg)
	if err != nil {
		log.Errorf("Failed to open database: \"%s\"", err.Error())
		usage(2)
	}
	defer db.Close()

	s := codefetcher.Storage{DB: db}
	err = s.Init(ctx)
	if err != nil {
		log.Errorf("Failed to initialize database: \"%s\"", err.Error())
		usage(3)
	}

	log.Infof("Connected to database %s", *databaseArg)
	log.Infof("Fetching code from github.com for language %s with query \"%s\"", language.String(), *queryArg)

	fetcher := codefetcher.NewGithubFetcher(*githubUserArg, *githubTokenArg, s, requestTimeout)
	err = fetcher.FetchCodes(ctx, language, *queryArg, *maxCodeSizeArg)
	if err != nil {
		log.Fatalf("Failed to fetch codes: %s", err.Error())
		usage(4)
	}
}
