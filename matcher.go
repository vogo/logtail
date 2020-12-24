package logtail

type Matcher interface {
	Match(bytes []byte) [][]byte
}

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

func (cm *ContainsMatcher) Match(bytes []byte) [][]byte {
	var matches [][]byte

	length := len(bytes)

	if length == 0 {
		return matches
	}

	j := -1
	start := 0

	for i := 0; i < length; i++ {
		if isNewLineTag(bytes[i]) {
			j = -1
			start = i + 1

			continue
		}

		for j > -1 && cm.pattern[j+1] != bytes[i] {
			j = cm.kmp[j]
		}

		if cm.pattern[j+1] == bytes[i] {
			j++
		}

		if j+1 == cm.plen {
			i = indexLineEnd(bytes, length, i)

			end := i

			i = ignoreNewLines(bytes, length, i)

			// append following lines
			for i < length && (bytes[i] == ' ' || bytes[i] == '\t') {
				i = indexLineEnd(bytes, length, i)

				end = i

				i = ignoreNewLines(bytes, length, i)
			}

			matches = append(matches, bytes[start:end])

			start = i

			j = cm.kmp[j]
		}
	}

	return matches
}

func isNewLineTag(b byte) bool {
	return b == '\n' || b == '\r'
}

func indexLineEnd(bytes []byte, length, i int) int {
	for ; i < length && !isNewLineTag(bytes[i]); i++ {
	}
	return i
}

func ignoreNewLines(bytes []byte, length, i int) int {
	for ; i < length && isNewLineTag(bytes[i]); i++ {
	}
	return i
}
