package syntax

import (
	"errors"
)

type Scanner struct {
	source string
	tokens []*Token
	start  int
	cur    int
	line   int
}

func NewScanner(source string) *Scanner {
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
		scanner.string(c)

	case '\n':
		scanner.addToken(TOKEN_NEW_LINE)
		scanner.line++
	case ' ', '\r', '\t':
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

func (scanner *Scanner) advance() byte {
	c := scanner.source[scanner.cur]
	scanner.cur++
	return c
}

func (scanner *Scanner) match(expected byte) bool {
	if scanner.isEnd() {
		return false
	}
	if scanner.source[scanner.cur] != expected {
		return false
	}

	scanner.cur++
	return true
}

func (scanner *Scanner) peek() byte {
	if scanner.isEnd() {
		return 0
	}
	return scanner.source[scanner.cur]
}

func (scanner *Scanner) peekNext() byte {
	if scanner.cur+1 >= len(scanner.source) {
		return 0
	}
	return scanner.source[scanner.cur+1]
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

	scanner.addTokenWithLiteral(TOKEN_NUMBER, scanner.source[scanner.start:scanner.cur])
}

func (scanner *Scanner) string(quote byte) error {
	for scanner.peek() != quote && !scanner.isEnd() {
		if scanner.peek() == '\n' {
			return errors.New("unexpected new line")
		}
		scanner.advance()
	}

	if scanner.isEnd() {
		return errors.New("unterminated string")
	}

	scanner.advance()

	value := scanner.source[scanner.start+1 : scanner.cur-1]
	scanner.addTokenWithLiteral(TOKEN_STRING, value)

	return nil
}

func (scanner *Scanner) identifier() {
	for scanner.isAlphaNumeric(scanner.peek()) {
		scanner.advance()
	}

	text := scanner.source[scanner.start:scanner.cur]
	tokenType, ok := keywords[text]
	if ok {
		scanner.addToken(tokenType)
	} else {
		scanner.addTokenWithLiteral(TOKEN_IDENTIFIER, text)
	}
}

func (scanner *Scanner) isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (scanner *Scanner) isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_' || c == '$'
}

func (scanner *Scanner) isSpace(c byte) bool {
	return c == ' ' || c == '\r' || c == '\t'
}

func (scanner *Scanner) isAlphaNumeric(c byte) bool {
	return scanner.isAlpha(c) || scanner.isDigit(c)
}
