package ui

import (
	lua "github.com/yuin/gopher-lua"
)

// Handler defines Go callbacks for Lua UI events.
type Handler interface {
	SendCommand(cmd string)
	ShowModal(msg string)
	UpdateChat(msg string)
	ClearChat()
	SetTitle(title string)
}

// AttachHandlers binds a Handler’s methods into the Lua stubs.
func (ui *UI) AttachHandlers(h Handler) {
	L := ui.L
	L.SetGlobal("send_command", L.NewFunction(func(L *lua.LState) int {
		h.SendCommand(L.CheckString(1))
		return 0
	}))
	L.SetGlobal("show_modal", L.NewFunction(func(L *lua.LState) int {
		h.ShowModal(L.CheckString(1))
		return 0
	}))
	L.SetGlobal("update_chat", L.NewFunction(func(L *lua.LState) int {
		h.UpdateChat(L.CheckString(1))
		return 0
	}))
	L.SetGlobal("clear_chat", L.NewFunction(func(L *lua.LState) int {
		h.ClearChat()
		return 0
	}))
	L.SetGlobal("set_title", L.NewFunction(func(L *lua.LState) int {
		h.SetTitle(L.CheckString(1))
		return 0
	}))
}
