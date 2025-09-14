package db

import "github.com/jackc/pgx/v5/pgtype"

func ToPGText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func ToPGInt4(i int32) pgtype.Int4 {
	if i == 0 {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: i, Valid: true}
}
