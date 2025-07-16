package lua_api

import (
	"os"
	"strconv"
	"strings"

	"aiolimas/accounts"
	"aiolimas/logging"
	db_types "aiolimas/types"
	globals "aiolimas/util"

	"github.com/yuin/gopher-lua"
)

func Fullsetup(state *lua.LState) {
	aioTble := state.NewTable()

	// helper funcs {{{
	hasPrefix := state.NewFunction(func(s *lua.LState) int {
		testStr := s.ToString(1)
		prefix := s.ToString(2)
		s.Push(lua.LBool(strings.HasPrefix(testStr, prefix)))
		return 1
	})
	state.SetField(aioTble, "hasprefix", hasPrefix)

	title := state.NewFunction(func(s *lua.LState) int {
		str := s.ToString(1)
		if len(str) == 0 {
			s.Push(lua.LString(""))
		} else {
			s.Push(lua.LString(strings.ToTitle(string(str[0])) + str[1:]))
		}
		return 1
	})
	state.SetField(aioTble, "title", title)

	split := state.NewFunction(func(s *lua.LState) int {
		str := s.ToString(1)
		sep := s.ToString(2)
		tbl := lua.LTable{}
		for _, item := range strings.Split(str, sep) {
			tbl.Append(lua.LString(item))
		}

		s.Push(&tbl)
		return 1
	})
	state.SetField(aioTble, "split", split)

	bor := state.NewFunction(func(s *lua.LState) int {
		n1 := s.ToInt64(1)
		n2 := s.ToInt64(2)
		s.Push(lua.LNumber(n1 | n2))
		return 1
	})
	state.SetField(aioTble, "bor", bor)
	// }}}

	// event funcs {{{
	listen := state.NewFunction(func(s *lua.LState) int {
		eventName := s.ToString(1)
		onEvent := s.ToFunction(2)

		globals.LuaEventRegistry[eventName] = append(globals.LuaEventRegistry[eventName], onEvent)
		return 0
	})
	state.SetField(aioTble, "listen", listen)
	//}}}

	// api funcs {{{
	accID := state.NewFunction(func(s *lua.LState) int {
		username := s.ToString(1)

		id, err := accounts.Username2Id(os.Getenv("AIO_PATH"), username)

		if err != nil {
			logging.ELog(err)
			s.Push(lua.LNumber(1))
			return 1
		}

		s.Push(lua.LNumber(id))
		return 1
	})
	state.SetField(aioTble, "username2id", accID)
	// }}}

	// types{{{
	artStylesTble := state.NewTable()

	for k, v := range db_types.ListArtStyles() {
		state.Push(lua.LString(v))
		state.SetField(artStylesTble, strconv.Itoa(int(k)), lua.LString(v))
	}
	state.SetField(aioTble, "artStyles", artStylesTble)

	typesTbl := state.NewTable()
	for _, v := range db_types.ListMediaTypes() {
		typesTbl.Append(lua.LString(v))
	}
	state.SetField(aioTble, "types", typesTbl)
	// }}}


	state.SetGlobal("aio", aioTble)
}

func ReloadLuaInstance(fromFile string) (*lua.LState, error) {
	globals.LuaEventRegistry = map[string][]*lua.LFunction{}

	return InitGlobalLuaInstance(fromFile)
}

func InitGlobalLuaInstance(fromFile string) (*lua.LState, error) {
	l := lua.NewState()

	Fullsetup(l)

	err := l.DoFile(fromFile)
	if err != nil {
		logging.ELog(err)
		return l, err
	}

	return l, nil
}
