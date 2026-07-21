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
	NULL      = &value.Null{}
	TRUE      = &value.Boolean{Value: true}
	FALSE     = &value.Boolean{Value: false}
	UNDEFINED = &value.Undefined{}
)

type Evaluator struct {
	env *runtime.Runtime
}

func New(env *runtime.Runtime) *Evaluator {
	e := &Evaluator{env: env}
	e.registerBuiltins()
	return e
}

func (e *Evaluator) registerBuiltins() {
	// console.log - we use a special marker
	consoleObj := &value.Object{
		Properties: map[string]value.Value{
			"log": &value.Function{
				Parameters: []string{"__builtin__", "console.log"},
				Body:       nil, // nil body means built-in
				Env:        nil,
			},
		},
	}
	e.env.Set("console", consoleObj)
}

func (e *Evaluator) Eval(node ast.Node) value.Value {
	if node == nil {
		return UNDEFINED
	}

	switch node := node.(type) {
	case *ast.Program:
		return e.evalProgram(node)

	// Statements
	case *ast.BlockStatement:
		return e.evalBlockStatement(node)
	case *ast.VarStatement:
		return e.evalVarStatement(node)
	case *ast.ExpressionStatement:
		return e.Eval(node.Expression)
	case *ast.IfStatement:
		return e.evalIfStatement(node)
	case *ast.ReturnStatement:
		return e.evalReturnStatement(node)
	case *ast.WhileStatement:
		return e.evalWhileStatement(node)
	case *ast.DoWhileStatement:
		return e.evalDoWhileStatement(node)
	case *ast.ForStatement:
		return e.evalForStatement(node)
	case *ast.FunctionStatement:
		return e.evalFunctionStatement(node)
	case *ast.EmptyStatement:
		return nil
	case *ast.BreakStatement:
		return &ReturnValue{Value: UNDEFINED, isBreak: true}
	case *ast.ContinueStatement:
		return &ReturnValue{Value: UNDEFINED, isContinue: true}
	case *ast.ThrowStatement:
		val := e.Eval(node.Argument)
		return &ReturnValue{Value: val, isThrow: true}

	// Declarations
	case *ast.VariableDeclaration:
		var val value.Value
		if node.Value != nil {
			val = e.Eval(node.Value)
		} else {
			val = UNDEFINED
		}
		e.env.Set(node.Name.Value, val)
		return val

	// Expressions
	case *ast.Literal:
		return e.evalLiteral(node)
	case *ast.Identifier:
		return e.evalIdentifier(node)
	case *ast.UnaryExpression:
		return e.evalUnaryExpression(node)
	case *ast.BinaryExpression:
		return e.evalBinaryExpression(node)
	case *ast.TernaryExpression:
		return e.evalTernaryExpression(node)
	case *ast.CallExpression:
		return e.evalCallExpression(node)
	case *ast.FunctionExpression:
		return e.evalFunctionExpression(node)
	case *ast.AssignmentExpression:
		return e.evalAssignmentExpression(node)
	case *ast.GroupExpression:
		return e.Eval(node.Expression)
	case *ast.ArrayLiteral:
		return e.evalArrayLiteral(node)
	case *ast.MemberExpression:
		return e.evalMemberExpression(node)
	}

	return nil
}

func (e *Evaluator) evalProgram(program *ast.Program) value.Value {
	var res value.Value
	for _, statement := range program.Statements {
		res = e.Eval(statement)
		if rv, ok := res.(*ReturnValue); ok {
			return rv.Value
		}
	}
	return res
}

func (e *Evaluator) evalBlockStatement(block *ast.BlockStatement) value.Value {
	var res value.Value
	for _, statement := range block.StatementList {
		res = e.Eval(statement)
		if res != nil {
			if rv, ok := res.(*ReturnValue); ok {
				return rv
			}
		}
	}
	return res
}

func (e *Evaluator) evalVarStatement(vs *ast.VarStatement) value.Value {
	for _, decl := range vs.Declarations {
		e.Eval(decl)
	}
	return nil
}

