package edit

// filterSpan returns the [start, end) byte span of the segment of buffer that
// should seed a histlist filter, by splitting at the last unquoted shell
// separator (';', '|', '&', or newline). A trailing separator yields an empty
// span at the end of the buffer.
//
// Separators inside quotes or parentheses do not split.
func filterSpan(buffer string) (int, int) {
	start := 0
	for i := 0; i < len(buffer); i++ {
		switch c := buffer[i]; c {
		case '\\':
			i++
		case '\'', '"':
			i = skipQuoted(buffer, i)
		case '(', '`':
			i = skipNested(buffer, i)
		case ';', '|', '&', '\n':
			// Trailing whitespace of the current segment belongs to the prefix.
			start = i + 1
		}
	}
	// Keep leading whitespace in the prefix, but only skip it when there is a
	// non-blank filter left; a trailing separator yields an empty span at the
	// end of the buffer so that accepting chains instead of replacing.
	trimmed := start
	for trimmed < len(buffer) && (buffer[trimmed] == ' ' || buffer[trimmed] == '\t') {
		trimmed++
	}
	if trimmed < len(buffer) {
		start = trimmed
	}
	return start, len(buffer)
}

// skipQuoted returns the index of the closing quote matching buffer[i], or the
// last index if unterminated.
func skipQuoted(s string, i int) int {
	q := s[i]
	for j := i + 1; j < len(s); j++ {
		if s[j] == '\\' && q == '"' {
			j++
			continue
		}
		if s[j] == q {
			return j
		}
	}
	return len(s) - 1
}

// skipNested returns the index of the closing bracket matching s[i]. Inside
// backticks, quotes are not special; inside parens they are.
func skipNested(s string, i int) int {
	open := s[i]
	closeCh := byte(')')
	if open == '`' {
		closeCh = '`'
	}
	depth := 0
	for j := i + 1; j < len(s); j++ {
		switch s[j] {
		case '\\':
			if open == '`' {
				j++
			}
		case '\'', '"':
			if open != '`' {
				j = skipQuoted(s, j)
			}
		default:
			if s[j] == open {
				depth++
			} else if s[j] == closeCh {
				if depth == 0 {
					return j
				}
				depth--
			}
		}
	}
	return len(s) - 1
}
