package main

import( 
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"snippetbox.mike9708.net/ui"
)


func (app *application) routes() http.Handler{
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { 
		app.notFound(w)
	})

	fileServer := http.FileServer(http.FS(ui.Files)) 
	router.Handler(http.MethodGet,"/static/*filepath", fileServer)

	dynamic := alice.New(app.SessionsManager.LoadAndSave, noSurf, app.authenticate)	

	router.Handler(http.MethodGet,"/", dynamic.ThenFunc(app.home))
	router.HandlerFunc(http.MethodGet,"/ping", ping)
	router.Handler(http.MethodPost,"/user/signup", dynamic.ThenFunc(app.user_signupPost))
	router.Handler(http.MethodGet,"/user/signup", dynamic.ThenFunc(app.user_signup))
	router.Handler(http.MethodGet,"/user/login", dynamic.ThenFunc(app.user_login))
	router.Handler(http.MethodPost,"/user/login", dynamic.ThenFunc(app.user_loginPost))
	router.Handler(http.MethodGet,"/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	
	protected := dynamic.Append(app.requiredAuth)

	router.Handler(http.MethodPost,"/user/logout", protected.ThenFunc(app.user_logout))
	router.Handler(http.MethodPost,"/snippet/create", protected.ThenFunc(app.snippetCreatePost))
	router.Handler(http.MethodGet,"/snippet/create", protected.ThenFunc(app.SnippetCreate))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders ) 
	return standard.Then(router) 
}
