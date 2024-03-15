package usecase

import (
	"context"

	"github.com/brshpl/otl/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	// OneTimeLink -.
	OneTimeLink interface {
		Create(ctx context.Context, data string) (string, error)
		Get(ctx context.Context, link string) (string, error)
	}

	// OneTimeLinkRepo -.
	OneTimeLinkRepo interface {
		// Store one-time link in repo
		Store(context.Context, entity.OneTimeLink) error
		// Get one-time link from repo and expire record
		Get(ctx context.Context, link string) (entity.OneTimeLink, error)
		// Check link in repo and return if it exists or not
		Check(ctx context.Context, link string) (bool, error)
	}
)
