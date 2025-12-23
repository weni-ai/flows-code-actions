package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
	"github.com/weni-ai/flows-code-actions/internal/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDeleteOlder(t *testing.T) {
	cfg := config.NewConfig()
	db, err := db.GetMongoDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewCodeLogRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = repo.Create(ctx, &codelog.CodeLog{
		CodeID:    primitive.NewObjectID().Hex(),
		RunID:     primitive.NewObjectID().Hex(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.NoError(t, err)
	_, err = repo.Create(ctx, &codelog.CodeLog{
		CodeID:    primitive.NewObjectID().Hex(),
		RunID:     primitive.NewObjectID().Hex(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.NoError(t, err)

	total, err := repo.DeleteOlder(ctx, time.Now(), 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
}
