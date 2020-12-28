package logtail

type Matcher interface {
	Match(format *Format, bytes []byte) [][]byte
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

func (cm *ContainsMatcher) Match(format *Format, bytes []byte) [][]byte {
	var matches [][]byte

	length := len(bytes)

	if length == 0 {
		return matches
	}

	j := -1
	start := 0

	for i := 0; i < length; i++ {
		if isLineEnd(bytes[i]) {
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

			i = ignoreLineEnd(bytes, length, i)

			// append following lines
			for i < length && isFollowingLine(format, bytes[i:]) {
				i = indexLineEnd(bytes, length, i)

				end = i

				i = ignoreLineEnd(bytes, length, i)
			}

			matches = append(matches, bytes[start:end])

			start = i

			j = cm.kmp[j]
		}
	}

	return matches
}

func isFollowingLine(format *Format, bytes []byte) bool {
	if format == nil {
		format = defaultFormat
	}

	if format != nil {
		return !format.PrefixMatch(bytes)
	}

	return bytes[0] == ' ' || bytes[0] == '\t'
}

func isLineEnd(b byte) bool {
	return b == '\n' || b == '\r'
}

func indexLineEnd(bytes []byte, length, i int) int {
	for ; i < length && !isLineEnd(bytes[i]); i++ {
	}
	return i
}

func ignoreLineEnd(bytes []byte, length, i int) int {
	for ; i < length && isLineEnd(bytes[i]); i++ {
	}
	return i
}
