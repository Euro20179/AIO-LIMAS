package search

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

type TT int

const (
	TT_WORD TT = iota

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

	parseWord := func(quote string) string {
		inQuote := true

		final := ""
		if quote == "" {
			final = string(search[i])
		}

		escape := false

		endOfWordChars := " \t\n;<>=&~^|"

		for next() && inQuote {
			ch := search[i]
			if ch == '\\' {
				escape = true
				continue
			}
			if quote == "" && strings.Contains(endOfWordChars, string(ch)) {
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
			case '|':
				ty = TT_OR
				val = "|"
			case '^':
				ty = TT_IN
				val = "^"
			case '&':
				ty = TT_AND
				val = "&"
			case '=':
				if len(search) > 1 && search[1] == '=' {
					ty = TT_EQ
					val = "=="
				} else {
					ty = TT_EQ
					val = "="
				}
			case '>':
				if len(search) > 1 && search[1] == '=' {
					ty = TT_GE
					val = ">="
				} else {
					ty = TT_GT
					val = ">"
				}
			case '<':
				if len(search) > 1 && search[1] == '=' {
					ty = TT_LE
					val = "<="
				} else {
					ty = TT_LT
					val = "<"
				}
			case '!':
				ty = TT_NOT
				val = "!"
			case '"':
				ty = TT_STRING
				val = parseWord("\"")
			case '\'':
				ty = TT_STRING
				val = parseWord("'")
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				ty = TT_NUMBER
				val = parseNumber()
			default:
				ty = TT_WORD
				val = parseWord("")
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
	ToString() string
}

type ListNode struct {
	Items []Node
}

func (self ListNode) ToString() string {
	str := "("
	for _, item := range self.Items {
		str += item.ToString() + ","
	}
	str = str[0 : len(str)-1]
	return str + ")"
}

type StringNode struct {
	Value string
}

func (self StringNode) ToString() string {
	newVal, err := json.Marshal(self.Value)
	if err != nil {
		panic("Could not string into json string")
	}
	return string(newVal)
}

type NumberNode struct {
	Value string
}

func (self NumberNode) ToString() string {
	return self.Value
}

type PlainWordNode struct {
	Value string
}

func (self PlainWordNode) ToString() string {
	return strings.ReplaceAll(self.Value, ";", "")
}

type OperatorNode struct {
	Operator TT
	Negate   bool
}

func (self OperatorNode) ToString() string {
	negatedOps := map[string]string{
		"==": "!=",
		"<=": ">",
		">=": "<",
		"~":  "!~",
		"&":  "|",
	}

	for k, v := range negatedOps {
		negatedOps[v] = k
	}

	strOp := ""
	switch self.Operator {
	case TT_EQ:
		strOp = "=="
	case TT_LT:
		strOp = "<="
	case TT_LE:
		strOp = "<"
	case TT_GT:
		strOp = ">="
	case TT_GE:
		strOp = ">"
	case TT_SIMILAR:
		strOp = "~"
	case TT_AND:
		strOp = "&"
	default:
		panic("unknown operator")
	}

	if self.Negate {
		return negatedOps[strOp]
	}
	return strOp
}

type BinOpNode struct {
	Left     Node
	Right    Node
	Operator OperatorNode
}

func (self BinOpNode) ToString() string {
	left := self.Left.ToString()
	right := self.Right.ToString()

	switch self.Operator.Operator {
	case TT_IN:
		return "(" + left + " IN " + right + ")"
	case TT_AND:
		return "(" + left + " AND " + right + ")"
	case TT_OR:
		return "(" + left + " OR " + right + ")"
	case TT_DIV:
		return "(" + left + " / " + right + ")"
	case TT_MUL:
		return "(" + left + " * " + right + ")"
	case TT_SUB:
		return "(" + left + " - " + right + ")"
	case TT_PLUS:
		return "(" + left + " + " + right + ")"
	case TT_SIMILAR:
		return "(" + left + " LIKE " + right + ")"
	case TT_GE:
		return "(" + left + " >= " + right + ")"
	case TT_LE:
		return "(" + left + " <= " + right + ")"
	case TT_GT:
		return "(" + left + " > " + right + ")"
	case TT_LT:
		return "(" + left + " < " + right + ")"
	default:
		panic(fmt.Sprintf("Unimplemented operator: %+v", self.Operator))
	}
}

func Parse(tokens []Token) (string, error) {
	i := -1

	var search func() (Node, error)
	var atom func() Node
	var comparison func() Node

	next := func() bool {
		i++
		return i < len(tokens)
	}

	back := func() {
		i--
	}

	atEnd := func() bool {
		return i >= len(tokens)
	}

	atom = func() Node {
		switch tokens[i].Ty {
		case TT_STRING:
			return StringNode{
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
			n, err := search()
			if err != nil {
				return StringNode{}
			}
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
				// the user put a trailing colon, they probably want a list with 1 irem
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
		fallback := BinOpNode{
			Left: PlainWordNode{
				Value: "en_title",
			},
			Operator: OperatorNode{
				Operator: TT_SIMILAR,
				Negate:   false,
			},
			Right: StringNode{
				Value: tokens[i].Value,
			},
		}

		next()
		if atEnd() {
			return fallback
		}
		back()

		for next() {
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
					Negate:   false,
				},
				Right: right,
			}
		}
		return left
	}

	search = func() (Node, error) {
		next()
		logicToks := []TT{TT_AND, TT_OR}
		left := comparison()
		for next() {
			if !slices.Contains(logicToks, tokens[i].Ty) {
				continue
			}
			op := tokens[i]
			if !next() {
				return StringNode{}, fmt.Errorf("Expected expr after %s", op.Value)
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
		return left, nil
	}

	sn, err := search()
	if err != nil {
		return "", err
	}

	return sn.ToString(), nil
}

func Search2String(search string) string {
	tokens := Lex(search)
	tree, err := Parse(tokens)
	if err != nil {
		panic(err.Error())
	}
	return tree
}

// func (self BinOpNode) ToString() string {
// 	return self.Left.ToString() + self.Operator.ToString() + self.Right.ToString()
// }
//
// func parseExpr(tokens []Token, i int) (expr Expr, err error) {
// 	return nil, err
// }
