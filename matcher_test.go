package logtail_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail"
)

func TestMatch(t *testing.T) {
	data := []byte(`2020-12-25 14:54:38.523  ERROR exception occurs`)

	assert.True(t, logtail.NewContainsMatcher("ERROR", true).Match(data))
	assert.True(t, logtail.NewContainsMatcher("exception", true).Match(data))

	assert.False(t, logtail.NewContainsMatcher("ERROR", false).Match(data))
	assert.False(t, logtail.NewContainsMatcher("exception", false).Match(data))
	assert.False(t, logtail.NewContainsMatcher("WARN", true).Match(data))
}

func TestMatch2(t *testing.T) {
	data := []byte(`2020-12-25 14:54:38.523  错误 error 异常 exception 数据找不到信息`)

	assert.True(t, logtail.NewContainsMatcher("error", true).Match(data))
	assert.True(t, logtail.NewContainsMatcher("exception", true).Match(data))
	assert.True(t, logtail.NewContainsMatcher("错误", true).Match(data))
	assert.True(t, logtail.NewContainsMatcher("异常", true).Match(data))

	assert.False(t, logtail.NewContainsMatcher("error", false).Match(data))
	assert.False(t, logtail.NewContainsMatcher("exception", false).Match(data))
	assert.False(t, logtail.NewContainsMatcher("错误", false).Match(data))
	assert.False(t, logtail.NewContainsMatcher("异常", false).Match(data))
	assert.False(t, logtail.NewContainsMatcher("找不到", false).Match(data))

	assert.True(t, logtail.NewContainsMatcher("没问题", false).Match(data))
}
