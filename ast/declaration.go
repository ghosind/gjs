package ast

import "bytes"

type Declaration interface {
	Node
}

type VarDeclaration struct {
	Declaration
	Name  *Identifier
	Value Expression
}

func (d *VarDeclaration) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(d.Name.String())
	if d.Value != nil {
		buf.WriteString(" = ")
		buf.WriteString(d.Value.String())
	}
	return buf.String()
}
