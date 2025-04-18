package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf8"

	"aiolimas/accounts"
	"aiolimas/db"
	"aiolimas/metadata"
	"aiolimas/types"
)

type (
	Parser      = func(uid int64, in string) (any, error)
	QueryParams = map[string]QueryParamInfo
)

type QueryParamInfo struct {
	Parser   Parser
	Required bool
}

type RequestContext struct {
	Uid        int64
	Req        *http.Request
	W          http.ResponseWriter
	Authorized bool
	PP         ParsedParams
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
	DELETE Method = "DELETE"
)

type ApiEndPoint struct {
	Handler        func(ctx RequestContext)
	EndPoint       string
	QueryParams    QueryParams
	Method         Method
	Description    string
	Returns        string
	PossibleErrors []string
	// whether or not auth is required, false == auth required, true == auth not required
	// it's named this way, so that by default, auth is intuitively required
	// because by default this will be false
	GuestAllowed bool

	// whether or not a user id is a required parameter
	UserIndependant bool
}

func (self *ApiEndPoint) GenerateDocHTML(root string) string {
	return fmt.Sprintf(`
		<div>
			<h2>%s/%s</h2>
				<h3>Description</h3>
					<p>%s</p>
				<h3>Returns</h3>
					<p>%s</p>
		</div>
	`, root, self.EndPoint, self.Description, self.Returns)
}


func ckAuthorizationHeader(text string) (string, error) {
	var estring string

	if b64L := strings.SplitN(text, "Basic ", 2); len(b64L) > 1 {
		b64 := b64L[1]
		info, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			estring = "You're bad at encoding base64 😀\n"
			println(err.Error())
			goto unauthorized
		}
		username, password, found := strings.Cut(string(info), ":")
		if !found {
			estring = "Invalid credentials\n"
			goto unauthorized
		}

		uid, err := accounts.CkLogin(username, password)

		if err != nil {
			println(err.Error())
			return "", err
		}

		if uid == "" {
			return "", err
		}

		return uid, nil
	}

unauthorized:
	return "", errors.New(estring)
}

func (self *ApiEndPoint) Listener(w http.ResponseWriter, req *http.Request) {
	parsedParams := ParsedParams{}

	ctx := RequestContext{}
	ctx.Req = req
	ctx.W = w

	method := self.Method
	if method == "" {
		method = "GET"
	}

	//"preflight" request
	if req.Method == "OPTIONS" {
		w.WriteHeader(200)
		return
	}

	if req.Method != string(method){
		w.WriteHeader(405)
		fmt.Fprintf(w, "Invalid method: '%s', expected: %s", req.Method, method)
		return
	}

	var uidStr = req.URL.Query().Get("uid")

	authorized := true

	privateAll := os.Getenv("AIO_PRIVATE")
	if !self.GuestAllowed || privateAll != "" {
		auth := req.Header.Get("Authorization")

		if auth == "" {
			authorized = false
		}

		newUid, err := ckAuthorizationHeader(auth)

		if newUid == "" || err != nil {
			authorized = false
		}

		if newUid != uidStr && uidStr != "" {
			//mismatch between requested uid, and the account that the user is logged into
			//this is important because the user is trying to perform a priveledged action for UID while not logged into that UID
			authorized = false
		}

		uidStr = newUid
	}

	if !authorized {
		w.Header().Add("WWW-Authenticate", "Basic realm=\"/\"")
		w.WriteHeader(401)
		return
	}

	query := req.URL.Query()

	uidInt := int64(-1)

	// uid is required on almost all endpoints, so
	// have UserIndependant that keeps track of if it's **not** required
	// this also saves the hassle of having to say it is required in QueryParams
	if !self.UserIndependant && uidStr == "" {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Missing parameter: 'uid'")
		return
	} else if !self.UserIndependant {
		var err error
		uidInt, err = strconv.ParseInt(uidStr, 10, 64)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Invalid user id, %s", err.Error())
			return
		}
	}

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

		val, err := info.Parser(uidInt, queryVal)
		if err != nil {
			w.WriteHeader(400)
			funcName := runtime.FuncForPC(reflect.ValueOf(info.Parser).Pointer()).Name()
			fmt.Fprintf(w, "%s\nInvalid value for: '%s'\nexpected to pass: '%s'", err.Error(), name, funcName)
			return
		}

		parsedParams[name] = val
	}

	ctx.PP = parsedParams
	ctx.Uid = uidInt

	self.Handler(ctx)
}