func (e *Evaluator) evalIfStatement(ie *ast.IfStatement) value.Value {
	condition := e.Eval(ie.Condition)
	if isTruthy(condition) {
		return e.Eval(ie.TrueBranch)
	} else if ie.FalseBranch != nil {
		return e.Eval(ie.FalseBranch)
	}
	return nil
}

func (e *Evaluator) evalReturnStatement(rs *ast.ReturnStatement) value.Value {
	var val value.Value
	if rs.Result != nil {
		val = e.Eval(rs.Result)
	} else {
		val = UNDEFINED
	}
	return &ReturnValue{Value: val}
}

func (e *Evaluator) evalWhileStatement(ws *ast.WhileStatement) value.Value {
	for {
		condition := e.Eval(ws.Condition)
		if !isTruthy(condition) {
			break
		}
		result := e.Eval(ws.Body)
		if result != nil {
			switch rv := result.(type) {
			case *ReturnValue:
				if rv.isBreak {
					break
				}
				if rv.isContinue {
					continue
				}
				return result
			}
		}
	}
	return nil
}

func (e *Evaluator) evalDoWhileStatement(dw *ast.DoWhileStatement) value.Value {
	for {
		result := e.Eval(dw.Body)
		if result != nil {
			switch rv := result.(type) {
			case *ReturnValue:
				if rv.isBreak {
					break
				}
				if rv.isContinue {
					continue
				}
				return result
			}
		}
		condition := e.Eval(dw.Condition)
		if !isTruthy(condition) {
			break
		}
	}
	return nil
}

func (e *Evaluator) evalForStatement(fs *ast.ForStatement) value.Value {
	if fs.Init != nil {
		e.Eval(fs.Init)
	}

	for {
		if fs.Condition != nil {
			condition := e.Eval(fs.Condition)
			if !isTruthy(condition) {
				break
			}
		}

		result := e.Eval(fs.Body)
		if result != nil {
			switch rv := result.(type) {
			case *ReturnValue:
				if rv.isBreak {
					break
				}
				if rv.isContinue {
					// fall through to update
				} else {
					return result
				}
			}
		}

		if fs.Update != nil {
			e.Eval(fs.Update)
		}
	}
	return nil
}

func (e *Evaluator) evalFunctionStatement(fs *ast.FunctionStatement) value.Value {
	params := make([]string, len(fs.Parameters))
	for i, p := range fs.Parameters {
		params[i] = p.Value
	}
	fn := &value.Function{
		Parameters: params,
		Body:       fs.Body,
		Env:        e.env,
	}
	e.env.Set(fs.Name.Value, fn)
	return fn
}

func (e *Evaluator) evalLiteral(lit *ast.Literal) value.Value {
	switch lit.Kind {
	case ast.LitNumber:
		val, err := strconv.ParseFloat(lit.Value, 64)
		if err != nil {
			return newError("could not parse %q as number", lit.Value)
		}
		return &value.Number{Value: val}
	case ast.LitString:
		return &value.String{Value: lit.Value}
	case ast.LitBoolean:
		if lit.Value == "true" {
			return TRUE
		}
		return FALSE
	case ast.LitNull:
		return NULL
	default:
		return newError("unknown literal kind: %d", lit.Kind)
	}
}

func (e *Evaluator) evalIdentifier(node *ast.Identifier) value.Value {
	if val, ok := e.env.Get(node.Value); ok {
		return val
	}
	return newError("identifier not found: " + node.Value)
}

