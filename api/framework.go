package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"aiolimas/db"
	"aiolimas/metadata"
)

type (
	Parser      = func(in string) (any, error)
	QueryParams = map[string]QueryParamInfo
)

type QueryParamInfo struct {
	Parser   Parser
	Required bool
}

func MkQueryInfo(parser Parser, required bool) QueryParamInfo {
	return QueryParamInfo{
		Parser: parser, Required: required,
	}
}

type ParsedParams map[string]any

func (self *ParsedParams) Get(name string, backup any) any {
	if v, exists := (*self)[name]; exists {
		return v
	}
	return backup
}

type Method string

const (
	GET  Method = "GET"
	POST Method = "POST"
)

type ApiEndPoint struct {
	Handler        func(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams)
	EndPoint       string
	QueryParams    QueryParams
	Method         Method
	Description    string
	Returns        string
	PossibleErrors []string
}

func (self *ApiEndPoint) GenerateDocHTML() string {
	htStr := "<div>"
	htStr += fmt.Sprintf("<h2>/%s</h2>", self.EndPoint)
	htStr += fmt.Sprintf("<h3>Description</h3><p>%s</p>", self.Description)
	htStr += fmt.Sprintf("<h3>Returns</h3><p>%s</p>", self.Returns)
	return htStr + "</div>"
}

func (self *ApiEndPoint) Listener(w http.ResponseWriter, req *http.Request) {
	parsedParams := ParsedParams{}

	method := self.Method
	if method == "" {
		method = "GET"
	}

	if req.Method != string(method) {
		w.WriteHeader(401)
		fmt.Fprintf(w, "Invalid method: %s", method)
		return
	}

	query := req.URL.Query()
	for name, info := range self.QueryParams {
		if !query.Has(name) {
			if info.Required {
				w.WriteHeader(400)
				fmt.Fprintf(w, "Missing parameter: '%s'", name)
				return
			}
			continue
		}

		queryVal := query.Get(name)

		val, err := info.Parser(queryVal)
		if err != nil {
			w.WriteHeader(400)
			funcName := runtime.FuncForPC(reflect.ValueOf(info.Parser).Pointer()).Name()
			fmt.Fprintf(w, "%s\nInvalid value for: '%s'\nexpected to pass: '%s'", err.Error(), name, funcName)
			return
		}

		parsedParams[name] = val
	}

	self.Handler(w, req, parsedParams)
}

func P_Int64(in string) (any, error) {
	i, err := strconv.ParseInt(in, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func P_Float64(in string) (any, error) {
	f, err := strconv.ParseFloat(in, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}

func P_VerifyIdAndGetUserEntry(id string) (any, error) {
	var out db.UserViewingEntry
	i, err := P_Int64(id)
	if err != nil {
		return out, err
	}
	entry, err := db.GetUserViewEntryById(i.(int64))
	if err != nil {
		return out, err
	}
	return entry, nil
}

func P_VerifyIdAndGetInfoEntry(id string) (any, error) {
	var out db.InfoEntry
	i, err := P_Int64(id)
	if err != nil {
		return out, err
	}
	entry, err := db.GetInfoEntryById(i.(int64))
	if err != nil {
		return out, err
	}
	return entry, nil
}

func P_VerifyIdAndGetMetaEntry(id string) (any, error) {
	var out db.MetadataEntry
	i, err := P_Int64(id)
	if err != nil {
		return out, err
	}
	entry, err := db.GetMetadataEntryById(i.(int64))
	if err != nil {
		return out, err
	}
	return entry, nil
}

func P_True(in string) (any, error) {
	return in, nil
}

func P_NotEmpty(in string) (any, error) {
	if in != "" {
		return in, nil
	}
	return in, errors.New("Empty")
}

func P_SqlSafe(in string) (any, error) {
	if in == "" {
		return in, errors.New("Empty")
	}
	match, err := regexp.Match("[0-9A-Za-z-_\\.]", []byte(in))
	if err != nil {
		return "", err
	} else if match {
		return in, nil
	}
	return in, fmt.Errorf("'%s' contains invalid characters", in)
}

func P_EntryFormat(in string) (any, error) {
	i, err := P_Int64(in)
	if err != nil {
		return 0, err
	}
	if !db.IsValidFormat(i.(int64)) {
		return 0, fmt.Errorf("Invalid format '%s'", in)
	}
	return db.Format(i.(int64)), nil
}

func P_EntryType(in string) (any, error) {
	if db.IsValidType(in) {
		return db.MediaTypes(in), nil
	}
	return db.MediaTypes("Show"), fmt.Errorf("Invalid entry type: '%s'", in)
}

func P_ArtStyle(in string) (any, error) {
	val, err := strconv.ParseUint(in, 10, 64)
	if err != nil{
		return uint(0), err
	}
	return uint(val), nil
}

func P_MetaProvider(in string) (any, error) {
	if metadata.IsValidProvider(in) {
		return in, nil
	}
	return in, fmt.Errorf("Invalid metadata provider: '%s'", in)
}

func P_UserStatus(in string) (any, error) {
	if db.IsValidStatus(in) {
		return db.Status(in), nil
	}
	return "Planned", fmt.Errorf("Invalid user status: '%s'", in)
}

func P_TList[T any](sep string, toT func(in string) T) func(string) (any, error){
	return func(in string) (any, error) {
		var arr []T
		items := strings.Split(in, sep)
		for _, i := range items {
			arr = append(arr, toT(i))
		}
		return arr, nil
	}
}

func P_Uint64Array(in string) (any, error) {
	var arr []uint64
	err := json.Unmarshal([]byte(in), &arr)
	if err != nil {
		return arr, err
	}
	return arr, nil
}

func P_Bool(in string) (any, error) {
	if in == "true" || in == "on" {
		return true, nil
	}
	if in == "" || in == "false" {
		return false, nil
	}
	return false, fmt.Errorf("Not a boolean: '%s'", in)
}

func P_IdIdentifier(in string) (any, error) {
	if metadata.IsValidIdIdentifier(in) {
		return in, nil
	}
	return "", fmt.Errorf("Invalid id identifier: '%s'", in)
}

func P_Identifier(in string) (any, error) {
	if metadata.IsValidIdentifier(in) {
		return in, nil
	}
	return "", fmt.Errorf("Invalid identifier: '%s'", in)
}

func As_JsonMarshal(parser Parser) Parser {
	return func(in string) (any, error) {
		v, err := parser(in)
		if err != nil {
			return "", err
		}
		res, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(res), nil
	}
}
