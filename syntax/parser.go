package syntax

import (
	"errors"
)

type Parser struct {
	tokens []*Token
	cur    int
}

func (p *Parser) Init(tokens []*Token) {
	p.tokens = tokens
	p.cur = 0
}

func (p *Parser) expression() (Expr, error) {
	// TODO: comma operation
	return p.assignmentExpr()
}

func (p *Parser) assignmentExpr() (Expr, error) {
	cur := p.cur
	expr, err := p.leftHandSideExpr()
	if err != nil {
		return nil, err
	} else if expr != nil {
		p.skip()
		if p.match(TOKEN_EQUAL,
			TOKEN_STAR_EQUAL,
			TOKEN_SLASH_EQUAL,
			TOKEN_PERCENT_EQUAL,
			TOKEN_PLUS_EQUAL,
			TOKEN_MINUS_EQUAL,
			TOKEN_LESS_LESS_EQUAL,
			TOKEN_GREATER_GREATER_EQUAL,
			TOKEN_GREATER_GREATER_GREATER_EQUAL,
			TOKEN_AND_EQUAL,
			TOKEN_HAT_EQUAL,
			TOKEN_PIPE_EQUAL,
			TOKEN_STAR_STAR_EQUAL,
			TOKEN_AND_AND_EQUAL,
			TOKEN_PIPE_PIPE_EQUAL,
			TOKEN_QUESTION_QUESTION_EQUAL,
		) {
			op := p.previous()
			right, err := p.assignmentExpr()
			if err != nil {
				return nil, err
			}

			return &OpExpr{Op: op.ToKenType, Left: expr, Right: right}, nil
		}
	}
	p.cur = cur

	return p.conditionalExpr()
}

func (p *Parser) conditionalExpr() (Expr, error) {
	expr, err := p.shortCircuitExpr()
	if err != nil {
		return nil, err
	}

	p.skip()
	if p.match(TOKEN_QUESTION) {
		trueExpr, err := p.assignmentExpr()
		if err != nil {
			return nil, err
		}

		p.skip()
		_, err = p.consume(TOKEN_COLON)
		if err != nil {
			return nil, err
		}

		falseExpr, err := p.assignmentExpr()
		if err != nil {
			return nil, err
		}

		return &TernaryExpr{Cond: expr, True: trueExpr, False: falseExpr}, nil
	}

	return expr, nil
}

func (p *Parser) shortCircuitExpr() (Expr, error) {
	// TODO: CoalesceExpression
	return p.logicalOrExpr()
}

func (p *Parser) logicalOrExpr() (Expr, error) {
	expr, err := p.logicalAndExpr()
	if err != nil {
		return nil, err
	}

	p.skip()
	if p.match(TOKEN_PIPE_PIPE) {
		op := p.previous()
		right, err := p.logicalAndExpr()
		if err != nil {
			return nil, err
		}
		return &OpExpr{Op: op.ToKenType, Left: expr, Right: right}, nil
	}

	return expr, nil
}

func (p *Parser) logicalAndExpr() (Expr, error) {
	expr, err := p.bitwiseOrExpr()
	if err != nil {
		return nil, err
	}

	p.skip()
	if p.match(TOKEN_AND_AND) {
		op := p.previous()
		right, err := p.bitwiseOrExpr()
		if err != nil {
			return nil, err
		}
		return &OpExpr{Op: op.ToKenType, Left: expr, Right: right}, nil
	}

	return expr, nil
}

func (p *Parser) bitwiseOrExpr() (Expr, error) {
	expr, err := p.bitwiseXorExpr()
	if err != nil {
		return nil, err
	}

	p.skip()
	if p.match(TOKEN_PIPE) {
		op := p.previous()
		right, err := p.bitwiseXorExpr()
		if err != nil {
			return nil, err
		}
		return &OpExpr{Op: op.ToKenType, Left: expr, Right: right}, nil
	}

	return expr, nil
}

