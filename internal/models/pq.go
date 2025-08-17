package models

import "time"

type User struct {
	ID        int64
	Login     string
	PassHash  []byte
	CreatedAt time.Time
}

type Document struct {
	ID          string
	OwnerID     int64
	Name        string
	Mime        string
	HashFile    bool
	Public      bool
	JsonDate    []byte
	StoragePath string
	CreatedAt   time.Time
	IsDeleted   bool
}

type Grands struct {
	DocID  string
	UserID int64
}

type DocumentResponse struct {
	Doc    Document
	Grants []string
}
