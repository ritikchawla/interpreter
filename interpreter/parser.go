package interpreter

import (
	"fmt"

	"programminglang/constants"
	"programminglang/helpers"
	"programminglang/interpreter/errors"
	"programminglang/types"
)

type Parser struct {
	Lexer        LexicalAnalyzer
	CurrentToken types.Token
}

func (p *Parser) Init(text string) {
	p.Lexer = LexicalAnalyzer{
		Text: text,
	}

	p.Lexer.Init()

	p.CurrentToken = p.Lexer.GetNextToken()
}

func (p *Parser) Error(errorCode string, token types.Token) {
	// log.Fatal(
	// 	"Bad Token",
	// 	"\nCurrent Token: ", p.CurrentToken.Print(),
	// 	"\nToken Type to check with ", tokenType,
	// )

	parseError := errors.ParseError{
		ErrorCode: errorCode,
		Token:     token,
		Message:   fmt.Sprintf("%s -> %s", errorCode, token.Print()),
	}

	parseError.Print()
}

/*
	Validate whether the current token maches the token type passed in.

	If valid advances the parser pointer.

	If not valid, prints a fatal error and exits
*/
func (p *Parser) ValidateToken(tokenType string) {
	// fmt.Println("Validating Token ", p.CurrentToken)
	// fmt.Println("Validating against ", tokenType, "\n\n")

	if p.CurrentToken.Type == tokenType {
		p.CurrentToken = p.Lexer.GetNextToken()
		// fmt.Println("\n\n", p.CurrentToken, "\n\n")
	} else {
		p.Error(constants.ERROR_UNEXPECTED_TOKEN, p.CurrentToken)
	}
}

/*
	1. Gets the current token

	2. Validates the current token as integer

	3. Returns the IntegerValue of the token

	TERM --> FACTOR ((MUL | DIV) FACTOR)*
*/
func (p *Parser) Term() AbstractSyntaxTree {
	returningValue := p.Factor()

	for helpers.ValueInSlice(p.CurrentToken.Type, constants.MUL_DIV_SLICE) {
		currentToken := p.CurrentToken

		// fmt.Println("current token in term is saved")

		switch p.CurrentToken.Type {
		case constants.INTEGER_DIV:
			p.ValidateToken(constants.INTEGER_DIV)

		case constants.FLOAT_DIV:
			p.ValidateToken(constants.FLOAT_DIV)

		case constants.MUL:
			p.ValidateToken(constants.MUL)
		}

		returningValue = BinaryOperationNode{
			Left:      returningValue,
			Operation: currentToken,
			Right:     p.Factor(),
		}
	}

	// fmt.Println("\nreturinig from p.Term = ", returningValue)

	return returningValue
}

/*
	FACTOR --> ((PLUS | MINUS) FACTOR) | INTEGER | LPAREN EXPRESSION RPAREN
*/
func (p *Parser) Factor() AbstractSyntaxTree {
	token := p.CurrentToken

	var returningValue AbstractSyntaxTree

	switch token.Type {
	case constants.PLUS:
		p.ValidateToken(constants.PLUS)
		returningValue = UnaryOperationNode{
			Operation: token,
			Operand:   p.Factor(),
		}

	case constants.MINUS:
		p.ValidateToken(constants.MINUS)
		returningValue = UnaryOperationNode{
			Operation: token,
			Operand:   p.Factor(),
		}

	case constants.INTEGER:
		p.ValidateToken(constants.INTEGER)
		returningValue = IntegerNumber{
			Token: token,
			Value: token.IntegerValue,
		}

	case constants.FLOAT:
		p.ValidateToken(constants.FLOAT)
		returningValue = FloatNumber{
			Token: token,
			Value: token.FloatValue,
		}

	case constants.LPAREN:
		p.ValidateToken(constants.LPAREN)
		returningValue = p.Expression()
		p.ValidateToken(constants.RPAREN)

	default:
		returningValue = p.Variable()
	}

	// fmt.Println("\nreturining from Factor = ", returningValue)

	return returningValue
}

/*
	Parser / Parser

	EXPRESSION --> TERM ((PLUS | MINUS) TERM)*
*/
func (p *Parser) Expression() AbstractSyntaxTree {
	result := p.Term()

	// fmt.Println("\nin Expression p.Term = ", result)

	for helpers.ValueInSlice(p.CurrentToken.Type, constants.PLUS_MINUS_SLICE) {
		currentToken := p.CurrentToken

		switch p.CurrentToken.Value {
		case constants.PLUS_SYMBOL:
			// this will advance the pointer
			p.ValidateToken(constants.PLUS)

		case constants.MINUS_SYMBOL:
			// this will advance the pointer
			p.ValidateToken(constants.MINUS)
		}

		result = BinaryOperationNode{
			Left:      result,
			Operation: currentToken,
			Right:     p.Term(),
		}
	}

	return result
}

func (p *Parser) Program() AbstractSyntaxTree {
	declarationNodes := p.Declarations()
	compoundStatementNodes := p.CompoundStatement()

	node := Program{
		Declarations:      declarationNodes,
		CompoundStatement: compoundStatementNodes,
	}

	return node
}

// declarations --> LET (variable_declaration SEMI)+ | blank
func (p *Parser) Declarations() []AbstractSyntaxTree {
	var declarations []AbstractSyntaxTree

	// variables are defined as, let varialble_name(s) : variable_type;
	if p.CurrentToken.Type == constants.LET {
		// this is messed up. there is no type called constants.LET
		p.ValidateToken(constants.LET)

		for p.CurrentToken.Type == constants.IDENTIFIER {
			varDeclaration := p.VariableDeclaration()
			declarations = append(declarations, varDeclaration...)
			p.ValidateToken(constants.SEMI_COLON)
		}

	}

	for p.CurrentToken.Type == constants.DEFINE {
		p.ValidateToken(constants.DEFINE)

		functionName := p.CurrentToken.Value

		p.ValidateToken(constants.IDENTIFIER)

		functionBlock := p.Program()

		function := FunctionDeclaration{
			FunctionName:  functionName,
			FunctionBlock: functionBlock,
		}

		declarations = append(declarations, function)
	}

	return declarations
}

