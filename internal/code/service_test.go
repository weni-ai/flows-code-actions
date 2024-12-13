package code

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weni-ai/flows-code-actions/config"
)

func TestCreateCodeWithBlackList(t *testing.T) {
	os.Setenv("FLOWS_CODE_ACTIONS_BLACKLIST", "foo,bar,baz,qux")

	cfg := config.NewConfig()
	// s := server.NewServer(cfg)

	// db, err := db.GetMongoDatabase(cfg)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// s.DB = db

	// confs := config.NewConfig()

	codeService := NewCodeService(cfg, nil, nil)

	_, err := codeService.Create(context.TODO(), &Code{
		Name:        "Test Code",
		Source:      "foo bar baz qux",
		Language:    TypePy,
		Type:        TypeEndpoint,
		URL:         "https://example.com",
		ProjectUUID: "5e82df29-f731-4861-8836-1b047ce03506",
	})

	assert.Equal(t, err.Error(), "source code contains blacklisted term")

}
