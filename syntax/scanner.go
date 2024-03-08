package syntax

import (
	"errors"
	"unicode/utf8"
)

type Scanner struct {
	source []byte
	tokens []*Token
	start  int
	cur    int
	line   int
}

func NewScanner(source []byte) *Scanner {
	scanner := new(Scanner)

	scanner.source = source
	scanner.tokens = make([]*Token, 0)
	scanner.start = 0
	scanner.cur = 0
	scanner.line = 1

	return scanner
}

func (scanner *Scanner) ScanTokens() ([]*Token, error) {
	for !scanner.isEnd() {
		scanner.start = scanner.cur
		if err := scanner.scanToken(); err != nil {
			return nil, err
		}
	}

	scanner.tokens = append(scanner.tokens, &Token{
		ToKenType: TOKEN_EOF,
		Line:      scanner.line,
	})

	return scanner.tokens, nil
}

func (scanner *Scanner) scanToken() error {
	c := scanner.advance()
	switch c {
	case '(':
		scanner.addToken(TOKEN_LEFT_PAREN)
	case ')':
		scanner.addToken(TOKEN_RIGHT_PAREN)
	case '{':
		scanner.addToken(TOKEN_LEFT_BRACE)
	case '}':
		scanner.addToken(TOKEN_RIGHT_BRACE)
	case '[':
		scanner.addToken(TOKEN_LEFT_BRACKET)
	case ']':
		scanner.addToken(TOKEN_RIGHT_BRACKET)

	case '&':
		if scanner.match('&') {
			if scanner.match('=') {
				scanner.addToken(TOKEN_AND_AND_EQUAL)
			} else {
				scanner.addToken(TOKEN_AND_AND)
			}
		} else if scanner.match('=') {
			scanner.addToken(TOKEN_AND_EQUAL)
		} else {
			scanner.addToken(TOKEN_AND)
		}
	case '!':
		if scanner.match('=') {
			if scanner.match('=') {
				scanner.addToken(TOKEN_BANG_EQUAL_EQUAL)
			} else {
				scanner.addToken(TOKEN_BANG_EQUAL)
			}
		} else {
			scanner.addToken(TOKEN_BANG)
		}
	case ':':
		scanner.addToken(TOKEN_COLON)
	case ',':
		scanner.addToken(TOKEN_COMMA)
	case '.':
		if scanner.peek() == '.' && scanner.peekNext() == '.' {
			scanner.addToken(TOKEN_DOT_DOT_DOT)
			scanner.advance()
			scanner.advance()
		} else {
			scanner.addToken(TOKEN_DOT)
		}
	case '=':
		if scanner.match('=') {
			if scanner.match('=') {
				scanner.addToken(TOKEN_EQUAL_EQUAL_EQUAL)
			} else {
				scanner.addToken(TOKEN_EQUAL_EQUAL)
			}
		} else {
			scanner.addToken(TOKEN_EQUAL)
		}
	case '>':
		if scanner.match('=') {
			scanner.addToken(TOKEN_GREATER_EQUAL)
		} else if scanner.match('>') {
			if scanner.match('=') {
				scanner.addToken(TOKEN_GREATER_GREATER_EQUAL)
			} else if scanner.match('>') {
				if scanner.match('=') {
					scanner.addToken(TOKEN_GREATER_GREATER_GREATER_EQUAL)
				} else {
					scanner.addToken(TOKEN_GREATER_GREATER_GREATER)
				}
			} else {
				scanner.addToken(TOKEN_GREATER_GREATER)
			}
		} else {
			scanner.addToken(TOKEN_GREATER)
		}
	case '#':
		if scanner.match('!') {
			for scanner.peek() != '\n' && !scanner.isEnd() {
				scanner.advance()
			}
			text := string(scanner.source[scanner.start:scanner.cur])
			scanner.addTokenWithLiteral(TOKEN_HASH_BANG, text)
		} else {
			scanner.addToken(TOKEN_HASH)
		}
	case '^':
		if scanner.match('=') {
			scanner.addToken(TOKEN_HAT_EQUAL)
		} else {
			scanner.addToken(TOKEN_HAT)
		}
	case '<':
		if scanner.match('=') {
			scanner.addToken(TOKEN_LESS_EQUAL)
		} else if scanner.match('<') {
			if scanner.match('=') {
				scanner.addToken(TOKEN_LESS_LESS_EQUAL)
			} else {
				scanner.addToken(TOKEN_LESS_LESS)
			}
		} else {
			scanner.addToken(TOKEN_LESS)
		}
	case '-':
		if scanner.match('-') {
			scanner.addToken(TOKEN_MINUS_MINUS)
		} else if scanner.match('=') {
			scanner.addToken(TOKEN_MINUS_EQUAL)
		} else {
			scanner.addToken(TOKEN_MINUS)
		}
	case '%':
		if scanner.match('=') {
			scanner.addToken(TOKEN_PERCENT_EQUAL)
		} else {
			scanner.addToken(TOKEN_PERCENT)
		}
	case '|':
		if scanner.match('|') {
			if scanner.match('=') {
				scanner.addToken(TOKEN_PIPE_PIPE_EQUAL)
			} else {
				scanner.addToken(TOKEN_PIPE_PIPE)
			}
		} else if scanner.match('=') {
			scanner.addToken(TOKEN_PIPE_EQUAL)
		} else {
			scanner.addToken(TOKEN_PIPE)
		}
	case '+':
		if scanner.match('+') {
			scanner.addToken(TOKEN_PLUS_PLUS)
		} else if scanner.match('=') {
			scanner.addToken(TOKEN_PLUS_EQUAL)
		} else {
			scanner.addToken(TOKEN_PLUS)
		}
	case '?':
		if scanner.match('?') {
			if scanner.match('=') {
				scanner.addToken(TOKEN_QUESTION_QUESTION_EQUAL)
			} else {
				scanner.addToken(TOKEN_QUESTION_QUESTION)
			}
		} else if scanner.match('.') {
			scanner.addToken(TOKEN_QUESTION_DOT)
		} else {
			scanner.addToken(TOKEN_QUESTION)
		}
	case ';':
		scanner.addToken(TOKEN_SEMICOLON)
	case '/':
		switch {
		case scanner.match('/'):
			for scanner.peek() != '\n' && !scanner.isEnd() {
				scanner.advance()
			}
			text := string(scanner.source[scanner.start:scanner.cur])
			scanner.addTokenWithLiteral(TOKEN_SINGLE_LINE_COMMENT, text)
		case scanner.match('*'):
			isClosed := false
			for !scanner.isEnd() {
				if scanner.match('*') && scanner.match('/') {
					isClosed = true
					break
				}
				scanner.advance()
			}
			if !isClosed {
				return errors.New("invalid or unexpected token")
			}
			text := string(scanner.source[scanner.start:scanner.cur])
			scanner.addTokenWithLiteral(TOKEN_MULTI_LINE_COMMENT, text)
		case scanner.match('='):
			scanner.addToken(TOKEN_SLASH_EQUAL)
		default:
			scanner.addToken(TOKEN_SLASH)
		}
	case '*':
		if scanner.match('*') {
			if scanner.match('=') {
				scanner.addToken(TOKEN_STAR_STAR_EQUAL)
			} else {
				scanner.addToken(TOKEN_STAR_STAR)
			}
		} else if scanner.match('=') {
			scanner.addToken(TOKEN_STAR_EQUAL)
		} else {
			scanner.addToken(TOKEN_STAR)
		}
	case '~':
		scanner.addToken(TOKEN_TILDE)

	case '"', '\'', '`':
		if err := scanner.string(c); err != nil {
			return err
		}

	case '\n':
		scanner.addToken(TOKEN_NEW_LINE)
		scanner.line++
	case ' ', '\t', '\v', '\f', 0xA0, 0xFEFF:
		// skip white-spaces
		if scanner.isSpace(scanner.peek()) {
			scanner.advance()
		}
		scanner.addToken(TOKEN_SPACE)

	default:
		if scanner.isDigit(c) {
			scanner.number()
		} else if scanner.isAlpha(c) {
			scanner.identifier()
		} else {
			return errors.New("unexpected character")
		}
	}

	return nil
}

