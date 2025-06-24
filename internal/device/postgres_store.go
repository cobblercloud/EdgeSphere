package device

import (
	"context"
	"database/sql"
	"time"
	
	_ "github.com/lib/pq"
	"edgesphere/internal/pkg/types"
)

const (
	createTableSQL = `
	CREATE TABLE IF NOT EXISTS devices (
		id VARCHAR(64) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(50) NOT NULL,
		status SMALLINT NOT NULL DEFAULT 0,
		gateway_id VARCHAR(64),
		last_seen TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		metadata JSONB
	);
	
	CREATE INDEX IF NOT EXISTS idx_gateway_id ON devices(gateway_id);
	CREATE INDEX IF NOT EXISTS idx_status ON devices(status);`
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	
	// 初始化表结构
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}
	
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Save(ctx context.Context, device *types.Device) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO devices 
		(id, name, type, status, gateway_id, last_seen, metadata) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			status = EXCLUDED.status,
			gateway_id = EXCLUDED.gateway_id,
			last_seen = EXCLUDED.last_seen,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()`,
		device.ID, device.Name, device.Type, device.Status, 
		device.GatewayID, device.LastSeen, device.Metadata)
	return err
}

func (s *PostgresStore) BatchSave(devices []*types.Device) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	stmt, err := tx.Prepare(`
		INSERT INTO devices 
		(id, name, type, status, gateway_id, last_seen, metadata) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	
	for _, d := range devices {
		_, err := stmt.Exec(d.ID, d.Name, d.Type, d.Status, 
			d.GatewayID, d.LastSeen, d.Metadata)
		if err != nil {
			return err
		}
	}
	
	return tx.Commit()
}

func (s *PostgresStore) UpdateStatus(ctx context.Context, id string, status types.DeviceStatus) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE devices SET 
			status = $1,
			updated_at = NOW()
		WHERE id = $2`, status, id)
	return err
}

func (s *PostgresStore) FindByGateway(gatewayID string) ([]*types.Device, error) {
	rows, err := s.db.Query(`
		SELECT id, name, type, status, last_seen, metadata 
		FROM devices 
		WHERE gateway_id = $1`, gatewayID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var devices []*types.Device
	for rows.Next() {
		var d types.Device
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Type, &d.Status, &d.LastSeen, &d.Metadata,
		); err != nil {
			return nil, err
		}
		devices = append(devices, &d)
	}
	return devices, nil
}

// 分页查询
func (s *PostgresStore) List(limit, offset int) ([]*types.Device, error) {
	rows, err := s.db.Query(`
		SELECT id, name, type, status, gateway_id, last_seen, created_at 
		FROM devices 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var devices []*types.Device
	for rows.Next() {
		var d types.Device
		if err := rows.Scan(
			&d.ID, &d.Name, &d.Type, &d.Status, 
			&d.GatewayID, &d.LastSeen, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		devices = append(devices, &d)
	}
	return devices, nil
}