package parser

import (
	"errors"
	"fmt"

	"github.com/ghosind/gjs/ast"
	"github.com/ghosind/gjs/lexer"
	"github.com/ghosind/gjs/token"
)

type Parser struct {
	l   *lexer.Lexer
	err error

	prevToken *token.Token
	curToken  *token.Token
	peekToken *token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := new(Parser)
	p.l = l

	return p
}

func (p *Parser) ParseProgram() (*ast.Program, error) {
	for p.current() == nil {
		err := p.nextToken()
		if err != nil {
			return nil, err
		}
	}

	program := new(ast.Program)
	program.Statements = make([]ast.Statement, 0)

	for p.current().TokenType != token.TOKEN_EOF {
		stmt, err := p.statement()
		if err != nil {
			return nil, err
		} else if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		err = p.nextToken()
		if err != nil {
			return nil, err
		}
	}

	return program, nil
}

func (p *Parser) nextToken() error {
	p.prevToken = p.curToken
	p.curToken = p.peekToken
	tok, err := p.l.ScanToken()
	if err != nil {
		return err
	}
	p.peekToken = tok
	return nil
}

func (p *Parser) statement() (ast.Statement, error) {
	p.skip()
	tok := p.current()

	switch tok.TokenType {
	case token.TOKEN_BREAK:
		return p.breakStmt()
	case token.TOKEN_CONTINUE:
		return p.continueStmt(), nil
	case token.TOKEN_DEBUGGER:
		if _, err := p.consume(token.TOKEN_SEMICOLON); err != nil {
			return nil, err
		}
		return new(ast.DebuggerStatement), nil
	case token.TOKEN_DO:
		return p.doWhileStmt()
	case token.TOKEN_FOR:
		return p.forStmt()
	case token.TOKEN_IF:
		return p.ifStat()
	case token.TOKEN_LEFT_BRACE:
		return p.blockStmt()
	case token.TOKEN_RETURN:
		return p.returnStmt()
	case token.TOKEN_SEMICOLON:
		if _, err := p.consume(token.TOKEN_SEMICOLON); err != nil {
			return nil, err
		}
		return new(ast.EmptyStatement), nil
	case token.TOKEN_SWITCH:
		return p.switchStmt()
	case token.TOKEN_THROW:
		return p.throwStmt()
	case token.TOKEN_TRY:
		return p.tryStmt()
	case token.TOKEN_VAR:
		return p.variableStatement()
	case token.TOKEN_WHILE:
		return p.whileStmt()
	default:
		return p.exprStmt()
	}
}

func (p *Parser) declaration() (ast.Statement, error) {
	// TODO
	return nil, nil
}

func (p *Parser) tryStmt() (ast.Statement, error) {
	// TODO
	return nil, nil
}

func (p *Parser) throwStmt() (ast.Statement, error) {
	if _, err := p.consume(token.TOKEN_THROW); err != nil {
		return nil, err
	}

	p.skip(token.TOKEN_SPACE, token.TOKEN_MULTI_LINE_COMMENT)

	result, err := p.expression()
	if err != nil {
		return nil, err
	}

	p.skipAndConsume(token.TOKEN_SEMICOLON)

	return &ast.ThrowStatement{
		Argument: result,
	}, nil
}

func (p *Parser) switchStmt() (ast.Statement, error) {
	// TODO

	// p.consume(TOKEN_SWITCH)

	// if _, err := p.skipAndConsume(TOKEN_LEFT_PAREN); err != nil {
	// 	return nil, err
	// }
	// tag, err := p.expression()
	// if err != nil {
	// 	return nil, err
	// }
	// if _, err := p.skipAndConsume(TOKEN_RIGHT_PAREN); err != nil {
	// 	return nil, err
	// }
	// if _, err := p.skipAndConsume(TOKEN_LEFT_BRACE); err != nil {
	// 	return nil, err
	// }

	// body := make([]*CaseClause, 0)
	// for !p.skipAndMatch(TOKEN_RIGHT_BRACE) {
	// 	// TODO
	// }

	// return &SwitchStmt{
	// 	Tag:  tag,
	// 	Body: body,
	// }, nil

	return nil, nil
}

func (p *Parser) returnStmt() (ast.Statement, error) {
	p.consume(token.TOKEN_RETURN)

	p.skip(token.TOKEN_SPACE, token.TOKEN_MULTI_LINE_COMMENT)

	// TODO
	cur := p.curToken
	result, err := p.expression()
	if err != nil {
		return nil, err
	} else if result == nil {
		p.curToken = cur
	}

	p.skipAndConsume(token.TOKEN_SEMICOLON)

	return &ast.ReturnStatement{
		Result: result,
	}, nil
}

