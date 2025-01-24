package cmd

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [flags] trace_file constraint_file(s)",
	Short: "Inspect a trace file",
	Long:  `Inspect a trace file using an interactive (terminal-based) environment`,
	Run: func(cmd *cobra.Command, args []string) {
		if term.IsTerminal(0) {
			println("in a term")
		} else {
			println("not in a term")
		}
		width, height, err := term.GetSize(0)
		if err != nil {
			return
		}
		state, err := term.MakeRaw(0)
		if err != nil {
			return
		}
		screen := struct {
			io.Reader
			io.Writer
		}{os.Stdin, os.Stdout}

		terminal := term.NewTerminal(screen, "")
		for j := 0; j < width; j++ {
			frame := createFrame(j, width, height)
			terminal.Write(frame)
			time.Sleep(50 * time.Millisecond)
		}
		//
		term.Restore(0, state)
	},
}

func createFrame(offset int, width int, height int) []byte {
	var buf bytes.Buffer
	for i := 0; i < height; i++ {
		buf.Write(createLine(offset, width))
	}

	return buf.Bytes()
}

func createLine(offset int, width int) []byte {
	line := make([]byte, width)
	for i := range line {
		if i == offset {
			line[i] = '*'
		} else {
			line[i] = ' '
		}
	}
	return line
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(inspectCmd)
}