func P_Int64(uid int64, in string) (any, error) {
	i, err := strconv.ParseInt(in, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func P_Float64(uid int64, in string) (any, error) {
	f, err := strconv.ParseFloat(in, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}

func P_VerifyIdAndGetUserEntry(uid int64, id string) (any, error) {
	var out db_types.UserViewingEntry
	i, err := P_Int64(uid, id)
	if err != nil {
		return out, err
	}
	entry, err := db.GetUserViewEntryById(uid, i.(int64))
	if err != nil {
		return out, err
	}
	return entry, nil
}

func P_VerifyIdAndGetInfoEntry(uid int64, id string) (any, error) {
	var out db_types.InfoEntry
	i, err := P_Int64(uid, id)
	if err != nil {
		return out, err
	}
	entry, err := db.GetInfoEntryById(uid, i.(int64))
	if err != nil {
		return out, err
	}
	return entry, nil
}

func P_VerifyIdAndGetMetaEntry(uid int64, id string) (any, error) {
	var out db_types.MetadataEntry
	i, err := P_Int64(uid, id)
	if err != nil {
		return out, err
	}
	entry, err := db.GetMetadataEntryById(uid, i.(int64))
	if err != nil {
		return out, err
	}
	return entry, nil
}

func P_True(uid int64, in string) (any, error) {
	if !utf8.ValidString(in) {
		return in, errors.New("Invalid utf8")
	}
	return in, nil
}

func P_NotEmpty(uid int64, in string) (any, error) {
	if !utf8.ValidString(in) {
		return in, errors.New("Invalid utf8")
	}
	if in != "" {
		return in, nil
	}
	return in, errors.New("Empty")
}

func P_SqlSafe(uid int64, in string) (any, error) {
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

func P_EntryFormat(uid int64, in string) (any, error) {
	i, err := P_Int64(uid, in)
	if err != nil {
		return 0, err
	}
	if !db_types.IsValidFormat(i.(int64)) {
		return 0, fmt.Errorf("Invalid format '%s'", in)
	}
	return db_types.Format(i.(int64)), nil
}

func P_EntryType(uid int64, in string) (any, error) {
	if db_types.IsValidType(in) {
		return db_types.MediaTypes(in), nil
	}
	return db_types.MediaTypes("Show"), fmt.Errorf("Invalid entry type: '%s'", in)
}

func P_ArtStyle(uid int64, in string) (any, error) {
	val, err := strconv.ParseUint(in, 10, 64)
	if err != nil {
		return uint(0), err
	}
	return uint(val), nil
}

func P_MetaProvider(uid int64, in string) (any, error) {
	if metadata.IsValidProvider(in) {
		return in, nil
	}
	return in, fmt.Errorf("Invalid metadata provider: '%s'", in)
}

func P_UserStatus(uid int64, in string) (any, error) {
	if db_types.IsValidStatus(in) {
		return db_types.Status(in), nil
	}
	return "Planned", fmt.Errorf("Invalid user status: '%s'", in)
}

func P_TList[T any](sep string, toT func(in string) T) func(int64, string) (any, error) {
	return func(uid int64, in string) (any, error) {
		var arr []T
		items := strings.Split(in, sep)
		for _, i := range items {
			arr = append(arr, toT(i))
		}
		return arr, nil
	}
}

func P_Uint64Array(uid int64, in string) (any, error) {
	var arr []uint64
	err := json.Unmarshal([]byte(in), &arr)
	if err != nil {
		return arr, err
	}
	return arr, nil
}

func P_Bool(uid int64, in string) (any, error) {
	if in == "true" || in == "on" {
		return true, nil
	}
	if in == "" || in == "false" {
		return false, nil
	}
	return false, fmt.Errorf("Not a boolean: '%s'", in)
}

func P_IdIdentifier(uid int64, in string) (any, error) {
	if metadata.IsValidIdIdentifier(in) {
		return in, nil
	}
	return "", fmt.Errorf("Invalid id identifier: '%s'", in)
}

func P_Identifier(uid int64, in string) (any, error) {
	if metadata.IsValidIdentifier(in) {
		return in, nil
	}
	return "", fmt.Errorf("Invalid identifier: '%s'", in)
}

func As_JsonMarshal(parser Parser) Parser {
	return func(uid int64, in string) (any, error) {
		v, err := parser(uid, in)
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