func (scanner *Scanner) addToken(tok TokenType) {
	scanner.addTokenWithLiteral(tok, "")
}

func (scanner *Scanner) addTokenWithLiteral(tok TokenType, lit string) {
	scanner.tokens = append(scanner.tokens, &Token{
		ToKenType: tok,
		Line:      scanner.line,
		Literal:   lit,
	})
}

func (scanner *Scanner) isEnd() bool {
	return scanner.cur >= len(scanner.source)
}

func (scanner *Scanner) advance() rune {
	r, width := utf8.DecodeRune(scanner.source[scanner.cur:])
	scanner.cur += width
	return r
}

func (scanner *Scanner) match(expected rune) bool {
	if scanner.isEnd() {
		return false
	}
	r, width := utf8.DecodeRune(scanner.source[scanner.cur:])
	if r != expected {
		return false
	}

	scanner.cur += width
	return true
}

func (scanner *Scanner) peek() rune {
	if scanner.isEnd() {
		return 0
	}
	r, _ := utf8.DecodeRune(scanner.source[scanner.cur:])
	return r
}

func (scanner *Scanner) peekNext() rune {
	if scanner.cur+1 >= len(scanner.source) {
		return 0
	}
	_, width := utf8.DecodeRune(scanner.source[scanner.cur:])
	r, _ := utf8.DecodeRune(scanner.source[scanner.cur+width:])
	return r
}

func (scanner *Scanner) number() {
	for scanner.isDigit(scanner.peek()) {
		scanner.advance()
	}

	if scanner.peek() == '.' && scanner.isDigit(scanner.peekNext()) {
		scanner.advance()

		for scanner.isDigit(scanner.peek()) {
			scanner.advance()
		}
	}

	scanner.addTokenWithLiteral(TOKEN_NUMBER, string(scanner.source[scanner.start:scanner.cur]))
}

func (scanner *Scanner) string(quote rune) error {
	isEscape := false
	tok := scanner.peek()
	for !scanner.isEnd() {
		if tok == '\n' {
			return errors.New("unexpected new line")
		}

		if !isEscape {
			if tok == quote {
				break
			} else if tok == '\\' {
				isEscape = true
			}
		} else {
			isEscape = false
		}

		scanner.advance()
		tok = scanner.peek()
	}

	if scanner.isEnd() {
		return errors.New("unterminated string")
	}

	scanner.advance()

	value := scanner.source[scanner.start+1 : scanner.cur-1]
	scanner.addTokenWithLiteral(TOKEN_STRING, string(value))

	return nil
}

func (scanner *Scanner) identifier() {
	for scanner.isAlphaNumeric(scanner.peek()) {
		scanner.advance()
	}

	text := string(scanner.source[scanner.start:scanner.cur])
	tokenType, ok := keywords[text]
	if ok {
		scanner.addToken(tokenType)
	} else {
		scanner.addTokenWithLiteral(TOKEN_IDENTIFIER, text)
	}
}

func (scanner *Scanner) isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func (scanner *Scanner) isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_' || c == '$'
}

func (scanner *Scanner) isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\v' || c == '\f' || c == 0xA0 || c == 0xFEFF
}

func (scanner *Scanner) isAlphaNumeric(c rune) bool {
	return scanner.isAlpha(c) || scanner.isDigit(c)
}
