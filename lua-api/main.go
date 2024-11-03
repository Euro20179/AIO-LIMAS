package lua_api

import (
	"strconv"
	"strings"

	db_types "aiolimas/types"

	"github.com/yuin/gopher-lua"
)

func Fullsetup(state *lua.LState) {
	aioTble := state.NewTable()

	// functions {{{
	hasPrefix := state.NewFunction(func(s *lua.LState) int {
		testStr := s.ToString(1)
		prefix := s.ToString(2)
		s.Push(lua.LBool(strings.HasPrefix(testStr, prefix)))
		return 1
	})
	state.SetField(aioTble, "hasprefix", hasPrefix)

	// split := state.NewFunction(func(s *lua.LState) int {
	// 	str := s.ToString(1)
	// 	sep := s.ToString(2)
	// 	tbl := s.NewTable()
	// 	for i, v := range strings.Split(str, sep) {
	// 		state.SetTable(tbl, lua.LNumber(i), lua.LString(v))
	// 	}
	// 	s.Push(tbl)
	// 	return 1
	// })
	// state.SetField(aioTble, "split", split)

	title := state.NewFunction(func(s *lua.LState) int {
		str := s.ToString(1)
		s.Push(lua.LString(strings.ToTitle(string(str[0])) + str[1:]))
		return 1
	})
	state.SetField(aioTble, "title", title)
	// }}}

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
