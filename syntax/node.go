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

type LitKind int

const (
	LitNull LitKind = iota
	LitBoolean
	LitNumber
	LitString
)
