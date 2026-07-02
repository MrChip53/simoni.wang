package main

import "github.com/jackc/pgx/v5/pgtype"

func PointerStringToPgText(str *string) pgtype.Text {
	if str == nil {
		return pgtype.Text{Valid: false}
	}

	return pgtype.Text{String: *str, Valid: true}
}

func PgTextToPointerString(text pgtype.Text) *string {
	if !text.Valid {
		return nil
	}

	return &text.String
}