func (p *Parser) bitwiseXorExpr() (Expr, error) {
	expr, err := p.bitwiseAndExpr()
	if err != nil {
		return nil, err
	}

	p.skip()
	if p.match(TOKEN_HAT) {
		op := p.previous()
		right, err := p.bitwiseAndExpr()
		if err != nil {
			return nil, err
		}
		return &OpExpr{Op: op.ToKenType, Left: expr, Right: right}, nil
	}

	return expr, nil
}

func (p *Parser) bitwiseAndExpr() (Expr, error) {
	expr, err := p.equalityExpr()
	if err != nil {
		return nil, err
	}

	p.skip()
	if p.match(TOKEN_AND) {
		op := p.previous()
		right, err := p.equalityExpr()
		if err != nil {
			return nil, err
		}
		return &OpExpr{Op: op.ToKenType, Left: expr, Right: right}, nil
	}

	return expr, nil
}

func (p *Parser) equalityExpr() (Expr, error) {
	expr, err := p.relationalExpr()
	if err != nil {
		return nil, err
	}

	p.skip()
	if p.match(
		TOKEN_EQUAL_EQUAL,
		TOKEN_BANG_EQUAL,
		TOKEN_EQUAL_EQUAL_EQUAL,
		TOKEN_BANG_EQUAL_EQUAL,
	) {
		op := p.previous()
		right, err := p.relationalExpr()
		if err != nil {
			return nil, err
		}
		return &OpExpr{Op: op.ToKenType, Left: expr, Right: right}, nil
	}

	return expr, nil
}

func (p *Parser) relationalExpr() (Expr, error) {
	expr, err := p.shiftExpr()
	if err != nil {
		return nil, err
	}

	p.skip()
	if p.match(
		TOKEN_LESS,
		TOKEN_GREATER,
		TOKEN_LESS_EQUAL,
		TOKEN_GREATER_EQUAL,
		TOKEN_INSTANCEOF,
		TOKEN_IN,
	) {
		op := p.previous()
		right, err := p.shiftExpr()
		if err != nil {
			return nil, err
		}
		return &OpExpr{Op: op.ToKenType, Left: expr, Right: right}, nil
	}

	return expr, nil
}

func (p *Parser) shiftExpr() (Expr, error) {
	expr, err := p.additiveExpr()
	if err != nil {
		return nil, err
	}

	p.skip()
	if p.match(TOKEN_LESS_LESS, TOKEN_GREATER_GREATER, TOKEN_GREATER_GREATER_GREATER) {
		op := p.previous()
		right, err := p.additiveExpr()
		if err != nil {
			return nil, err
		}
		return &OpExpr{Op: op.ToKenType, Left: expr, Right: right}, nil
	}

	return expr, nil
}

func (p *Parser) additiveExpr() (Expr, error) {
	expr, err := p.multiplicativeExpr()
	if err != nil {
		return nil, err
	}

	p.skip()
	if p.match(TOKEN_PLUS, TOKEN_MINUS) {
		op := p.previous()
		right, err := p.multiplicativeExpr()
		if err != nil {
			return nil, err
		}
		return &OpExpr{Op: op.ToKenType, Left: expr, Right: right}, nil
	}

	return expr, nil
}

func (p *Parser) multiplicativeExpr() (Expr, error) {
	expr, err := p.exponentiationExpr()
	if err != nil {
		return nil, err
	}

	p.skip()
	if p.match(TOKEN_STAR, TOKEN_SLASH, TOKEN_PERCENT) {
		op := p.previous()
		right, err := p.exponentiationExpr()
		if err != nil {
			return nil, err
		}
		return &OpExpr{Op: op.ToKenType, Left: expr, Right: right}, nil
	}

	return expr, nil
}

func (p *Parser) exponentiationExpr() (Expr, error) {
	expr, err := p.updateExpr()
	if err != nil {
		return nil, err
	}

	p.skip()
	if p.match(TOKEN_STAR_STAR) {
		op := p.previous()
		right, err := p.unaryExpr()
		if err != nil {
			return nil, err
		}
		return &OpExpr{Op: op.ToKenType, Left: expr, Right: right}, nil
	}

	return expr, nil
}

