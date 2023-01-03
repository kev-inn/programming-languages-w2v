package codefetcher

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"github.com/softlandia/cpd"
	"golang.org/x/sync/errgroup"
	"io"
	"strconv"
	"strings"
	"time"
)

// CodeSizeLimit code size limit for a single file, 0 for no limit
const CodeSizeLimit = 256 * 1024
const MaxRequestsParallel = 1

var (
	ErrorCodeSizeLimitExceeded = errors.New("code size limit exceeded")
)

type GithubFetcher struct {
	client         *github.Client
	storage        Storage
	requestTimeout time.Duration
}

func NewGithubFetcher(githubUser, githubAccessToken string, storage Storage, requestTimeout time.Duration) GithubFetcher {

	tp := github.BasicAuthTransport{
		Username: githubUser,
		Password: githubAccessToken,
	}

	return GithubFetcher{
		client:         github.NewClient(tp.Client()),
		storage:        storage,
		requestTimeout: requestTimeout,
	}
}

func (f GithubFetcher) DownloadCode(ctx context.Context, codeResult *github.CodeResult) ([]byte, error) {
	if ctx.Err() != nil {
		return []byte{}, ctx.Err()
	}

	reader, err := f.client.Repositories.DownloadContents(ctx, codeResult.Repository.GetOwner().GetLogin(), codeResult.Repository.GetName(), codeResult.GetPath(), nil)
	if err != nil {
		return []byte{}, err
	}
	defer reader.Close()

	uft8Reader, err := cpd.NewReader(reader)
	if err != nil {
		return []byte{}, err
	}

	if CodeSizeLimit > 0 {
		code, err := io.ReadAll(&io.LimitedReader{R: uft8Reader, N: CodeSizeLimit})
		if err != nil {
			return []byte{}, err
		}

		if len(code) == CodeSizeLimit {
			_, err = uft8Reader.Read(make([]byte, 1))
			if err != io.EOF {
				return []byte{}, ErrorCodeSizeLimitExceeded
			}
		}

		return code, nil
	}

	code, err := io.ReadAll(uft8Reader)
	if err != nil {
		return []byte{}, err
	}

	return code, nil
}

func secondsToTime(seconds string) string {
	secondsInt, err := strconv.Atoi(seconds)
	if err != nil {
		return seconds
	}
	time := time.Unix(int64(secondsInt), 0)
	return time.Format("2006-01-02 15:04:05")
}

