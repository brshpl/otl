package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/brshpl/otl/internal/entity"
)

var (
	// ErrLinkExpired -.
	ErrLinkExpired = errors.New("expired link")
	// ErrInvalidLink - no data found by this link
	ErrInvalidLink = errors.New("invalid link")
)

// OneTimeLinkUseCase implements OneTimeLink usecase
type OneTimeLinkUseCase struct {
	repo      OneTimeLinkRepo
	generator Generator
	linkLen   int
}

// Generator generates random string having length len
type Generator func(len int) string

// New -.
func New(r OneTimeLinkRepo, generator Generator, linkLen int) *OneTimeLinkUseCase {
	return &OneTimeLinkUseCase{
		repo:      r,
		generator: generator,
		linkLen:   linkLen,
	}
}

// Create new one-time link
func (uc *OneTimeLinkUseCase) Create(ctx context.Context, data string) (string, error) {
	var (
		link string
		err  error
	)

	exists := true
	for exists {
		link = uc.generator(uc.linkLen)
		if link == "metrics" || link == "healthz" { // you never know...
			continue
		}

		exists, err = uc.repo.Check(ctx, link)
		if err != nil {
			return "", fmt.Errorf("OneTimeLinkUseCase - Create - uc.repo.Check: %w", err)
		}
	}

	otl := entity.OneTimeLink{Data: data, Link: link}

	err = uc.repo.Store(ctx, otl)
	if err != nil {
		return "", fmt.Errorf("OneTimeLinkUseCase - Create - uc.repo.Store: %w", err)
	}

	return otl.Link, nil
}

// Get value by one-time link. Returns ErrInvalidLink if no data found, ErrLinkExpired if link expired
func (uc *OneTimeLinkUseCase) Get(ctx context.Context, link string) (string, error) {
	zero := entity.OneTimeLink{}

	otl, err := uc.repo.Get(ctx, link)
	if err != nil {
		return "", fmt.Errorf("OneTimeLinkUseCase - Get - uc.repo.Get: %w", err)
	}

	if otl == zero {
		return "", ErrInvalidLink
	}

	if otl.Expired {
		return "", ErrLinkExpired
	}

	return otl.Data, nil
}