func (p *Parser) unaryExpr() (Expr, error) {
	p.skip()
	if p.match(TOKEN_DELETE,
		TOKEN_VOID,
		TOKEN_TYPEOF,
		TOKEN_PLUS,
		TOKEN_MINUS,
		TOKEN_TILDE,
		TOKEN_BANG,
	) {
		op := p.previous()
		expr, err := p.unaryExpr()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: op.ToKenType, Value: expr}, nil
	}

	return p.updateExpr()
}

func (p *Parser) updateExpr() (Expr, error) {
	p.skip()
	if p.match(TOKEN_PLUS_PLUS, TOKEN_MINUS) {
		op := p.previous()
		expr, err := p.unaryExpr()
		if err != nil {
			return nil, err
		}
		return &OpExpr{Op: op.ToKenType, Right: expr}, nil
	}

	expr, err := p.leftHandSideExpr()
	if err != nil {
		return nil, err
	}
	p.skip(TOKEN_SPACE, TOKEN_MULTI_LINE_COMMENT)
	if p.match(TOKEN_PLUS_PLUS, TOKEN_MINUS) {
		op := p.previous()
		return &OpExpr{Op: op.ToKenType, Left: expr}, nil
	}
	return expr, nil
}

func (p *Parser) leftHandSideExpr() (Expr, error) {
	// TODO: call and optional expression
	return p.newExpr()
}

func (p *Parser) newExpr() (Expr, error) {
	p.skip()
	if p.match(TOKEN_NEW) {
		op := p.previous()
		expr, err := p.newExpr()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: op.ToKenType, Value: expr}, nil
	}

	return p.memberExpr()
}

func (p *Parser) memberExpr() (Expr, error) {
	expr, err := p.primaryExpr()
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func (p *Parser) primaryExpr() (Expr, error) {
	var expr Expr

	p.skip()

	tok := p.peek()

	switch tok.ToKenType {
	case TOKEN_IDENTIFIER:
		expr = &Identifier{Value: tok.Literal}
	case TOKEN_NULL:
		expr = &Literal{Value: tok.Literal, Kind: LitNull}
	case TOKEN_TRUE, TOKEN_FALSE:
		expr = &Literal{Value: tok.Literal, Kind: LitBoolean}
	case TOKEN_NUMBER:
		expr = &Literal{Value: tok.Literal, Kind: LitNumber}
	case TOKEN_STRING:
		expr = &Literal{Value: tok.Literal, Kind: LitString}
	default:
		return nil, nil
	}

	p.advance()

	return expr, nil
}

func (p *Parser) consume(tok TokenType) (*Token, error) {
	if p.isEnd() {
		return nil, errors.New("unexpected termination")
	}
	if p.peek().ToKenType == tok {
		return p.advance(), nil
	}
	return nil, errors.New("unexpected token")
}

func (p *Parser) skip(skipTypes ...TokenType) {
	if len(skipTypes) == 0 {
		skipTypes = []TokenType{
			TOKEN_SPACE,
			TOKEN_NEW_LINE,
			TOKEN_SINGLE_LINE_COMMENT,
			TOKEN_MULTI_LINE_COMMENT,
		}
	}
	for p.match(skipTypes...) {
	}
}

func (p *Parser) match(tokTypes ...TokenType) bool {
	tok := p.peek()
	for _, tokType := range tokTypes {
		if tok.ToKenType == tokType {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) advance() *Token {
	if !p.isEnd() {
		p.cur++
	}
	return p.previous()
}

func (p *Parser) isEnd() bool {
	return p.cur >= len(p.tokens) || p.peek().ToKenType == TOKEN_EOF
}

func (p *Parser) peek() *Token {
	return p.tokens[p.cur]
}

func (p *Parser) previous() *Token {
	return p.tokens[p.cur-1]
}
