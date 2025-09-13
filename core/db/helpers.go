package db

import "github.com/jackc/pgx/v5/pgtype"

func ToPGText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}
