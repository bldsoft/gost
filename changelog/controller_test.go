package changelog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type stubChangeLogService struct {
	params *RecordsParams
}

func (s *stubChangeLogService) GetRecords(_ context.Context, params *RecordsParams) (*Records, error) {
	s.params = params

	return &Records{}, nil
}

func (s *stubChangeLogService) FindByID(context.Context, string) (*Record, error) {
	return nil, nil
}

func (s *stubChangeLogService) FindByIDs(context.Context, []string, bool) ([]*Record, error) {
	return nil, nil
}

func TestControllerMountClampsLimitAndOffset(t *testing.T) {
	tests := []struct {
		name   string
		target string
		limit  int64
		offset int64
	}{
		{name: "omitted limit", target: "/", limit: 100, offset: 0},
		{name: "negative limit and offset", target: "/?limit=-10&offset=-5", limit: 100, offset: 0},
		{name: "huge limit", target: "/?limit=2147483647", limit: 1000, offset: 0},
		{name: "valid limit and offset", target: "/?limit=25&offset=10", limit: 25, offset: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &stubChangeLogService{}
			router := chi.NewRouter()
			NewControllerByService(svc).Mount(router)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.target, nil))

			require.Equal(t, http.StatusOK, rec.Code)
			require.NotNil(t, svc.params)
			require.Equal(t, tt.limit, svc.params.Limit)
			require.Equal(t, tt.offset, svc.params.Offset)
		})
	}
}
