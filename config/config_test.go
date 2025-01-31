package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlackList(t *testing.T) {
	os.Setenv("FLOWS_CODE_ACTIONS_BLACKLIST", "foo,bar,baz,qux")

	confs := NewConfig()

	blacklist := confs.GetBlackListTerms()

	assert.Equal(t, 4, len(blacklist))
	assert.Equal(t, "bar", blacklist[0])
}
