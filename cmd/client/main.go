package main

import (
	"log"

	"github.com/yuin/gopher-lua"
	"github.com/PranavKharche24/chat-app/internal/networking"
	"github.com/PranavKharche24/chat-app/internal/ui"
)

func main() {
	// 1) Connect to the Go chat server
	client, err := networking.NewClient("localhost:9000")
	if err != nil {
		log.Fatalf("network error: %v", err)
	}
	defer client.Close()
	client.Receive() // start reading incoming (and decrypted) messages

	// 2) Initialize the Lua UI engine
	uiEngine, err := ui.NewUI()
	if err != nil {
		log.Fatalf("failed to init UI: %v", err)
	}
	defer uiEngine.Close()

	// 3) Attach Go → Lua callbacks
	//    Implement ui.Handler interface in your TUI code to call client.Send(...) etc.
	//    Here we use a simple handler that just logs—we’ll stub it out for you:
	handler := ui.NewStubHandler(client)
	uiEngine.AttachHandlers(handler)

	// 4) Load all Lua scripts (config, keymaps, widgets, actions)
	if err := uiEngine.LoadScripts(
		"lua/config.lua",
		"lua/keymaps.lua",
		"lua/widgets.lua",
		"lua/actions.lua",
	); err != nil {
		log.Fatalf("failed loading Lua scripts: %v", err)
	}

	// 5) Call the Lua entrypoint to build the Welcome/Login screen:
	if err := uiEngine.L.CallByParam(lua.P{
		Fn:      uiEngine.L.GetGlobal("build_welcome_screen"),
		NRet:    0,
		Protect: true,
	}); err != nil {
		log.Fatalf("Lua error: %v", err)
	}

	// 6) Start your TUI main loop here (e.g. with tview or a custom loop).
	//    This is application-specific—hook into uiEngine and client.Incoming to update UI.
	select {}
}
