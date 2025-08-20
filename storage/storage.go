package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "modernc.org/sqlite"
)

type Storage struct {
	db *bun.DB
}

type Note struct {
	ID        uint      `bun:"id,pk,autoincrement"` // Автоинкрементный первичный ключ
	UserID    int64     `bun:"user_id"`
	Type      string    `bun:"type"`
	Text      string    `bun:"text"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"` // Автоматическая установка времени
}

func NewStorage(dbPath string) (*Storage, error) {
	// Открываем соединение с SQLite через modernc.org/sqlite
	sqldb, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Создаём Bun DB
	db := bun.NewDB(sqldb, sqlitedialect.New())

	// Создаём таблицу, если она не существует
	ctx := context.Background()
	_, err = db.NewCreateTable().Model((*Note)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) AddNote(userID int64, noteType, text string) (int64, error) {
	ctx := context.Background()
	note := &Note{
		UserID:    userID,
		Type:      noteType,
		Text:      text,
		CreatedAt: time.Now(), // Устанавливаем текущее время
	}

	// Вставляем запись
	_, err := s.db.NewInsert().Model(note).Exec(ctx)
	if err != nil {
		return 0, err
	}

	return int64(note.ID), nil
}

func (s *Storage) GetNotes(userID int64, noteType string) ([]Note, error) {
	ctx := context.Background()
	var notes []Note

	// Строим запрос
	query := s.db.NewSelect().Model(&notes).Where("user_id = ?", userID)
	if noteType != "" {
		query = query.Where("type = ?", noteType)
	}

	// Выполняем запрос с сортировкой по created_at
	err := query.Order("created_at DESC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	return notes, nil
}

func (s *Storage) DeleteNote(userID, id int64) error {
	ctx := context.Background()

	// Удаляем заметку по ID и UserID
	result, err := s.db.NewDelete().Model((*Note)(nil)).Where("id = ? AND user_id = ?", id, userID).Exec(ctx)
	if err != nil {
		return err
	}

	// Проверяем, была ли удалена запись
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows // Используем sql.ErrNoRows вместо bun.ErrNoRows
	}

	return nil
}

func (s *Storage) DeleteAllByType(userID int64, noteType string) error {
	ctx := context.Background()

	// Удаляем все заметки определённого типа для пользователя
	_, err := s.db.NewDelete().Model((*Note)(nil)).Where("user_id = ? AND type = ?", userID, noteType).Exec(ctx)
	return err
}
