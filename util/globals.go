package util

import lua "github.com/yuin/gopher-lua"

var GlobalLuaInstance *lua.LState

var LuaEventRegistry = map[string][]*lua.LFunction{}
