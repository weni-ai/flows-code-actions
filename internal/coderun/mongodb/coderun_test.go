package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
	"github.com/weni-ai/flows-code-actions/internal/db"
)

func TestDeleteOlder(t *testing.T) {
	cfg := config.NewConfig()
	db, err := db.GetMongoDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewCodeRunRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = repo.Create(ctx, &coderun.CodeRun{
		Status:    coderun.StatusCompleted,
		Result:    "success",
		Extra:     map[string]interface{}{"status_code": 200},
		Params:    map[string]interface{}{"id": "1"},
		Body:      "{}",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.NoError(t, err)

	total, err := repo.DeleteOlder(ctx, time.Now(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
}
