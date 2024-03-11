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
	col    int
	width  int
}

func (s *Scanner) Init(source []byte) {
	s.source = source
	s.tokens = make([]*Token, 0)
	s.start = 0
	s.cur = 0
	s.line = 1
	s.col = 1
	s.width = 0
}

func (s *Scanner) ScanTokens() ([]*Token, error) {
	for !s.isEnd() {
		s.start = s.cur
		s.width = 0
		if err := s.scanToken(); err != nil {
			return nil, err
		}

		s.col += s.width
	}

	s.tokens = append(s.tokens, &Token{
		ToKenType: TOKEN_EOF,
		Line:      s.line,
		Col:       s.col,
	})

	return s.tokens, nil
}

func (s *Scanner) scanToken() error {
	c := s.advance()
	switch c {
	case '(':
		s.addToken(TOKEN_LEFT_PAREN)
	case ')':
		s.addToken(TOKEN_RIGHT_PAREN)
	case '{':
		s.addToken(TOKEN_LEFT_BRACE)
	case '}':
		s.addToken(TOKEN_RIGHT_BRACE)
	case '[':
		s.addToken(TOKEN_LEFT_BRACKET)
	case ']':
		s.addToken(TOKEN_RIGHT_BRACKET)
	case '&':
		if s.match('&') {
			if s.match('=') {
				s.addToken(TOKEN_AND_AND_EQUAL)
			} else {
				s.addToken(TOKEN_AND_AND)
			}
		} else if s.match('=') {
			s.addToken(TOKEN_AND_EQUAL)
		} else {
			s.addToken(TOKEN_AND)
		}
	case '!':
		if s.match('=') {
			if s.match('=') {
				s.addToken(TOKEN_BANG_EQUAL_EQUAL)
			} else {
				s.addToken(TOKEN_BANG_EQUAL)
			}
		} else {
			s.addToken(TOKEN_BANG)
		}
	case ':':
		s.addToken(TOKEN_COLON)
	case ',':
		s.addToken(TOKEN_COMMA)
	case '.':
		if s.peek() == '.' && s.peekNext() == '.' {
			s.addToken(TOKEN_DOT_DOT_DOT)
			s.advance()
			s.advance()
		} else {
			s.addToken(TOKEN_DOT)
		}
	case '=':
		if s.match('=') {
			if s.match('=') {
				s.addToken(TOKEN_EQUAL_EQUAL_EQUAL)
			} else {
				s.addToken(TOKEN_EQUAL_EQUAL)
			}
		} else {
			s.addToken(TOKEN_EQUAL)
		}
	case '>':
		if s.match('=') {
			s.addToken(TOKEN_GREATER_EQUAL)
		} else if s.match('>') {
			if s.match('=') {
				s.addToken(TOKEN_GREATER_GREATER_EQUAL)
			} else if s.match('>') {
				if s.match('=') {
					s.addToken(TOKEN_GREATER_GREATER_GREATER_EQUAL)
				} else {
					s.addToken(TOKEN_GREATER_GREATER_GREATER)
				}
			} else {
				s.addToken(TOKEN_GREATER_GREATER)
			}
		} else {
			s.addToken(TOKEN_GREATER)
		}
	case '#':
		if s.match('!') {
			for !s.isLineTerminator(s.peek()) && !s.isEnd() {
				s.advance()
			}
			text := string(s.source[s.start+2 : s.cur])
			s.addTokenWithLiteral(TOKEN_HASH_BANG, text)
		} else {
			s.addToken(TOKEN_HASH)
		}
	case '^':
		if s.match('=') {
			s.addToken(TOKEN_HAT_EQUAL)
		} else {
			s.addToken(TOKEN_HAT)
		}
	case '<':
		if s.match('=') {
			s.addToken(TOKEN_LESS_EQUAL)
		} else if s.match('<') {
			if s.match('=') {
				s.addToken(TOKEN_LESS_LESS_EQUAL)
			} else {
				s.addToken(TOKEN_LESS_LESS)
			}
		} else {
			s.addToken(TOKEN_LESS)
		}
	case '-':
		if s.match('-') {
			s.addToken(TOKEN_MINUS_MINUS)
		} else if s.match('=') {
			s.addToken(TOKEN_MINUS_EQUAL)
		} else {
			s.addToken(TOKEN_MINUS)
		}
	case '%':
		if s.match('=') {
			s.addToken(TOKEN_PERCENT_EQUAL)
		} else {
			s.addToken(TOKEN_PERCENT)
		}
	case '|':
		if s.match('|') {
			if s.match('=') {
				s.addToken(TOKEN_PIPE_PIPE_EQUAL)
			} else {
				s.addToken(TOKEN_PIPE_PIPE)
			}
		} else if s.match('=') {
			s.addToken(TOKEN_PIPE_EQUAL)
		} else {
			s.addToken(TOKEN_PIPE)
		}
	case '+':
		if s.match('+') {
			s.addToken(TOKEN_PLUS_PLUS)
		} else if s.match('=') {
			s.addToken(TOKEN_PLUS_EQUAL)
		} else {
			s.addToken(TOKEN_PLUS)
		}
	case '?':
		if s.match('?') {
			if s.match('=') {
				s.addToken(TOKEN_QUESTION_QUESTION_EQUAL)
			} else {
				s.addToken(TOKEN_QUESTION_QUESTION)
			}
		} else if s.match('.') {
			s.addToken(TOKEN_QUESTION_DOT)
		} else {
			s.addToken(TOKEN_QUESTION)
		}
	case ';':
		s.addToken(TOKEN_SEMICOLON)
	case '/':
		switch {
		case s.match('/'):
			for s.peek() != '\n' && !s.isEnd() {
				s.advance()
			}
			s.addToken(TOKEN_SINGLE_LINE_COMMENT)
		case s.match('*'):
			isClosed := false
			for !s.isEnd() {
				if s.match('*') && s.match('/') {
					isClosed = true
					break
				} else if s.isLineTerminator(s.peek()) {
					s.line++
					s.col = 0
					s.width = 0
				}
				s.advance()
			}
			if !isClosed {
				return errors.New("invalid or unexpected token")
			}
			s.addToken(TOKEN_MULTI_LINE_COMMENT)
		case s.match('='):
			s.addToken(TOKEN_SLASH_EQUAL)
		default:
			s.addToken(TOKEN_SLASH)
		}
	case '*':
		if s.match('*') {
			if s.match('=') {
				s.addToken(TOKEN_STAR_STAR_EQUAL)
			} else {
				s.addToken(TOKEN_STAR_STAR)
			}
		} else if s.match('=') {
			s.addToken(TOKEN_STAR_EQUAL)
		} else {
			s.addToken(TOKEN_STAR)
		}
	case '~':
		s.addToken(TOKEN_TILDE)
	case '"', '\'', '`':
		if err := s.string(c); err != nil {
			return err
		}
	case '\n', '\r', 0x2028, 0x2029:
		if c == '\r' {
			s.match('\n')
		}
		s.addToken(TOKEN_NEW_LINE)
		s.line++
		s.col = 0
	case ' ', '\t', '\v', '\f', 0xA0, 0xFEFF:
		// skip white-spaces
		if s.isSpace(s.peek()) {
			s.advance()
		}
		s.addToken(TOKEN_SPACE)
	default:
		if s.isDigit(c) {
			s.number()
		} else if s.isAlpha(c) {
			s.identifier()
		} else {
			return errors.New("unexpected character")
		}
	}

	return nil
}

