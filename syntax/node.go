package syntax

type (
	Expr interface{}

	Identifier struct {
		Value string
		Expr
	}

	Literal struct {
		Value string
		Kind  LitKind
		Expr
	}

	Elision struct{}

	SpreadElem struct {
		Value Expr
	}

	ArrayLiteral struct {
		ElemList []Expr
	}

	UnaryExpr struct {
		Op    TokenType
		Value Expr
		Expr
	}

	OpExpr struct {
		Op          TokenType
		Left, Right Expr
		Expr
	}

	TernaryExpr struct {
		Cond        Expr
		True, False Expr
		Expr
	}
)

type (
	Stmt interface{}

	BlockStmt struct {
		List []Stmt
	}

	EmptyStmt struct{}

	ExprStmt struct {
		Expr Expr
	}

	IfStmt struct {
		Cond Expr
		Then Stmt
		Else Stmt
	}

	ForStmt struct {
		Init Expr
		Cond Expr
		Post Expr
		Body Stmt
	}

	DoWhileStmt struct {
		Body Stmt
		Cond Expr
	}

	ContinueStmt struct {
		Label Expr
	}

	BreakStmt struct {
		Label Expr
	}

	ReturnStmt struct {
		Result Expr
	}

	SwitchStmt struct {
		Tag  Expr
		Body []*CaseClause
	}

	ThrowStmt struct {
		Result Expr
	}

	TryStmt struct {
		Try        Stmt
		CatchParam Identifier
		Catch      Stmt
		Finally    Stmt
	}

	DebuggerStmt struct{}
)

type (
	CaseClause struct {
		Case Expr
		Body []Stmt
	}
)

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
