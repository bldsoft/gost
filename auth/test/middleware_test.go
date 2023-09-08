package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bldsoft/gost/auth"
	"github.com/bldsoft/gost/auth/mocks"
	"github.com/bldsoft/gost/controller"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthMiddleware(t *testing.T) {
	reg := &auth.User{auth.Creds{"user", auth.EntityPassword{UserPassword: "password"}}}
	changePassReg := &auth.User{auth.Creds{"user", auth.EntityPassword{UserPassword: "password", ChangePassword: true}}}
	testCases := []struct {
		name         string
		user         *auth.User
		expectedCode int
	}{
		{"Authorized", reg, http.StatusOK},
		{"ChangePass", changePassReg, http.StatusOK},
		{"Unauthorized", nil, http.StatusOK},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := mocks.NewStore(t)
			authController := auth.NewAuthController[*auth.User](nil, store, "")

			r := chi.NewRouter()
			r.Use(authController.AuthenticateMiddleware())
			r.Get("/ping", controller.GetPingHandler)

			req, err := http.NewRequest("GET", "/ping", nil)
			assert.NoError(t, err)
			session := sessions.NewSession(store, "")
			if testCase.user != nil {
				session.Values[auth.SessionUserKey] = *testCase.user
				store.EXPECT().Save(mock.Anything, mock.Anything, mock.Anything).Return(nil)
			}
			store.EXPECT().Get(mock.Anything, mock.Anything).Return(session, nil).Maybe()
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedCode, w.Code)
		})
	}
}
