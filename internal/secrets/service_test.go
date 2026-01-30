package secrets

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, secret *Secret) (*Secret, error) {
	args := m.Called(ctx, secret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Secret), args.Error(1)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*Secret, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Secret), args.Error(1)
}

func (m *MockRepository) GetByProjectUUID(ctx context.Context, projectUUID string) ([]Secret, error) {
	args := m.Called(ctx, projectUUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Secret), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, id string, secret *Secret) (*Secret, error) {
	args := m.Called(ctx, id, secret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Secret), args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestNewSecretService(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)

	assert.NotNil(t, service)
}

func TestServiceCreate_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	secret := &Secret{
		Name:        "API_KEY",
		Value:       "secret-value-123",
		ProjectUUID: "project-uuid-123",
	}

	expectedSecret := &Secret{
		ID:          "secret-uuid-123",
		Name:        "API_KEY",
		Value:       "secret-value-123",
		ProjectUUID: "project-uuid-123",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("Create", ctx, secret).Return(expectedSecret, nil)

	result, err := service.Create(ctx, secret)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSecret.ID, result.ID)
	assert.Equal(t, expectedSecret.Name, result.Name)
	mockRepo.AssertExpectations(t)
}

func TestServiceCreate_EmptyName(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	secret := &Secret{
		Name:        "",
		Value:       "secret-value-123",
		ProjectUUID: "project-uuid-123",
	}

	result, err := service.Create(ctx, secret)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "secret name is required", err.Error())
	mockRepo.AssertNotCalled(t, "Create")
}

func TestServiceCreate_EmptyValue(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	secret := &Secret{
		Name:        "API_KEY",
		Value:       "",
		ProjectUUID: "project-uuid-123",
	}

	result, err := service.Create(ctx, secret)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "secret value is required", err.Error())
	mockRepo.AssertNotCalled(t, "Create")
}

func TestServiceCreate_EmptyProjectUUID(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	secret := &Secret{
		Name:        "API_KEY",
		Value:       "secret-value-123",
		ProjectUUID: "",
	}

	result, err := service.Create(ctx, secret)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "project_uuid is required", err.Error())
	mockRepo.AssertNotCalled(t, "Create")
}

func TestServiceCreate_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	secret := &Secret{
		Name:        "API_KEY",
		Value:       "secret-value-123",
		ProjectUUID: "project-uuid-123",
	}

	mockRepo.On("Create", ctx, secret).Return(nil, errors.New("database error"))

	result, err := service.Create(ctx, secret)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database error", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestServiceGetByID_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	expectedSecret := &Secret{
		ID:          "secret-uuid-123",
		Name:        "API_KEY",
		Value:       "secret-value-123",
		ProjectUUID: "project-uuid-123",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetByID", ctx, "secret-uuid-123").Return(expectedSecret, nil)

	result, err := service.GetByID(ctx, "secret-uuid-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSecret.ID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetByID_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, "non-existent-id").Return(nil, errors.New("secret not found"))

	result, err := service.GetByID(ctx, "non-existent-id")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "secret not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestServiceGetByProjectUUID_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	expectedSecrets := []Secret{
		{
			ID:          "secret-uuid-1",
			Name:        "API_KEY",
			Value:       "value-1",
			ProjectUUID: "project-uuid-123",
		},
		{
			ID:          "secret-uuid-2",
			Name:        "DB_PASSWORD",
			Value:       "value-2",
			ProjectUUID: "project-uuid-123",
		},
	}

	mockRepo.On("GetByProjectUUID", ctx, "project-uuid-123").Return(expectedSecrets, nil)

	result, err := service.GetByProjectUUID(ctx, "project-uuid-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "API_KEY", result[0].Name)
	assert.Equal(t, "DB_PASSWORD", result[1].Name)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetByProjectUUID_Empty(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByProjectUUID", ctx, "project-uuid-123").Return([]Secret{}, nil)

	result, err := service.GetByProjectUUID(ctx, "project-uuid-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
	mockRepo.AssertExpectations(t)
}

func TestServiceUpdate_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	existingSecret := &Secret{
		ID:          "secret-uuid-123",
		Name:        "API_KEY",
		Value:       "old-value",
		ProjectUUID: "project-uuid-123",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	updatedSecret := &Secret{
		ID:          "secret-uuid-123",
		Name:        "NEW_API_KEY",
		Value:       "new-value",
		ProjectUUID: "project-uuid-123",
		CreatedAt:   existingSecret.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetByID", ctx, "secret-uuid-123").Return(existingSecret, nil)
	mockRepo.On("Update", ctx, "secret-uuid-123", mock.AnythingOfType("*secrets.Secret")).Return(updatedSecret, nil)

	result, err := service.Update(ctx, "secret-uuid-123", "NEW_API_KEY", "new-value")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "NEW_API_KEY", result.Name)
	assert.Equal(t, "new-value", result.Value)
	mockRepo.AssertExpectations(t)
}

func TestServiceUpdate_OnlyName(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	existingSecret := &Secret{
		ID:          "secret-uuid-123",
		Name:        "API_KEY",
		Value:       "old-value",
		ProjectUUID: "project-uuid-123",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetByID", ctx, "secret-uuid-123").Return(existingSecret, nil)
	mockRepo.On("Update", ctx, "secret-uuid-123", mock.AnythingOfType("*secrets.Secret")).Return(existingSecret, nil)

	result, err := service.Update(ctx, "secret-uuid-123", "NEW_NAME", "")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestServiceUpdate_OnlyValue(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	existingSecret := &Secret{
		ID:          "secret-uuid-123",
		Name:        "API_KEY",
		Value:       "old-value",
		ProjectUUID: "project-uuid-123",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetByID", ctx, "secret-uuid-123").Return(existingSecret, nil)
	mockRepo.On("Update", ctx, "secret-uuid-123", mock.AnythingOfType("*secrets.Secret")).Return(existingSecret, nil)

	result, err := service.Update(ctx, "secret-uuid-123", "", "new-value")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestServiceUpdate_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, "non-existent-id").Return(nil, errors.New("secret not found"))

	result, err := service.Update(ctx, "non-existent-id", "NEW_NAME", "new-value")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "secret not found", err.Error())
	mockRepo.AssertNotCalled(t, "Update")
}

func TestServiceDelete_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Delete", ctx, "secret-uuid-123").Return(nil)

	err := service.Delete(ctx, "secret-uuid-123")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServiceDelete_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewSecretService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Delete", ctx, "non-existent-id").Return(errors.New("secret not found"))

	err := service.Delete(ctx, "non-existent-id")

	assert.Error(t, err)
	assert.Equal(t, "secret not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestNewSecret(t *testing.T) {
	secret := NewSecret("API_KEY", "secret-value", "project-uuid-123")

	assert.NotNil(t, secret)
	assert.Equal(t, "API_KEY", secret.Name)
	assert.Equal(t, "secret-value", secret.Value)
	assert.Equal(t, "project-uuid-123", secret.ProjectUUID)
	assert.Empty(t, secret.ID)
}
