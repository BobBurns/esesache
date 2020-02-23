package main

import (
	"fmt"
	"io"

	"github.com/jroimartin/gocui"
)

type selection struct {
	prompt    string
	usage     string
	choices   []string
	action    string
	selection int
}

func (s *selection) layout(g *gocui.Gui) error {
	maxx, maxy := g.Size()
	v, err := g.SetView("header", 0, 0, max(80, len(s.prompt)), 2)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
	}

	v.Clear()
	s.renderHeader(v, maxx)

	v, err = g.SetView("list", 0, 3, maxx-1, maxy-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Highlight = true
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack

		if _, err := g.SetCurrentView("list"); err != nil {
			return err
		}
	}

	v.Clear()
	s.render(v, maxx)

	v, err = g.SetView("footer", 0, maxy-3, max(80, len(s.usage)), maxy-1)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
	}

	v.Clear()
	s.renderFooter(v, maxx)

	return nil
}

func (s *selection) renderHeader(v io.Writer, maxx int) {
	fmt.Fprintf(v, "\u001b[1m Select Host to Connect to \u001b[0m\n")
}

func (s *selection) render(v io.Writer, maxx int) {
	for _, item := range s.choices {
		fmt.Fprintf(v, "%s\n", item)
	}
}

func (s *selection) renderFooter(v io.Writer, maxx int) {
	fmt.Fprintf(v, "\u001b[1m%s\u001b[0m\n", s.usage)
}

func (s *selection) keybindings(g *gocui.Gui) error {
	for _, kb := range []struct {
		name   string
		key    interface{}
		action func(*gocui.Gui, *gocui.View) error
	}{
		{"", 'q', s.quit},
		{"list", gocui.KeyArrowDown, s.cursorDown},
		{"list", gocui.KeyArrowUp, s.cursorUp},
		{"list", gocui.KeyEnter, s.def},
	} {
		if err := g.SetKeybinding(kb.name, kb.key, gocui.ModNone, kb.action); err != nil {
			return err
		}
	}
	return nil
}

func (s *selection) quit(g *gocui.Gui, v *gocui.View) error {
	s.action = "quit"
	return gocui.ErrQuit
}

func (s *selection) cursorDown(g *gocui.Gui, v *gocui.View) error {
	y := getSelectedLine(v)
	if y < len(s.choices)-1 {
		v.MoveCursor(0, 1, false)
	}
	return nil
}

func (s *selection) cursorUp(g *gocui.Gui, v *gocui.View) error {
	v.MoveCursor(0, -1, false)
	return nil
}

func (s *selection) def(g *gocui.Gui, v *gocui.View) error {
	s.action = "ok"
	s.selection = getSelectedLine(v)
	return gocui.ErrQuit
}

// GetSelection

// act, sel := cui.GetSelection(ctx, "Select action", "<↑/↓> to change the
// selection, <→> to select, <ESC> to quit", choices)
func GetSelection(choices []string) (string, int) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	g.InputEsc = true
	g.Mouse = false
	g.Cursor = false

	s := selection{
		prompt:    "Choose host to connect to",
		usage:     "<↑/↓> to change the selection, Enter to select, q to quit",
		choices:   choices,
		selection: 0,
	}

	g.SetManagerFunc(s.layout)

	if err := s.keybindings(g); err != nil {
		panic(err)
	}

	if err := g.MainLoop(); err != nil {
		if err != gocui.ErrQuit {
			return "aborted", s.selection
		}
	}
	return s.action, s.selection
}

func getSelectedLine(v *gocui.View) int {
	_, y := v.Cursor()
	_, oy := v.Origin()

	return y + oy
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
