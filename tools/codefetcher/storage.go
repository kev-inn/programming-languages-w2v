package codefetcher

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

const (
	sqlite3CreateTableCode = `CREATE TABLE IF NOT EXISTS "code" (
	"id"	INTEGER,
	"language"	TEXT NOT NULL,
	"url"	TEXT NOT NULL,
	"content"	TEXT NOT NULL,
	"hash"	TEXT NOT NULL UNIQUE,
	"size"	INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("id" AUTOINCREMENT)
);`
	sqlite3CreateTableProgress = `CREATE TABLE IF NOT EXISTS "progress" (
    	"language"	TEXT NOT NULL,
    	"query"	TEXT NOT NULL,
    	"last_page"	INTEGER NOT NULL DEFAULT 0,
    	PRIMARY KEY("language", "query")
);`
	sqlite3DropTables            = `DROP TABLE IF EXISTS "code"; DROP TABLE IF EXISTS "progress";`
	sqlite3InsertCode            = `INSERT INTO code (language, url, content, hash, size) VALUES (?, ?, ?, ?, ?)`
	sqlite3CountCodes            = `SELECT COUNT(id) as row_count FROM code;`
	sqlite3GetFirstFromTable     = `SELECT * FROM %s LIMIT 1;`
	sqlite3GetCodeSizeByLanguage = `SELECT IFNULL(SUM(size), 0) as total_size FROM code WHERE language = ?;`
	sqlite3CodeExists            = `SELECT COUNT(1) FROM code WHERE hash = ?;`
	sqlite3GetProgress           = `SELECT last_page FROM progress WHERE language = ? AND query = ?;`
	sqlite3UpdateProgress        = `INSERT OR REPLACE INTO progress (language, query, last_page) VALUES (?, ?, ?);`
)

type Storage interface {
	SaveCodefile(ctx context.Context, language Language, url string, content []byte, hash string) error
	CountCodefiles(ctx context.Context) (int, error)
	GetTotalCodeSizeByLanguage(ctx context.Context, language Language) (int, error)
	Close() error
	Url() string
	GetProgress(ctx context.Context, language Language, query string) (int, error)
	UpdateProgress(ctx context.Context, language Language, query string, lastPage int) error
	CodeExistsByHash(ctx context.Context, hash string) (bool, error)

	// for testing
	queryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	tableExists(ctx context.Context, tableName string) (bool, error)
	dropTables() error
}

type Sqlite3Storage struct {
	db      *sql.DB
	path    string
	testing bool // true for testing only (used to drop database after close)
}

func NewSqlite3Storage(path string) (Storage, error) {
	s := Sqlite3Storage{
		path: path,
	}

	var err error
	s.db, err = sql.Open("sqlite3", fmt.Sprintf("file:%s?_mutex=full", path))
	if err != nil {
		log.Debugf("Failed to open sqlite3 database [%s]: %s", path, err.Error())
		return nil, err
	}
	s.db.SetMaxOpenConns(1)

	// init database
	for _, query := range []string{sqlite3CreateTableCode, sqlite3CreateTableProgress} {
		_, err = s.db.Exec(query)
		if err != nil {
			log.Debugf("Failed to execute query [%s]: %s", query, err.Error())
			return nil, err
		}
	}

	return s, nil
}

func newSqlite3TestStorage(path string) (Storage, error) {
	s := Sqlite3Storage{
		path:    path,
		testing: true,
	}

	var err error
	s.db, err = sql.Open("sqlite3", fmt.Sprintf("file:%s?_mutex=full", path))
	if err != nil {
		log.Debugf("Failed to open sqlite3 database [%s]: %s", path, err.Error())
		return nil, err
	}
	s.db.SetMaxOpenConns(1)

	// init database
	for _, query := range []string{sqlite3CreateTableCode, sqlite3CreateTableProgress} {
		_, err = s.db.Exec(query)
		if err != nil {
			log.Debugf("Failed to execute query [%s]: %s", query, err.Error())
			return nil, err
		}
	}

	return s, nil
}

func (s Sqlite3Storage) Url() string {
	return s.path
}

func (s Sqlite3Storage) Close() error {

	var err error
	if s.testing {
		// drop database for testing
		err = s.dropTables()
		if err != nil {
			log.Errorf("Failed to drop tables: %s", err.Error())
		}
	}

	if closeErr := s.db.Close(); err == nil {
		err = closeErr
	}

	return err
}

func (s Sqlite3Storage) SaveCodefile(ctx context.Context, language Language, url string, content []byte, hash string) error {

	if len(hash) == 0 {
		hash = hex.EncodeToString(sha1.New().Sum(content))
	}

	_, err := s.db.ExecContext(ctx, sqlite3InsertCode, language.String(), url, content, hash, len(content))
	if err != nil {
		if err, ok := err.(sqlite3.Error); ok {
			if err.ExtendedCode == sqlite3.ErrConstraintUnique {
				// ignore duplicate code
				return nil
			}
		}
		log.Debugf("Failed to save codefile VALUES(%s, %s): %s", language.String(), url, err.Error())
		return err
	}
	return nil
}

func (s Sqlite3Storage) CountCodefiles(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, sqlite3CountCodes).Scan(&count)
	if err != nil {
		log.Debugf("Failed to count codefiles: %s", err.Error())
		return 0, err
	}
	return count, nil
}

func (s Sqlite3Storage) GetTotalCodeSizeByLanguage(ctx context.Context, language Language) (int, error) {
	var size int
	err := s.db.QueryRowContext(ctx, sqlite3GetCodeSizeByLanguage, language.String()).Scan(&size)
	if err != nil {
		log.Debugf("Failed to get total code size for language %s: %s", language, err.Error())
		return 0, err
	}
	return size, nil
}

func (s Sqlite3Storage) GetProgress(ctx context.Context, language Language, query string) (int, error) {
	var lastPage int
	err := s.db.QueryRowContext(ctx, sqlite3GetProgress, language.String(), query).Scan(&lastPage)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no progress found for language %s and query %s", language, query)
		}
		log.Debugf("Failed to get progress VALUES(%s, %s): %s", language, query, err.Error())
		return 0, err
	}
	return lastPage, nil
}

func (s Sqlite3Storage) UpdateProgress(ctx context.Context, language Language, query string, lastPage int) error {
	_, err := s.db.ExecContext(ctx, sqlite3UpdateProgress, language.String(), query, lastPage)
	if err != nil {
		log.Debugf("Failed to update progress VALUES(%s, %s, %d): %s", language, query, lastPage, err.Error())
		return err
	}
	return err
}

func (s Sqlite3Storage) CodeExistsByHash(ctx context.Context, hash string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, sqlite3CodeExists, hash).Scan(&exists)
	if err != nil {
		log.Debugf("Failed to check if code exists: %s", err.Error())
		return false, err
	}
	return exists, nil
}

func (s Sqlite3Storage) queryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}

func (s Sqlite3Storage) dropTables() error {
	_, err := s.db.Exec(sqlite3DropTables)
	return err
}

func (s Sqlite3Storage) tableExists(ctx context.Context, tableName string) (bool, error) {
	_, err := s.db.QueryContext(ctx, fmt.Sprintf(sqlite3GetFirstFromTable, tableName))
	if err != nil {
		return false, err
	}
	return true, nil
}