func (f GithubFetcher) FetchCodes(ctx context.Context, language Language, query string, maxTotalSizeBytes int) error {

	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 30},
	}

	lastPage, err := f.storage.GetProgress(ctx, language, query)
	if err == nil {
		opt.Page = lastPage
		log.Infof("Resuming from page %d", lastPage)
		if opt.Page == -1 { // -1 indicates that the search is complete
			log.Infof("Search for language %s and query %s is already complete", language.String(), query)
			return nil
		}
	}
	defer func() {
		if opt.Page != 0 {
			f.storage.UpdateProgress(ctx, language, query, opt.Page)
		}
	}()

	for {
		time.Sleep(f.requestTimeout) // sleep to avoid rate limit
		log.Infof("Fetching page %d with %d entries per page", opt.Page, opt.PerPage)
		result, response, err := f.client.Search.Code(ctx, fmt.Sprintf("%s+%s", query, language.GithubQueryFilter()), opt)
		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				log.Errorf("Rate limit error: %s", err.Error())

				limit := response.Header.Get("X-RateLimit-Limit")
				remaining := response.Header.Get("X-RateLimit-Remaining")
				used := response.Header.Get("X-RateLimit-Used")
				reset := secondsToTime(response.Header.Get("X-RateLimit-Reset"))
				log.Infof("Status: limitReqPerH=%s, remainingReq=%s, usedReq=%s, resetAt=%s", limit, remaining, used, reset)

				rateLimitSleepTime := response.Rate.Reset.Sub(time.Now())
				log.Infof("Status: Sleeping for %s", rateLimitSleepTime)
				time.Sleep(rateLimitSleepTime)
				continue
			} else if errResponse, ok := err.(*github.ErrorResponse); ok {
				isSeconaryRateLimit := strings.Index(strings.ToLower(errResponse.Message), "secondary rate limit") != -1
				if isSeconaryRateLimit {
					log.Errorf("Secondary rate limit error: %s", err.Error())

					limit := response.Header.Get("X-RateLimit-Limit")
					remaining := response.Header.Get("X-RateLimit-Remaining")
					used := response.Header.Get("X-RateLimit-Used")
					reset := secondsToTime(response.Header.Get("X-RateLimit-Reset"))
					log.Infof("Status: limitReqPerH=%s, remainingReq=%s, usedReq=%s, resetAt=%s", limit, remaining, used, reset)

					rateLimitSleepTime := 10 * time.Minute
					log.Infof("Status: Sleeping for %s", rateLimitSleepTime)
					time.Sleep(rateLimitSleepTime)
					continue
				}
			}
			return err
		}

		// stop fetching code if total size limit is reached
		totalSizeBytes, err := f.storage.GetTotalCodeSizeByLanguage(ctx, language)
		if err != nil {
			return err
		}
		if totalSizeBytes > maxTotalSizeBytes {
			log.Infof("Total code size limit for language %s reached: %d bytes", language.String(), totalSizeBytes)
			return nil
		}

		log.Infof("Status: Downloading %d new code files...", len(result.CodeResults))
		if len(result.CodeResults) == 0 {
			log.Errorf("No code files found for language %s and query %s", language.String(), query)
			rateLimitSleepTime := 10 * time.Minute
			log.Infof("Status: Sleeping for %s", rateLimitSleepTime)
			time.Sleep(rateLimitSleepTime)
			continue
		}

		g, errCtx := errgroup.WithContext(ctx)
		g.SetLimit(MaxRequestsParallel) // limit number of parallel requests, set to 1 to avoid github rate limit!
		for _, codeResult := range result.CodeResults {
			if err = language.ValidFileExtension(codeResult.GetPath()); err != nil {
				log.Infof("Skip: %s - %s", codeResult.GetHTMLURL(), err.Error())
				continue
			}

			codeAlreadyExists, err := f.storage.CodeExistsByHash(ctx, codeResult.GetSHA())
			if err == nil && codeAlreadyExists {
				log.Infof("Skip: %s - Code already exists", codeResult.GetHTMLURL())
				continue
			}

			g.Go(func() error {
				time.Sleep(f.requestTimeout) // sleep to avoid rate limit
				code, err := f.DownloadCode(errCtx, &codeResult)
				if err != nil {
					if err == ErrorCodeSizeLimitExceeded {
						log.Infof("Skip: %s - %s", codeResult.GetHTMLURL(), err.Error())
						return nil
					}
					log.Infof("Error downloading code: %s", err.Error())
					return err
				}

				err = f.storage.SaveCodefile(errCtx, language, codeResult.GetHTMLURL(), code, codeResult.GetSHA())
				if err != nil {
					return err
				}

				log.Infof("OK: %s", codeResult.GetHTMLURL())
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			// this error is not critical, just log it
			if _, ok := err.(*github.RateLimitError); ok {
				log.Errorf("Error: Rate limit: %s", err.Error())
				log.Infof("Status: Pausing fetch process for 1 minute...")
				time.Sleep(1 * time.Minute)
			} else {
				log.Errorf("Error: fetching codes: %s", err.Error())
			}
		}

		if response.NextPage == 0 {
			log.Infof("Status: No more pages left for query %s", query)
			opt.Page = -1
			break
		}

		opt.Page = response.NextPage
		f.storage.UpdateProgress(ctx, language, query, opt.Page)
	}

	return nil
}
