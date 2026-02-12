package lexer

import (
	"errors"
	"unicode/utf8"

	"github.com/ghosind/gjs/token"
)

type Lexer struct {
	source []byte
	start  int
	cur    int
	line   int
	col    int
	width  int
}

func New(source []byte) *Lexer {
	l := new(Lexer)
	l.source = source
	l.start = 0
	l.cur = 0
	l.line = 1
	l.col = 1
	l.width = 0
	return l
}

func (s *Lexer) ScanToken() (*token.Token, error) {
	if !s.isEnd() {
		s.start = s.cur
		s.width = 0
		tok, err := s.scanToken()
		if err != nil {
			return nil, err
		}

		s.col += s.width
		return tok, nil
	}

	return &token.Token{
		TokenType: token.TOKEN_EOF,
		Line:      s.line,
		Col:       s.col,
	}, nil
}

func (s *Lexer) scanToken() (*token.Token, error) {
	var tok *token.Token

	c := s.advance()
	switch c {
	case '(':
		tok = s.newToken(token.TOKEN_LEFT_PAREN)
	case ')':
		tok = s.newToken(token.TOKEN_RIGHT_PAREN)
	case '{':
		tok = s.newToken(token.TOKEN_LEFT_BRACE)
	case '}':
		tok = s.newToken(token.TOKEN_RIGHT_BRACE)
	case '[':
		tok = s.newToken(token.TOKEN_LEFT_BRACKET)
	case ']':
		tok = s.newToken(token.TOKEN_RIGHT_BRACKET)
	case '&':
		if s.match('&') {
			if s.match('=') {
				tok = s.newToken(token.TOKEN_AND_AND_EQUAL)
			} else {
				tok = s.newToken(token.TOKEN_AND_AND)
			}
		} else if s.match('=') {
			tok = s.newToken(token.TOKEN_AND_EQUAL)
		} else {
			tok = s.newToken(token.TOKEN_AND)
		}
	case '!':
		if s.match('=') {
			if s.match('=') {
				tok = s.newToken(token.TOKEN_BANG_EQUAL_EQUAL)
			} else {
				tok = s.newToken(token.TOKEN_BANG_EQUAL)
			}
		} else {
			tok = s.newToken(token.TOKEN_BANG)
		}
	case ':':
		tok = s.newToken(token.TOKEN_COLON)
	case ',':
		tok = s.newToken(token.TOKEN_COMMA)
	case '.':
		if s.peek() == '.' && s.peekNext() == '.' {
			tok = s.newToken(token.TOKEN_DOT_DOT_DOT)
			s.advance()
			s.advance()
		} else {
			tok = s.newToken(token.TOKEN_DOT)
		}
	case '=':
		if s.match('=') {
			if s.match('=') {
				tok = s.newToken(token.TOKEN_EQUAL_EQUAL_EQUAL)
			} else {
				tok = s.newToken(token.TOKEN_EQUAL_EQUAL)
			}
		} else {
			tok = s.newToken(token.TOKEN_EQUAL)
		}
	case '>':
		if s.match('=') {
			tok = s.newToken(token.TOKEN_GREATER_EQUAL)
		} else if s.match('>') {
			if s.match('=') {
				tok = s.newToken(token.TOKEN_GREATER_GREATER_EQUAL)
			} else if s.match('>') {
				if s.match('=') {
					tok = s.newToken(token.TOKEN_GREATER_GREATER_GREATER_EQUAL)
				} else {
					tok = s.newToken(token.TOKEN_GREATER_GREATER_GREATER)
				}
			} else {
				tok = s.newToken(token.TOKEN_GREATER_GREATER)
			}
		} else {
			tok = s.newToken(token.TOKEN_GREATER)
		}
	case '#':
		if s.match('!') {
			for !s.isLineTerminator(s.peek()) && !s.isEnd() {
				s.advance()
			}
			text := string(s.source[s.start+2 : s.cur])
			tok = s.newTokenWithLiteral(token.TOKEN_HASH_BANG, text)
		} else {
			tok = s.newToken(token.TOKEN_HASH)
		}
	case '^':
		if s.match('=') {
			tok = s.newToken(token.TOKEN_HAT_EQUAL)
		} else {
			tok = s.newToken(token.TOKEN_HAT)
		}
	case '<':
		if s.match('=') {
			tok = s.newToken(token.TOKEN_LESS_EQUAL)
		} else if s.match('<') {
			if s.match('=') {
				tok = s.newToken(token.TOKEN_LESS_LESS_EQUAL)
			} else {
				tok = s.newToken(token.TOKEN_LESS_LESS)
			}
		} else {
			tok = s.newToken(token.TOKEN_LESS)
		}
	case '-':
		if s.match('-') {
			tok = s.newToken(token.TOKEN_MINUS_MINUS)
		} else if s.match('=') {
			tok = s.newToken(token.TOKEN_MINUS_EQUAL)
		} else {
			tok = s.newToken(token.TOKEN_MINUS)
		}
	case '%':
		if s.match('=') {
			tok = s.newToken(token.TOKEN_PERCENT_EQUAL)
		} else {
			tok = s.newToken(token.TOKEN_PERCENT)
		}
	case '|':
		if s.match('|') {
			if s.match('=') {
				tok = s.newToken(token.TOKEN_PIPE_PIPE_EQUAL)
			} else {
				tok = s.newToken(token.TOKEN_PIPE_PIPE)
			}
		} else if s.match('=') {
			tok = s.newToken(token.TOKEN_PIPE_EQUAL)
		} else {
			tok = s.newToken(token.TOKEN_PIPE)
		}
	case '+':
		if s.match('+') {
			tok = s.newToken(token.TOKEN_PLUS_PLUS)
		} else if s.match('=') {
			tok = s.newToken(token.TOKEN_PLUS_EQUAL)
		} else {
			tok = s.newToken(token.TOKEN_PLUS)
		}
	case '?':
		if s.match('?') {
			if s.match('=') {
				tok = s.newToken(token.TOKEN_QUESTION_QUESTION_EQUAL)
			} else {
				tok = s.newToken(token.TOKEN_QUESTION_QUESTION)
			}
		} else if s.match('.') {
			tok = s.newToken(token.TOKEN_QUESTION_DOT)
		} else {
			tok = s.newToken(token.TOKEN_QUESTION)
		}
	case ';':
		tok = s.newToken(token.TOKEN_SEMICOLON)
	case '/':
		switch {
		case s.match('/'):
			for s.peek() != '\n' && !s.isEnd() {
				s.advance()
			}
			tok = s.newToken(token.TOKEN_SINGLE_LINE_COMMENT)
		case s.match('*'):
			isClosed := false
			line := s.line
			col := s.col
			width := s.width
			for !s.isEnd() {
				if s.match('*') && s.match('/') {
					isClosed = true
					break
				} else if s.isLineTerminator(s.peek()) {
					line++
					col = 0
					width = 0
				}
				s.advance()
			}
			if !isClosed {
				return nil, errors.New("invalid or unexpected token")
			}
			tok = s.newToken(token.TOKEN_MULTI_LINE_COMMENT)
			s.line = line
			s.col = col + 2
			s.width = width + 2
		case s.match('='):
			tok = s.newToken(token.TOKEN_SLASH_EQUAL)
		default:
			tok = s.newToken(token.TOKEN_SLASH)
		}
	case '*':
		if s.match('*') {
			if s.match('=') {
				tok = s.newToken(token.TOKEN_STAR_STAR_EQUAL)
			} else {
				tok = s.newToken(token.TOKEN_STAR_STAR)
			}
		} else if s.match('=') {
			tok = s.newToken(token.TOKEN_STAR_EQUAL)
		} else {
			tok = s.newToken(token.TOKEN_STAR)
		}
	case '~':
		tok = s.newToken(token.TOKEN_TILDE)
	case '"', '\'', '`':
		if t, err := s.string(c); err != nil {
			return nil, err
		} else {
			tok = t
		}
	case '\n', '\r', 0x2028, 0x2029:
		if c == '\r' {
			s.match('\n')
		}
		tok = s.newToken(token.TOKEN_NEW_LINE)
		s.line++
		s.col = 0
	case ' ', '\t', '\v', '\f', 0xA0, 0xFEFF:
		// skip white-spaces
		if s.isSpace(s.peek()) {
			s.advance()
		}
		tok = s.newToken(token.TOKEN_SPACE)
	default:
		if s.isDigit(c) {
			tok = s.number()
		} else if s.isAlpha(c) {
			tok = s.identifier()
		} else {
			return nil, errors.New("unexpected character")
		}
	}

	return tok, nil
}

