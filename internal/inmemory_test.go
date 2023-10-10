//go:build inmemory

package internal_test

import (
	"github.com/xor111xor/pomodoro-go/internal/models"
	"github.com/xor111xor/pomodoro-go/internal/repository"
	"testing"
)

func getRepo(t *testing.T) (models.Repository, func()) {
	t.Helper()
	return repository.NewInMemoryRepo(), func() {}
}
