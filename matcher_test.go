package logtail_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail"
)

func TestMatch(t *testing.T) {
	s := `
2020-12-25 14:54:38.523  ERROR exception occurs
2020-12-25 14:54:38.532  INFO hello world
2020-12-25 14:54:38.532  ERROR err2
	stack1
	stack2
2020-12-25 14:54:38.532  INFO hello world 2
`

	// match: ERROR
	m := logtail.NewContainsMatcher("ERROR")
	matches := m.Match([]byte(s))

	assert.Equal(t, 2, len(matches))

	if len(matches) != 2 {
		assert.FailNow(t, "expect 2 matches")
	}

	assert.Equal(t, `2020-12-25 14:54:38.523  ERROR exception occurs`, string(matches[0]))
	assert.Equal(t, `2020-12-25 14:54:38.532  ERROR err2
	stack1
	stack2`, string(matches[1]))

	// match: hello world
	m = logtail.NewContainsMatcher("hello world")
	matches = m.Match([]byte(s))

	assert.Equal(t, 2, len(matches))

	if len(matches) != 2 {
		assert.FailNow(t, "expect 2 matches")
	}

	assert.Equal(t, `2020-12-25 14:54:38.532  INFO hello world`, string(matches[0]))
	assert.Equal(t, `2020-12-25 14:54:38.532  INFO hello world 2`, string(matches[1]))
}
