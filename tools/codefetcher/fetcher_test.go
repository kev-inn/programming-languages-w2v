package codefetcher

import (
	flag "github.com/spf13/pflag"
	"testing"
	"time"
)

var (
	fetcher        GithubFetcher
	githubUserArg  *string = flag.String("github-user", "", "Github username")
	githubTokenArg *string = flag.String("github-token", "", "Github access token")
)

func init() {
	flag.Parse()
	fetcher = NewGithubFetcher(*githubUserArg, *githubTokenArg, nil, 0)
}

func TestDownloadError(t *testing.T) {
	err := func() error {
		return ErrorCodeSizeLimitExceeded
	}()

	if err != ErrorCodeSizeLimitExceeded {
		t.Fatalf("Expected error %s, got %s", ErrorCodeSizeLimitExceeded, err)
	}
}

func TestCodeResultHash(t *testing.T) {
	if len(*githubUserArg) == 0 || len(*githubTokenArg) == 0 {
		t.Skip("Missing argument github username and access token")
	}

}

func TestUnixEpochTime(t *testing.T) {
	timestamp := time.Unix(time.Now().Unix(), 0)
	t.Logf("%s", timestamp)
}
