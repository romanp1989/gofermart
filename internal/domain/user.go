package domain

import "errors"

type UserID int64

var ErrLoginExists = errors.New("Данные логин уже используется")

type User struct {
	ID       UserID
	Login    string
	Password string
}