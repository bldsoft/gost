package rest

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/examples/notes/pkg/controller"
	"github.com/bldsoft/gost/examples/notes/pkg/entity"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
)

const defaultLimit int64 = 50

type NoteController struct {
	gost.BaseController
	notes controller.NoteService
}

func NewNoteController(notes controller.NoteService) *NoteController {
	return &NoteController{notes: notes}
}

func (c *NoteController) Mount(r chi.Router) {
	r.Get("/", c.List)
	r.Post("/", c.Create)
	r.Get("/{id}", c.Get)
	r.Put("/{id}", c.Update)
	r.Delete("/{id}", c.Delete)
}

func (c *NoteController) List(w http.ResponseWriter, r *http.Request) {
	limit := defaultLimit
	if !gost.ParseQueryOption(r, w, "limit", &limit) {
		return
	}
	var archived bool
	if !gost.ParseQueryOption(r, w, gost.ArchivedQueryName, &archived) {
		return
	}

	notes, err := c.notes.List(r.Context(), limit, archived)
	if err != nil {
		c.responseError(w, r, err)

		return
	}
	c.ResponseJson(w, r, notes)
}

func (c *NoteController) Create(w http.ResponseWriter, r *http.Request) {
	var note entity.Note
	if !c.GetObjectFromBody(w, r, &note) {
		return
	}
	if err := c.notes.Create(r.Context(), &note); err != nil {
		c.responseError(w, r, err)

		return
	}
	c.ResponseJson(w, r, note)
}

func (c *NoteController) Get(w http.ResponseWriter, r *http.Request) {
	var archived bool
	if !gost.ParseQueryOption(r, w, gost.ArchivedQueryName, &archived) {
		return
	}

	note, err := c.notes.Get(r.Context(), chi.URLParam(r, "id"), archived)
	if err != nil {
		c.responseError(w, r, err)

		return
	}
	c.ResponseJson(w, r, note)
}

func (c *NoteController) Update(w http.ResponseWriter, r *http.Request) {
	var note entity.Note
	if !c.GetObjectFromBody(w, r, &note) {
		return
	}
	if err := c.notes.Update(r.Context(), chi.URLParam(r, "id"), &note); err != nil {
		c.responseError(w, r, err)

		return
	}
	c.ResponseJson(w, r, note)
}

func (c *NoteController) Delete(w http.ResponseWriter, r *http.Request) {
	if err := c.notes.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		c.responseError(w, r, err)

		return
	}
	c.ResponseOK(w)
}

func (c *NoteController) responseError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		c.ResponseError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	case errors.Is(err, entity.ErrInvalidNote):
		c.ResponseError(w, err.Error(), http.StatusBadRequest)
	default:
		log.FromContext(r.Context()).ErrorWithFields(log.Fields{"error": err}, "Note request failed")
		c.ResponseError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
