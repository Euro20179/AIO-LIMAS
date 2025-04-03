package search

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"aiolimas/lua-api"

	"github.com/yuin/gopher-lua"
)

type TT int

const (
	TT_WORD TT = iota

	TT_MACRO TT = iota

	TT_STRING TT = iota

	TT_COLON TT = iota

	TT_SIMILAR TT = iota

	TT_EQ  TT = iota
	TT_LT  TT = iota
	TT_LE  TT = iota
	TT_GT  TT = iota
	TT_GE  TT = iota
	TT_NOT TT = iota

	TT_IN TT = iota

	TT_LPAREN TT = iota
	TT_RPAREN TT = iota

	TT_PRESERVED TT = iota

	TT_MUL  TT = iota
	TT_DIV  TT = iota
	TT_SUB  TT = iota
	TT_PLUS TT = iota

	TT_AND TT = iota
	TT_OR  TT = iota

	TT_NUMBER TT = iota
)

type Token struct {
	Ty    TT
	Value string
}

func Lex(search string) []Token {
	i := -1

	next := func() bool {
		i += 1
		return i < len(search)
	}

	back := func() {
		i--
	}

	parseNumber := func() string {
		hasDot := false

		final := string(search[i])
		for next() {
			ch := search[i]
			if ch == '.' && !hasDot {
				hasDot = true
				final += string(ch)
			} else if ch >= '0' && ch <= '9' {
				final += string(ch)
			} else {
				break
			}
		}

		return final
	}

	parseWord := func(quote string, forbiddenchars []rune) string {
		inQuote := true

		final := ""
		if quote == "" {
			final = string(search[i])
		}

		escape := false

		endOfWordChars := " \t\n;<>=&~^|"

		for next() && inQuote {
			ch := search[i]
			if slices.Contains(forbiddenchars, rune(ch)) {
				break
			}
			if ch == '\\' {
				escape = true
				continue
			}
			if !escape && quote == "" && strings.Contains(endOfWordChars, string(ch)) {
				back()
				break
			}
			if !escape && string(ch) == quote {
				break
			}
			escape = false
			final += string(ch)
		}

		return final
	}

	parseBrace := func() string {
		final := ""
		braceCount := 1
		for next() {
			ch := search[i]
			switch ch {
			case '{':
				braceCount++
			case '}':
				braceCount--
			}
			if braceCount != 0 {
				final += string(ch)
			} else {
				break
			}
		}

		return final
	}

	lexSearch := func() []Token {
		var tokens []Token
		for next() {
			ch := search[i]

			var ty TT
			var val string

			switch ch {
			case ' ', '\t', '\n':
				continue
			case '~':
				ty = TT_SIMILAR
				val = "~"
			case '*':
				ty = TT_MUL
				val = "*"
			case '/':
				ty = TT_DIV
				val = "/"
			case '-':
				ty = TT_SUB
				val = "-"
			case '+':
				ty = TT_PLUS
				val = "+"
			case ':':
				ty = TT_COLON
				val = ":"
			case '(':
				ty = TT_LPAREN
				val = "("
			case ')':
				ty = TT_RPAREN
				val = ")"
			case '{':
				ty = TT_PRESERVED
				val = parseBrace()
			case '?':
				fallthrough
			case '|':
				ty = TT_OR
				val = string(ch)
			case '^':
				ty = TT_IN
				val = "^"
			case '.':
				fallthrough
			case '&':
				ty = TT_AND
				val = string(ch)
			case '=':
				if len(search) > 1 && search[i+1] == '=' {
					next()
					ty = TT_EQ
					val = "=="
				} else {
					ty = TT_EQ
					val = "="
				}
			case '>':
				if len(search) > 1 && search[i+1] == '=' {
					next()
					ty = TT_GE
					val = ">="
				} else {
					ty = TT_GT
					val = ">"
				}
			case '<':
				if len(search) > 1 && search[i+1] == '=' {
					next()
					ty = TT_LE
					val = "<="
				} else {
					next()
					ty = TT_LT
					val = "<"
				}
			case '!':
				ty = TT_NOT
				val = "!"
			case '"':
				ty = TT_STRING
				val = parseWord("\"", []rune{})
			case '\'':
				ty = TT_STRING
				val = parseWord("'", []rune{})
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				ty = TT_NUMBER
				val = parseNumber()
			case '@': fallthrough
			case '#':
				ty = TT_MACRO
				// parseWord includes the first char, ignore it
				val = parseWord("", []rune{'\'', '|', '&', '"', ')', '('})[1:]
			default:
				ty = TT_WORD
				val = parseWord("", []rune{})
			}

			tokens = append(tokens, Token{
				Ty:    ty,
				Value: val,
			})
		}
		return tokens
	}
	return lexSearch()
}

type Node interface {
	ToString() (string, error)
}

type ListNode struct {
	Items []Node
}

func (self ListNode) ToString() (string, error) {
	str := "("
	for _, item := range self.Items {
		newText, err := item.ToString()
		if err != nil {
			return "", err
		}
		str += newText + ","
	}
	str = str[0 : len(str)-1]
	return str + ")", nil
}

type MacroNode struct {
	Value string
}

func (self MacroNode) ToString() (string, error) {
	onExpand, exists := lua_api.LuaEventRegistry["MacroExpand"]
	if !exists {
		return "", errors.New("Could not expand macro")
	}

	curMacro := self.Value

	state := lua_api.GlobalLuaInstance

	for _, expansion := range onExpand {
		state.Push(expansion)
		state.Push(lua.LString(curMacro))
		state.Call(1, 2)
		userErr := state.Get(-1).(lua.LString)
		if userErr != "" {
			return curMacro, errors.New(string(userErr))
		}

		text := state.Get(-2).(lua.LString)
		if text != "" {
			curMacro = string(text)
		}
	}

	return curMacro, nil
}

