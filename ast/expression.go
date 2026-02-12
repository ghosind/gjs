package ast

import (
	"bytes"

	"github.com/ghosind/gjs/token"
)

type Expression interface {
	Node
}

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) String() string {
	return i.Value
}

type Literal struct {
	Token token.Token
	Value string
	Kind  LitKind
}

func (l *Literal) String() string {
	return l.Value
}

type Elision struct{}

func (e *Elision) String() string {
	return ""
}

type SpreadElement struct {
	Value Expression
}

func (s *SpreadElement) String() string {
	return "..." + s.Value.String()
}

type ArrayLiteral struct {
	ElementList []Expression
}

func (a *ArrayLiteral) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("[")
	for i, elem := range a.ElementList {
		buf.WriteString(elem.String())
		if i < len(a.ElementList)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString("]")
	return buf.String()
}

type UnaryExpression struct {
	Token    token.Token
	Operator *token.Token
	Value    Expression
}

func (u *UnaryExpression) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(u.Operator.Literal)
	buf.WriteString(u.Value.String())
	return buf.String()
}

type BinaryExpression struct {
	Token    token.Token
	Operator *token.Token
	Left     Expression
	Right    Expression
}

func (b *BinaryExpression) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(b.Left.String())
	buf.WriteString(" " + b.Operator.Literal + " ")
	buf.WriteString(b.Right.String())
	return buf.String()
}

type TernaryExpression struct {
	Token       token.Token
	Condition   Expression
	TrueBranch  Expression
	FalseBranch Expression
}

func (t *TernaryExpression) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(t.Condition.String())
	buf.WriteString(" ? ")
	buf.WriteString(t.TrueBranch.String())
	buf.WriteString(" : ")
	buf.WriteString(t.FalseBranch.String())
	return buf.String()
}