func (s *Scanner) addToken(tok TokenType) {
	text := string(s.source[s.start:s.cur])
	s.addTokenWithLiteral(tok, text)
}

func (s *Scanner) addTokenWithLiteral(tok TokenType, lit string) {
	s.tokens = append(s.tokens, &Token{
		ToKenType: tok,
		Line:      s.line,
		Col:       s.col,
		Literal:   lit,
	})
}

func (s *Scanner) isEnd() bool {
	return s.cur >= len(s.source)
}

func (s *Scanner) advance() rune {
	r, width := utf8.DecodeRune(s.source[s.cur:])
	s.cur += width
	s.width++
	return r
}

func (s *Scanner) match(expected rune) bool {
	if s.isEnd() {
		return false
	}
	r, width := utf8.DecodeRune(s.source[s.cur:])
	if r != expected {
		return false
	}

	s.cur += width
	s.width++
	return true
}

func (s *Scanner) peek() rune {
	if s.isEnd() {
		return 0
	}
	r, _ := utf8.DecodeRune(s.source[s.cur:])
	return r
}

func (s *Scanner) peekNext() rune {
	if s.cur+1 >= len(s.source) {
		return 0
	}
	_, width := utf8.DecodeRune(s.source[s.cur:])
	r, _ := utf8.DecodeRune(s.source[s.cur+width:])
	return r
}

func (s *Scanner) number() {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && s.isDigit(s.peekNext()) {
		s.advance()

		for s.isDigit(s.peek()) {
			s.advance()
		}
	}

	s.addToken(TOKEN_NUMBER)
}

func (s *Scanner) string(quote rune) error {
	isEscape := false
	tok := s.peek()
	for !s.isEnd() {
		if s.isLineTerminator(tok) {
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

		s.advance()
		tok = s.peek()
	}

	if s.isEnd() {
		return errors.New("unterminated string")
	}

	s.advance()

	value := s.source[s.start+1 : s.cur-1]
	s.addTokenWithLiteral(TOKEN_STRING, string(value))

	return nil
}

func (s *Scanner) identifier() {
	for s.isAlphaNumeric(s.peek()) {
		s.advance()
	}

	tokenType, ok := keywords[string(s.source[s.start:s.cur])]
	if ok {
		s.addToken(tokenType)
	} else {
		s.addToken(TOKEN_IDENTIFIER)
	}
}

func (s *Scanner) isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func (s *Scanner) isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_' || c == '$'
}

func (s *Scanner) isAlphaNumeric(c rune) bool {
	return s.isAlpha(c) || s.isDigit(c)
}

func (s *Scanner) isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\v' || c == '\f' || c == 0xA0 || c == 0xFEFF
}

func (s *Scanner) isLineTerminator(c rune) bool {
	if c == '\n' || c == 0x2028 || c == 0x2029 {
		return true
	} else if c == '\r' {
		s.match('\n')
		return true
	}
	return false
}
