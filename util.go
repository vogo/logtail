package logtail

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
		indexLineEnd(message, &l, &i)
		ignoreLineEnd(message, &l, &i)

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

func indexLineEnd(bytes []byte, length, index *int) {
	for ; *index < *length && !isLineEnd(bytes[*index]); *index++ {
	}
}

func ignoreLineEnd(bytes []byte, length, index *int) {
	for ; *index < *length && isLineEnd(bytes[*index]); *index++ {
	}
}
