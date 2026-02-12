package ast

import "bytes"

type Statement interface {
	Node
}

type BlockStatement struct {
	StatementList []Statement
}

func (s *BlockStatement) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("{\n")
	for _, stmt := range s.StatementList {
		buf.WriteString(stmt.String())
		buf.WriteString("\n")
	}
	buf.WriteString("}")
	return buf.String()
}

type VarStatement struct {
	Declarations []Declaration
}

func (s *VarStatement) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("var ")
	for i, decl := range s.Declarations {
		buf.WriteString(decl.String())
		if i < len(s.Declarations)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString(";")
	return buf.String()
}

type EmptyStatement struct{}

func (s *EmptyStatement) String() string {
	return ";"
}

type ExpressionStatement struct {
	Expression Expression
}

func (s *ExpressionStatement) String() string {
	return s.Expression.String() + ";"
}

type IfStatement struct {
	Condition   Expression
	TrueBranch  Statement
	FalseBranch Statement
}

func (s *IfStatement) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("if ")
	buf.WriteString(s.Condition.String())
	buf.WriteString(" ")
	buf.WriteString(s.TrueBranch.String())
	if s.FalseBranch != nil {
		buf.WriteString(" else ")
		buf.WriteString(s.FalseBranch.String())
	}

	return buf.String()
}

type ForStatement struct {
	Init      Statement
	Condition Expression
	Update    Expression
	Body      Statement
}

func (s *ForStatement) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("for (")
	if s.Init != nil {
		buf.WriteString(s.Init.String())
	} else {
		buf.WriteString(";")
	}
	if s.Condition != nil {
		buf.WriteString(" " + s.Condition.String() + ";")
	} else {
		buf.WriteString(";")
	}
	if s.Update != nil {
		buf.WriteString(" " + s.Update.String())
	}
	buf.WriteString(") ")
	buf.WriteString(s.Body.String())
	return buf.String()
}

type WhileStatement struct {
	Condition Expression
	Body      Statement
}

func (s *WhileStatement) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("while (")
	buf.WriteString(s.Condition.String())
	buf.WriteString(") ")
	buf.WriteString(s.Body.String())
	return buf.String()
}

type DoWhileStatement struct {
	Body      Statement
	Condition Expression
}

func (s *DoWhileStatement) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("do ")
	buf.WriteString(s.Body.String())
	buf.WriteString(" while (")
	buf.WriteString(s.Condition.String())
	buf.WriteString(");")
	return buf.String()
}

type ContinueStatement struct {
	Label Expression
}

func (s *ContinueStatement) String() string {
	if s.Label != nil {
		return "continue " + s.Label.String() + ";"
	}
	return "continue;"
}

type BreakStatement struct {
	Label Expression
}

func (s *BreakStatement) String() string {
	if s.Label != nil {
		return "break " + s.Label.String() + ";"
	}
	return "break;"
}

type ReturnStatement struct {
	Result Expression
}

func (s *ReturnStatement) String() string {
	if s.Result != nil {
		return "return " + s.Result.String() + ";"
	}
	return "return;"
}

type SwitchStatement struct {
	Discriminant Expression
	Cases        []SwitchCase
	DefaultCase  *SwitchCase
}

func (s *SwitchStatement) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("switch (")
	buf.WriteString(s.Discriminant.String())
	buf.WriteString(") {\n")
	for _, switchCase := range s.Cases {
		buf.WriteString(switchCase.String() + "\n")
	}
	if s.DefaultCase != nil {
		buf.WriteString("default:\n")
		for _, stmt := range s.DefaultCase.Consequent {
			buf.WriteString(stmt.String() + "\n")
		}
	}
	buf.WriteString("}")
	return buf.String()
}

type SwitchCase struct {
	Test       Expression
	Consequent []Statement
}

func (c *SwitchCase) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("case " + c.Test.String() + ":\n")
	for _, stmt := range c.Consequent {
		buf.WriteString(stmt.String() + "\n")
	}
	return buf.String()
}

type LabeledStatement struct {
	Label     Expression
	Statement Statement
}

func (s *LabeledStatement) String() string {
	return s.Label.String() + ": " + s.Statement.String()
}

type ThrowStatement struct {
	Argument Expression
}

func (s *ThrowStatement) String() string {
	return "throw " + s.Argument.String() + ";"
}

type TryStatement struct {
	Block       *BlockStatement
	CatchClause *CatchClause
	Finally     *BlockStatement
}

func (s *TryStatement) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("try ")
	buf.WriteString(s.Block.String())
	if s.CatchClause != nil {
		buf.WriteString(s.CatchClause.String())
	}
	if s.Finally != nil {
		buf.WriteString(" finally ")
		buf.WriteString(s.Finally.String())
	}
	return buf.String()
}

type CatchClause struct {
	Param *Identifier
	Body  *BlockStatement
}

func (c *CatchClause) String() string {
	if c.Param != nil {
		return "catch (" + c.Param.String() + ") " + c.Body.String()
	}
	return "catch " + c.Body.String()
}

type DebuggerStatement struct {
}

func (s *DebuggerStatement) String() string {
	return "debugger;"
}
