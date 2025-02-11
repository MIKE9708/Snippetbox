package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
	"github.com/justinas/nosurf"
	"github.com/go-playground/form"
)

func (app *application) decodePostForm(r *http.Request, dst any) error{
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.FormDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecodeError *form.InvalidDecoderError
		if errors.As(err,&invalidDecodeError) {
			panic(err)
		}
		return err
	}
	return nil
}


func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

func (app *application) newTemplateData(r *http.Request) *TemplateData{
	return &TemplateData{
		CurrentYear: time.Now().Year(),
		Flash: app.SessionsManager.PopString(r.Context(), "flash"),
		IsAuth: app.isAuthenticated(r),
		CSRFToken: nosurf.Token(r),
	}
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s",err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status_error int) {
	http.Error(w, http.StatusText(status_error), status_error)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *TemplateData){
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("Page %s was not found for render", page)
		app.serverError(w, err)
		return 
	}
	buff := new (bytes.Buffer)

	err := ts.ExecuteTemplate(buff, "base", data)
	if err != nil {
		app.serverError(w, err)
	}
	w.WriteHeader(status)
	buff.WriteTo(w)
}