func (p *Parser) breakStmt() (ast.Statement, error) {
	var label ast.Expression

	if _, err := p.consume(token.TOKEN_BREAK); err != nil {
		return nil, err
	}

	p.skip(token.TOKEN_SPACE, token.TOKEN_MULTI_LINE_COMMENT)
	if p.match(token.TOKEN_IDENTIFIER) {
		tok := p.previous()
		label = &ast.Identifier{Value: tok.Literal}
	}

	p.skipAndConsume(token.TOKEN_SEMICOLON)

	if p.isSyntaxError() {
		return nil, p.err
	}

	return &ast.BreakStatement{
		Label: label,
	}, nil
}

func (p *Parser) continueStmt() ast.Statement {
	var label ast.Expression

	p.consume(token.TOKEN_CONTINUE)

	p.skip(token.TOKEN_SPACE, token.TOKEN_MULTI_LINE_COMMENT)
	if p.match(token.TOKEN_IDENTIFIER) {
		tok := p.previous()
		label = &ast.Identifier{Value: tok.Literal}
	}

	p.skipAndConsume(token.TOKEN_SEMICOLON)

	return &ast.ContinueStatement{
		Label: label,
	}
}

func (p *Parser) forStmt() (ast.Statement, error) {
	p.consume(token.TOKEN_FOR)

	if _, err := p.skipAndConsume(token.TOKEN_LEFT_PAREN); err != nil {
		return nil, err
	}

	var init ast.Expression
	var post ast.Expression
	var err error

	if !p.skipAndMatch(token.TOKEN_SEMICOLON) {
		init, err = p.expression()
		if err != nil {
			return nil, err
		}

		if _, err := p.skipAndConsume(token.TOKEN_SEMICOLON); err != nil {
			return nil, err
		}
	}

	cond, err := p.expression()
	if err != nil {
		return nil, err
	}

	if _, err := p.skipAndConsume(token.TOKEN_SEMICOLON); err != nil {
		return nil, err
	}

	if !p.skipAndMatch(token.TOKEN_RIGHT_PAREN) {
		post, err = p.expression()
		if err != nil {
			return nil, err
		}

		if _, err := p.skipAndConsume(token.TOKEN_RIGHT_PAREN); err != nil {
			return nil, err
		}
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	return &ast.ForStatement{
		Init:      init,
		Condition: cond,
		Update:    post,
		Body:      body,
	}, nil
}

func (p *Parser) whileStmt() (ast.Statement, error) {
	p.consume(token.TOKEN_WHILE)

	if _, err := p.skipAndConsume(token.TOKEN_LEFT_PAREN); err != nil {
		return nil, err
	}

	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	if _, err = p.skipAndConsume(token.TOKEN_RIGHT_PAREN); err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	return &ast.WhileStatement{
		Condition: expr,
		Body:      body,
	}, nil
}

func (p *Parser) doWhileStmt() (ast.Statement, error) {
	p.consume(token.TOKEN_DO)

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	if _, err := p.skipAndConsume(token.TOKEN_WHILE); err != nil {
		return nil, err
	}
	if _, err = p.skipAndConsume(token.TOKEN_LEFT_PAREN); err != nil {
		return nil, err
	}

	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	if _, err = p.skipAndConsume(token.TOKEN_RIGHT_PAREN); err != nil {
		return nil, err
	}

	p.skipAndConsume(token.TOKEN_SEMICOLON)

	return &ast.DoWhileStatement{
		Body:      body,
		Condition: expr,
	}, nil
}

func (p *Parser) ifStat() (ast.Statement, error) {
	p.consume(token.TOKEN_IF)

	if _, err := p.skipAndConsume(token.TOKEN_LEFT_PAREN); err != nil {
		return nil, err
	}
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	if _, err := p.skipAndConsume(token.TOKEN_RIGHT_PAREN); err != nil {
		return nil, err
	}

	thenStmt, err := p.statement()
	if err != nil {
		return nil, err
	}

	var elseStmt ast.Statement
	if p.skipAndMatch(token.TOKEN_ELSE) {
		elseStmt, err = p.statement()
		if err != nil {
			return nil, err
		}
	}

	return &ast.IfStatement{
		Condition:   expr,
		TrueBranch:  thenStmt,
		FalseBranch: elseStmt,
	}, nil
}

func (p *Parser) initializer() (ast.Expression, error) {
	if p.skipAndMatch(token.TOKEN_EQUAL) {
		return p.assignmentExpr()
	}

	return nil, nil
}

func (p *Parser) variableDeclaration() (ast.Declaration, error) {
	name := p.previous()
	decl := &ast.VariableDeclaration{
		Name: &ast.Identifier{
			Value: name.Literal,
		},
	}
	initializer, err := p.initializer()
	if err != nil {
		return nil, err
	}
	decl.Value = initializer
	return decl, nil
}

func (p *Parser) variableStatement() (ast.Statement, error) {
	p.consume(token.TOKEN_VAR)
	decls := make([]ast.Declaration, 0)

	for p.skipAndMatch(token.TOKEN_IDENTIFIER) {
		decl, err := p.variableDeclaration()
		if err != nil {
			return nil, err
		}

		decls = append(decls, decl)
		if !p.skipAndMatch(token.TOKEN_COMMA) {
			break
		}
	}

	if len(decls) == 0 {
		return nil, p.newSyntaxError(p.curToken)
	}

	p.skipAndConsume(token.TOKEN_SEMICOLON)

	return &ast.VarStatement{
		Declarations: decls,
	}, nil
}

func (p *Parser) blockStmt() (ast.Statement, error) {
	var stmt ast.Statement
	var err error

	p.consume(token.TOKEN_LEFT_BRACE)
	list := make([]ast.Statement, 0)

	for !p.skipAndMatch(token.TOKEN_RIGHT_BRACE) {
		stmt, err = p.statement()
		if err != nil {
			return nil, err
		}
		list = append(list, stmt)
	}

	return &ast.BlockStatement{StatementList: list}, nil
}

func (p *Parser) exprStmt() (ast.Statement, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	p.skipAndConsume(token.TOKEN_SEMICOLON)

	return &ast.ExpressionStatement{Expression: expr}, nil
}

func (p *Parser) expression() (ast.Expression, error) {
	// TODO: comma operation
	return p.assignmentExpr()
}

func (p *Parser) assignmentExpr() (ast.Expression, error) {
	expr, err := p.conditionalExpr()
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func (p *Parser) conditionalExpr() (ast.Expression, error) {
	expr, err := p.shortCircuitExpr()
	if err != nil {
		return nil, err
	}

	if p.skipAndMatch(token.TOKEN_QUESTION) {
		trueExpr, err := p.assignmentExpr()
		if err != nil {
			return nil, err
		}

		_, err = p.skipAndConsume(token.TOKEN_COLON)
		if err != nil {
			return nil, err
		}

		falseExpr, err := p.assignmentExpr()
		if err != nil {
			return nil, err
		}

		return &ast.TernaryExpression{
			Condition:   expr,
			TrueBranch:  trueExpr,
			FalseBranch: falseExpr,
		}, nil
	}

	return expr, nil
}

func (p *Parser) shortCircuitExpr() (ast.Expression, error) {
	// TODO: CoalesceExpression
	return p.logicalOrExpr()
}

func (p *Parser) logicalOrExpr() (ast.Expression, error) {
	expr, err := p.logicalAndExpr()
	if err != nil {
		return nil, err
	}

	if p.skipAndMatch(token.TOKEN_PIPE_PIPE) {
		op := p.previous()
		right, err := p.logicalAndExpr()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     expr,
			Right:    right,
		}, nil
	}

	return expr, nil
}

func (p *Parser) logicalAndExpr() (ast.Expression, error) {
	expr, err := p.bitwiseOrExpr()
	if err != nil {
		return nil, err
	}

	if p.skipAndMatch(token.TOKEN_AND_AND) {
		op := p.previous()
		right, err := p.bitwiseOrExpr()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     expr,
			Right:    right,
		}, nil
	}

	return expr, nil
}

