package codefetcher

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/glebarez/go-sqlite"
	_ "github.com/glebarez/go-sqlite"
	log "github.com/sirupsen/logrus"
)

const (
	sqlCreateTableCode = `CREATE TABLE IF NOT EXISTS "code" (
	"id"	INTEGER,
	"language"	TEXT NOT NULL,
	"url"	TEXT NOT NULL,
	"content"	TEXT NOT NULL,
	"hash"	TEXT NOT NULL UNIQUE,
	"size"	INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY("id" AUTOINCREMENT)
);`
	sqlCreateTableProgress = `CREATE TABLE IF NOT EXISTS "progress" (
    	"language"	TEXT NOT NULL,
    	"query"	TEXT NOT NULL,
    	"last_page"	INTEGER NOT NULL DEFAULT 0,
    	PRIMARY KEY("language", "query")
);`
	sqlDropTables            = `DROP TABLE IF EXISTS "code"; DROP TABLE IF EXISTS "progress";`
	sqlInsertCode            = `INSERT INTO code (language, url, content, hash, size) VALUES (?, ?, ?, ?, ?)`
	sqlCountCodes            = `SELECT COUNT(id) as row_count FROM code;`
	sqlGetFirstFromTable     = `SELECT * FROM %s LIMIT 1;`
	sqlGetCodeSizeByLanguage = `SELECT IFNULL(SUM(size), 0) as total_size FROM code WHERE language = ?;`
	sqlCodeExists            = `SELECT COUNT(1) FROM code WHERE hash = ?;`
	sqlGetProgress           = `SELECT last_page FROM progress WHERE language = ? AND query = ?;`
	sqlUpdateProgress        = `INSERT OR REPLACE INTO progress (language, query, last_page) VALUES (?, ?, ?);`
)

var (
	ErrorNoDatabase = fmt.Errorf("no database initialized")
)

type Storage struct {
	DB *sql.DB
}

func (s Storage) Init(ctx context.Context) error {
	for _, query := range []string{sqlCreateTableCode, sqlCreateTableProgress} {
		_, err := s.DB.ExecContext(ctx, query)
		if err != nil {
			log.Debugf("Failed to execute query [%s]: %s", query, err.Error())
			return err
		}
	}
	return nil
}

func (s Storage) StoreCodefile(ctx context.Context, language Language, url string, content []byte, hash string) error {
	if s.DB == nil {
		return ErrorNoDatabase
	}

	if len(hash) == 0 {
		hash = hex.EncodeToString(sha1.New().Sum(content))
	}

	_, err := s.DB.ExecContext(ctx, sqlInsertCode, language.String(), url, content, hash, len(content))
	if err != nil {
		if errSql, ok := err.(*sqlite.Error); ok {
			if errSql.Code() == 2067 {
				// threat duplicate hashes as already stored
				return nil
			}
		}
		log.Debugf("Failed to save codefile VALUES(%s, %s): %s", language.String(), url, err.Error())
		return err
	}
	return nil
}

func (s Storage) CountCodefiles(ctx context.Context) (int, error) {
	if s.DB == nil {
		return 0, ErrorNoDatabase
	}
	var count int
	err := s.DB.QueryRowContext(ctx, sqlCountCodes).Scan(&count)
	if err != nil {
		log.Debugf("Failed to count codefiles: %s", err.Error())
		return 0, err
	}
	return count, nil
}

func (s Storage) GetTotalCodeSizeByLanguage(ctx context.Context, language Language) (int, error) {
	if s.DB == nil {
		return 0, ErrorNoDatabase
	}
	var size int
	err := s.DB.QueryRowContext(ctx, sqlGetCodeSizeByLanguage, language.String()).Scan(&size)
	if err != nil {
		log.Debugf("Failed to get total code size for language %s: %s", language, err.Error())
		return 0, err
	}
	return size, nil
}

func (s Storage) GetProgress(ctx context.Context, language Language, query string) (int, error) {
	if s.DB == nil {
		return 0, ErrorNoDatabase
	}
	var lastPage int
	err := s.DB.QueryRowContext(ctx, sqlGetProgress, language.String(), query).Scan(&lastPage)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Debugf("Failed to get progress VALUES(%s, %s): %s", language, query, err.Error())
		return 0, err
	}
	return lastPage, nil
}

func (s Storage) UpdateProgress(ctx context.Context, language Language, query string, lastPage int) error {
	if s.DB == nil {
		return ErrorNoDatabase
	}
	_, err := s.DB.ExecContext(ctx, sqlUpdateProgress, language.String(), query, lastPage)
	if err != nil {
		log.Debugf("Failed to update progress VALUES(%s, %s, %d): %s", language, query, lastPage, err.Error())
		return err
	}
	return err
}

func (s Storage) CodeExistsByHash(ctx context.Context, hash string) (bool, error) {
	if s.DB == nil {
		return false, ErrorNoDatabase
	}
	var exists bool
	err := s.DB.QueryRowContext(ctx, sqlCodeExists, hash).Scan(&exists)
	if err != nil {
		log.Debugf("Failed to check if code exists: %s", err.Error())
		return false, err
	}
	return exists, nil
}

func (s Storage) queryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if s.DB == nil {
		return nil
	}
	return s.DB.QueryRowContext(ctx, query, args...)
}

func (s Storage) tableExists(ctx context.Context, tableName string) (bool, error) {
	if s.DB == nil {
		return false, ErrorNoDatabase
	}
	_, err := s.DB.QueryContext(ctx, fmt.Sprintf(sqlGetFirstFromTable, tableName))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s Storage) dropTables() error {
	if s.DB == nil {
		return ErrorNoDatabase
	}
	_, err := s.DB.Exec(sqlDropTables)
	return err
}
