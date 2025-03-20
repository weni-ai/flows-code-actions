package secret_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/weni-ai/flows-code-actions/internal/secret"
)

// MockRepository is a mock implementation of secret.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, s *secret.Secret) (*secret.Secret, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secret.Secret), args.Error(1)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*secret.Secret, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secret.Secret), args.Error(1)
}

func (m *MockRepository) ListByProjectUUID(ctx context.Context, projectUUID string) ([]secret.Secret, error) {
	args := m.Called(ctx, projectUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]secret.Secret), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, id string, s *secret.Secret) (*secret.Secret, error) {
	args := m.Called(ctx, id, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*secret.Secret), args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestSecretService(t *testing.T) {
	mockRepo := new(MockRepository)
	service := secret.NewService(mockRepo)
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		secretObj := secret.NewSecret("test-secret", "test-value", "test-project")
		mockRepo.On("Create", ctx, secretObj).Return(secretObj, nil)

		result, err := service.Create(ctx, secretObj)

		assert.NoError(t, err)
		assert.Equal(t, secretObj, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByID", func(t *testing.T) {
		secretObj := secret.NewSecret("test-secret", "test-value", "test-project")
		mockRepo.On("GetByID", ctx, "123").Return(secretObj, nil)

		result, err := service.GetByID(ctx, "123")

		assert.NoError(t, err)
		assert.Equal(t, secretObj, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ListByProjectUUID", func(t *testing.T) {
		secrets := []secret.Secret{
			*secret.NewSecret("secret1", "value1", "test-project"),
			*secret.NewSecret("secret2", "value2", "test-project"),
		}
		mockRepo.On("ListByProjectUUID", ctx, "test-project").Return(secrets, nil)

		results, err := service.ListByProjectUUID(ctx, "test-project")

		assert.NoError(t, err)
		assert.Equal(t, secrets, results)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update", func(t *testing.T) {
		secretObj := secret.NewSecret("updated-secret", "updated-value", "test-project")
		mockRepo.On("Update", ctx, "123", secretObj).Return(secretObj, nil)

		result, err := service.Update(ctx, "123", secretObj)

		assert.NoError(t, err)
		assert.Equal(t, secretObj, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete", func(t *testing.T) {
		mockRepo.On("Delete", ctx, "123").Return(nil)

		err := service.Delete(ctx, "123")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
