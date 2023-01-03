package codefetcher

import (
	"testing"
)

var (
	testLanguage, _ = ParseLanguage(string(githubLanguagePython))
)

func TestFileExtension(t *testing.T) {

	if err := testLanguage.ValidFileExtension("main.py"); err != nil {
		t.Errorf("Expected main.py to be a valid file extension for language %s", testLanguage.String())
	}

	if err := testLanguage.ValidFileExtension("test/main.py"); err != nil {
		t.Errorf("Expected test/main.py to be a valid file extension for language %s", testLanguage.String())
	}

	if err := testLanguage.ValidFileExtension("/.test/main.py"); err != nil {
		t.Errorf("Expected /.test/main.py to be a valid file extension for language %s", testLanguage.String())
	}

	if err := testLanguage.ValidFileExtension("main.c"); err == nil {
		t.Errorf("Expected main.c to be a valid file extension for language %s", testLanguage.String())
	}

	if err := testLanguage.ValidFileExtension("test/main.c"); err == nil {
		t.Errorf("Expected test/main.c to be a valid file extension for language %s", testLanguage.String())
	}

	if err := testLanguage.ValidFileExtension("/.test/main.c"); err == nil {
		t.Errorf("Expected /.test/main.c to be a valid file extension for language %s", testLanguage.String())
	}
}

func TestGithubExtensionQuery(t *testing.T) {
	for _, l := range AvailableLanguages {
		language, err := ParseLanguage(string(l))
		if err != nil {
			t.Fatalf("Expected no error when parsing language %s", l)
		}

		t.Logf("%s: \"%v\"", language, language.GithubQueryFilter())
	}
}
