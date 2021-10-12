package interpreter

import (
	"log"
	"programminglang/interpreter/symbols"
	"programminglang/types"
)

type VariableDeclaration struct {
	VariableNode AbstractSyntaxTree // a Variable struct
	TypeNode     AbstractSyntaxTree // a VariableType struct
}

type VariableType struct {
	Token types.Token
}

type Variable struct {
	Token types.Token
	Value string
}

func (v VariableDeclaration) Op() types.Token {
	return types.Token{}
}
func (v VariableDeclaration) LeftOperand() AbstractSyntaxTree {
	return v.VariableNode
}
func (v VariableDeclaration) RightOperand() AbstractSyntaxTree {
	return v.TypeNode
}
func (v VariableDeclaration) Visit(i *Interpreter) {
	typeName := v.TypeNode.Op().Value

	typeSymbol, _ := i.CurrentScope.LookupSymbol(typeName, false)

	variableName := v.VariableNode.Op().Value

	if alreadyDeclaredVarName, exists := i.CurrentScope.LookupSymbol(variableName, true); exists {
		// variable alreadyDeclaredVarName has already been declared
		log.Fatal("Error: Variable, ", alreadyDeclaredVarName, " has already been declared")
	}

	i.CurrentScope.DefineSymbol(symbols.Symbol{
		Name: variableName,
		Type: typeSymbol.Name,
	})

}

func (v VariableType) Op() types.Token {
	return v.Token
}
func (v VariableType) LeftOperand() AbstractSyntaxTree {
	return v
}
func (v VariableType) RightOperand() AbstractSyntaxTree {
	return v
}
func (v VariableType) Visit(s *Interpreter) {}

func (v Variable) Op() types.Token {
	return v.Token
}
func (v Variable) LeftOperand() AbstractSyntaxTree {
	return v
}
func (v Variable) RightOperand() AbstractSyntaxTree {
	return v
}
func (v Variable) Visit(i *Interpreter) {
	varName := v.Value
	_, exists := i.CurrentScope.LookupSymbol(varName, false)

	if !exists {
		log.Fatal("Variable, ", varName, " is not defined")
	}

}
