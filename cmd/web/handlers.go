package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"snippetbox.mike9708.net/internal/models"
	"snippetbox.mike9708.net/internal/validator"
	"strconv"
)

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

type userSignUpForm struct {
	Email               string `form:"email"`
	Name                string `form:"name"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ok"))
}

func (app *application) user_login(w http.ResponseWriter, r *http.Request) {
	template_data := app.newTemplateData(r)
	template_data.Form = userSignUpForm{}
	app.render(w, http.StatusOK, "login.tmpl", template_data)
}

func (app *application) user_loginPost(w http.ResponseWriter, r *http.Request) {
	var login_user userLoginForm
	err := app.decodePostForm(r, &login_user)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	login_user.CheckFields(login_user.CheckBlanck(login_user.Password), "content", "This field cannot be blank")
	login_user.CheckFields(login_user.CheckBlanck(login_user.Email), "title", "This field cannot be blank")
	login_user.CheckFields(login_user.Matches(login_user.Email, validator.EmailRX), "Email", "The email is not valid")

	if !login_user.Valid() {
		data := app.newTemplateData(r)
		data.Form = login_user
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}
	id, err := app.users.Authenticate(login_user.Email, login_user.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			login_user.AddNonFieldError("Invalid credentials for email or password")
			data := app.newTemplateData(r)
			data.Form = login_user
			app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			app.serverError(w, err)
			return
		}
	}
	err = app.SessionsManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.SessionsManager.Put(r.Context(), "authenticatedUserID", id)
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) user_logout(w http.ResponseWriter, r *http.Request) {
	err := app.SessionsManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.SessionsManager.Remove(r.Context(), "AuthUserId")
	app.SessionsManager.Put(r.Context(), "flash", "You sucessfully logged out")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) user_signup(w http.ResponseWriter, r *http.Request) {
	template_data := app.newTemplateData(r)
	template_data.Form = userSignUpForm{}
	app.render(w, http.StatusOK, "signup.tmpl", template_data)
}

func (app *application) user_signupPost(w http.ResponseWriter, r *http.Request) {
	var create_user userSignUpForm
	err := app.decodePostForm(r, &create_user)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	create_user.CheckFields(create_user.CheckBlanck(create_user.Name), "title", "This field cannot be blank")
	create_user.CheckFields(create_user.CheckBlanck(create_user.Password), "content", "This field cannot be blank")
	create_user.CheckFields(create_user.CheckBlanck(create_user.Email), "title", "This field cannot be blank")
	create_user.CheckFields(create_user.Matches(create_user.Email, validator.EmailRX), "Email", "The email is not valid")
	create_user.CheckFields(create_user.MinChars(create_user.Password, 8), "Password", "This field cannot be more than 8 chararcters")

	app.infoLog.Println(create_user.FieldsErr)

	if !create_user.Valid() {
		data := app.newTemplateData(r)
		data.Form = create_user
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}
	err = app.users.Insert(create_user.Name, create_user.Email, create_user.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			create_user.AddFieldError("Email", "Email address allready in use")
			data := app.newTemplateData(r)
			data.Form = create_user
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.SessionsManager.Put(r.Context(), "flash", "Signed sucessfuly, please login")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}
	template_data := app.newTemplateData(r)
	template_data.Snippets = snippets
	app.render(w, http.StatusOK, "home.tmpl", template_data)

}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
			return
		}
		app.serverError(w, err)
		return
	}
	template_data := app.newTemplateData(r)
	template_data.Snippet = snippet
	app.render(w, http.StatusOK, "view.tmpl", template_data)
}

func (app *application) SnippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = snippetCreateForm{
		Expires: 365,
	}
	app.render(w, http.StatusOK, "create.tmpl", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(405)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	var create_form snippetCreateForm
	err := app.decodePostForm(r, &create_form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	create_form.CheckFields(create_form.CheckBlanck(create_form.Title), "title", "This field cannot be blank")
	create_form.CheckFields(create_form.CheckBlanck(create_form.Content), "content", "This field cannot be blank")
	create_form.CheckFields(create_form.CheckLen(create_form.Title, 100), "title", "This field cannot be more than 100 chararcters")
	create_form.CheckFields(validator.PermittedValues(create_form.Expires, 1, 7, 365), "expires", "This field must be equal to 1, 7 or 365")

	if !create_form.Valid() {
		data := app.newTemplateData(r)
		data.Form = create_form
		app.render(w, http.StatusOK, "create.tmpl", data)
		return
	}

	id, err := app.snippets.Insert(create_form.Title, create_form.Content, create_form.Expires)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.SessionsManager.Put(r.Context(), "flash", "Snippet Successfully created")
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