func (e *Evaluator) evalUnaryExpression(ue *ast.UnaryExpression) value.Value {
	right := e.Eval(ue.Value)

	// prefix ++ and --
	switch ue.Operator.TokenType {
	case token.TOKEN_PLUS_PLUS:
		if right.Type() == value.DataType_Number {
			val := right.(*value.Number).Value + 1
			// try to update the variable
			e.updateVariable(ue.Value, &value.Number{Value: val})
			return &value.Number{Value: val}
		}
		return newError("unknown operator: ++%s", right.Type())
	case token.TOKEN_MINUS_MINUS:
		if right.Type() == value.DataType_Number {
			val := right.(*value.Number).Value - 1
			e.updateVariable(ue.Value, &value.Number{Value: val})
			return &value.Number{Value: val}
		}
		return newError("unknown operator: --%s", right.Type())
	}

	return evalUnaryOperator(ue.Operator, right)
}

func (e *Evaluator) updateVariable(expr ast.Expression, val value.Value) {
	if ident, ok := expr.(*ast.Identifier); ok {
		e.env.Set(ident.Value, val)
	}
}

func (e *Evaluator) evalBinaryExpression(be *ast.BinaryExpression) value.Value {
	left := e.Eval(be.Left)

	// short-circuit for && and ||
	if be.Operator.TokenType == token.TOKEN_AND_AND {
		if !isTruthy(left) {
			return left
		}
		return e.Eval(be.Right)
	}
	if be.Operator.TokenType == token.TOKEN_PIPE_PIPE {
		if isTruthy(left) {
			return left
		}
		return e.Eval(be.Right)
	}

	right := e.Eval(be.Right)

	// string concatenation with +
	if be.Operator.TokenType == token.TOKEN_PLUS {
		if left.Type() == value.DataType_String || right.Type() == value.DataType_String {
			return &value.String{Value: left.Inspect() + right.Inspect()}
		}
	}

	return evalBinaryOperator(be.Operator, left, right)
}

func (e *Evaluator) evalTernaryExpression(te *ast.TernaryExpression) value.Value {
	condition := e.Eval(te.Condition)
	if isTruthy(condition) {
		return e.Eval(te.TrueBranch)
	}
	return e.Eval(te.FalseBranch)
}

func (e *Evaluator) evalCallExpression(ce *ast.CallExpression) value.Value {
	callee := e.Eval(ce.Callee)

	args := make([]value.Value, len(ce.Arguments))
	for i, arg := range ce.Arguments {
		args[i] = e.Eval(arg)
	}

	if callee.Type() == value.DataType_Object {
		if fn, ok := callee.(*value.Function); ok {
			if fn.Body == nil {
				// Built-in function
				return e.applyBuiltin(fn, args)
			}
			return e.applyFunction(fn, args)
		}
	}

	return newError("not a function: %s", callee.Inspect())
}

func (e *Evaluator) evalFunctionExpression(fe *ast.FunctionExpression) value.Value {
	params := make([]string, len(fe.Parameters))
	for i, p := range fe.Parameters {
		params[i] = p.Value
	}
	return &value.Function{
		Parameters: params,
		Body:       fe.Body,
		Env:        e.env,
	}
}

func (e *Evaluator) evalAssignmentExpression(ae *ast.AssignmentExpression) value.Value {
	right := e.Eval(ae.Right)

	switch ident := ae.Left.(type) {
	case *ast.Identifier:
		if ae.Operator == "=" {
			e.env.Set(ident.Value, right)
		} else {
			// compound assignment: +=, -=, *=, /=
			leftVal, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier not found: " + ident.Value)
			}
			newVal := evalCompoundAssignment(ae.Operator, leftVal, right)
			e.env.Set(ident.Value, newVal)
			return newVal
		}
	default:
		return newError("invalid assignment target")
	}
	return right
}

func (e *Evaluator) evalArrayLiteral(al *ast.ArrayLiteral) value.Value {
	elements := make([]value.Value, len(al.ElementList))
	for i, elem := range al.ElementList {
		if elision, ok := elem.(*ast.Elision); ok {
			_ = elision
			elements[i] = UNDEFINED
		} else {
			elements[i] = e.Eval(elem)
		}
	}
	// Store as object with numeric keys for simplicity
	obj := &value.Object{Properties: make(map[string]value.Value)}
	for i, elem := range elements {
		obj.Properties[strconv.Itoa(i)] = elem
	}
	obj.Properties["length"] = &value.Number{Value: float64(len(elements))}
	return obj
}

