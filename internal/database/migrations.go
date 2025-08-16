package database

import (
	"database/sql"
)

func Migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id VARCHAR(36) PRIMARY KEY,
			content TEXT NOT NULL,
			type VARCHAR(20) DEFAULT 'text',
			user_id VARCHAR(100),
			username VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON chat_messages(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_user_id ON chat_messages(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_type ON chat_messages(type)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}