func (p *Parser) bitwiseOrExpr() (ast.Expression, error) {
	expr, err := p.bitwiseXorExpr()
	if err != nil {
		return nil, err
	}

	if p.skipAndMatch(token.TOKEN_PIPE) {
		op := p.previous()
		right, err := p.bitwiseXorExpr()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     expr,
			Right:    right,
		}, nil
	}

	return expr, nil
}

func (p *Parser) bitwiseXorExpr() (ast.Expression, error) {
	expr, err := p.bitwiseAndExpr()
	if err != nil {
		return nil, err
	}

	if p.skipAndMatch(token.TOKEN_HAT) {
		op := p.previous()
		right, err := p.bitwiseAndExpr()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     expr,
			Right:    right,
		}, nil
	}

	return expr, nil
}

func (p *Parser) bitwiseAndExpr() (ast.Expression, error) {
	expr, err := p.equalityExpr()
	if err != nil {
		return nil, err
	}

	if p.skipAndMatch(token.TOKEN_AND) {
		op := p.previous()
		right, err := p.equalityExpr()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     expr,
			Right:    right,
		}, nil
	}

	return expr, nil
}

func (p *Parser) equalityExpr() (ast.Expression, error) {
	expr, err := p.relationalExpr()
	if err != nil {
		return nil, err
	}

	if p.skipAndMatch(
		token.TOKEN_EQUAL_EQUAL,
		token.TOKEN_BANG_EQUAL,
		token.TOKEN_EQUAL_EQUAL_EQUAL,
		token.TOKEN_BANG_EQUAL_EQUAL,
	) {
		op := p.previous()
		right, err := p.relationalExpr()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     expr,
			Right:    right,
		}, nil
	}

	return expr, nil
}

