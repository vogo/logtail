package logtail_test

import (
	"fmt"
	"testing"
)

func TestSubBytes(t *testing.T) {
	data := []byte(`你好`)
	length := len(data)
	s := string(data[:length-1])
	fmt.Println(s)
}
