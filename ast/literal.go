package ast

type LitKind int

const (
	LitNull LitKind = iota
	LitBoolean
	LitNumber
	LitString
)

var litKindString = "nullboolnumberstring"

var litKindIndex = [...]int{0, 4, 8, 14, 20}

func (ty LitKind) String() string {
	return "literal<" + litKindString[litKindIndex[ty]:litKindIndex[ty+1]] + ">"
}