func (p *Parser) relationalExpr() (ast.Expression, error) {
	expr, err := p.shiftExpr()
	if err != nil {
		return nil, err
	}

	if p.skipAndMatch(
		token.TOKEN_LESS,
		token.TOKEN_GREATER,
		token.TOKEN_LESS_EQUAL,
		token.TOKEN_GREATER_EQUAL,
		token.TOKEN_INSTANCEOF,
		token.TOKEN_IN,
	) {
		op := p.previous()
		right, err := p.shiftExpr()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     expr,
			Right:    right,
		}, nil
	}

	return expr, nil
}

func (p *Parser) shiftExpr() (ast.Expression, error) {
	expr, err := p.additiveExpr()
	if err != nil {
		return nil, err
	}

	if p.skipAndMatch(
		token.TOKEN_LESS_LESS,
		token.TOKEN_GREATER_GREATER,
		token.TOKEN_GREATER_GREATER_GREATER,
	) {
		op := p.previous()
		right, err := p.additiveExpr()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     expr,
			Right:    right,
		}, nil
	}

	return expr, nil
}

func (p *Parser) additiveExpr() (ast.Expression, error) {
	expr, err := p.multiplicativeExpr()
	if err != nil {
		return nil, err
	}

	if p.skipAndMatch(token.TOKEN_PLUS, token.TOKEN_MINUS) {
		op := p.previous()
		right, err := p.multiplicativeExpr()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     expr,
			Right:    right,
		}, nil
	}

	return expr, nil
}

func (p *Parser) multiplicativeExpr() (ast.Expression, error) {
	expr, err := p.exponentiationExpr()
	if err != nil {
		return nil, err
	}

	if p.skipAndMatch(token.TOKEN_STAR, token.TOKEN_SLASH, token.TOKEN_PERCENT) {
		op := p.previous()
		right, err := p.exponentiationExpr()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     expr,
			Right:    right,
		}, nil
	}

	return expr, nil
}

func (p *Parser) exponentiationExpr() (ast.Expression, error) {
	expr, err := p.updateExpr()
	if err != nil {
		return nil, err
	}

	if p.skipAndMatch(token.TOKEN_STAR_STAR) {
		op := p.previous()
		right, err := p.unaryExpr()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     expr,
			Right:    right,
		}, nil
	}

	return expr, nil
}

func (p *Parser) unaryExpr() (ast.Expression, error) {
	if p.skipAndMatch(token.TOKEN_DELETE,
		token.TOKEN_VOID,
		token.TOKEN_TYPEOF,
		token.TOKEN_PLUS,
		token.TOKEN_MINUS,
		token.TOKEN_TILDE,
		token.TOKEN_BANG,
	) {
		op := p.previous()
		expr, err := p.unaryExpr()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{
			Operator: op,
			Value:    expr,
		}, nil
	}

	return p.updateExpr()
}

func (p *Parser) updateExpr() (ast.Expression, error) {
	if p.skipAndMatch(token.TOKEN_PLUS_PLUS, token.TOKEN_MINUS) {
		op := p.previous()
		expr, err := p.unaryExpr()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{Operator: op, Value: expr}, nil
	}

	expr, err := p.leftHandSideExpr()
	if err != nil {
		return nil, err
	}
	p.skip(token.TOKEN_SPACE, token.TOKEN_MULTI_LINE_COMMENT)
	if p.match(token.TOKEN_PLUS_PLUS, token.TOKEN_MINUS) {
		op := p.previous()
		return &ast.UnaryExpression{Operator: op, Value: expr}, nil
	}
	return expr, nil
}

func (p *Parser) leftHandSideExpr() (ast.Expression, error) {
	// TODO: call and optional expression
	return p.newExpr()
}

