package store

import (
	"errors"
	"time"
)

type Store interface {
	NewLogSession(key time.Time) error
	ListLogSessions() []time.Time
	ReadFromLogSession(key time.Time) ([]time.Time, []string, error)
	WriteToLogSession(key time.Time, line string) error
	DeleteLogSession(key time.Time) error
	CloseLogSession(key time.Time) error
}

var (
	ErrLogSessionAlreadyExists = errors.New("Log Session with key already exists")
	ErrLogSessionDoesNotExist = errors.New("Log Session with key does not exist")
	ErrLogSessionIsOpenCannotDelete = errors.New("Log Session is open, it cannot be deleted")
)