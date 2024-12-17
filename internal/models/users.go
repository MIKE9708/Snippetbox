package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"     // New import
)

type UserModelInterface interface {
	Insert(name, email, password string) error 
	Authenticate(email, password string) (int, error) 
	Exists(id int) (bool, error) 
}

type User struct {
	ID           int
	Name         string
	Email        string
	HashedPasswd []byte
	Created      time.Time
}

type UserModel struct {
	DB *sql.DB
}


func(m *UserModel) Insert(name, email, password string) error {
	hash_password, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return err
    }
	stmt := "INSERT INTO users(name, email, hashed_pw, created) VALUES (?, ?, ?, UTC_TIMESTAMP())"
	_, err = m.DB.Exec(stmt, name, email, string(hash_password)) 
	if err != nil {
		var mySQLerror *mysql.MySQLError
		if errors.As(err, &mySQLerror){
			if mySQLerror.Number == 1062 && strings.Contains(mySQLerror.Message, "user_uc_email") {
				return ErrDuplicateEmail
			}
			return err
		}
	}
	return nil
}

func(m *UserModel) Authenticate(email, password string) (int, error) {
	var hash_password []byte 
	var id int

	stmt := "SELECT id, hashed_pw FROM users WHERE email = ?"
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hash_password) 
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hash_password, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials 
		} else {
			return 0, err
		}
	} 
	return id, nil
}

func(m *UserModel) Exists(id int) (bool, error) {
	var exist bool

	stmt := "SELECT EXISTS(SELECT true FROM users where id = ?)"
	
	err := m.DB.QueryRow(stmt, id).Scan(&exist)

	return exist, err
}
