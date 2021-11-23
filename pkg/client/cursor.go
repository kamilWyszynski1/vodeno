package client

import (
	"net/http"
	"strconv"
)

const defaultLimit = 20

// Cursor represents a cursor to query a list of clients.
//
// Cursor takes limit and after_id fields.
// First request should be sent with ony limit field.
// After receiving first data, user will send another request
// with limit and after_id which is last id of received previously data.
type Cursor struct {
	Limit   int  `json:"limit"`
	AfterID *int `json:"after_id"`
}

func CursorFromRequest(r *http.Request) (*Cursor, error) {
	var (
		cursor Cursor
		err    error
	)

	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	limit := r.Form.Get("limit")
	if limit == "" {
		cursor.Limit = defaultLimit
	} else {
		cursor.Limit, err = strconv.Atoi(limit)
		if err != nil {
			return nil, err
		}
	}

	afterIDStr := r.Form.Get("after_id")
	if afterIDStr != "" {
		afterID, err := strconv.Atoi(afterIDStr)
		if err != nil {
			return nil, err
		}
		cursor.AfterID = &afterID
	}
	return &cursor, nil
}
