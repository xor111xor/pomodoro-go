//go:build inmemory

package cmd

import (
	"github.com/xor111xor/pomodoro-go/internal/models"
	"github.com/xor111xor/pomodoro-go/internal/repository"
)

func getRepo() (models.Repository, error) {
	return repository.NewInMemoryRepo(), nil
}
