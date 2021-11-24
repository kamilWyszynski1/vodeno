package client_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"vodeno/pkg/client"
	"vodeno/pkg/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// newTestServer returns new test server and router.
func newTestServer() (*httptest.Server, chi.Router) {
	router := chi.NewRouter()

	server := httptest.NewServer(router)
	return server, router
}

func intToPtrInt(i int) *int {
	return &i
}

func TestHandler_NewHandler(t *testing.T) {
	mock := mocks.NewMockService(gomock.NewController(t))
	handler := client.NewHandler(nil, mock)
	require.NotNil(t, handler)
}

func TestHandler_addRoute(t *testing.T) {
	log := logrus.New()
	log.Out = io.Discard

	t0 := time.Now().Truncate(time.Second)

	for _, tt := range []struct {
		name         string
		request      map[string]interface{}
		prep         func(service *mocks.MockService)
		wantedStatus int
	}{
		{
			name: "Basic",
			request: map[string]interface{}{
				"email":       "email@test.com",
				"title":       "title",
				"content":     "content",
				"mailing_id":  1,
				"insert_time": t0,
			},
			prep: func(mock *mocks.MockService) {
				mock.EXPECT().Add(gomock.Any(), client.Entry{
					ID:         0,
					Email:      "email@test.com",
					Title:      "title",
					Content:    "content",
					MailingID:  1,
					InsertTime: t0,
				})
			},
			wantedStatus: http.StatusNoContent,
		},
		{
			name: "Returns400OnBadEmail",
			request: map[string]interface{}{
				"email":       "bademail",
				"title":       "title",
				"content":     "content",
				"mailing_id":  1,
				"insert_time": t0,
			},
			prep:         func(mock *mocks.MockService) {},
			wantedStatus: http.StatusBadRequest,
		},
		{
			name: "Returns400OnMissingEmail",
			request: map[string]interface{}{
				"email":       "",
				"title":       "title",
				"content":     "content",
				"mailing_id":  1,
				"insert_time": t0,
			},
			prep:         func(mock *mocks.MockService) {},
			wantedStatus: http.StatusBadRequest,
		},
		{
			name: "Returns400OnMissingTitle",
			request: map[string]interface{}{
				"email":       "email@text.com",
				"title":       "",
				"content":     "content",
				"mailing_id":  1,
				"insert_time": t0,
			},
			prep:         func(mock *mocks.MockService) {},
			wantedStatus: http.StatusBadRequest,
		},
		{
			name: "Returns400OnMissingContent",
			request: map[string]interface{}{
				"email":       "email@text.com",
				"title":       "qwe",
				"content":     "",
				"mailing_id":  1,
				"insert_time": t0,
			},
			prep:         func(mock *mocks.MockService) {},
			wantedStatus: http.StatusBadRequest,
		},
		{
			name: "Returns400OnMissingMailingID",
			request: map[string]interface{}{
				"email":       "email@text.com",
				"title":       "qwe",
				"content":     "qwe",
				"insert_time": t0,
			},
			prep:         func(mock *mocks.MockService) {},
			wantedStatus: http.StatusBadRequest,
		},
		{
			name: "Returns400OnMissingInsertTime",
			request: map[string]interface{}{
				"email":      "email@text.com",
				"title":      "qwe",
				"content":    "qwe",
				"mailing_id": 1,
			},
			prep:         func(mock *mocks.MockService) {},
			wantedStatus: http.StatusBadRequest,
		},
		{
			name: "Returns400OnDuplicate",
			request: map[string]interface{}{
				"email":       "email@test.com",
				"title":       "title",
				"content":     "content",
				"mailing_id":  1,
				"insert_time": t0,
			},
			prep: func(mock *mocks.MockService) {
				mock.EXPECT().Add(gomock.Any(), client.Entry{
					ID:         0,
					Email:      "email@test.com",
					Title:      "title",
					Content:    "content",
					MailingID:  1,
					InsertTime: t0,
				}).Return(client.ErrDuplicate)
			},
			wantedStatus: http.StatusBadRequest,
		},
		{
			name: "Returns400InvalidRequest",
			request: map[string]interface{}{
				"field": "value",
			},
			prep:         func(mock *mocks.MockService) {},
			wantedStatus: http.StatusBadRequest,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server, router := newTestServer()
			defer server.Close()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMockService(ctrl)

			tt.prep(mock)
			handler := client.NewHandler(log, mock)
			handler.AddRoutes(router)

			b, err := json.Marshal(tt.request)
			require.NoError(t, err)

			resp, err := http.Post(
				fmt.Sprintf("%s/clients/", server.URL),
				"application/json",
				bytes.NewReader(b),
			)
			require.NoError(t, err)
			require.Equal(t, tt.wantedStatus, resp.StatusCode)
		})
	}
}

func TestHandler_listRoute(t *testing.T) {
	log := logrus.New()
	log.Out = io.Discard

	entry := client.Entry{
		ID:         1,
		Email:      "email",
		Title:      "title",
		Content:    "conten",
		MailingID:  1,
		InsertTime: time.Now(),
	}
	for _, tt := range []struct {
		name         string
		path         string
		prep         func(service *mocks.MockService)
		wantedStatus int
	}{
		{
			name: "Returns204OnEmptyList",
			prep: func(mock *mocks.MockService) {
				mock.EXPECT().List(gomock.Any(), client.Cursor{Limit: 20}).Return(nil, nil)
			},
			path:         "/clients",
			wantedStatus: http.StatusNoContent,
		},
		{
			name: "Returns200WithDefaultLimit",
			prep: func(mock *mocks.MockService) {
				mock.EXPECT().List(gomock.Any(), client.Cursor{Limit: 20}).Return([]client.Entry{entry}, nil)
			},
			path:         "/clients",
			wantedStatus: http.StatusOK,
		},
		{
			name: "Returns200WithGivenCursor",
			prep: func(mock *mocks.MockService) {
				mock.EXPECT().List(gomock.Any(), gomock.Any()).Return([]client.Entry{entry}, nil)
			},
			path:         "/clients?limit=10&after_id=2",
			wantedStatus: http.StatusOK,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server, router := newTestServer()
			defer server.Close()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMockService(ctrl)

			tt.prep(mock)
			handler := client.NewHandler(log, mock)
			handler.AddRoutes(router)

			resp, err := http.Get(server.URL + tt.path)
			require.NoError(t, err)
			require.Equal(t, tt.wantedStatus, resp.StatusCode)
		})
	}
}
