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

type CallExpression struct {
	Token     token.Token
	Callee    Expression
	Arguments []Expression
}

func (c *CallExpression) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(c.Callee.String())
	buf.WriteString("(")
	for i, arg := range c.Arguments {
		buf.WriteString(arg.String())
		if i < len(c.Arguments)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString(")")
	return buf.String()
}

type FunctionExpression struct {
	Token      token.Token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (f *FunctionExpression) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("function(")
	for i, p := range f.Parameters {
		buf.WriteString(p.String())
		if i < len(f.Parameters)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString(") ")
	buf.WriteString(f.Body.String())
	return buf.String()
}

type AssignmentExpression struct {
	Token    token.Token
	Operator string
	Left     Expression
	Right    Expression
}

func (a *AssignmentExpression) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(a.Left.String())
	buf.WriteString(" ")
	buf.WriteString(a.Operator)
	buf.WriteString(" ")
	buf.WriteString(a.Right.String())
	return buf.String()
}

type GroupExpression struct {
	Expression Expression
}

func (g *GroupExpression) String() string {
	return "(" + g.Expression.String() + ")"
}

type MemberExpression struct {
	Token    token.Token
	Object   Expression
	Property Expression
	Computed bool // true for bracket notation obj[expr], false for dot notation obj.prop
}

func (m *MemberExpression) String() string {
	if m.Computed {
		return m.Object.String() + "[" + m.Property.String() + "]"
	}
	return m.Object.String() + "." + m.Property.String()
}
