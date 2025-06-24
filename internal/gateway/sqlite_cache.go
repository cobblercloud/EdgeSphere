package gateway

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	
	_ "github.com/mattn/go-sqlite3"
	"edgesphere/internal/pkg/types"
)

type SQLiteCache struct {
	db  *sql.DB
	mu  sync.Mutex
}

func NewSQLiteCache(path string) (*SQLiteCache, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	
	// 创建表结构
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS sessions (
		device_id TEXT PRIMARY KEY,
		connection BLOB,
		expires_at DATETIME
	);
	
	CREATE TABLE IF NOT EXISTS commands (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT,
		command BLOB,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
	
	return &SQLiteCache{db: db}, err
}

func (c *SQLiteCache) SaveSession(deviceID string, conn *types.DeviceConnection) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	data, err := json.Marshal(conn)
	if err != nil {
		return err
	}
	
	_, err = c.db.Exec(`
		INSERT OR REPLACE INTO sessions (device_id, connection, expires_at)
		VALUES (?, ?, datetime('now','+7 days'))`,
		deviceID, data)
	return err
}

func (c *SQLiteCache) SaveCommand(deviceID string, cmd []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	_, err := c.db.Exec(`
		INSERT INTO commands (device_id, command) 
		VALUES (?, ?)`, deviceID, cmd)
	return err
}

func (c *SQLiteCache) GetCommands(deviceID string) ([][]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	rows, err := c.db.Query(`
		SELECT command FROM commands 
		WHERE device_id = ? 
		ORDER BY created_at ASC`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var commands [][]byte
	for rows.Next() {
		var cmd []byte
		if err := rows.Scan(&cmd); err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}
	
	// 删除已读取命令
	_, _ = c.db.Exec("DELETE FROM commands WHERE device_id = ?", deviceID)
	return commands, nil
}

// 清理过期会话
func (c *SQLiteCache) Cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		c.db.Exec("DELETE FROM sessions WHERE expires_at < datetime('now')")
	}
}