func (s *Lexer) newToken(tok token.TokenType) *token.Token {
	text := string(s.source[s.start:s.cur])
	return s.newTokenWithLiteral(tok, text)
}

func (s *Lexer) newTokenWithLiteral(tok token.TokenType, lit string) *token.Token {
	return &token.Token{
		TokenType: tok,
		Line:      s.line,
		Col:       s.col,
		Literal:   lit,
	}
}

func (s *Lexer) isEnd() bool {
	return s.cur >= len(s.source)
}

func (s *Lexer) advance() rune {
	r, width := utf8.DecodeRune(s.source[s.cur:])
	s.cur += width
	s.width++
	return r
}

func (s *Lexer) match(expected rune) bool {
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

func (s *Lexer) peek() rune {
	if s.isEnd() {
		return 0
	}
	r, _ := utf8.DecodeRune(s.source[s.cur:])
	return r
}

func (s *Lexer) peekNext() rune {
	if s.cur+1 >= len(s.source) {
		return 0
	}
	_, width := utf8.DecodeRune(s.source[s.cur:])
	r, _ := utf8.DecodeRune(s.source[s.cur+width:])
	return r
}

func (s *Lexer) number() *token.Token {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && s.isDigit(s.peekNext()) {
		s.advance()

		for s.isDigit(s.peek()) {
			s.advance()
		}
	}

	return s.newToken(token.TOKEN_NUMBER)
}

func (s *Lexer) string(quote rune) (*token.Token, error) {
	isEscape := false
	tok := s.peek()
	for !s.isEnd() {
		if s.isLineTerminator(tok) {
			return nil, errors.New("unexpected new line")
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
		return nil, errors.New("unterminated string")
	}

	s.advance()

	value := s.source[s.start+1 : s.cur-1]
	return s.newTokenWithLiteral(token.TOKEN_STRING, string(value)), nil
}

func (s *Lexer) identifier() *token.Token {
	for s.isAlphaNumeric(s.peek()) {
		s.advance()
	}

	tokenType := token.LookupIdent(string(s.source[s.start:s.cur]))
	return s.newToken(tokenType)
}

func (s *Lexer) isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func (s *Lexer) isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_' || c == '$'
}

func (s *Lexer) isAlphaNumeric(c rune) bool {
	return s.isAlpha(c) || s.isDigit(c)
}

func (s *Lexer) isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\v' || c == '\f' || c == 0xA0 || c == 0xFEFF
}

func (s *Lexer) isLineTerminator(c rune) bool {
	switch c {
	case '\n', 0x2028, 0x2029:
		return true
	case '\r':
		s.match('\n')
		return true
	}
	return false
}
