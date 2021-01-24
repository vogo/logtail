package logtail_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail"
)

func TestMatch(t *testing.T) {
	data := []byte(`2020-12-25 14:54:38.523  ERROR exception occurs`)

	assert.True(t, logtail.NewContainsMatcher("ERROR").Match(data))
	assert.True(t, logtail.NewContainsMatcher("exception").Match(data))

	assert.False(t, logtail.NewContainsMatcher("WARN").Match(data))
}
