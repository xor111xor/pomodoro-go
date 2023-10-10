//go:build !inmemory

package internal_test

import (
	"os"
	"testing"

	"github.com/xor111xor/pomodoro-go/internal/models"
	"github.com/xor111xor/pomodoro-go/internal/repository"
)

func getRepo(t *testing.T) (models.Repository, func()) {
	t.Helper()
	tf, err := os.CreateTemp("/tmp", "pomo-")
	if err != nil {
		t.Fatal(err)
	}
	tf.Close()
	dbRepo, err := repository.NewSQLite3Repo(tf.Name())
	if err != nil {
		t.Fatal(err)
	}

	return dbRepo, func() {
		os.Remove(tf.Name())
	}

}