func (e *Evaluator) applyFunction(fn *value.Function, args []value.Value) value.Value {
	// Create new environment for function scope
	outerEnv := fn.Env.(*runtime.Runtime)
	funcEnv := runtime.NewEnclosedEnvironment(outerEnv)

	// Bind parameters
	for i, param := range fn.Parameters {
		if i < len(args) {
			funcEnv.Set(param, args[i])
		} else {
			funcEnv.Set(param, UNDEFINED)
		}
	}

	// Save current env, set function env
	oldEnv := e.env
	e.env = funcEnv
	defer func() { e.env = oldEnv }()

	// Execute function body
	body := fn.Body.(*ast.BlockStatement)
	result := e.evalBlockStatement(body)

	if rv, ok := result.(*ReturnValue); ok {
		return rv.Value
	}
	return UNDEFINED
}

func (e *Evaluator) applyBuiltin(fn *value.Function, args []value.Value) value.Value {
	// Check built-in name stored in Parameters[1]
	if len(fn.Parameters) >= 2 && fn.Parameters[0] == "__builtin__" {
		switch fn.Parameters[1] {
		case "console.log":
			for _, arg := range args {
				fmt.Print(arg.Inspect())
				fmt.Print(" ")
			}
			fmt.Println()
			return UNDEFINED
		}
	}
	return newError("unknown built-in function")
}

func (e *Evaluator) evalMemberExpression(me *ast.MemberExpression) value.Value {
	obj := e.Eval(me.Object)
	if obj.Type() != value.DataType_Object {
		return newError("cannot access property of non-object: %s", obj.Inspect())
	}
	objVal := obj.(*value.Object)

	var propName string
	if me.Computed {
		prop := e.Eval(me.Property)
		propName = prop.Inspect()
	} else {
		ident := me.Property.(*ast.Identifier)
		propName = ident.Value
	}

	if val, ok := objVal.Properties[propName]; ok {
		return val
	}
	return UNDEFINED
}

// --- helper functions ---

func evalUnaryOperator(operator *token.Token, right value.Value) value.Value {
	switch operator.TokenType {
	case token.TOKEN_BANG:
		return nativeBoolToBooleanObject(!isTruthy(right))
	case token.TOKEN_MINUS:
		if right.Type() == value.DataType_Number {
			val := right.(*value.Number).Value
			return &value.Number{Value: -val}
		}
		return newError("unknown operator: -%s", right.Type())
	case token.TOKEN_PLUS:
		if right.Type() == value.DataType_Number {
			return right
		}
		if right.Type() == value.DataType_String {
			val, err := strconv.ParseFloat(right.(*value.String).Value, 64)
			if err != nil {
				return newError("cannot convert string to number: %s", right.Inspect())
			}
			return &value.Number{Value: val}
		}
		return newError("unknown operator: +%s", right.Type())
	default:
		return newError("unknown operator: %s%s", operator.TokenType, right.Type())
	}
}

