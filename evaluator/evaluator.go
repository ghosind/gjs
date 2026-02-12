package evaluator

import (
	"fmt"
	"strconv"

	"github.com/ghosind/gjs/ast"
	"github.com/ghosind/gjs/runtime"
	"github.com/ghosind/gjs/token"
	"github.com/ghosind/gjs/value"
)

var (
	NULL  = &value.Null{}
	TRUE  = &value.Boolean{Value: true}
	FALSE = &value.Boolean{Value: false}
)

type Evaluator struct {
	env *runtime.Runtime
}

func New(env *runtime.Runtime) *Evaluator {
	return &Evaluator{env: env}
}

func (e *Evaluator) Eval(node ast.Node) value.Value {
	switch node := node.(type) {
	case *ast.Program:
		return e.evalProgram(node)
	// Statement
	case *ast.BlockStatement:
		return e.evalBlockStatement(node)
	case *ast.VarStatement:
		for _, decl := range node.Declarations {
			e.Eval(decl)
		}
	case *ast.ExpressionStatement:
		return e.Eval(node.Expression)
	case *ast.IfStatement:
		return e.evalIfExpression(node)
	case *ast.ReturnStatement:
		return e.Eval(node.Result)

	// Expression
	case *ast.Literal:
		switch node.Kind {
		case ast.LitNumber:
			val, err := strconv.ParseFloat(node.Value, 64)
			if err != nil {
				return newError("could not parse %q as number", node.Value)
			}
			return &value.Number{Value: val}
		case ast.LitString:
			return &value.String{Value: node.Value}
		case ast.LitBoolean:
			if node.Value == "true" {
				return TRUE
			}
			return FALSE
		default:
			return newError("unknown literal kind: %d", node.Kind)
		}
	case *ast.Identifier:
		return e.evalIdentifier(node)
	case *ast.UnaryExpression:
		right := e.Eval(node.Value)
		return evalUnaryExpression(node.Operator, right)
	case *ast.BinaryExpression:
		left := e.Eval(node.Left)
		right := e.Eval(node.Right)
		return evalBinaryExpression(node.Operator, left, right)

	// Declaration
	case *ast.VariableDeclaration:
		val := e.Eval(node.Value)
		e.env.Set(node.Name.Value, val)
	}

	return nil
}

func (e *Evaluator) evalProgram(program *ast.Program) value.Value {
	var res value.Value
	for _, statement := range program.Statements {
		res = e.Eval(statement)
	}
	return res
}

func (e *Evaluator) evalBlockStatement(block *ast.BlockStatement) value.Value {
	var res value.Value
	for _, statement := range block.StatementList {
		res = e.Eval(statement)
	}
	return res
}

func (e *Evaluator) evalIfExpression(ie *ast.IfStatement) value.Value {
	condition := e.Eval(ie.Condition)
	if isTruthy(condition) {
		return e.Eval(ie.TrueBranch)
	} else if ie.FalseBranch != nil {
		return e.Eval(ie.FalseBranch)
	}
	return nil
}

func (e *Evaluator) evalIdentifier(node *ast.Identifier) value.Value {
	if val, ok := e.env.Get(node.Value); ok {
		return val
	}
	return newError("identifier not found: " + node.Value)
}

func evalUnaryExpression(operator *token.Token, right value.Value) value.Value {
	switch operator.TokenType {
	case token.TOKEN_BANG:
		return evalBangOperatorExpression(right)
	case token.TOKEN_MINUS:
		if right.Type() != value.DataType_Number {
			return newError("unknown operator: -%s", right.Type())
		}
		val := right.(*value.Number).Value
		return &value.Number{Value: -val}
	default:
		return newError("unknown operator: %s%s", operator.TokenType, right.Type())
	}
}

func evalBangOperatorExpression(right value.Value) value.Value {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		switch right.Type() {
		case value.DataType_Number:
			if right.(*value.Number).Value == 0 {
				return TRUE
			}
			return FALSE
		case value.DataType_String:
			if right.(*value.String).Value == "" {
				return TRUE
			}
			return FALSE
		default:
			return FALSE
		}
	}
}

func evalBinaryExpression(operator *token.Token, left, right value.Value) value.Value {
	// integer ops
	switch {
	case left.Type() == value.DataType_Number && right.Type() == value.DataType_Number:
		return evalNumberBinaryExpression(operator, left, right)
	case operator.TokenType == token.TOKEN_EQUAL_EQUAL:
		return nativeBoolToBooleanObject(left == right)
	case operator.TokenType == token.TOKEN_BANG_EQUAL:
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator.TokenType, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator.TokenType, right.Type())
	}
}

func evalNumberBinaryExpression(operator *token.Token, left, right value.Value) value.Value {
	lv := left.(*value.Number).Value
	rv := right.(*value.Number).Value

	switch operator.TokenType {
	case token.TOKEN_PLUS:
		return &value.Number{Value: lv + rv}
	case token.TOKEN_MINUS:
		return &value.Number{Value: lv - rv}
	case token.TOKEN_STAR:
		return &value.Number{Value: lv * rv}
	case token.TOKEN_SLASH:
		return &value.Number{Value: lv / rv}
	case token.TOKEN_LESS:
		return nativeBoolToBooleanObject(lv < rv)
	case token.TOKEN_GREATER:
		return nativeBoolToBooleanObject(lv > rv)
	case token.TOKEN_EQUAL_EQUAL:
		return nativeBoolToBooleanObject(lv == rv)
	case token.TOKEN_BANG_EQUAL:
		return nativeBoolToBooleanObject(lv != rv)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator.TokenType, right.Type())
	}
}

func isTruthy(obj value.Value) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func nativeBoolToBooleanObject(input bool) *value.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func newError(format string, a ...interface{}) value.Value {
	// TODO: define an error object type
	return &value.Object{
		Properties: map[string]value.Value{
			"message": &value.String{Value: fmt.Sprintf(format, a...)},
		},
	}
}

func isError(obj value.Value) bool {
	// TODO: define an error object type and check for it here
	if obj != nil {
		return obj.Type() == value.DataType_Object && obj.(*value.Object).Properties["message"] != nil
	}
	return false
}
