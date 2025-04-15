package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/PranavKharche24/chat-app/internal/encryption"
)

// Global variables
var conn net.Conn
var app *tview.Application
var pages *tview.Pages
var chatView *tview.TextView
var chatInput *tview.InputField
var currentUser string

// welcomeScreen creates the initial menu screen.
func welcomeScreen() tview.Primitive {
	header := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Welcome to ChatApp CLI\nPlease select an option below").
		SetTextColor(tcell.ColorGreen)

	menu := tview.NewList().
		AddItem("Login", "Existing users log in", 'l', func() {
			pages.SwitchToPage("login")
		}).
		AddItem("Register", "New user registration", 'r', func() {
			pages.SwitchToPage("register")
		}).
		AddItem("Quit", "Exit application", 'q', func() {
			app.Stop()
		})

	// Arrange header and menu vertically.
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 1, false).
		AddItem(menu, 0, 2, true)
	return flex
}

// loginForm creates a form for logging in.
func loginForm() tview.Primitive {
	form := tview.NewForm()
	form.AddInputField("Username", "", 20, nil, nil)
	form.AddPasswordField("Password", "", 20, '*', nil)
	form.AddButton("Login", func() {
		username := form.GetFormItemByLabel("Username").(*tview.InputField).GetText()
		password := form.GetFormItemByLabel("Password").(*tview.InputField).GetText()
		if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
			modalError("Both username and password are required.")
			return
		}
		// Send login command to server.
		fmt.Fprintf(conn, "LOGIN %s %s\n", username, password)
	})
	form.AddButton("Back", func() {
		pages.SwitchToPage("welcome")
	})
	form.SetBorder(true).SetTitle(" Login ").SetTitleAlign(tview.AlignCenter)
	return form
}

// registerForm creates a form for user registration.
func registerForm() tview.Primitive {
	form := tview.NewForm()
	form.AddInputField("Username", "", 20, nil, nil)
	form.AddPasswordField("Password", "", 20, '*', nil)
	form.AddButton("Register", func() {
		username := form.GetFormItemByLabel("Username").(*tview.InputField).GetText()
		password := form.GetFormItemByLabel("Password").(*tview.InputField).GetText()
		if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
			modalError("Both username and password are required.")
			return
		}
		// Send registration command to server.
		fmt.Fprintf(conn, "REGISTER %s %s\n", username, password)
	})
	form.AddButton("Back", func() {
		pages.SwitchToPage("welcome")
	})
	form.SetBorder(true).SetTitle(" Register ").SetTitleAlign(tview.AlignCenter)
	return form
}

// chatScreen creates the main chat interface.
func chatScreen() tview.Primitive {
	// Chat view: displays messages.
	chatView = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw() // Force redraw when new messages are added.
		})
	chatView.SetBorder(true).SetTitle(" Chat Room ")

	// Chat input: for sending messages.
	chatInput = tview.NewInputField().
		SetLabel("Enter message: ").
		SetFieldWidth(0)
	chatInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			msg := chatInput.GetText()
			chatInput.SetText("")
			if strings.TrimSpace(msg) != "" {
				// Send the message command.
				fmt.Fprintf(conn, "MSG %s\n", msg)
			}
		}
	})
	chatInput.SetBorder(true)

	// Status bar: provides quick hotkey instructions.
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[F2] History   [F3] Logout").
		SetTextAlign(tview.AlignCenter)

	// Arrange chat view, input field, and status bar in a vertical layout.
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(chatView, 0, 1, false).
		AddItem(chatInput, 3, 0, true).
		AddItem(statusBar, 1, 0, false)

	// Global key bindings for history (F2) and logout (F3).
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF2:
			// Send the HISTORY command to the server.
			fmt.Fprintf(conn, "HISTORY\n")
			return nil
		case tcell.KeyF3:
			// Logout: clear username, clear chat view, send QUIT, and show welcome screen.
			currentUser = ""
			chatView.SetText("")
			fmt.Fprintf(conn, "QUIT\n")
			pages.SwitchToPage("welcome")
			return nil
		}
		return event
	})

	return flex
}

// modalError displays an error modal with the specified message.
func modalError(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Ok"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("modal")
		})
	pages.AddPage("modal", modal, true, true)
}

// readFromServer continuously reads from the TCP connection and updates the UI.
func readFromServer() {
	reader := bufio.NewScanner(conn)
	for reader.Scan() {
		line := reader.Text()
		// Process broadcast chat messages (starting with "CHAT: ") or history entries (starting with "[").
		if strings.HasPrefix(line, "CHAT: ") || strings.HasPrefix(line, "[") {
			var decrypted string
			if strings.HasPrefix(line, "CHAT: ") {
				encryptedMsg := strings.TrimPrefix(line, "CHAT: ")
				var err error
				decrypted, err = encryption.Decrypt(encryptedMsg)
				if err != nil {
					decrypted = "[Error decrypting message]"
				}
			} else {
				// Assume history or plain text messages are sent as-is.
				decrypted = line
			}

			// If the message is in the expected history format "[username @ timestamp]: message"
			// then highlight the username portion in yellow.
			if strings.HasPrefix(decrypted, "[") {
				endIdx := strings.Index(decrypted, "]")
				if endIdx != -1 {
					usernamePart := decrypted[1:endIdx]
					// Format the message: username in yellow followed by the rest of the message.
					decrypted = fmt.Sprintf("[yellow]%s[white]%s", usernamePart, decrypted[endIdx:])
				}
			}

			app.QueueUpdateDraw(func() {
				fmt.Fprintf(chatView, "%s\n", decrypted)
			})
		} else {
			// Process informational messages (login confirmations, errors, etc.).
			if strings.Contains(line, "Login successful") {
				parts := strings.Split(line, " ")
				if len(parts) > 1 {
					currentUser = parts[1]
				}
				app.QueueUpdateDraw(func() {
					pages.SwitchToPage("chat")
				})
			}
			app.QueueUpdateDraw(func() {
				modal := tview.NewModal().
					SetText(line).
					AddButtons([]string{"Ok"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						pages.RemovePage("info")
					})
				pages.AddPage("info", modal, true, true)
			})
		}
	}
	// When the connection closes, display a disconnect modal.
	app.QueueUpdateDraw(func() {
		modal := tview.NewModal().
			SetText("Disconnected from server.").
			AddButtons([]string{"Quit"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				app.Stop()
			})
		pages.AddPage("disconnected", modal, true, true)
	})
}

func main() {
	var err error
	// Establish connection to the chat server (adjust the address if needed).
	conn, err = net.Dial("tcp", "localhost:9000")
	if err != nil {
		log.Fatalf("Unable to connect to server: %v", err)
	}
	defer conn.Close()

	// Initialize the tview application and page container.
	app = tview.NewApplication()
	pages = tview.NewPages()
	pages.AddPage("welcome", welcomeScreen(), true, true)
	pages.AddPage("login", loginForm(), true, false)
	pages.AddPage("register", registerForm(), true, false)
	pages.AddPage("chat", chatScreen(), true, false)

	// Start a goroutine to read incoming server messages.
	go readFromServer()

	// Run the TUI application.
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
