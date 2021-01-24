package logtail

type ContainsMatcher struct {
	pattern string
	plen    int
	kmp     []int
}

func NewContainsMatcher(pattern string) *ContainsMatcher {
	if pattern == "" {
		panic("pattern nil")
	}

	cm := &ContainsMatcher{
		pattern: pattern,
	}

	cm.plen = len(pattern)
	cm.kmp = make([]int, cm.plen+1)
	cm.kmp[0] = -1

	for i := 1; i < cm.plen; i++ {
		j := cm.kmp[i-1]
		for j > -1 && cm.pattern[j+1] != cm.pattern[i] {
			j = cm.kmp[j]
		}

		if cm.pattern[j+1] == cm.pattern[i] {
			j++
		}

		cm.kmp[i] = j
	}

	return cm
}

func (cm *ContainsMatcher) Match(bytes []byte) bool {
	length := len(bytes)

	if length == 0 {
		return false
	}

	j := -1

	for i := 0; i < length; i++ {
		for j > -1 && cm.pattern[j+1] != bytes[i] {
			j = cm.kmp[j]
		}

		if cm.pattern[j+1] == bytes[i] {
			j++
		}

		if j+1 == cm.plen {
			return true
		}
	}

	return false
}
