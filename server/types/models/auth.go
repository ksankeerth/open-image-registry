package models

type RevokedToken struct {
	SignatureHash string
	ExpiresAt     int64
	IssuedAt      int64
	UserID        string
}