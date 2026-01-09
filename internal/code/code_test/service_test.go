package code_test

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/code"
	coderepos "github.com/weni-ai/flows-code-actions/internal/code/mongodb"
	"github.com/weni-ai/flows-code-actions/internal/codelib"
	librepos "github.com/weni-ai/flows-code-actions/internal/codelib/mongodb"
	"github.com/weni-ai/flows-code-actions/internal/db"
)

func TestCreateCodeWithBlackList(t *testing.T) {
	os.Setenv("FLOWS_CODE_ACTIONS_BLACKLIST", "foo,bar,baz,qux")

	cfg := config.NewConfig()

	db, err := db.GetMongoDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}

	libRepo := librepos.NewCodeLibRepo(db)
	libService := codelib.NewCodeLibService(libRepo)

	repo := coderepos.NewCodeRepository(db)
	codeService := code.NewCodeService(cfg, repo, libService)

	_, err = codeService.Create(context.TODO(), &code.Code{
		Name:        "Test Code",
		Source:      "foo bar baz qux",
		Language:    code.TypePy,
		Type:        code.TypeEndpoint,
		URL:         "https://example.com",
		ProjectUUID: "5e82df29-f731-4861-8836-1b047ce03506",
	})
	assert.Equal(t, err.Error(), "source code contains blacklisted term")

	cd, err := codeService.Create(context.TODO(), &code.Code{
		Name:        "Test Code",
		Source:      "def Run(engine):\nprint('ahoy')",
		Language:    code.TypePy,
		Type:        code.TypeEndpoint,
		URL:         "https://example.com",
		ProjectUUID: "5e82df29-f731-4861-8836-1b047ce03506",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Test Code", cd.Name)

}

func TestUpdateCodelibWithBlackList(t *testing.T) {
	os.Setenv("FLOWS_CODE_ACTIONS_BLACKLIST", "foo,bar,baz,qux")

	cfg := config.NewConfig()

	db, err := db.GetMongoDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}

	libRepo := librepos.NewCodeLibRepo(db)
	libService := codelib.NewCodeLibService(libRepo)

	repo := coderepos.NewCodeRepository(db)
	codeService := code.NewCodeService(cfg, repo, libService)

	cd, _ := codeService.Create(context.TODO(), &code.Code{
		Name:        "Test Code",
		Source:      "def Run(engine):\nprint('ahoy')",
		Language:    code.TypePy,
		Type:        code.TypeEndpoint,
		URL:         "https://example.com",
		ProjectUUID: "5e82df29-f731-4861-8836-1b047ce03506",
	})

	_, err = codeService.Update(context.TODO(), cd.ID, "Test Code", "foo bar baz qux", string(code.TypeEndpoint), 60)

	assert.Equal(t, err.Error(), "source code contains blacklisted term")

	id := cd.ID
	cdu, err := codeService.Update(context.TODO(), id, "Test Code", "def Run(engine):\nprint('ahoy2')", string(code.TypeEndpoint), 60)

	assert.NoError(t, err)
	assert.True(t, strings.Contains(cdu.Source, "ahoy2"))
}
