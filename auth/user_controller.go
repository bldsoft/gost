package auth

import (
	"errors"
	"net/http"

	"github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
	"github.com/go-chi/chi/v5"
)

var ErrForbidden = errors.New("forbidden")

type UserController[PT IUserPtr[T], T any] struct {
	controller.BaseController
	service IUserService[PT, T]
}

func NewUserController[PT IUserPtr[T], T any](service IUserService[PT, T]) *UserController[PT, T] {
	return &UserController[PT, T]{service: service}
}

func (c *UserController[PT, T]) GetHandler(w http.ResponseWriter, r *http.Request) {
	archived := false
	if !controller.ParseQueryOption(r, w, controller.ArchivedQueryName, &archived) {
		return
	}
	users, err := c.service.GetAll(r.Context(), archived)
	switch {
	case err == nil:
		c.ResponseJson(w, r, users)
	case errors.Is(err, ErrForbidden):
		c.ResponseError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	default:
		log.FromContext(r.Context()).Error(err.Error())
		c.ResponseError(w, err.Error(), http.StatusBadRequest)
	}
}

func (c *UserController[PT, T]) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := c.service.GetByID(r.Context(), id)
	switch {
	case err == nil:
		c.ResponseJson(w, r, user)
	case errors.Is(err, ErrForbidden):
		c.ResponseError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	case errors.Is(err, utils.ErrObjectNotFound):
		c.ResponseError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		log.FromContext(r.Context()).Error(err.Error())
		c.ResponseError(w, err.Error(), http.StatusBadRequest)
	}
}

func (c *UserController[PT, T]) PostHandler(w http.ResponseWriter, r *http.Request) {
	var user T
	if !c.GetObjectFromBody(w, r, &user) {
		return
	}
	err := c.service.Create(r.Context(), &user, true)
	switch {
	case err == nil:
		c.ResponseJson(w, r, &user)
	case errors.Is(err, ErrForbidden):
		c.ResponseError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	default:
		log.FromContext(r.Context()).Error(err.Error())
		c.ResponseError(w, err.Error(), http.StatusBadRequest)
	}
}

func (c *UserController[PT, T]) PutHandler(w http.ResponseWriter, r *http.Request) {
	var user T
	if !c.GetObjectFromBody(w, r, &user) {
		return
	}
	PT(&user).SetIDFromString(chi.URLParam(r, "id"))
	err := c.service.Update(r.Context(), &user)
	switch {
	case err == nil:
		c.ResponseJson(w, r, &user)
	case errors.Is(err, ErrForbidden):
		c.ResponseError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	default:
		log.FromContext(r.Context()).Error(err.Error())
		c.ResponseError(w, err.Error(), http.StatusBadRequest)
	}
}

func (c *UserController[PT, T]) PasswordPutHandler(w http.ResponseWriter, r *http.Request) {
	var pass EntityPassword
	if !c.GetObjectFromBody(w, r, &pass) {
		return
	}
	err := c.service.UpdatePassword(r.Context(), chi.URLParam(r, "id"), pass.Password())
	switch {
	case err == nil:
		c.ResponseOK(w)
	case errors.Is(err, ErrForbidden):
		c.ResponseError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	default:
		log.FromContext(r.Context()).Error(err.Error())
		c.ResponseError(w, err.Error(), http.StatusBadRequest)
	}
}

func (c *UserController[PT, T]) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	archived := true
	if !controller.ParseQueryOption(r, w, controller.ArchivedQueryName, &archived) {
		return
	}
	err := c.service.Delete(r.Context(), chi.URLParam(r, "id"), archived)
	switch {
	case err == nil:
		c.ResponseOK(w)
	case errors.Is(err, utils.ErrObjectNotFound):
		c.ResponseError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	case errors.Is(err, ErrForbidden):
		c.ResponseError(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	default:
		log.FromContext(r.Context()).Error(err.Error())
		c.ResponseError(w, err.Error(), http.StatusBadRequest)
	}
}

func (c *UserController[PT, T]) Mount(r chi.Router) {
	r.Get("/", c.GetHandler)
	r.Get("/{id}", c.GetByID)
	r.Post("/", c.PostHandler)
	r.Put("/{id}", c.PutHandler)
	r.Put("/{id}/password", c.PasswordPutHandler)
	r.Delete("/{id}", c.DeleteHandler)
}
