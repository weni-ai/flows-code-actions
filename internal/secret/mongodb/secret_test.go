package secret

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weni-ai/flows-code-actions/internal/secret"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestSecretRepository(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Create", func(mt *mtest.T) {
		repo := NewSecretRepository(mt.DB)
		ctx := context.Background()

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "secrets", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "name", Value: "test-secret"},
			{Key: "value", Value: "test-value"},
			{Key: "project_uuid", Value: "test-project"},
			{Key: "created_at", Value: time.Now()},
			{Key: "updated_at", Value: time.Now()},
		}))

		secret := &secret.Secret{
			Name:        "test-secret",
			Value:       "test-value",
			ProjectUUID: "test-project",
		}

		result, err := repo.Create(ctx, secret)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, "test-secret", result.Name)
		assert.Equal(t, "test-value", result.Value)
		assert.Equal(t, "test-project", result.ProjectUUID)
	})

	mt.Run("GetByID", func(mt *mtest.T) {
		repo := NewSecretRepository(mt.DB)
		ctx := context.Background()
		secretID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "secrets", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: secretID},
			{Key: "name", Value: "test-secret"},
			{Key: "value", Value: "test-value"},
			{Key: "project_uuid", Value: "test-project"},
			{Key: "created_at", Value: time.Now()},
			{Key: "updated_at", Value: time.Now()},
		}))

		result, err := repo.GetByID(ctx, secretID.Hex())
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, secretID, result.ID)
		assert.Equal(t, "test-secret", result.Name)
		assert.Equal(t, "test-value", result.Value)
		assert.Equal(t, "test-project", result.ProjectUUID)
	})

	mt.Run("ListByProjectUUID", func(mt *mtest.T) {
		repo := NewSecretRepository(mt.DB)
		ctx := context.Background()

		mt.AddMockResponses(mtest.CreateCursorResponse(2, "secrets", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "name", Value: "secret1"},
			{Key: "value", Value: "value1"},
			{Key: "project_uuid", Value: "test-project"},
			{Key: "created_at", Value: time.Now()},
			{Key: "updated_at", Value: time.Now()},
		}, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "name", Value: "secret2"},
			{Key: "value", Value: "value2"},
			{Key: "project_uuid", Value: "test-project"},
			{Key: "created_at", Value: time.Now()},
			{Key: "updated_at", Value: time.Now()},
		}))

		results, err := repo.ListByProjectUUID(ctx, "test-project")
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "secret1", results[0].Name)
		assert.Equal(t, "secret2", results[1].Name)
	})

	mt.Run("Update", func(mt *mtest.T) {
		repo := NewSecretRepository(mt.DB)
		ctx := context.Background()
		secretID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		secret := &secret.Secret{
			Name:        "updated-secret",
			Value:       "updated-value",
			ProjectUUID: "test-project",
		}

		result, err := repo.Update(ctx, secretID.Hex(), secret)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "updated-secret", result.Name)
		assert.Equal(t, "updated-value", result.Value)
		assert.Equal(t, "test-project", result.ProjectUUID)
	})

	mt.Run("Delete", func(mt *mtest.T) {
		repo := NewSecretRepository(mt.DB)
		ctx := context.Background()
		secretID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Delete(ctx, secretID.Hex())
		require.NoError(t, err)
	})
}
