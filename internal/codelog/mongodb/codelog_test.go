package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/db"
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
	total, err := repo.DeleteOlder(ctx, time.Now(), 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
}
