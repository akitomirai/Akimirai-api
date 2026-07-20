package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type createAPIKeyRepoStub struct {
	APIKeyRepository
}

func (s *createAPIKeyRepoStub) Create(_ context.Context, key *APIKey) error {
	key.ID = 101
	return nil
}

func (s *createAPIKeyRepoStub) ExistsByKey(context.Context, string) (bool, error) {
	return false, nil
}

type createAPIKeyUserRepoStub struct {
	UserRepository
	user *User
}

func (s *createAPIKeyUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	return s.user, nil
}

type createAPIKeyGroupRepoStub struct {
	GroupRepository
	group *Group
}

func (s *createAPIKeyGroupRepoStub) GetByID(context.Context, int64) (*Group, error) {
	return s.group, nil
}

func TestAPIKeyServiceCreateReturnsSelectedGroup(t *testing.T) {
	groupID := int64(42)
	group := &Group{
		ID:       groupID,
		Name:     "OpenAI",
		Platform: PlatformOpenAI,
		Status:   StatusActive,
	}
	customKey := "sk-create-response-group"
	svc := NewAPIKeyService(
		&createAPIKeyRepoStub{},
		&createAPIKeyUserRepoStub{user: &User{ID: 7, Status: StatusActive}},
		&createAPIKeyGroupRepoStub{group: group},
		nil,
		nil,
		nil,
		&config.Config{},
	)

	created, err := svc.Create(context.Background(), 7, CreateAPIKeyRequest{
		Name:      "created-key",
		GroupID:   &groupID,
		CustomKey: &customKey,
	})

	require.NoError(t, err)
	require.Equal(t, groupID, *created.GroupID)
	require.Same(t, group, created.Group)
	require.Equal(t, PlatformOpenAI, created.Group.Platform)
}
