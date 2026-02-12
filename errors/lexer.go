package errors

import "bytes"

type LexerError struct {
	line string
	col  int
}

func (e *LexerError) Error() string {
	buf := new(bytes.Buffer)
	buf.WriteString(e.line)
	buf.WriteString("\n")
	for i := 0; i < e.col-1; i++ {
		buf.WriteString(" ")
	}
	buf.WriteString("^")
	buf.WriteString("\n")

	buf.WriteString("Uncaught SyntaxError: Invalid or unexpected token")
	return buf.String()
}

func NewLexerError(line string, col int) error {
	return &LexerError{
		line: line,
		col:  col,
	}
}
