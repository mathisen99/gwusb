// Package components provides reusable GUI components for WoeUSB-go
package components

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// PasswordResult holds the result of a password dialog
type PasswordResult struct {
	Password  string
	Cancelled bool
}

// ShowPasswordDialog displays a password entry dialog and returns the result via callback
func ShowPasswordDialog(parent fyne.Window, callback func(result PasswordResult)) {
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.PlaceHolder = "Enter your password"

	// Create form items
	formItems := []*widget.FormItem{
		{Text: "Password", Widget: passwordEntry},
	}

	// Create and show the dialog
	d := dialog.NewForm(
		"Administrator Password Required",
		"Authenticate",
		"Cancel",
		formItems,
		func(submitted bool) {
			if submitted {
				callback(PasswordResult{
					Password:  passwordEntry.Text,
					Cancelled: false,
				})
			} else {
				callback(PasswordResult{Cancelled: true})
			}
		},
		parent,
	)

	// Make the dialog a reasonable size
	d.Resize(fyne.NewSize(400, 150))
	d.Show()

	// Focus the password entry
	parent.Canvas().Focus(passwordEntry)
}

// ShowPasswordDialogSync shows a password dialog and blocks until user responds
// Returns the password and whether it was cancelled
func ShowPasswordDialogSync(parent fyne.Window) (string, bool) {
	var result PasswordResult
	var wg sync.WaitGroup
	wg.Add(1)

	// Must run dialog on main thread
	ShowPasswordDialog(parent, func(r PasswordResult) {
		result = r
		wg.Done()
	})

	wg.Wait()
	return result.Password, result.Cancelled
}

// PasswordDialogWithInfo shows a password dialog with additional info text
func ShowPasswordDialogWithInfo(parent fyne.Window, info string, callback func(result PasswordResult)) {
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.PlaceHolder = "Enter your password"

	infoLabel := widget.NewLabel(info)
	infoLabel.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		infoLabel,
		widget.NewSeparator(),
		widget.NewLabel("Password:"),
		passwordEntry,
	)

	d := dialog.NewCustomConfirm(
		"Administrator Password Required",
		"Authenticate",
		"Cancel",
		content,
		func(submitted bool) {
			if submitted {
				callback(PasswordResult{
					Password:  passwordEntry.Text,
					Cancelled: false,
				})
			} else {
				callback(PasswordResult{Cancelled: true})
			}
		},
		parent,
	)

	d.Resize(fyne.NewSize(450, 200))
	d.Show()

	// Focus the password entry
	parent.Canvas().Focus(passwordEntry)
}
