package api

import (
	"aiolimas/db"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
)

type Parser = func(in string) (any, error)
type QueryParams = map[string]QueryParamInfo

type QueryParamInfo struct {
	Parser Parser
	Required bool
}

func MkQueryInfo(parser Parser, required bool) QueryParamInfo {
	return QueryParamInfo{
		Parser: parser, Required: required,
	}
}

type ApiEndPoint struct {
	Handler func(w http.ResponseWriter, req *http.Request, parsedParams map[string]any)
	QueryParams QueryParams
}

func (self *ApiEndPoint) Listener(w http.ResponseWriter, req *http.Request) {
	parsedParams := map[string]any{}

	query := req.URL.Query()
	for name, info := range self.QueryParams {
		if !query.Has(name){
			if info.Required {
				w.WriteHeader(400)
				fmt.Fprintf(w, "Missing parameter: '%s'", name)
				return
			}
			continue
		}

		queryVal := query.Get(name)

		val, err := info.Parser(queryVal)
		if err != nil{
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
	if err != nil{
		return 0, err
	}
	return i, nil
}

func P_Float64(in string) (any, error) {
	f, err := strconv.ParseFloat(in, 64)
	if err != nil{
		return 0, err
	}
	return f, nil
}

func P_VerifyIdAndGetUserEntry(id string) (any, error) {
	var out db.UserViewingEntry
	i, err := P_Int64(id)
	if err != nil{
		return out, err
	}
	entry, err := db.GetUserViewEntryById(i.(int64))
	if err != nil{
		return out, err
	}
	return entry, nil
}

func P_VerifyIdAndGetInfoEntry(id string) (any, error) {
	var out db.InfoEntry
	i, err := P_Int64(id)
	if err != nil{
		return out, err
	}
	entry, err := db.GetInfoEntryById(i.(int64))
	if err != nil{
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

func P_EntryFormat(in string) (any, error) {
	i, err := P_Int64(in)
	if err != nil{
		return 0, err
	}
	if !db.IsValidFormat(i.(int64)) {
		return 0, fmt.Errorf("Invalid format '%s'", in)
	}
	return db.Format(i.(int64)), nil
}