type StringNode struct {
	Value string
}

func (self StringNode) ToString() (string, error) {
	newVal, err := json.Marshal(self.Value)
	if err != nil {
		return "", err
	}
	return string(newVal), nil
}

type NumberNode struct {
	Value string
}

func (self NumberNode) ToString() (string, error) {
	return self.Value, nil
}

type NegateNode struct {
	Right Node
}

func (self NegateNode) ToString() (string, error) {
	r, err := self.Right.ToString()
	if err != nil {
		return "!", err
	}
	return "not " + r, nil
}

type PlainWordNode struct {
	Value string
}

func (self PlainWordNode) ToString() (string, error) {
	return strings.ReplaceAll(self.Value, ";", ""), nil
}

type OperatorNode struct {
	Operator TT
	Negate   bool
}

func (self OperatorNode) ToString() (string, error) {
	negatedOps := map[string]string{
		"=":  "!=",
		"<=": ">",
		">=": "<",
		"~":  "!~",
		"&":  "|",
		"^":  "!^",
	}

	for k, v := range negatedOps {
		negatedOps[v] = k
	}

	opNames := map[string]string{
		"&":  " AND ",
		"|":  " OR ",
		"~":  " LIKE ",
		"!~": " NOT LIKE ",
		"^":  " IN ",
		"!^": " NOT IN ",
	}

	strOp := ""
	switch self.Operator {
	case TT_EQ:
		strOp = "="
	case TT_LE:
		strOp = "<="
	case TT_LT:
		strOp = "<"
	case TT_GE:
		strOp = ">="
	case TT_GT:
		strOp = ">"
	case TT_SIMILAR:
		strOp = "~"
	case TT_AND:
		strOp = "&"
	case TT_OR:
		strOp = "|"
	case TT_IN:
		strOp = "^"
	default:
		panic(fmt.Sprintf("Unknown operator: %d", self.Operator))
	}

	if self.Negate {
		strOp = negatedOps[strOp]
	}

	name, hasName := opNames[strOp]
	if !hasName {
		name = strOp
	}

	return name, nil
}

type BinOpNode struct {
	Left     Node
	Right    Node
	Operator OperatorNode
}

func (self BinOpNode) ToString() (string, error) {
	op, err := self.Operator.ToString()
	if err != nil {
		return "", err
	}
	left, err := self.Left.ToString()
	if err != nil {
		return "", err
	}
	right, err := self.Right.ToString()
	if err != nil {
		return "", err
	}

	return "(" + left + op + right + ")", nil
}

func Parse(tokens []Token) (string, error) {
	i := -1

	var search func() Node
	var atom func() Node
	var comparison func() Node

	next := func() bool {
		i++
		return i < len(tokens)
	}

	back := func() {
		i--
	}

	atom = func() Node {
		if tokens[i].Ty == TT_NOT {
			if !next() {
				return NegateNode{
					Right: StringNode{
						Value: "",
					},
				}
			}
			return NegateNode{
				Right: atom(),
			}
		}

		switch tokens[i].Ty {
		case TT_STRING:
			return StringNode{
				Value: tokens[i].Value,
			}
		case TT_MACRO:
			return MacroNode{
				Value: tokens[i].Value,
			}
		case TT_WORD:
			return PlainWordNode{
				Value: tokens[i].Value,
			}
		case TT_PRESERVED:
			return PlainWordNode{
				Value: tokens[i].Value,
			}
		case TT_NUMBER:
			return NumberNode{
				Value: tokens[i].Value,
			}
		case TT_LPAREN:
			n := search()
			next()
			return n
		}

		return StringNode{}
	}

	atomList := func() Node {
		items := []Node{
			atom(),
		}

		wantsList := false

		for next() {
			if tokens[i].Ty != TT_COLON {
				back()
				break
			}
			if next() {
				items = append(items, atom())
			} else {
				// the user put a trailing colon,
				// they probably want a list with 1 irem
				wantsList = true
			}
		}

		if len(items) == 1 && !wantsList {
			return items[0]
		}

		return ListNode{
			Items: items,
		}
	}

	comparison = func() Node {
		left := atomList()
		compToks := []TT{
			TT_LT, TT_GT, TT_LE, TT_GE, TT_EQ, TT_SIMILAR, TT_IN,
		}

		negated := false
		for next() {
			if tokens[i].Ty == TT_NOT {
				negated = true
				continue
			}
			if !slices.Contains(compToks, tokens[i].Ty) {
				back()
				break
			}
			op := tokens[i]
			if !next() {
				return StringNode{}
			}
			right := atomList()
			left = BinOpNode{
				Left: left,
				Operator: OperatorNode{
					Operator: op.Ty,
					Negate:   negated,
				},
				Right: right,
			}
			negated = false
		}
		return left
	}

	gate := func() Node {
		next()
		logicToks := []TT{TT_AND, TT_OR}
		left := comparison()
		for next() {
			op := tokens[i]

			if !slices.Contains(logicToks, op.Ty) {
				continue
			}

			if !next() {
				return StringNode{}
			}

			right := comparison()
			left = BinOpNode{
				Left: left,
				Operator: OperatorNode{
					Operator: op.Ty,
					Negate:   false,
				},
				Right: right,
			}
		}
		return left
	}

	search = func() Node {
		return gate()
	}

	sn := search()

	return sn.ToString()
}

func Search2String(search string) (string, error) {
	tokens := Lex(search)
	return Parse(tokens)
}
