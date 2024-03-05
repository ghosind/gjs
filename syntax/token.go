package syntax

type TokenType int

const (
	TOKEN_INVALID TokenType = iota

	TOKEN_LEFT_PAREN    // (
	TOKEN_RIGHT_PAREN   // )
	TOKEN_LEFT_BRACE    // {
	TOKEN_RIGHT_BRACE   // }
	TOKEN_LEFT_BRACKET  // [
	TOKEN_RIGHT_BRACKET // ]

	TOKEN_AND                           // &
	TOKEN_AND_AND                       // &&
	TOKEN_AND_AND_EQUAL                 // &&=
	TOKEN_AND_EQUAL                     // &=
	TOKEN_BANG                          // !
	TOKEN_BANG_EQUAL                    // !=
	TOKEN_BANG_EQUAL_EQUAL              // !==
	TOKEN_COLON                         // :
	TOKEN_COMMA                         // ,
	TOKEN_DOT                           // .
	TOKEN_DOT_DOT_DOT                   // ...
	TOKEN_EQUAL                         // =
	TOKEN_EQUAL_EQUAL                   // ==
	TOKEN_EQUAL_EQUAL_EQUAL             // ===
	TOKEN_GREATER                       // >
	TOKEN_GREATER_EQUAL                 // >=
	TOKEN_GREATER_GREATER               // >>
	TOKEN_GREATER_GREATER_EQUAL         // >>=
	TOKEN_GREATER_GREATER_GREATER       // >>>
	TOKEN_GREATER_GREATER_GREATER_EQUAL // >>>=
	TOKEN_HAT                           // ^
	TOKEN_HAT_EQUAL                     // ^=
	TOKEN_LESS                          // <
	TOKEN_LESS_EQUAL                    // <=
	TOKEN_LESS_LESS                     // <<
	TOKEN_LESS_LESS_EQUAL               // <<=
	TOKEN_MINUS                         // -
	TOKEN_MINUS_EQUAL                   // -=
	TOKEN_MINUS_MINUS                   // --
	TOKEN_PERCENT                       // %
	TOKEN_PERCENT_EQUAL                 // %=
	TOKEN_PIPE                          // |
	TOKEN_PIPE_EQUAL                    // |=
	TOKEN_PIPE_PIPE                     // ||
	TOKEN_PIPE_PIPE_EQUAL               // ||=
	TOKEN_PLUS                          // +
	TOKEN_PLUS_EQUAL                    // +=
	TOKEN_PLUS_PLUS                     // ++
	TOKEN_QUESTION                      // ?
	TOKEN_QUESTION_DOT                  // ?.
	TOKEN_QUESTION_QUESTION             // ??
	TOKEN_QUESTION_QUESTION_EQUAL       // ??=
	TOKEN_SEMICOLON                     // ;
	TOKEN_SLASH                         // /
	TOKEN_SLASH_EQUAL                   // /=
	TOKEN_STAR                          // *
	TOKEN_STAR_EQUAL                    // *=
	TOKEN_STAR_STAR                     // **
	TOKEN_STAR_STAR_EQUAL               // **=
	TOKEN_TILDE                         // ~

	TOKEN_IDENTIFIER
	TOKEN_STRING
	TOKEN_NUMBER

	TOKEN_ARGUMENTS
	TOKEN_AS
	TOKEN_ASYNC
	TOKEN_AWAIT
	TOKEN_BREAK
	TOKEN_CASE
	TOKEN_CATCH
	TOKEN_CLASS
	TOKEN_CONST
	TOKEN_CONTINUE
	TOKEN_DEBUGGER
	TOKEN_DEFAULT
	TOKEN_DELETE
	TOKEN_DO
	TOKEN_ELSE
	TOKEN_ENUM
	TOKEN_EVAL
	TOKEN_EXPORT
	TOKEN_EXTENDS
	TOKEN_FALSE
	TOKEN_FINALLY
	TOKEN_FOR
	TOKEN_FROM
	TOKEN_FUNCTION
	TOKEN_GET
	TOKEN_IF
	TOKEN_IMPLEMENTS
	TOKEN_IMPORT
	TOKEN_IN
	TOKEN_INSTANCEOF
	TOKEN_INTERFACE
	TOKEN_LET
	TOKEN_META
	TOKEN_NEW
	TOKEN_NULL
	TOKEN_OF
	TOKEN_PACKAGE
	TOKEN_PRIVATE
	TOKEN_PROTECTED
	TOKEN_PUBLIC
	TOKEN_RETURN
	TOKEN_SET
	TOKEN_STATIC
	TOKEN_SUPER
	TOKEN_SWITCH
	TOKEN_TARGET
	TOKEN_THIS
	TOKEN_THROW
	TOKEN_TRUE
	TOKEN_TRY
	TOKEN_TYPEOF
	TOKEN_UNDEFINED
	TOKEN_VAR
	TOKEN_VOID
	TOKEN_WHILE
	TOKEN_WITH
	TOKEN_YIELD

	TOKEN_NEW_LINE
	TOKEN_SPACE

	TOKEN_EOF
)

type Token struct {
	ToKenType TokenType
	Line      int
	Literal   string
}
