package starcomplete

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type CompletionItemKind int

const (
	CIKMethod CompletionItemKind = iota
	CIKFunction
	CIKConstructor
	CIKField
	CIKVariable
	CIKClass
	CIKStruct
	CIKInterface
	CIKModule
	CIKProperty
	CIKEvent
	CIKOperator
	CIKUnit
	CIKValue
	CIKConstant
	CIKEnum
	CIKEnumMember
	CIKKeyword
	CIKText
	CIKColor
	CIKFile
	CIKReference
	CIKCustomcolor
	CIKFolder
	CIKTypeParameter
	CIKSnippet
)

func (k CompletionItemKind) String() string {
	return [...]string{
		"Method",
		"Function",
		"Constructor",
		"Field",
		"Variable",
		"Class",
		"Struct",
		"Interface",
		"Module",
		"Property",
		"Event",
		"Operator",
		"Unit",
		"Value",
		"Constant",
		"Enum",
		"EnumMember",
		"Keyword",
		"Text",
		"Color",
		"File",
		"Reference",
		"Customcolor",
		"Folder",
		"TypeParameter",
		"Snippet",
	}[k]
}

// Completion is a completion suggestion
type Completion struct {
	InsertText    string             `json:"insertText"`
	Detail        string             `json:"detail"`
	Kind          CompletionItemKind `json:"kind"`
	Label         string             `json:"label"`
	Documentation string             `json:"documentation"`
	Range         Range              `json:"range"`
}

func (c Completion) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"insertText":    c.InsertText,
		"detail":        c.Detail,
		"kind":          c.Kind,
		"label":         c.Label,
		"documentation": c.Documentation,
		"range":         c.Range.ToMap(),
	}
}

type Range struct {
	StartLineNumber int32 `json:"startLineNumber"`
	StartColumn     int32 `json:"startColumn"`
	EndLineNumber   int32 `json:"endLineNumber"`
	EndColumn       int32 `json:"endColumn"`
}

func NewRange(start, end Position) Range {
	return Range{
		StartLineNumber: start.LineNumber,
		StartColumn:     start.Column,
		EndLineNumber:   end.LineNumber,
		EndColumn:       end.Column,
	}
}

func (r Range) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"startLineNumber": r.StartLineNumber,
		"startColumn":     r.StartColumn,
		"endLineNumber":   r.EndLineNumber,
		"endColumn":       r.EndColumn,
	}
}

type Position struct {
	LineNumber int32 `json:"lineNumber"`
	Column     int32 `json:"column"`
}

func NewPosition(l, c int32) Position {
	return Position{
		LineNumber: l,
		Column:     c,
	}
}

type ModuleInfo struct {
	Name                string
	Documentation       string
	DefaultImportSymbol string
}

func (mi ModuleInfo) LoadString() string {
	if mi.DefaultImportSymbol != "" {
		return fmt.Sprintf("load(%q,%q)", mi.Name, mi.DefaultImportSymbol)
	}
	return fmt.Sprintf("load(%q)", mi.Name)
}

func (mi ModuleInfo) Completion(p Position) Completion {
	return Completion{
		InsertText:    mi.LoadString(),
		Detail:        "module",
		Kind:          CIKFolder,
		Label:         mi.Name,
		Documentation: mi.Documentation,
		Range:         NewRange(p, p),
	}
}

func Completions(filename string, src interface{}, p Position, predeclared starlark.StringDict, modules []ModuleInfo) ([]Completion, error) {
	// 1. parse input program
	file, err := syntax.Parse(filename, src, 0)
	if err != nil {
		if _, ok := err.(syntax.Error); ok {
			// ignore syntax errors. We're ok with programs the don't compile
			return nil, nil
		} else {
			return nil, err
		}
	}

	// file, _, err := starlark.SourceProgram(filename, src, predeclared.Has)
	// if err != nil {
	// }

	// 2. get token in the AST the position cursor is intersecting
	stmt, err := statementAtPosition(file, p)
	if err != nil {
		return nil, err
	}

	// at this point we have a positional set of *statements*, but not their
	// connection to an executable program...

	// 3. map AST symbol to completion prefixes
	return completionsForStatment(stmt, p, predeclared, modules)
}

func CompletionsToMap(cmpls []Completion) []map[string]interface{} {
	r := make([]map[string]interface{}, 0, len(cmpls))
	for _, cmpl := range cmpls {
		r = append(r, cmpl.ToMap())
	}
	return r
}

func statementAtPosition(file *syntax.File, p Position) (syntax.Stmt, error) {
	for _, stmt := range file.Stmts {
		start, end := stmt.Span()
		if end.Line == p.LineNumber && start.Col <= p.Column && end.Col >= p.Column {
			return stmt, nil
		} else if end.Line > p.LineNumber {
			break
		}
	}
	return nil, fmt.Errorf("token not found")
}

func completionsForStatment(stmt syntax.Stmt, p Position, predeclared starlark.StringDict, modules []ModuleInfo) ([]Completion, error) {
	fmt.Printf("statement: %#v\n", stmt)
	switch stmt.(type) {
	case *syntax.AssignStmt:
	case *syntax.BranchStmt:
	case *syntax.DefStmt:
	case *syntax.ExprStmt:
	case *syntax.ForStmt:
	case *syntax.WhileStmt:
	case *syntax.IfStmt:
	case *syntax.LoadStmt:
		return moduleInfosToCompletions(modules, p), nil
	case *syntax.ReturnStmt:
	}
	return nil, nil
}

func moduleInfosToCompletions(mods []ModuleInfo, p Position) []Completion {
	if len(mods) == 0 {
		return nil
	}

	cmpls := make([]Completion, 0, len(mods))
	for _, mi := range mods {
		cmpls = append(cmpls, mi.Completion(p))
	}

	return cmpls
}
