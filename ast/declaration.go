package ast

import "bytes"

type Declaration interface {
	Node
}

type VariableDeclaration struct {
	Declaration
	Name  *Identifier
	Value Expression
}

func (d *VariableDeclaration) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(d.Name.String())
	if d.Value != nil {
		buf.WriteString(" = ")
		buf.WriteString(d.Value.String())
	}
	return buf.String()
}
