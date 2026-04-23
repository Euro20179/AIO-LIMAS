package search

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	db_types "aiolimas/types"
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

func runeAt(text []rune, pos int) rune {
	return text[pos]
}

func Lex(search []rune) []Token {
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

		final := string(runeAt(search, i))
		for next() {
			ch := runeAt(search, i)
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
			final = string(runeAt(search, i))
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

			if ch == '\'' {
				final += "''"
			} else {
				final += string(ch)
			}
			escape = false
		}

		return final
	}

	parseBrace := func() string {
		final := ""
		braceCount := 1
		for next() {
			ch := runeAt(search, i)
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
			ch := runeAt(search, i)

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
				if len(search) > 1 && runeAt(search, i+1) == '=' {
					next()
					ty = TT_EQ
					val = "=="
				} else {
					ty = TT_EQ
					val = "="
				}
			case '>':
				if len(search) > 1 && runeAt(search, i+1) == '=' {
					next()
					ty = TT_GE
					val = ">="
				} else {
					ty = TT_GT
					val = ">"
				}
			case '<':
				if len(search) > 1 && runeAt(search, i+1) == '=' {
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
			case '@':
				fallthrough
			case '#':
				ty = TT_MACRO
				// parseWord includes the first char, ignore it
				val = parseWord("", []rune{'\'', '|', '&', '"', ')', '('})[1:]
			default:
				ty = TT_WORD
				val = parseWord("", []rune{})
				if strings.ToLower(val) == "and" {
					ty = TT_AND
				} else if strings.ToLower(val) == "or" {
					ty = TT_OR
				}
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
	// onExpand, exists := globals.LuaEventRegistry["MacroExpand"]
	// if !exists {
	// 	return "", errors.New("Could not expand macro")
	// }
	//
	// curMacro := self.Value
	//
	// // TODO reimplement the macro expansion in go
	// // there isn't really a reason for it to be in lua
	// // the user-config for lua never really manifested
	//
	// state := globals.GlobalLuaInstance
	//
	// for _, expansion := range onExpand {
	// 	state.Push(expansion)
	// 	state.Push(lua.LString(curMacro))
	// 	state.Call(1, 2)
	// 	userErr := state.Get(-1).(lua.LString)
	// 	if userErr != "" {
	// 		return curMacro, errors.New(string(userErr))
	// 	}
	//
	// 	text := state.Get(-2).(lua.LString)
	// 	if text != "" {
	// 		curMacro = string(text)
	// 	}
	// }

	macro := self.Value

	comp := func(left string, right string) string {
		return left + "==" + right
	}

	parseDateParams := func(paramString string, startOrEnd string) string {
		month := 1
		if startOrEnd != "start" {
			month = 12
		}

		timeC := map[string]int {
			"year": 2025,
			"month": month,
			"day": 1,
			"hour": 0,
			"minute": 0,
			"second": 0,
		}

		curKey := ""
		curVal := ""

		parsing := "key"

		fmt.Printf("date: %s\n", paramString)
		for _, ch := range paramString {
			if ch == ':' {
				parsing = "val"
				continue
			} else if ch == '/' {
				parsing = "key"

				full, ok := map[string]string {
					"y": "year",
					"m": "month",
					"d": "day",
					"H": "hour",
					"M": "minute",
					"S": "second",
				}[curKey]

				if !ok {
					_, ok = timeC[curKey]
					if ok {
						full = curKey
					} else {
						full = "year"
					}
				}

				i, err := strconv.ParseInt(curVal, 10, 64)
				if err == nil {
					timeC[full] = int(i)
					curKey = ""
					curVal = ""
				} else {
					return "0"
				}
				continue
			}

			if parsing == "key" {
				curKey = curKey + fmt.Sprintf("%c", ch)
			} else {
				curVal = curVal + fmt.Sprintf("%c", ch)
			}
		}

		t := time.Date(
			timeC["year"],
			time.Month(timeC["month"]),
			timeC["day"],
			timeC["hour"],
			timeC["minute"],
			timeC["second"],
			0,
			time.UTC,
		)
		fmt.Printf("%+v %d\n", timeC, t.UnixMilli())
		return fmt.Sprintf("%d", t.UnixMilli())
	}

	statuses := db_types.ListStatuses()
	mediaTypes := db_types.ListMediaTypes()
	formatIds := db_types.ListFormats()
	formats := map[string]db_types.Format {}
	for k, v := range formatIds {
		formats[v] = k
	}

	as := db_types.ListArtStyles()
	asName2I := map[string]db_types.ArtStyle{}
	for k, v := range as {
		asName2I[v] = k
	}

	prefixMacros := map[string] func(string) (string, error) {
		"s": func(macro string) (string, error){
			text := strings.Title(macro[2:])
			return comp("status", "\"" + text + "\""), nil
		},
		"t": func(macro string) (string, error) {
			return comp("type", "\"" + strings.Title(macro[2:]) + "\""), nil
		},
		"a": func(macro string) (string, error) {
			itemList := macro[2:]
			items := strings.Split(itemList, "+")
			query := ""
			for _, item := range items {
				titledArg := strings.Title(item)
				as_int, ok := asName2I[titledArg]
				if !ok {
					return "", errors.New("invalid art style " + titledArg)
				}

				if query != "" {
					query = query + fmt.Sprintf(
						"and (artStyle & %d == %d)",
						as_int,
						as_int,
					)
				} else {
					//extra ( because i want to encase the whole thing with ()
					query = fmt.Sprintf("((argStyle & %d == %d)",
						as_int,
						as_int,
					)
				}
			}
			return query + ")", nil
		},
		"f": func(macro string) (string, error) {
			reqFmt := strings.ToUpper(macro[2:])
			if strings.Contains(macro, "+d") {
				reqFmt = reqFmt[0:len(reqFmt) - 2]
				reqFmt = fmt.Sprintf("%d", formats[reqFmt] | db_types.F_MOD_DIGITAL)
				return comp("Format", reqFmt), nil
			}

			if strings.Contains(macro, "-d") {
				reqFmt = reqFmt[0:len(reqFmt) - 2]
				return comp("Format", fmt.Sprintf("%d", formats[reqFmt])), nil
			}

			return "(" + comp(
							"Format",
							fmt.Sprintf("%d", formats[reqFmt]),
						) + " or " +
						comp(
							"Format",
							fmt.Sprintf("%d", formats[reqFmt] | db_types.F_MOD_DIGITAL),
						) + ")", nil
		},

		"tag": func(macro string) (string, error) {
			tag := macro[4:]
			return "Collection LIKE ('%' || char(31) || '" + tag + "' || char(31) || '%')", nil
		},

		"md": func(macro string) (string, error) {
			name := macro[3:]
			return fmt.Sprintf("mediaDependant != '' and json_extract(mediaDependant, '$.%s')", name), nil
		},
		"mdi": func(macro string) (string, error) {
			name := macro[4:]
			return fmt.Sprintf("mediaDependant != '' AND CAST(jsoN_extract(mediaDependant, '$.%s') as decimal)", name), nil
		},

		"g": func(macro string) (string, error) {
			genre := macro[2:]
			return fmt.Sprintf("EXISTS (SELECT * FROM json_each(json_extract(genres, '$')) WHERE genres != '' AND json_each.value LIKE '%s')", genre), nil
		},
	}

	basicMacros := map[string]string{
		"isAnime": fmt.Sprintf("(artStyle & %d == %d)", db_types.AS_ANIME, db_types.AS_ANIME),
		"r":       "userRating",
		"R":       "rating",
		"t":       "en_title",
		"T":       "title",
		"d":       "description",
		"ts":      "timestamp",
		"y":       "releaseyear",
		"s:v":     comp("status", "\"Viewing\"") + " or " + comp("status", "\"ReViewing\""),
		"ep":      "CAST(json_extract(mediaDependant, format('$.%s-episodes', type)) as DECIMAL)",
		"len":     "CAST(json_extract(mediaDependant, format('$.%s-length', type)) as DECIMAL)",
		"epd":     "CAST(json_extract(mediaDependant, format('$.%s-episode-duration', type)) as DECIMAL)",
	}

	for _, item := range mediaTypes {
		basicMacros[strings.ToLower(string(item))] = "(type = '" + string(item) + "')"
	}

	for _, item := range statuses {
		basicMacros[strings.ToLower(string(item))] = "(status = '" + string(item) + "')"
	}

	e := strings.Index(macro, ":")
	prefix := ""
	if e != -1 {
		prefix = macro[:e]
	}

	if v, has := basicMacros[macro]; has {
		return v, nil
	} else if v, has := prefixMacros[prefix]; has {
		return v(macro)
	} else if len(macro) > 2 && macro[0] == '#' {
		text := macro[2:]
		return fmt.Sprintf(`(
			En_Title LIKE '%%%s%%' OR
				entryInfo.Native_Title LIKE '%%%s%%' OR
				Title LIKE '%%%s%%' OR
				metadata.Native_Title LIKE '%%%s%%')`,
			text, text, text, text), nil
	} else if len(macro) > 3 && macro[0:3] == "ev-" {
		time := macro[3:] + "/"
		d := parseDateParams(time, "start")
		return fmt.Sprintf(`
			((%s > timestamp AND timestamp > 0) OR
			(%s > after AND after > 0) OR
			(%s > beforeTS AND beforeTS > 0))
		`, d, d, d), nil
	} else if len(macro) > 3 && macro[0:3] == "ev+" {
		time := macro[3:] + "/"
		d := parseDateParams(time, "end")
		return fmt.Sprintf(`
			((%s < timestamp AND timestamp > 0) OR
			(%s < after AND after > 0) OR
			(%s < beforeTS AND beforeTS != 0))
		`, d, d, d), nil
	} else if len(macro) > 6 && macro[0:4] == "date" {
		beginOrEnd := "start"
		if macro[4] == '+' {
			beginOrEnd = "end"
		}

		time := macro[5:]
		if time == "" {
			return "false", nil
		}

		return parseDateParams(time, beginOrEnd), nil
	} else {
		return fmt.Sprintf("(en_title LIKE '%%%s%%')", macro[1:]), nil
	}
}

type StringNode struct {
	Value string
}

func (self StringNode) ToString() (string, error) {
	return "'" + self.Value + "'", nil
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
	tokens := Lex([]rune(search))
	return Parse(tokens)
}
