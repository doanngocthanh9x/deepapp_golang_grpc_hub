package repository

import (
	"database/sql"

	"deepapp_golang_grpc_hub/internal/models"
)

type MessagesRepo struct {
	db *sql.DB
}

func NewMessagesRepo(db *sql.DB) *MessagesRepo {
	return &MessagesRepo{db: db}
}

func (r *MessagesRepo) Save(msg *models.Message) error {
	_, err := r.db.Exec("INSERT INTO messages (id, from_client, to_client, channel, content, timestamp) VALUES (?, ?, ?, ?, ?, ?)",
		msg.ID, msg.From, msg.To, msg.Channel, msg.Content, msg.Timestamp)
	return err
}

func (r *MessagesRepo) GetByID(id string) (*models.Message, error) {
	var msg models.Message
	err := r.db.QueryRow("SELECT id, from_client, to_client, channel, content, timestamp FROM messages WHERE id = ?", id).
		Scan(&msg.ID, &msg.From, &msg.To, &msg.Channel, &msg.Content, &msg.Timestamp)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *MessagesRepo) GetAll() ([]*models.Message, error) {
	rows, err := r.db.Query("SELECT id, from_client, to_client, channel, content, timestamp FROM messages")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.From, &msg.To, &msg.Channel, &msg.Content, &msg.Timestamp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}