func evalBinaryOperator(operator *token.Token, left, right value.Value) value.Value {
	switch {
	case left.Type() == value.DataType_Number && right.Type() == value.DataType_Number:
		return evalNumberBinaryExpression(operator, left, right)
	case left.Type() == value.DataType_String && right.Type() == value.DataType_String:
		return evalStringBinaryExpression(operator, left, right)
	case operator.TokenType == token.TOKEN_EQUAL_EQUAL:
		return nativeBoolToBooleanObject(valuesEqual(left, right))
	case operator.TokenType == token.TOKEN_BANG_EQUAL:
		return nativeBoolToBooleanObject(!valuesEqual(left, right))
	case operator.TokenType == token.TOKEN_EQUAL_EQUAL_EQUAL:
		return nativeBoolToBooleanObject(left == right)
	case operator.TokenType == token.TOKEN_BANG_EQUAL_EQUAL:
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
	case token.TOKEN_PERCENT:
		return &value.Number{Value: float64(int64(lv) % int64(rv))}
	case token.TOKEN_LESS:
		return nativeBoolToBooleanObject(lv < rv)
	case token.TOKEN_GREATER:
		return nativeBoolToBooleanObject(lv > rv)
	case token.TOKEN_LESS_EQUAL:
		return nativeBoolToBooleanObject(lv <= rv)
	case token.TOKEN_GREATER_EQUAL:
		return nativeBoolToBooleanObject(lv >= rv)
	case token.TOKEN_EQUAL_EQUAL:
		return nativeBoolToBooleanObject(lv == rv)
	case token.TOKEN_BANG_EQUAL:
		return nativeBoolToBooleanObject(lv != rv)
	case token.TOKEN_EQUAL_EQUAL_EQUAL:
		return nativeBoolToBooleanObject(lv == rv)
	case token.TOKEN_BANG_EQUAL_EQUAL:
		return nativeBoolToBooleanObject(lv != rv)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator.TokenType, right.Type())
	}
}

func evalStringBinaryExpression(operator *token.Token, left, right value.Value) value.Value {
	lv := left.(*value.String).Value
	rv := right.(*value.String).Value

	switch operator.TokenType {
	case token.TOKEN_PLUS:
		return &value.String{Value: lv + rv}
	case token.TOKEN_EQUAL_EQUAL:
		return nativeBoolToBooleanObject(lv == rv)
	case token.TOKEN_BANG_EQUAL:
		return nativeBoolToBooleanObject(lv != rv)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator.TokenType, right.Type())
	}
}

func evalCompoundAssignment(operator string, left, right value.Value) value.Value {
	if left.Type() == value.DataType_Number && right.Type() == value.DataType_Number {
		lv := left.(*value.Number).Value
		rv := right.(*value.Number).Value
		switch operator {
		case "+=":
			return &value.Number{Value: lv + rv}
		case "-=":
			return &value.Number{Value: lv - rv}
		case "*=":
			return &value.Number{Value: lv * rv}
		case "/=":
			return &value.Number{Value: lv / rv}
		}
	}
	if operator == "+=" && left.Type() == value.DataType_String {
		return &value.String{Value: left.Inspect() + right.Inspect()}
	}
	return newError("unknown compound assignment: %s", operator)
}

func isTruthy(obj value.Value) bool {
	if obj == nil {
		return false
	}
	switch obj.Type() {
	case value.DataType_Null, value.DataType_Undefined:
		return false
	case value.DataType_Boolean:
		return obj.(*value.Boolean).Value
	case value.DataType_Number:
		return obj.(*value.Number).Value != 0
	case value.DataType_String:
		return obj.(*value.String).Value != ""
	default:
		return true
	}
}

func valuesEqual(left, right value.Value) bool {
	if left.Type() != right.Type() {
		return false
	}
	switch left.Type() {
	case value.DataType_Null, value.DataType_Undefined:
		return true
	case value.DataType_Boolean:
		return left.(*value.Boolean).Value == right.(*value.Boolean).Value
	case value.DataType_Number:
		return left.(*value.Number).Value == right.(*value.Number).Value
	case value.DataType_String:
		return left.(*value.String).Value == right.(*value.String).Value
	default:
		return left == right
	}
}

func nativeBoolToBooleanObject(input bool) *value.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func newError(format string, a ...interface{}) value.Value {
	return &value.ErrorValue{Message: fmt.Sprintf(format, a...)}
}

func isError(obj value.Value) bool {
	if obj != nil {
		_, ok := obj.(*value.ErrorValue)
		return ok
	}
	return false
}

// ReturnValue wraps a return value to propagate it through nested blocks
type ReturnValue struct {
	Value      value.Value
	isBreak    bool
	isContinue bool
	isThrow    bool
}

func (rv *ReturnValue) Type() value.DataType {
	return rv.Value.Type()
}

func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}