// variable_declaration --> ID (COMMA ID)* COLON var_type
func (p *Parser) VariableDeclaration() []AbstractSyntaxTree {
	// current node is a variable node
	variableNodes := []AbstractSyntaxTree{Variable{Token: p.CurrentToken, Value: p.CurrentToken.Value}}
	p.ValidateToken(constants.IDENTIFIER)

	// variables can be separated by comma so keep iterating while there's a comma
	for p.CurrentToken.Type == constants.COMMA {
		p.ValidateToken(constants.COMMA)

		variableNodes = append(variableNodes, Variable{Token: p.CurrentToken, Value: p.CurrentToken.Value})

		p.ValidateToken(constants.IDENTIFIER)
	}

	// var variableName : variableType
	// variable name and type will be separated by a colon
	p.ValidateToken(constants.COLON)

	variableType := p.VarType()

	// make a new slice to store all the variable declarations
	var variableDeclarations []AbstractSyntaxTree

	for _, node := range variableNodes {
		newVarDeclr := VariableDeclaration{
			VariableNode: node,
			TypeNode:     variableType,
		}

		variableDeclarations = append(variableDeclarations, newVarDeclr)
	}

	return variableDeclarations
}

// formal_parameter_list --> formal_parameters | formal_parameters SEMI_COLON formal_parameter_list
func (p *Parser) FormalParametersList() []AbstractSyntaxTree {
	var paramNodes []AbstractSyntaxTree

	if p.CurrentToken.Type != constants.IDENTIFIER {
		return paramNodes
	}

	paramNodes = p.FormalParameters()

	for p.CurrentToken.Type == constants.SEMI_COLON {
		p.ValidateToken(constants.SEMI_COLON)
		paramNodes = append(paramNodes, p.FormalParameters()...)
	}

	return paramNodes
}

// formal_parameters --> ID (COMMA ID)* COLON type_spec
func (p *Parser) FormalParameters() []AbstractSyntaxTree {
	var paramNodes []AbstractSyntaxTree

	paramTokens := []types.Token{p.CurrentToken}

	p.ValidateToken(constants.IDENTIFIER)

	for p.CurrentToken.Type == constants.COMMA {
		p.ValidateToken(constants.COMMA)
		paramTokens = append(paramTokens, p.CurrentToken)
		p.ValidateToken(constants.IDENTIFIER)
	}

	p.ValidateToken(constants.COLON)

	typeNode := p.VarType()

	for _, parameterToken := range paramTokens {
		paramNodes = append(paramNodes, FunctionParameters{
			VariableNode: Variable{
				Token: parameterToken,
				Value: parameterToken.Value,
			},
			TypeNode: typeNode,
		})
	}

	return paramNodes

}

// var_type --> INTEGER_TYPE | FLOAT_TYPE
func (p *Parser) VarType() AbstractSyntaxTree {
	token := p.CurrentToken

	if token.Type == constants.INTEGER_TYPE {
		p.ValidateToken(constants.INTEGER_TYPE)
	} else {
		p.ValidateToken(constants.FLOAT_TYPE)
	}

	return VariableType{
		Token: token,
	}

}

func (p *Parser) CompoundStatement() AbstractSyntaxTree {
	nodes := p.StatementList()

	root := CompoundStatement{}

	root.Children = append(root.Children, nodes...)

	return root
}

// statement_list --> statement SEMI_COLON | statement SEMI_COLON statement_list
func (p *Parser) StatementList() []AbstractSyntaxTree {
	node := p.Statement()

	results := []AbstractSyntaxTree{node}

	for p.CurrentToken.Type == constants.SEMI_COLON {
		p.ValidateToken(constants.SEMI_COLON)
		results = append(results, p.Statement())
	}

	// if p.CurrentToken.Type == constants.IDENTIFIER {
	// 	p.Error(constants.SEMI_COLON)
	// }

	return results
}

/*
	statement --> assignment_statement | blank
*/
func (p *Parser) Statement() AbstractSyntaxTree {
	var node AbstractSyntaxTree

	if p.CurrentToken.Type == constants.IDENTIFIER {
		node = p.AssignmentStatement()
	} else if p.CurrentToken.Type == constants.INTEGER || p.CurrentToken.Type == constants.FLOAT {
		node = p.Expression()
	} else {
		node = BlankStatement{
			Token: types.Token{
				Type:  constants.BLANK,
				Value: "",
			},
		}
	}

	return node

}

/*
	assignment_statement --> variable ASSIGN expression
*/
func (p *Parser) AssignmentStatement() AbstractSyntaxTree {
	left := p.Variable()

	token := p.CurrentToken
	p.ValidateToken(constants.ASSIGN)

	right := p.Expression()

	return AssignmentStatement{
		Left:  left,
		Token: token,
		Right: right,
	}
}

/*
	variable --> ID
*/
func (p *Parser) Variable() AbstractSyntaxTree {
	variable := Variable{
		Token: p.CurrentToken,
		Value: p.CurrentToken.Value,
	}

	p.ValidateToken(constants.IDENTIFIER)

	return variable
}

func (p *Parser) Parse() AbstractSyntaxTree {
	return p.Program()
	// return p.Expression()
}
