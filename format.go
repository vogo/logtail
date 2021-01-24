package logtail

import "fmt"

// Format the log format.
type Format struct {
	Prefix string `json:"prefix"` // the wildcard of the line prefix of a log record
}

// PrefixMatch whether the given data has a prefix of a new record.
func (f *Format) PrefixMatch(data []byte) bool {
	return WildcardMatch(f.Prefix, data)
}

// String format string info.
func (f *Format) String() string {
	return fmt.Sprintf("format{prefix:%s}", f.Prefix)
}

// WildcardMatch -  finds whether the bytes matches/satisfies the pattern wildcard.
// supports:
// - '?' as one byte char
// - '~' as one alphabet char
// - '!' as one number char
// NOT support '*' for none or many char.
func WildcardMatch(pattern string, data []byte) bool {
	var p, b byte

	for i, j := 0, 0; i < len(pattern); i++ {
		if j >= len(data) {
			return false
		}

		p = pattern[i]
		b = data[j]

		switch p {
		case '?':
			if len(data) == 0 {
				return false
			}
		case '~':
			if !isAlphabetChar(b) {
				return false
			}
		case '!':
			if !isNumberChar(b) {
				return false
			}
		default:
			if len(data) == 0 || b != p {
				return false
			}
		}

		j++
	}

	return true
}

func isNumberChar(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlphabetChar(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func indexToNextLineStart(format *Format, message []byte) []byte {
	l := len(message)
	i := 0

	for i < l {
		i = indexLineEnd(message, l, i)
		i = ignoreLineEnd(message, l, i)

		if format == nil || format.PrefixMatch(message[i:]) {
			return message[i:]
		}
	}

	return nil
}

func indexToLineStart(format *Format, data []byte) []byte {
	if format == nil || format.PrefixMatch(data) {
		return data
	}

	return indexToNextLineStart(format, data)
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
