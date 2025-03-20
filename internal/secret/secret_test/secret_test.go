package secret_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/weni-ai/flows-code-actions/internal/secret"
	secretRepo "github.com/weni-ai/flows-code-actions/internal/secret/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestSecretRepository(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Create", func(mt *mtest.T) {
		repo := secretRepo.NewSecretRepository(mt.DB)
		ctx := context.Background()

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		secretObj := secret.NewSecret("test-secret", "test-value", "test-project")
		result, err := repo.Create(ctx, secretObj)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, "test-secret", result.Name)
		assert.Equal(t, "test-value", result.Value)
		assert.Equal(t, "test-project", result.ProjectUUID)
		assert.NotZero(t, result.CreatedAt)
		assert.NotZero(t, result.UpdatedAt)

		// Verify the insert command
		mt.GetStartedEvent()
		mt.GetStartedEvent()
	})

	mt.Run("GetByID", func(mt *mtest.T) {
		repo := secretRepo.NewSecretRepository(mt.DB)
		ctx := context.Background()

		expectedID := primitive.NewObjectID()
		expectedSecret := &secret.Secret{
			ID:          expectedID,
			Name:        "test-secret",
			Value:       "test-value",
			ProjectUUID: "test-project",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.secrets", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expectedID},
			{Key: "name", Value: expectedSecret.Name},
			{Key: "value", Value: expectedSecret.Value},
			{Key: "project_uuid", Value: expectedSecret.ProjectUUID},
			{Key: "created_at", Value: expectedSecret.CreatedAt},
			{Key: "updated_at", Value: expectedSecret.UpdatedAt},
		}))

		result, err := repo.GetByID(ctx, expectedID.Hex())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedID, result.ID)
		assert.Equal(t, expectedSecret.Name, result.Name)
		assert.Equal(t, expectedSecret.Value, result.Value)
		assert.Equal(t, expectedSecret.ProjectUUID, result.ProjectUUID)
	})

	mt.Run("ListByProjectUUID", func(mt *mtest.T) {
		repo := secretRepo.NewSecretRepository(mt.DB)
		ctx := context.Background()

		expectedSecrets := []secret.Secret{
			{
				ID:          primitive.NewObjectID(),
				Name:        "secret1",
				Value:       "value1",
				ProjectUUID: "test-project",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          primitive.NewObjectID(),
				Name:        "secret2",
				Value:       "value2",
				ProjectUUID: "test-project",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(2, "test.secrets", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expectedSecrets[0].ID},
			{Key: "name", Value: expectedSecrets[0].Name},
			{Key: "value", Value: expectedSecrets[0].Value},
			{Key: "project_uuid", Value: expectedSecrets[0].ProjectUUID},
			{Key: "created_at", Value: expectedSecrets[0].CreatedAt},
			{Key: "updated_at", Value: expectedSecrets[0].UpdatedAt},
		}, bson.D{
			{Key: "_id", Value: expectedSecrets[1].ID},
			{Key: "name", Value: expectedSecrets[1].Name},
			{Key: "value", Value: expectedSecrets[1].Value},
			{Key: "project_uuid", Value: expectedSecrets[1].ProjectUUID},
			{Key: "created_at", Value: expectedSecrets[1].CreatedAt},
			{Key: "updated_at", Value: expectedSecrets[1].UpdatedAt},
		}))
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.secrets", mtest.NextBatch))

		results, err := repo.ListByProjectUUID(ctx, "test-project")

		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, expectedSecrets[0].ID, results[0].ID)
		assert.Equal(t, expectedSecrets[1].ID, results[1].ID)
	})

	mt.Run("Update", func(mt *mtest.T) {
		repo := secretRepo.NewSecretRepository(mt.DB)
		ctx := context.Background()

		expectedID := primitive.NewObjectID()
		updatedSecret := &secret.Secret{
			ID:          expectedID,
			Name:        "updated-secret",
			Value:       "updated-value",
			ProjectUUID: "test-project",
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		result, err := repo.Update(ctx, expectedID.Hex(), updatedSecret)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedID, result.ID)
		assert.Equal(t, updatedSecret.Name, result.Name)
		assert.Equal(t, updatedSecret.Value, result.Value)
		assert.Equal(t, updatedSecret.ProjectUUID, result.ProjectUUID)
		assert.NotZero(t, result.UpdatedAt)
	})

	mt.Run("Delete", func(mt *mtest.T) {
		repo := secretRepo.NewSecretRepository(mt.DB)
		ctx := context.Background()

		expectedID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Delete(ctx, expectedID.Hex())

		assert.NoError(t, err)
	})

	mt.Run("GetByID_NotFound", func(mt *mtest.T) {
		repo := secretRepo.NewSecretRepository(mt.DB)
		ctx := context.Background()

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.secrets", mtest.FirstBatch))

		result, err := repo.GetByID(ctx, primitive.NewObjectID().Hex())

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
