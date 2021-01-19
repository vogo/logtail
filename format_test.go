package logtail_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail"
)

func TestWildcardMatch(t *testing.T) {
	assert.False(t, logtail.WildcardMatch("!!!!", nil))
	assert.True(t, logtail.WildcardMatch("", nil))
	assert.True(t, logtail.WildcardMatch("", []byte("abcd")))

	assert.False(t, logtail.WildcardMatch("!!!!", []byte("a")))
	assert.False(t, logtail.WildcardMatch("!!!!", []byte("abcd")))
	assert.False(t, logtail.WildcardMatch("!!!!", []byte("123a")))

	assert.True(t, logtail.WildcardMatch("!!!!", []byte("1234")))
	assert.True(t, logtail.WildcardMatch("!!!!", []byte("1234abcd")))
	assert.True(t, logtail.WildcardMatch("!!!!-!!-!!", []byte("2021-01-01")))
	assert.False(t, logtail.WildcardMatch("!!!!-!!-!!", []byte("2021001001")))

	assert.False(t, logtail.WildcardMatch("~~~~", []byte("1234abcd")))
	assert.True(t, logtail.WildcardMatch("~~~~", []byte("abcd")))
	assert.True(t, logtail.WildcardMatch("~~~~", []byte("abcd1234")))

	assert.True(t, logtail.WildcardMatch("????", []byte("1234")))
	assert.True(t, logtail.WildcardMatch("????", []byte("abcd")))
}
