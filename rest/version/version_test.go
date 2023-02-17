package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultVersion(t *testing.T) {
	v := DefaultVersion()
	assert.Equal(t, "dev", v.BuildType)
	assert.Empty(t, v.BuildDate)
	assert.Empty(t, v.BuildVersion)
}
