package lua_api

import (
	"strconv"
	"strings"

	db_types "aiolimas/types"

	"github.com/yuin/gopher-lua"
)

var GlobalLuaInstance *lua.LState

var LuaEventRegistry = map[string][]*lua.LFunction{}

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

	bor := state.NewFunction(func(s *lua.LState) int {
		n1 := s.ToInt64(1)
		n2 := s.ToInt64(2)
		println(n1, n2)
		s.Push(lua.LNumber(n1 | n2))
		return 1
	})
	state.SetField(aioTble, "bor", bor)
	// }}}

	// event funcs {{{
	listen := state.NewFunction(func(s *lua.LState) int {
		eventName := s.ToString(1)
		onEvent := s.ToFunction(2)

		LuaEventRegistry[eventName] = append(LuaEventRegistry[eventName], onEvent)
		return 0
	})
	state.SetField(aioTble, "listen", listen)
	//}}}

	// types{{{
	artStylesTble := state.NewTable()

	for k, v := range db_types.ListArtStyles() {
		state.Push(lua.LString(v))
		state.SetField(artStylesTble, strconv.Itoa(int(k)), lua.LString(v))
	}
	state.SetField(aioTble, "artStyles", artStylesTble)
	// }}}

	state.SetGlobal("aio", aioTble)
}

func ReloadLuaInstance(fromFile string) (*lua.LState, error) {
	LuaEventRegistry = map[string][]*lua.LFunction{}

	return InitGlobalLuaInstance(fromFile)
}

func InitGlobalLuaInstance(fromFile string) (*lua.LState, error) {
	l := lua.NewState()

	Fullsetup(l)

	err := l.DoFile(fromFile)
	if err != nil {
		println(err.Error())
		return l, err
	}

	return l, nil
}
