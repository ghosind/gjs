package lexer

import (
	"unicode/utf8"

	"github.com/ghosind/gjs/errors"
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

func (l *Lexer) ScanToken() (*token.Token, error) {
	if !l.isEnd() {
		l.start = l.cur
		l.width = 0
		tok, err := l.scanToken()
		if err != nil {
			return nil, err
		}

		l.col += l.width
		return tok, nil
	}

	return &token.Token{
		TokenType: token.TOKEN_EOF,
		Line:      l.line,
		Col:       l.col,
	}, nil
}

func (l *Lexer) scanToken() (*token.Token, error) {
	var tok *token.Token

	c := l.advance()
	switch c {
	case '(':
		tok = l.newToken(token.TOKEN_LEFT_PAREN)
	case ')':
		tok = l.newToken(token.TOKEN_RIGHT_PAREN)
	case '{':
		tok = l.newToken(token.TOKEN_LEFT_BRACE)
	case '}':
		tok = l.newToken(token.TOKEN_RIGHT_BRACE)
	case '[':
		tok = l.newToken(token.TOKEN_LEFT_BRACKET)
	case ']':
		tok = l.newToken(token.TOKEN_RIGHT_BRACKET)
	case '&':
		if l.match('&') {
			if l.match('=') {
				tok = l.newToken(token.TOKEN_AND_AND_EQUAL)
			} else {
				tok = l.newToken(token.TOKEN_AND_AND)
			}
		} else if l.match('=') {
			tok = l.newToken(token.TOKEN_AND_EQUAL)
		} else {
			tok = l.newToken(token.TOKEN_AND)
		}
	case '!':
		if l.match('=') {
			if l.match('=') {
				tok = l.newToken(token.TOKEN_BANG_EQUAL_EQUAL)
			} else {
				tok = l.newToken(token.TOKEN_BANG_EQUAL)
			}
		} else {
			tok = l.newToken(token.TOKEN_BANG)
		}
	case ':':
		tok = l.newToken(token.TOKEN_COLON)
	case ',':
		tok = l.newToken(token.TOKEN_COMMA)
	case '.':
		if l.peek() == '.' && l.peekNext() == '.' {
			tok = l.newToken(token.TOKEN_DOT_DOT_DOT)
			l.advance()
			l.advance()
		} else {
			tok = l.newToken(token.TOKEN_DOT)
		}
	case '=':
		if l.match('=') {
			if l.match('=') {
				tok = l.newToken(token.TOKEN_EQUAL_EQUAL_EQUAL)
			} else {
				tok = l.newToken(token.TOKEN_EQUAL_EQUAL)
			}
		} else {
			tok = l.newToken(token.TOKEN_EQUAL)
		}
	case '>':
		if l.match('=') {
			tok = l.newToken(token.TOKEN_GREATER_EQUAL)
		} else if l.match('>') {
			if l.match('=') {
				tok = l.newToken(token.TOKEN_GREATER_GREATER_EQUAL)
			} else if l.match('>') {
				if l.match('=') {
					tok = l.newToken(token.TOKEN_GREATER_GREATER_GREATER_EQUAL)
				} else {
					tok = l.newToken(token.TOKEN_GREATER_GREATER_GREATER)
				}
			} else {
				tok = l.newToken(token.TOKEN_GREATER_GREATER)
			}
		} else {
			tok = l.newToken(token.TOKEN_GREATER)
		}
	case '#':
		if l.match('!') {
			for !l.isLineTerminator(l.peek()) && !l.isEnd() {
				l.advance()
			}
			text := string(l.source[l.start+2 : l.cur])
			tok = l.newTokenWithLiteral(token.TOKEN_HASH_BANG, text)
		} else {
			tok = l.newToken(token.TOKEN_HASH)
		}
	case '^':
		if l.match('=') {
			tok = l.newToken(token.TOKEN_HAT_EQUAL)
		} else {
			tok = l.newToken(token.TOKEN_HAT)
		}
	case '<':
		if l.match('=') {
			tok = l.newToken(token.TOKEN_LESS_EQUAL)
		} else if l.match('<') {
			if l.match('=') {
				tok = l.newToken(token.TOKEN_LESS_LESS_EQUAL)
			} else {
				tok = l.newToken(token.TOKEN_LESS_LESS)
			}
		} else {
			tok = l.newToken(token.TOKEN_LESS)
		}
	case '-':
		if l.match('-') {
			tok = l.newToken(token.TOKEN_MINUS_MINUS)
		} else if l.match('=') {
			tok = l.newToken(token.TOKEN_MINUS_EQUAL)
		} else {
			tok = l.newToken(token.TOKEN_MINUS)
		}
	case '%':
		if l.match('=') {
			tok = l.newToken(token.TOKEN_PERCENT_EQUAL)
		} else {
			tok = l.newToken(token.TOKEN_PERCENT)
		}
	case '|':
		if l.match('|') {
			if l.match('=') {
				tok = l.newToken(token.TOKEN_PIPE_PIPE_EQUAL)
			} else {
				tok = l.newToken(token.TOKEN_PIPE_PIPE)
			}
		} else if l.match('=') {
			tok = l.newToken(token.TOKEN_PIPE_EQUAL)
		} else {
			tok = l.newToken(token.TOKEN_PIPE)
		}
	case '+':
		if l.match('+') {
			tok = l.newToken(token.TOKEN_PLUS_PLUS)
		} else if l.match('=') {
			tok = l.newToken(token.TOKEN_PLUS_EQUAL)
		} else {
			tok = l.newToken(token.TOKEN_PLUS)
		}
	case '?':
		if l.match('?') {
			if l.match('=') {
				tok = l.newToken(token.TOKEN_QUESTION_QUESTION_EQUAL)
			} else {
				tok = l.newToken(token.TOKEN_QUESTION_QUESTION)
			}
		} else if l.match('.') {
			tok = l.newToken(token.TOKEN_QUESTION_DOT)
		} else {
			tok = l.newToken(token.TOKEN_QUESTION)
		}
	case ';':
		tok = l.newToken(token.TOKEN_SEMICOLON)
	case '/':
		switch {
		case l.match('/'):
			for l.peek() != '\n' && !l.isEnd() {
				l.advance()
			}
			tok = l.newToken(token.TOKEN_SINGLE_LINE_COMMENT)
		case l.match('*'):
			isClosed := false
			line := l.line
			col := l.col
			width := l.width
			for !l.isEnd() {
				if l.match('*') && l.match('/') {
					isClosed = true
					break
				} else if l.isLineTerminator(l.peek()) {
					line++
					col = 0
					width = 0
				}
				l.advance()
			}
			if !isClosed {
				return nil, l.newSyntaxError()
			}
			tok = l.newToken(token.TOKEN_MULTI_LINE_COMMENT)
			l.line = line
			l.col = col + 2
			l.width = width + 2
		case l.match('='):
			tok = l.newToken(token.TOKEN_SLASH_EQUAL)
		default:
			tok = l.newToken(token.TOKEN_SLASH)
		}
	case '*':
		if l.match('*') {
			if l.match('=') {
				tok = l.newToken(token.TOKEN_STAR_STAR_EQUAL)
			} else {
				tok = l.newToken(token.TOKEN_STAR_STAR)
			}
		} else if l.match('=') {
			tok = l.newToken(token.TOKEN_STAR_EQUAL)
		} else {
			tok = l.newToken(token.TOKEN_STAR)
		}
	case '~':
		tok = l.newToken(token.TOKEN_TILDE)
	case '"', '\'', '`':
		if t, err := l.string(c); err != nil {
			return nil, err
		} else {
			tok = t
		}
	case '\n', '\r', 0x2028, 0x2029:
		if c == '\r' {
			l.match('\n')
		}
		tok = l.newToken(token.TOKEN_NEW_LINE)
		l.line++
		l.col = 0
	case ' ', '\t', '\v', '\f', 0xA0, 0xFEFF:
		// skip white-spaces
		if l.isSpace(l.peek()) {
			l.advance()
		}
		tok = l.newToken(token.TOKEN_SPACE)
	default:
		if l.isAlpha(c) || c == '$' || c == '_' {
			tok = l.identifier()
		} else if l.isDigit(c) {
			t, err := l.number()
			if err != nil {
				return nil, err
			}
			tok = t
		} else {
			return nil, l.newSyntaxError()
		}
	}

	return tok, nil
}

func (l *Lexer) newToken(tok token.TokenType) *token.Token {
	text := string(l.source[l.start:l.cur])
	return l.newTokenWithLiteral(tok, text)
}

func (l *Lexer) newTokenWithLiteral(tok token.TokenType, lit string) *token.Token {
	return &token.Token{
		TokenType: tok,
		Line:      l.line,
		Col:       l.col,
		Literal:   lit,
	}
}

func (l *Lexer) isEnd() bool {
	return l.cur >= len(l.source)
}

func (l *Lexer) advance() rune {
	r, width := utf8.DecodeRune(l.source[l.cur:])
	l.cur += width
	l.width++
	return r
}

func (l *Lexer) match(expected rune) bool {
	if l.isEnd() {
		return false
	}
	r, width := utf8.DecodeRune(l.source[l.cur:])
	if r != expected {
		return false
	}

	l.cur += width
	l.width++
	return true
}

func (l *Lexer) peek() rune {
	if l.isEnd() {
		return 0
	}
	r, _ := utf8.DecodeRune(l.source[l.cur:])
	return r
}

func (l *Lexer) peekNext() rune {
	if l.cur+1 >= len(l.source) {
		return 0
	}
	_, width := utf8.DecodeRune(l.source[l.cur:])
	r, _ := utf8.DecodeRune(l.source[l.cur+width:])
	return r
}

func (l *Lexer) number() (*token.Token, error) {
	for l.isDigit(l.peek()) {
		l.advance()
	}

	if l.peek() == '.' && l.isDigit(l.peekNext()) {
		l.advance()

		for l.isDigit(l.peek()) {
			l.advance()
		}
	}

	if l.isAlpha(l.peek()) {
		return nil, l.newSyntaxError()
	}

	return l.newToken(token.TOKEN_NUMBER), nil
}

func (l *Lexer) string(quote rune) (*token.Token, error) {
	isEscape := false
	tok := l.peek()
	for !l.isEnd() {
		if l.isLineTerminator(tok) {
			return nil, l.newSyntaxError()
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

		l.advance()
		tok = l.peek()
	}

	if l.isEnd() {
		return nil, l.newSyntaxError()
	}

	l.advance()

	value := l.source[l.start+1 : l.cur-1]
	return l.newTokenWithLiteral(token.TOKEN_STRING, string(value)), nil
}

func (l *Lexer) identifier() *token.Token {
	for l.isAlphaNumeric(l.peek()) {
		l.advance()
	}

	tokenType := token.LookupIdent(string(l.source[l.start:l.cur]))
	return l.newToken(tokenType)
}

func (l *Lexer) isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func (l *Lexer) isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_' || c == '$'
}

func (l *Lexer) isAlphaNumeric(c rune) bool {
	return l.isAlpha(c) || l.isDigit(c)
}

func (l *Lexer) isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\v' || c == '\f' || c == 0xA0 || c == 0xFEFF
}

func (l *Lexer) isLineTerminator(c rune) bool {
	switch c {
	case '\n', 0x2028, 0x2029:
		return true
	case '\r':
		l.match('\n')
		return true
	}
	return false
}

func (l *Lexer) newSyntaxError() error {
	line := l.getCurrentLine()
	return errors.NewLexerError(line, l.col)
}

func (l *Lexer) getCurrentLine() string {
	lineStart := l.cur - l.width
	for lineStart > 0 && l.source[lineStart-1] != '\n' {
		lineStart--
	}
	lineEnd := l.cur
	for lineEnd < len(l.source) && l.source[lineEnd] != '\n' {
		lineEnd++
	}
	return string(l.source[lineStart:lineEnd])
}
