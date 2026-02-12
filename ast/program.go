package ast

import "bytes"

type Program struct {
	Statements []Statement
}

func (p *Program) String() string {
	buf := new(bytes.Buffer)
	for _, statement := range p.Statements {
		buf.WriteString(statement.String())
	}
	return buf.String()
}
