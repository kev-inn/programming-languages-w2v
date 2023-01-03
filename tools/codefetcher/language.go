package codefetcher

import (
	"fmt"
	"strings"
)

type githubLanguage string
type githubLanguages []githubLanguage

const (
	githubLanguagePython     githubLanguage = "Python"
	githubLanguageGolang                    = "Go"
	githubLanguageCSharp                    = "C#"
	githubLanguageCpp                       = "C++"
	githubLanguageC                         = "C"
	githubLanguageJava                      = "Java"
	githubLanguageJavascript                = "JavaScript"
	githubLanguageKotlin                    = "Kotlin"
)

var AvailableLanguages = githubLanguages{
	githubLanguagePython,
	githubLanguageGolang,
	githubLanguageCSharp,
	githubLanguageCpp,
	githubLanguageC,
	githubLanguageJava,
	githubLanguageJavascript,
	githubLanguageKotlin,
}

func (ls githubLanguages) String() string {
	var s []string
	for _, l := range AvailableLanguages {
		s = append(s, string(l))
	}
	return strings.Join(s, ", ")
}

func ParseLanguage(language string) (Language, error) {
	switch strings.ToLower(language) {
	case "python", "python3", "py", "py3":
		return newLanguage(githubLanguagePython, []string{".py", ".py3"}), nil
	case "go", "golang":
		return newLanguage(githubLanguageGolang, []string{".go"}), nil
	case "c#", "csharp":
		return newLanguage(githubLanguageCSharp, []string{".cs"}), nil
	case "c++", "cpp":
		return newLanguage(githubLanguageCpp, []string{".cpp", ".hpp", ".cxx", ".hxx", ".cc", ".hh", ".C", ".H", ".c", ".h"}), nil
	case "c":
		return newLanguage(githubLanguageC, []string{".c", ".h"}), nil
	case "java":
		return newLanguage(githubLanguageJava, []string{".java"}), nil
	case "javascript", "js":
		return newLanguage(githubLanguageJavascript, []string{".js"}), nil
	case "kotlin", "kt":
		return newLanguage(githubLanguageKotlin, []string{".kt"}), nil
	}

	return Language{}, fmt.Errorf("unknown language: %s", language)
}

type Language struct {
	name       githubLanguage
	extensions map[string]bool
}

func newLanguage(githubLanguage githubLanguage, fileExtensions []string) Language {
	l := Language{name: githubLanguage, extensions: make(map[string]bool)}
	for _, extension := range fileExtensions {
		l.extensions[strings.Replace(extension, ".", "", -1)] = true
	}
	return l
}

func (l Language) String() string {
	return string(l.name)
}

func (l Language) GithubQueryFilter() string {
	return strings.Join(append(l.GithubExtensionQuery(), "language:"+l.String()), "+")
}

func (l Language) GithubExtensionQuery() []string {
	var s []string
	for extension := range l.extensions {
		s = append(s, fmt.Sprintf("extension:%s", strings.Replace(extension, ".", "", -1)))
	}
	return s
}

func (l Language) ValidFileExtension(filepath string) error {
	dotPosition := strings.LastIndex(filepath, ".")
	if dotPosition == -1 {
		return fmt.Errorf("file %s has no extension", filepath)
	}

	fileExtension := filepath[dotPosition+1:]
	if _, ok := l.extensions[fileExtension]; ok {
		return nil
	}

	return fmt.Errorf("file %s has invalid extension for language %s", filepath, l)
}
