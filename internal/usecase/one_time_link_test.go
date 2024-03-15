package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brshpl/otl/pkg/random"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/brshpl/otl/internal/entity"
	"github.com/brshpl/otl/internal/usecase"
)

const (
	testLink = "testLink"
	testData = "testData"
)

var errRepo = errors.New("repo error")

type test struct {
	name string
	mock func()
	res  interface{}
	err  error
}

func oneTimeLink(t *testing.T, gen usecase.Generator) (*usecase.OneTimeLinkUseCase, *MockOneTimeLinkRepo) {
	t.Helper()

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	repo := NewMockOneTimeLinkRepo(mockCtl)

	uc := usecase.New(repo, gen, len(testLink))

	return uc, repo
}

func TestGet(t *testing.T) {
	t.Parallel()

	uc, repo := oneTimeLink(t, random.GenerateString)

	tests := []test{
		{
			name: "empty result",
			mock: func() {
				repo.EXPECT().Get(context.Background(), testLink).Return(entity.OneTimeLink{}, nil)
			},
			res: "",
			err: usecase.ErrInvalidLink,
		},
		{
			name: "test data not expired",
			mock: func() {
				repo.EXPECT().Get(context.Background(), testLink).Return(entity.OneTimeLink{
					Data:    testData,
					Link:    testLink,
					Expired: false,
				}, nil)
			},
			res: testData,
			err: nil,
		},
		{
			name: "test data not expired but link not exist at first",
			mock: func() {
				repo.EXPECT().Get(context.Background(), testLink).Return(entity.OneTimeLink{
					Data:    testData,
					Link:    testLink,
					Expired: false,
				}, nil)
			},
			res: testData,
			err: nil,
		},
		{
			name: "test data expired",
			mock: func() {
				repo.EXPECT().Get(context.Background(), testLink).Return(entity.OneTimeLink{
					Data:    testData,
					Link:    testLink,
					Expired: true,
				}, nil)
			},
			res: "",
			err: usecase.ErrLinkExpired,
		},
		{
			name: "result with error",
			mock: func() {
				repo.EXPECT().Get(context.Background(), testLink).Return(entity.OneTimeLink{}, errRepo)
			},
			res: "",
			err: errRepo,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()

			res, err := uc.Get(context.Background(), testLink)

			require.Equal(t, res, tc.res)
			require.ErrorIs(t, err, tc.err)
		})
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	uc, repo := oneTimeLink(t, dummyGenerator)

	tests := []test{
		{
			name: "created",
			mock: func() {
				repo.EXPECT().Check(context.Background(), testLink).Return(false, nil).Times(1)
				repo.EXPECT().Store(context.Background(), entity.OneTimeLink{
					Data: testData,
					Link: testLink,
				}).Return(nil)
			},
			res: testLink,
			err: nil,
		},
		{
			name: "created but link exists at first",
			mock: func() {
				atFirst := repo.EXPECT().Check(context.Background(), testLink).Return(true, nil).Times(1)
				repo.EXPECT().Check(context.Background(), testLink).Return(false, nil).After(atFirst).Times(1)
				repo.EXPECT().Store(context.Background(), entity.OneTimeLink{
					Data: testData,
					Link: testLink,
				}).Return(nil)
			},
			res: testLink,
			err: nil,
		},
		{
			name: "repo check error",
			mock: func() {
				repo.EXPECT().Check(context.Background(), testLink).Return(false, errRepo).Times(1)
			},
			res: "",
			err: errRepo,
		},
		{
			name: "repo store error",
			mock: func() {
				repo.EXPECT().Check(context.Background(), testLink).Return(false, nil).Times(1)
				repo.EXPECT().Store(context.Background(), entity.OneTimeLink{
					Data: testData,
					Link: testLink,
				}).Return(errRepo)
			},
			res: "",
			err: errRepo,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			tc.mock()

			res, err := uc.Create(context.Background(), testData)

			require.EqualValues(t, res, tc.res)
			require.ErrorIs(t, err, tc.err)
		})
	}
}

func dummyGenerator(int) string {
	return testLink
}
