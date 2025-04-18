package ui

import (
	"fmt"
	"os"

	lua "github.com/yuin/gopher-lua"
)

// UI is the Lua engine for frontend.
type UI struct {
	L *lua.LState
}

// NewUI creates a new Lua VM and registers APIs.
func NewUI() (*UI, error) {
	L := lua.NewState()
	ui := &UI{L}
	ui.registerAPIs()
	return ui, nil
}

func (ui *UI) registerAPIs() {
	L := ui.L
	L.SetGlobal("send_command", L.NewFunction(ui.luaSend))
	L.SetGlobal("show_modal", L.NewFunction(ui.luaModal))
	L.SetGlobal("update_chat", L.NewFunction(ui.luaChat))
	L.SetGlobal("clear_chat", L.NewFunction(ui.luaClear))
	L.SetGlobal("set_title", L.NewFunction(ui.luaTitle))
}

func (ui *UI) LoadScripts(files ...string) error {
	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("lua file missing: %s", f)
		}
		if err := ui.L.DoFile(f); err != nil {
			return fmt.Errorf("lua error in %s: %w", f, err)
		}
	}
	return nil
}

func (ui *UI) Close() { ui.L.Close() }

// Stubs: overridden by handlers
func (ui *UI) luaSend(L *lua.LState) int    { return 0 }
func (ui *UI) luaModal(L *lua.LState) int   { return 0 }
func (ui *UI) luaChat(L *lua.LState) int    { return 0 }
func (ui *UI) luaClear(L *lua.LState) int   { return 0 }
func (ui *UI) luaTitle(L *lua.LState) int   { return 0 }
