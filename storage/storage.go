package storage

import (
	"database/sql"
	"fmt"
	"slices"
	"sync"
	"time"

	// Required
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB

	registeredTables   []string
	muRegisteredTables sync.Mutex

	cleanupDone chan struct{}
	cleanupWg   sync.WaitGroup
}

func New(filename string) (*Storage, error) {
	db, err := sql.Open("sqlite3", filename+"?_journal=WAL&_synchronous=NORMAL")
	if err != nil {
		return nil, fmt.Errorf("storage.New: open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		// Ignore the error
		db.Close()
		return nil, fmt.Errorf("storage.New: ping database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(1 * time.Hour)

	_, err = db.Exec(`
        PRAGMA cache_size = -20000;     -- Use 20MB of memory for cache
        PRAGMA page_size = 8192;
        PRAGMA mmap_size = 134217728;   -- Memory map up to 128MB
        PRAGMA journal_mode = WAL;    	-- Use Write-Ahead Logging for better concurrency
        PRAGMA synchronous = NORMAL;
        PRAGMA busy_timeout = 5000;     -- 5 seconds
        PRAGMA auto_vacuum = INCREMENTAL;
    `)
	if err != nil {
		// Ignore the error
		db.Close()
		return nil, fmt.Errorf("storage.New: configure database: %w", err)
	}

	s := &Storage{
		db:               db,
		registeredTables: nil,
		cleanupDone:      make(chan struct{}),
	}

	s.cleanupWg.Add(1)
	go s.startCleanup()

	return s, nil
}

func (s *Storage) startCleanup() {
	defer s.cleanupWg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.cleanupDone:
			return
		case <-ticker.C:
			s.muRegisteredTables.Lock()
			tableNames := slices.Clone(s.registeredTables)
			s.muRegisteredTables.Unlock()

			for _, tableName := range tableNames {
				expired, err := s.cleanupExpired(tableName)
				fmt.Printf("storage.cleanupExpired: table = %s, count = %d, error = %+v\n", tableName, expired, err)
			}
		}
	}
}

func (s *Storage) Close() error {
	close(s.cleanupDone)
	s.cleanupWg.Wait()

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("storage.Close: close database: %w", err)
	}
	return nil
}

func (s *Storage) registerTable(tableName string) {
	s.muRegisteredTables.Lock()
	defer s.muRegisteredTables.Unlock()

	if !slices.Contains(s.registeredTables, tableName) {
		s.registeredTables = append(s.registeredTables, tableName)
	}
}

func (s *Storage) cleanupExpired(tableName string) (int, error) {
	query := fmt.Sprintf(
		`
			DELETE FROM %s
			WHERE expires_at != 0 AND expires_at <= ?
        `,
		tableName,
	)

	res, err := s.db.Exec(query, nowUnix())
	if err != nil {
		return 0, fmt.Errorf("storage.cleanupExpired: clean expired keys for table %s: %w", tableName, err)
	}

	affectedCount, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("storage.cleanupExpired: affected count: %w", err)
	}
	return int(affectedCount), nil
}
