package repository

import (
	"database/sql"
	"errors"
)

type Sub struct {
	db *sql.DB
}

func NewSubRepository(db *sql.DB) *Sub {
	return &Sub{db: db}
}

func (r *Sub) AddSubscription(userID int64, channel string) error {
	insertQuery := `
          INSERT INTO subscriptions (user_id, channel_username)
          VALUES ($1, $2)
          ON CONFLICT (user_id, channel_username) DO NOTHING`

	_, err := r.db.Exec(insertQuery, userID, channel)
	return err
}

func (r *Sub) RemoveSubscription(userID int64, channel string) error {
	removeQuery := `DELETE FROM subscriptions WHERE user_id = $1 AND channel_username = $2`
	_, err := r.db.Exec(removeQuery, userID, channel)
	return err
}

func (r *Sub) GetSubscribers(channel string) ([]int64, error) {
	selectQuery := `SELECT user_id FROM subscriptions WHERE channel_username = $1`

	rows, err := r.db.Query(selectQuery, channel)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0)

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (r *Sub) GetChannelLastPostId(channel string) int {
	selectQuery := `SELECT last_post_id FROM channel_states WHERE channel_username = $1`

	row := r.db.QueryRow(selectQuery, channel)

	var lastPostId int

	err := row.Scan(&lastPostId)
	if errors.Is(err, sql.ErrNoRows) {
		return 0
	}

	return lastPostId
}

func (r *Sub) AddChannel(channel string, lastPostId int) error {
	insertQuery := `INSERT INTO channel_states (channel_username, last_post_id) VALUES ($1, $2)`

	_, err := r.db.Exec(insertQuery, channel, lastPostId)
	return err
}

func (r *Sub) RemoveChannel(channel string) error {
	deleteQuery := `DELETE FROM channel_states WHERE channel_username = $1`

	_, err := r.db.Exec(deleteQuery, channel)
	return err
}

func (r *Sub) GetAllChannels() ([]string, error) {
	selectQuery := `SELECT channel_username FROM channel_states`
	rows, err := r.db.Query(selectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	channels := make([]string, 0)

	for rows.Next() {
		var channel string
		if err := rows.Scan(&channel); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

func (r *Sub) UpdateLastPostId(channel string, id int) error {
	updateQuery := `UPDATE channel_states SET last_post_id = $2 WHERE channel_username = $1`

	_, err := r.db.Exec(updateQuery, channel, id)
	return err
}
