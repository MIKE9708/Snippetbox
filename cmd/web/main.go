package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
	"path/filepath"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-playground/form"
	"snippetbox.mike9708.net/internal/models"
	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/mysqlstore"
)

// users, snippets : as long as an object has the methods to satisfy the interface
// is possible to use them in application

type application struct {
	errorLog *log.Logger
	infoLog *log.Logger
	snippets models.SnippetModelInterface
	users models.UserModelInterface
	templateCache map[string]*template.Template
	FormDecoder *form.Decoder
	SessionsManager *scs.SessionManager
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w,r, filepath.Clean("./ui/"))
}

func main() {
	tls_config := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "root:my-secret-pw@tcp(localhost:3307)/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()
	
	info_log := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	error_log := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	
	db, err := openDB(*dsn)
	
	if err != nil {
		error_log.Fatal(err)

	}
	defer db.Close()
	
	template_cache, err := newTemplateCache()
	if err != nil {
		error_log.Fatal(err)
	}
	
	formDecoder := form.NewDecoder()
	session_manager := scs.New()
	session_manager.Store = mysqlstore.New(db)
	session_manager.Lifetime = 12 * time.Hour
	session_manager.Cookie.Secure = true 

	app := &application{
		errorLog: error_log,
		infoLog:  info_log,
		snippets: &models.SnippetModel{DB: db},
		users: &models.UserModel{DB: db},
		templateCache: template_cache, 
		FormDecoder: formDecoder,
		SessionsManager: session_manager,
	}
	
	srv := &http.Server{
		Addr : *addr,
		ErrorLog: error_log,
		Handler: app.routes(),
		TLSConfig: tls_config,
		IdleTimeout: time.Minute,
		ReadTimeout: time.Second * 5,
		WriteTimeout: time.Second * 10,
	}

	info_log.Printf("Starting Server on: %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	error_log.Fatal(err)
}

func openDB(dsn string)(*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

