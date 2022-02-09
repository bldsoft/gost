package auth

import (
	"errors"
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

type testRole = int

const (
	user testRole = iota
	admin
)

type testUser struct {
	role testRole
}

func (u *testUser) Role() testRole {
	return u.role
}

func TestAuthMiddlewares(t *testing.T) {
	adm := &User[testRole]{Creds{"admin", "password"}, EntityRole[testRole]{admin}}
	reg := &User[testRole]{Creds{"user", "password"}, EntityRole[testRole]{user}}

	testCases := []struct {
		name         string
		user         *User[testRole]
		allowedRoles []testRole
		expectedCode int
	}{
		{"Admin, all allowed", adm, []testRole{admin, user}, http.StatusOK},
		{"User, all allowed", reg, []testRole{admin, user}, http.StatusOK},
		{"Admin, only admins allowed", adm, []testRole{admin}, http.StatusOK},
		{"User, only admins allowed", reg, []testRole{admin}, http.StatusForbidden},
		{"Unauthorized, only admins allowed", nil, []testRole{admin}, http.StatusUnauthorized},
		{"User,  authorization is off", reg, nil, http.StatusOK},
		{"Unauthorized, authorization is off", nil, nil, http.StatusUnauthorized},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			store := mocks.NewMockStore(mockCtrl)
			authController := NewAuthController[*User[testRole]](store, nil, "")

			r := chi.NewRouter()
			r.Use(authController.AuthenticateMiddleware())
			if testCase.allowedRoles != nil {
				r.Use(AuthorizationMiddleware(testCase.allowedRoles...))
			}
			r.Get("/ping", controller.GetPingHandler)

			req, err := http.NewRequest("GET", "/ping", nil)
			assert.NoError(t, err)
			if testCase.user != nil {
				session := sessions.NewSession(store, "")
				session.Values[SessionUserKey] = *testCase.user
				store.EXPECT().Get(gomock.Any(), gomock.Any()).Return(session, nil).AnyTimes()
			} else {
				store.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("")).AnyTimes()
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedCode, w.Code)
		})
	}

}
