package parser

import "github.com/ghosind/gjs/token"

type SyntaxError struct {
	tok *token.Token
}

func (e *SyntaxError) Error() string {
	return "SyntaxError: unexpected token " + e.tok.Literal
}

func (p *Parser) newSyntaxError(tok *token.Token) error {
	return &SyntaxError{tok: tok}
}
