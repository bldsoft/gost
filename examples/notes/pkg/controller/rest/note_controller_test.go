package rest_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/bldsoft/gost/examples/notes/pkg/controller/rest"
	"github.com/bldsoft/gost/examples/notes/pkg/entity"
	"github.com/bldsoft/gost/repository"
)

type fakeNoteService struct {
	create func(note *entity.Note) error
	get    func(id string, archived bool) (*entity.Note, error)
	list   func(limit int64, archived bool) ([]*entity.Note, error)
	update func(id string, note *entity.Note) error
	delete func(id string) error
}

func (f *fakeNoteService) Create(_ context.Context, note *entity.Note) error { return f.create(note) }
func (f *fakeNoteService) Get(_ context.Context, id string, archived bool) (*entity.Note, error) {
	return f.get(id, archived)
}
func (f *fakeNoteService) List(_ context.Context, limit int64, archived bool) ([]*entity.Note, error) {
	return f.list(limit, archived)
}
func (f *fakeNoteService) Update(_ context.Context, id string, note *entity.Note) error {
	return f.update(id, note)
}
func (f *fakeNoteService) Delete(_ context.Context, id string) error { return f.delete(id) }

func newRouter(svc *fakeNoteService) http.Handler {
	r := chi.NewRouter()
	r.Route("/notes", rest.NewNoteController(svc).Mount)

	return r
}

func do(t *testing.T, h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, target, reader)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

func TestNoteController_Create(t *testing.T) {
	t.Run("returns the created note", func(t *testing.T) {
		svc := &fakeNoteService{create: func(note *entity.Note) error {
			note.Title = "stored"

			return nil
		}}

		rec := do(t, newRouter(svc), http.MethodPost, "/notes/", `{"title":"hello"}`)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		require.Contains(t, rec.Body.String(), `"title":"stored"`)
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		svc := &fakeNoteService{create: func(*entity.Note) error { t.Fatal("service must not be called"); return nil }}

		rec := do(t, newRouter(svc), http.MethodPost, "/notes/", `{"title":`)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("maps ErrInvalidNote to 400", func(t *testing.T) {
		svc := &fakeNoteService{create: func(*entity.Note) error { return entity.ErrInvalidNote }}

		rec := do(t, newRouter(svc), http.MethodPost, "/notes/", `{}`)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), entity.ErrInvalidNote.Error())
	})
}

func TestNoteController_Get(t *testing.T) {
	t.Run("maps ErrNotFound to 404", func(t *testing.T) {
		svc := &fakeNoteService{get: func(string, bool) (*entity.Note, error) { return nil, repository.ErrNotFound }}

		rec := do(t, newRouter(svc), http.MethodGet, "/notes/abc", "")
		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("passes id and archived flag", func(t *testing.T) {
		svc := &fakeNoteService{get: func(id string, archived bool) (*entity.Note, error) {
			require.Equal(t, "abc", id)
			require.True(t, archived)

			return &entity.Note{Title: "t"}, nil
		}}

		rec := do(t, newRouter(svc), http.MethodGet, "/notes/abc?archived=true", "")
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("maps unknown errors to 500", func(t *testing.T) {
		svc := &fakeNoteService{get: func(string, bool) (*entity.Note, error) { return nil, errors.New("boom") }}

		rec := do(t, newRouter(svc), http.MethodGet, "/notes/abc", "")
		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.NotContains(t, rec.Body.String(), "boom")
	})
}

func TestNoteController_List(t *testing.T) {
	t.Run("uses the default limit", func(t *testing.T) {
		svc := &fakeNoteService{list: func(limit int64, archived bool) ([]*entity.Note, error) {
			require.EqualValues(t, 50, limit)
			require.False(t, archived)

			return []*entity.Note{}, nil
		}}

		rec := do(t, newRouter(svc), http.MethodGet, "/notes/", "")
		require.Equal(t, http.StatusOK, rec.Code)
		require.JSONEq(t, `[]`, rec.Body.String())
	})

	t.Run("rejects a non-numeric limit", func(t *testing.T) {
		svc := &fakeNoteService{}

		rec := do(t, newRouter(svc), http.MethodGet, "/notes/?limit=many", "")
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestNoteController_Delete(t *testing.T) {
	svc := &fakeNoteService{delete: func(id string) error {
		require.Equal(t, "abc", id)

		return nil
	}}

	rec := do(t, newRouter(svc), http.MethodDelete, "/notes/abc", "")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "OK", rec.Body.String())
}
