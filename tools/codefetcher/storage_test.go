package codefetcher

import (
	"bytes"
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	_ "github.com/glebarez/go-sqlite"
	"os"
	"testing"
	"time"
)

const (
	testGetUrlByContentQuery = "SELECT url from code WHERE content = ?;"
	testGetContentQuery      = "SELECT content from code;"
)

var (
	testCodefileHelloWorld     = []byte("#!/usr/bin/env python3\n\nprint(\"Hello World\")\n")
	testCodefileHelloWorldHash = hex.EncodeToString(sha1.New().Sum(testCodefileHelloWorld))

	testCodefileHelloWorld2     = []byte("#!/usr/bin/env python3\n\nprint(\"Hello World\")\n\n")
	testCodefileHelloWorld2Hash = hex.EncodeToString(sha1.New().Sum(testCodefileHelloWorld2))

	testLanguage1, _ = ParseLanguage("python")
	testLanguage2, _ = ParseLanguage("c#")
	testTimeout      = 3 * time.Second
	testDatabasePath = "test.db"
)

func createTempDatabase(t *testing.T) Storage {

	tempDatabasePath := t.TempDir() + testDatabasePath

	db, err := sql.Open("sqlite", tempDatabasePath)
	if err != nil {
		t.Fatalf("Error creating database: %v", err)
	}

	s := Storage{
		DB: db,
	}

	err = s.Init(context.Background())
	if err != nil {
		t.Fatalf("Error initializing database: %v", err)
	}

	return s
}

func TestInit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	s := createTempDatabase(t)
	defer s.DB.Close()

	var count int
	err := s.queryRowContext(ctx, `SELECT COUNT(name) FROM sqlite_schema WHERE name="code";`).Scan(&count)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}

	if count != 1 {
		t.Fatalf("Table 'code' not found in database")
	}

	count = 0
	err = s.queryRowContext(ctx, `SELECT COUNT(name) FROM sqlite_schema WHERE name="progress";`).Scan(&count)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}

	if count != 1 {
		t.Fatalf("Table 'progress' not found in database")
	}
}

func TestTableExists(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	s := createTempDatabase(t)
	defer s.DB.Close()

	exists, err := s.tableExists(ctx, "code")
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}

	if !exists {
		t.Fatalf("Table 'code' not found in database")
	}

	exists, err = s.tableExists(ctx, "progress")
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}

	if !exists {
		t.Fatalf("Table 'progress' not found in database")
	}

	exists, err = s.tableExists(ctx, "doesnotexist")
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}

	if exists {
		t.Fatalf("Table 'doesnotexist' found in database")
	}
}

func TestHash(t *testing.T) {
	hash := sha1.Sum(testCodefileHelloWorld)
	t.Log(hex.EncodeToString(hash[:]))
}

func TestHashUnique(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	s := createTempDatabase(t)
	defer s.DB.Close()

	err := s.StoreCodefile(ctx, testLanguage1, "http://localhost/main.py", testCodefileHelloWorld, testCodefileHelloWorldHash)
	if err != nil {
		t.Fatalf("Error inserting codefile: %v", err)
	}

	err = s.StoreCodefile(ctx, testLanguage1, "http://localhost/main3.py", testCodefileHelloWorld, testCodefileHelloWorldHash)
	if err != nil {
		t.Fatalf("Error inserting codefile: %v", err)
	}

	codeFiles, err := s.CountCodefiles(ctx)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}
	if codeFiles != 1 {
		t.Fatalf("Expected 1 codefiles, got %d", codeFiles)
	}

	var url string
	err = s.queryRowContext(ctx, testGetUrlByContentQuery, testCodefileHelloWorld).Scan(&url)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}

	if url != "http://localhost/main.py" {
		t.Fatalf("Expected url to be 'http://localhost/main.py', got '%s'", url)
	}
}

func TestInsertAndCodeCount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	s := createTempDatabase(t)
	defer s.DB.Close()

	count, err := s.CountCodefiles(ctx)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}

	if count != 0 {
		t.Fatalf("Expected 0 rows in table, got %d", count)
	}

	err = s.StoreCodefile(ctx, testLanguage1, "http://localhost/main.py", testCodefileHelloWorld, testCodefileHelloWorldHash)
	if err != nil {
		t.Fatalf("Error inserting codefile: %v", err)
	}

	var content string
	err = s.queryRowContext(ctx, testGetContentQuery).Scan(&content)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}

	if bytes.Compare([]byte(content), testCodefileHelloWorld) != 0 {
		t.Fatalf("Expected content to be '%s', got '%s'", testCodefileHelloWorld, content)
	}
}

