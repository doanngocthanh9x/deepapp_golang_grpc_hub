package repository

import (
	"database/sql"

	"deepapp_golang_grpc_hub/internal/models"
)

type ClientsRepo struct {
	db *sql.DB
}

func NewClientsRepo(db *sql.DB) *ClientsRepo {
	return &ClientsRepo{db: db}
}

func (r *ClientsRepo) Save(client *models.Client) error {
	_, err := r.db.Exec("INSERT OR REPLACE INTO clients (id, status, last_seen) VALUES (?, ?, ?)",
		client.ID, client.Status, client.LastSeen)
	return err
}

func (r *ClientsRepo) GetByID(id string) (*models.Client, error) {
	var client models.Client
	err := r.db.QueryRow("SELECT id, status, last_seen FROM clients WHERE id = ?", id).
		Scan(&client.ID, &client.Status, &client.LastSeen)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *ClientsRepo) GetAll() ([]*models.Client, error) {
	rows, err := r.db.Query("SELECT id, status, last_seen FROM clients")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*models.Client
	for rows.Next() {
		var client models.Client
		err := rows.Scan(&client.ID, &client.Status, &client.LastSeen)
		if err != nil {
			return nil, err
		}
		clients = append(clients, &client)
	}
	return clients, nil
}