func (p *Parser) newExpr() (ast.Expression, error) {
	if p.skipAndMatch(token.TOKEN_NEW) {
		op := p.previous()
		expr, err := p.newExpr()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{Operator: op, Value: expr}, nil
	}

	return p.memberExpr()
}

func (p *Parser) memberExpr() (ast.Expression, error) {
	expr, err := p.primaryExpr()
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func (p *Parser) arrayLiteral() (ast.Expression, error) {
	list := make([]ast.Expression, 0)

	for !p.skipAndMatch(token.TOKEN_RIGHT_BRACKET) {
		if p.isSyntaxError() {
			return nil, p.err
		}

		tok := p.peek()
		switch tok.TokenType {
		case token.TOKEN_COMMA:
			list = append(list, &ast.Elision{})
			p.advance()
			if p.isSyntaxError() {
				return nil, p.err
			}
		case token.TOKEN_DOT_DOT_DOT:
			p.advance()
			if p.isSyntaxError() {
				return nil, p.err
			}

			expr, err := p.assignmentExpr()
			if err != nil {
				return nil, err
			} else if expr == nil {
				return nil, fmt.Errorf("unexpected token %v", p.peek())
			}
			list = append(list, &ast.SpreadElement{
				Value: expr,
			})
		default:
			expr, err := p.assignmentExpr()
			if err != nil {
				return nil, err
			} else if expr == nil {
				return nil, fmt.Errorf("unexpected token %v", tok)
			}
			list = append(list, expr)
		}

		p.skipAndMatch(token.TOKEN_COMMA)
	}

	return &ast.ArrayLiteral{
		ElementList: list,
	}, nil
}

func (p *Parser) primaryExpr() (expr ast.Expression, err error) {
	p.skip()
	tok := p.current()
	if p.isSyntaxError() {
		return nil, p.err
	}

	switch tok.TokenType {
	case token.TOKEN_IDENTIFIER:
		expr = &ast.Identifier{Value: tok.Literal}
	case token.TOKEN_NULL:
		expr = &ast.Literal{Value: tok.Literal, Kind: ast.LitNull}
	case token.TOKEN_TRUE, token.TOKEN_FALSE:
		expr = &ast.Literal{Value: tok.Literal, Kind: ast.LitBoolean}
	case token.TOKEN_NUMBER:
		expr = &ast.Literal{Value: tok.Literal, Kind: ast.LitNumber}
	case token.TOKEN_STRING:
		expr = &ast.Literal{Value: tok.Literal, Kind: ast.LitString}
	case token.TOKEN_LEFT_BRACKET:
		p.advance()
		if p.isSyntaxError() {
			return nil, p.err
		}
		expr, err = p.arrayLiteral()
		return
	default:
		return nil, nil
	}

	p.advance()
	if p.isSyntaxError() {
		return nil, p.err
	}

	return
}

func (p *Parser) consume(tokType token.TokenType) (*token.Token, error) {
	if p.isEnd() {
		return nil, errors.New("unexpected termination")
	}
	if p.current().TokenType == tokType {
		tok := p.advance()
		return tok, p.err
	}
	return nil, fmt.Errorf("unexpected token %s", tokType)
}

func (p *Parser) skip(skipTypes ...token.TokenType) {
	if len(skipTypes) == 0 {
		skipTypes = []token.TokenType{
			token.TOKEN_SPACE,
			token.TOKEN_NEW_LINE,
			token.TOKEN_SINGLE_LINE_COMMENT,
			token.TOKEN_MULTI_LINE_COMMENT,
		}
	}
	for p.match(skipTypes...) {
	}
}

func (p *Parser) skipAndConsume(tok token.TokenType) (*token.Token, error) {
	p.skip()

	return p.consume(tok)
}

func (p *Parser) match(tokTypes ...token.TokenType) bool {
	tok := p.current()
	for _, tokType := range tokTypes {
		if tok.TokenType == tokType {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) skipAndMatch(tokTypes ...token.TokenType) bool {
	p.skip()

	return p.match(tokTypes...)
}

func (p *Parser) advance() *token.Token {
	if !p.isEnd() {
		err := p.nextToken()
		if err != nil {
			p.err = err
			return nil
		}
	}
	return p.previous()
}

func (p *Parser) isEnd() bool {
	tok := p.current()
	return tok == nil || tok.TokenType == token.TOKEN_EOF
}

func (p *Parser) current() *token.Token {
	return p.curToken
}

func (p *Parser) peek() *token.Token {
	return p.peekToken
}

func (p *Parser) previous() *token.Token {
	return p.prevToken
}

func (p *Parser) isSyntaxError() bool {
	return p.err != nil
}