func TestTotalCodeSize(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	s := createTempDatabase(t)
	defer s.DB.Close()

	totalSize, err := s.GetTotalCodeSizeByLanguage(ctx, testLanguage1)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}
	if totalSize != 0 {
		t.Fatalf("Expected 0 bytes, got %d", totalSize)
	}

	err = s.StoreCodefile(ctx, testLanguage1, "http://localhost/main.py", testCodefileHelloWorld, testCodefileHelloWorldHash)
	if err != nil {
		t.Fatalf("Error inserting codefile: %v", err)
	}
	totalSize, err = s.GetTotalCodeSizeByLanguage(ctx, testLanguage1)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}
	if totalSize != len(testCodefileHelloWorld) {
		t.Fatalf("Expected %d bytes, got %d", len(testCodefileHelloWorld), totalSize)
	}

	err = s.StoreCodefile(ctx, testLanguage1, "http://localhost/main2.py", testCodefileHelloWorld2, testCodefileHelloWorld2Hash)
	if err != nil {
		t.Fatalf("Error inserting codefile: %v", err)
	}
	totalSize, err = s.GetTotalCodeSizeByLanguage(ctx, testLanguage1)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}
	if totalSize != len(testCodefileHelloWorld)+len(testCodefileHelloWorld2) {
		t.Fatalf("Expected %d bytes, got %d", len(testCodefileHelloWorld)+len(testCodefileHelloWorld2), totalSize)
	}

	totalSize, err = s.GetTotalCodeSizeByLanguage(ctx, testLanguage2)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}
	if totalSize != 0 {
		t.Fatalf("Expected 0 bytes, got %d", totalSize)
	}
}

func TestLargeCode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	s := createTempDatabase(t)
	defer s.DB.Close()

	largeCodeFile, err := os.ReadFile("storage_test.go")
	if err != nil {
		t.Fatalf("Error reading test file: %v", err)
	}
	largeCodeFile = bytes.Repeat(largeCodeFile, 100)

	err = s.StoreCodefile(ctx, testLanguage1, "http://localhost/main.py", largeCodeFile, hex.EncodeToString(sha1.New().Sum(largeCodeFile)))
	if err != nil {
		t.Fatalf("Error inserting codefile: %v", err)
	}
}

func TestGetProcessEmpty(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	s := createTempDatabase(t)
	defer s.DB.Close()

	progress, err := s.GetProgress(ctx, testLanguage1, "*")
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}
	if progress != 0 {
		t.Fatalf("Expected 0, got %d", progress)
	}
}
func TestGetProcess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	s := createTempDatabase(t)
	defer s.DB.Close()

	err := s.UpdateProgress(ctx, testLanguage1, "*", 50)
	if err != nil {
		t.Fatalf("Error updating database: %v", err)
	}

	progress, err := s.GetProgress(ctx, testLanguage1, "*")
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}
	if progress != 50 {
		t.Fatalf("Expected 50, got %d", progress)
	}

	err = s.UpdateProgress(ctx, testLanguage1, "*", 42)
	if err != nil {
		t.Fatalf("Error updating database: %v", err)
	}

	progress, err = s.GetProgress(ctx, testLanguage1, "*")
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}
	if progress != 42 {
		t.Fatalf("Expected 42, got %d", progress)
	}
}

func TestCodeExistsByHash(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	s := createTempDatabase(t)
	defer s.DB.Close()

	exists, err := s.CodeExistsByHash(ctx, testCodefileHelloWorldHash)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}
	if exists {
		t.Fatalf("Expected code to not exist")
	}

	err = s.StoreCodefile(ctx, testLanguage1, "http://localhost/main.py", testCodefileHelloWorld, testCodefileHelloWorldHash)
	if err != nil {
		t.Fatalf("Error inserting codefile: %v", err)
	}

	exists, err = s.CodeExistsByHash(ctx, testCodefileHelloWorldHash)
	if err != nil {
		t.Fatalf("Error querying database: %v", err)
	}
	if !exists {
		t.Fatalf("Expected code to exist")
	}
}
