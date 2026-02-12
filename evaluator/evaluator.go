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

func Eval(node ast.Node, env *runtime.Runtime) value.Value {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.Result, env)
		return val
	case *ast.VarDeclaration:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
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
	case *ast.UnaryExpression:
		right := Eval(node.Value, env)
		if isError(right) {
			return right
		}
		return evalUnaryExpression(node.Operator, right)
	case *ast.BinaryExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalBinaryExpression(node.Operator, left, right)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.IfStatement:
		return evalIfExpression(node, env)
	}

	return nil
}

func evalProgram(program *ast.Program, env *runtime.Runtime) value.Value {
	var result value.Value
	for _, statement := range program.Statements {
		result = Eval(statement, env)
	}
	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *runtime.Runtime) value.Value {
	var result value.Value
	for _, statement := range block.StatementList {
		result = Eval(statement, env)
	}
	return result
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

func evalIfExpression(ie *ast.IfStatement, env *runtime.Runtime) value.Value {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(ie.TrueBranch, env)
	} else if ie.FalseBranch != nil {
		return Eval(ie.FalseBranch, env)
	} else {
		return NULL
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

func evalIdentifier(node *ast.Identifier, env *runtime.Runtime) value.Value {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	return newError("identifier not found: " + node.Value)
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
