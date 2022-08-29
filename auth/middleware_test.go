package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bldsoft/gost/auth/mocks"
	"github.com/bldsoft/gost/controller"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/sessions"

	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	reg := &User{Creds{"user", EntityPassword{"password"}}}
	testCases := []struct {
		name         string
		user         *User
		expectedCode int
	}{
		{"Authorized", reg, http.StatusOK},
		{"Unauthorized", nil, http.StatusOK},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			store := mocks.NewMockStore(mockCtrl)
			authController := NewAuthController[*User](nil, store, "")

			r := chi.NewRouter()
			r.Use(authController.AuthenticateMiddleware())
			r.Get("/ping", controller.GetPingHandler)

			req, err := http.NewRequest("GET", "/ping", nil)
			assert.NoError(t, err)
			session := sessions.NewSession(store, "")
			if testCase.user != nil {
				session.Values[SessionUserKey] = *testCase.user
				store.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any())
			}
			store.EXPECT().Get(gomock.Any(), gomock.Any()).Return(session, nil).AnyTimes()
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedCode, w.Code)
		})
	}

}
