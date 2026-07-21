package value

import (
	"bytes"
	"strconv"
)

type DataType int

const (
	DataType_Undefined DataType = iota
	DataType_Null
	DataType_Boolean
	DataType_String
	DataType_Symbol
	DataType_Number
	DataType_Object
)

var dataTypeString = "undefinednullboolstringsymbolnumberobject"

var dataTypeIndex = [...]int{0, 9, 13, 17, 23, 29, 35, 41}

func (dt DataType) String() string {
	return dataTypeString[dataTypeIndex[dt]:dataTypeIndex[dt+1]]
}

type Value interface {
	Type() DataType
	Inspect() string
}

type Undefined struct{}

func (u *Undefined) Type() DataType {
	return DataType_Undefined
}

func (u *Undefined) Inspect() string {
	return "undefined"
}

type Null struct{}

func (n *Null) Type() DataType {
	return DataType_Null
}

func (n *Null) Inspect() string {
	return "null"
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() DataType {
	return DataType_Boolean
}

func (b *Boolean) Inspect() string {
	if b.Value {
		return "true"
	}
	return "false"
}

type String struct {
	Value string
}

func (s *String) Type() DataType {
	return DataType_String
}

func (s *String) Inspect() string {
	return s.Value
}

type Symbol struct {
	Description string
}

func (s *Symbol) Type() DataType {
	return DataType_Symbol
}

func (s *Symbol) Inspect() string {
	if s.Description == "" {
		return "Symbol()"
	}
	return "Symbol(" + s.Description + ")"
}

type Number struct {
	Value float64
}

func (n *Number) Type() DataType {
	return DataType_Number
}

func (n *Number) Inspect() string {
	return strconv.FormatFloat(n.Value, 'f', -1, 64)
}

type Object struct {
	Properties map[string]Value
}

func (o *Object) Type() DataType {
	return DataType_Object
}

func (o *Object) Inspect() string {
	buf := new(bytes.Buffer)
	buf.WriteString("{")
	first := true
	for key, value := range o.Properties {
		if !first {
			buf.WriteString(", ")
		}
		buf.WriteString(key)
		buf.WriteString(": ")
		buf.WriteString(value.Inspect())
		first = false
	}
	buf.WriteString("}")
	return buf.String()
}

type Function struct {
	Parameters []string
	Body       interface{} // *ast.BlockStatement - stored as interface{} to avoid circular import
	Env        interface{} // *runtime.Runtime - stored as interface{} to avoid circular import
}

func (f *Function) Type() DataType {
	return DataType_Object
}

func (f *Function) Inspect() string {
	buf := new(bytes.Buffer)
	buf.WriteString("function(")
	for i, p := range f.Parameters {
		buf.WriteString(p)
		if i < len(f.Parameters)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString(") { ... }")
	return buf.String()
}

type ErrorValue struct {
	Message string
}

func (e *ErrorValue) Type() DataType {
	return DataType_Object
}

func (e *ErrorValue) Inspect() string {
	return "Error: " + e.Message
}
