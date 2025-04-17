package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/PranavKharche24/chat-app/internal/encryption"
)

const (
	historyDir = "history"
)

var (
	conn          net.Conn
	app           *tview.Application
	pages         *tview.Pages
	welcomeHeader *tview.TextView
	chatView      *tview.TextView
	partner       string
)

// marquee animates the header.
func marquee(txt string, tv *tview.TextView) {
	pos := 0
	for {
		app.QueueUpdateDraw(func() {
			tv.SetText(txt[pos:]+txt[:pos])
		})
		pos = (pos + 1) % len(txt)
		time.Sleep(200 * time.Millisecond)
	}
}

// welcomeScreen with flashy UI.
func welcomeScreen() tview.Primitive {
	welcomeHeader = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorYellow)
	go marquee("✨  WELCOME TO SUPER‑CHAT ✨   ", welcomeHeader)

	list := tview.NewList().
		AddItem("Login", "", 'l', func() { pages.SwitchToPage("login") }).
		AddItem("Register", "", 'r', func() { pages.SwitchToPage("register") }).
		AddItem("Quit", "", 'q', func() { app.Stop() })

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(welcomeHeader, 3, 1, false).
		AddItem(list, 0, 2, true)
}

// loginForm page.
func loginForm() tview.Primitive {
	f := tview.NewForm()
	f.AddInputField("UserID", "", 10, nil, nil)
	f.AddPasswordField("Password", "", 20, '*', nil)
	f.AddButton("Login", func() {
		uid := f.GetFormItemByLabel("UserID").(*tview.InputField).GetText()
		pw := f.GetFormItemByLabel("Password").(*tview.InputField).GetText()
		if uid == "" || pw == "" {
			modal("ID & Password required", false)
			return
		}
		fmt.Fprintf(conn, "LOGIN %s %s\n", uid, pw)
	})
	f.AddButton("Back", func() {
		pages.SwitchToPage("welcome")
	})

	f.SetBorder(true).SetTitle(" Login ")
	return f
}

// registerForm page.
func registerForm() tview.Primitive {
	f := tview.NewForm()
	f.AddInputField("Username", "", 20, nil, nil)
	f.AddInputField("Email", "", 30, nil, nil)
	f.AddInputField("DOB (YYYY-MM-DD)", "", 10, nil, nil)
	f.AddInputField("Full Name", "", 30, nil, nil)
	f.AddPasswordField("Password", "", 20, '*', nil)
	f.AddButton("Register", func() {
		u := f.GetFormItemByLabel("Username").(*tview.InputField).GetText()
		e := f.GetFormItemByLabel("Email").(*tview.InputField).GetText()
		d := f.GetFormItemByLabel("DOB (YYYY-MM-DD)").(*tview.InputField).GetText()
		n := f.GetFormItemByLabel("Full Name").(*tview.InputField).GetText()
		p := f.GetFormItemByLabel("Password").(*tview.InputField).GetText()
		if u == "" || e == "" || d == "" || n == "" || p == "" {
			modal("All fields required", false)
			return
		}
		fmt.Fprintf(conn, "REGISTER %s %s %s %s %s\n", u, e, d, n, p)
	})
	f.AddButton("Back", func() {
		pages.SwitchToPage("welcome")
	})

	f.SetBorder(true).SetTitle(" Register ")
	return f
}

// chatScreen page.
func chatScreen() tview.Primitive {
	chatView = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	chatView.SetBorder(true).SetTitle(" Chat with: "+partner)

	partnerInput := tview.NewInputField().
		SetLabel("Partner username: ").
		SetFieldWidth(20)
	partnerInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			p := partnerInput.GetText()
			if p != "" {
				partner = p
				chatView.Clear()
				loadHistory(p)
				chatView.SetTitle(" Chat with: " + p)
			}
		}
	})

	msgInput := tview.NewInputField().
		SetLabel("Message: ").
		SetFieldWidth(0)
	msgInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter && partner != "" {
			text := msgInput.GetText()
			msgInput.SetText("")
			fmt.Fprintf(conn, "SEND %s %s\n", partner, text)
		}
	})

	status := tview.NewTextView().
		SetText("[F2] Local History   [F3] Quit").
		SetTextAlign(tview.AlignCenter)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(partnerInput, 1, 0, true).
		AddItem(chatView, 0, 1, false).
		AddItem(msgInput, 1, 0, true).
		AddItem(status, 1, 0, false)

	// Key bindings
	flex.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		switch e.Key() {
		case tcell.KeyF2:
			showHistory(partner)
			return nil
		case tcell.KeyF3:
			app.Stop()
			return nil
		}
		return e
	})

	return flex
}

// modal shows a message; if toChat, switch to chat on OK.
func modal(text string, toChat bool) {
	m := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(_ int, _ string) {
			pages.RemovePage("modal")
			if toChat {
				pages.SwitchToPage("chat")
			}
		})
	pages.AddPage("modal", m, true, true)
}

// showHistory loads and displays local history for a partner.
func showHistory(user string) {
	data, err := ioutil.ReadFile(historyDir + "/" + user + ".txt")
	if err != nil || len(data) == 0 {
		data = []byte("No local history.")
	}
	m := tview.NewModal().
		SetText(string(data)).
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(_ int, _ string) {
			pages.RemovePage("hist")
		})
	pages.AddPage("hist", m, true, true)
}

// loadHistory populates chatView from local file.
func loadHistory(user string) {
	file := historyDir + "/" + user + ".txt"
	data, _ := ioutil.ReadFile(file)
	fmt.Fprint(chatView, string(data))
}

// storeLocal appends a message to the user’s local history.
func storeLocal(user, msg string) {
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return
	}
	f, err := os.OpenFile(historyDir+"/"+user+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		fmt.Fprintln(f, msg)
		f.Close()
	}
}

// reader pumps server messages into UI.
func reader() {
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "REGISTERED"):
			modal(line, false)
		case line == "LOGIN OK":
			modal("Login successful!", true)
		case line == "SENT":
			modal("Sent!", false)
		case strings.HasPrefix(line, "ERROR"):
			modal(line, false)
		case strings.HasPrefix(line, "CHAT:"):
			enc := strings.TrimPrefix(line, "CHAT:")
			dec, err := encryption.Decrypt(enc)
			if err != nil {
				dec = "[decrypt error]"
			}
			// show + store
			app.QueueUpdateDraw(func() {
				fmt.Fprintln(chatView, dec)
			})
			storeLocal(partner, dec)
		default:
			// ignore or show generic
		}
	}
}

func main() {
	var err error
	conn, err = net.Dial("tcp", "localhost:9000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	app = tview.NewApplication()
	pages = tview.NewPages()

	pages.AddPage("welcome", welcomeScreen(), true, true)
	pages.AddPage("login", loginForm(), true, false)
	pages.AddPage("register", registerForm(), true, false)
	pages.AddPage("chat", chatScreen(), true, false)

	go reader()

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		log.Fatal(err)
	}
}
