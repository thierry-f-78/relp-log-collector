package utilities

import "fmt"
import "strconv"
import "strings"

func EscapeNonASCIIPrintable(input []byte) []byte {
	var b byte
	var builder strings.Builder

	for _, b = range input {
		switch b {
		case '\a':
			builder.WriteString("\\a")
		case '\b':
			builder.WriteString("\\b")
		case '\t':
			builder.WriteString("\\t")
		case '\n':
			builder.WriteString("\\n")
		case '\v':
			builder.WriteString("\\v")
		case '\f':
			builder.WriteString("\\f")
		case '\r':
			builder.WriteString("\\r")
		case '\\':
			builder.WriteString("\\\\")
		default:
			if b < 32 || b > 126 {
				builder.WriteString(fmt.Sprintf("\\x%02X", b))
			} else {
				builder.WriteString(string(b))
			}
		}
	}
	return []byte(builder.String())
}

// In error case, the returned string will be the input string
func UnescapeNonASCIIPrintable(input string) (string, error) {
	var b uint64
	var builder strings.Builder
	var i int
	var inputBytes []byte
	var err error

	inputBytes = []byte(input)

	for i = 0; i < len(inputBytes); i++ {
		if inputBytes[i] == '\\' {
			if i+1 >= len(inputBytes) {
				return input, fmt.Errorf("wrong encoding")
			}
			i++
			switch inputBytes[i] {
			case 'a':
				builder.WriteByte('\a')
			case 'b':
				builder.WriteByte('\b')
			case 't':
				builder.WriteByte('\t')
			case 'n':
				builder.WriteByte('\n')
			case 'v':
				builder.WriteByte('\v')
			case 'f':
				builder.WriteByte('\f')
			case 'r':
				builder.WriteByte('\r')
			case '\\':
				builder.WriteByte('\\')
			case 'x':
				if i+2 >= len(inputBytes) {
					return input, fmt.Errorf("wrong encoding")
				}
				i++
				b, err = strconv.ParseUint(string(inputBytes[i:i+2]), 16, 8)
				if err != nil {
					return input, err
				}
				builder.WriteByte(byte(b))
				i++
			default:
				return input, fmt.Errorf("wrong encoding")
			}
		} else {
			builder.WriteByte(inputBytes[i])
		}
	}
	return builder.String(), nil
}
