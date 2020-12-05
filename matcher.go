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
	if len(pattern) == 0 {
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
		if bytes[i] == '\n' || bytes[i] == '\r' {
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
			for ; i < length && bytes[i] != '\n' && bytes[i] != '\r'; i++ {
			}
			end := i

			// append following lines
			for i < length {
				for ; i < length && (bytes[i] == '\n' || bytes[i] == '\r'); i++ {
				}

				if i < length && (bytes[i] == ' ' || bytes[i] == '\t') {
					for ; i < length && bytes[i] != '\n' && bytes[i] != '\r'; i++ {
					}
					end = i
					continue
				}

				break
			}

			matches = append(matches, bytes[start:end])

			start = i

			j = cm.kmp[j]
		}
	}
	return matches
}
