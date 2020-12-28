package running

import (
	"bytes"
	"io"
	"os"
	"regexp"

	"github.com/fatih/color"
)

var (
	currentPrefixColor = 0
	prefixColors       = []color.Attribute{
		color.FgHiBlue,
		color.FgHiCyan,
		color.FgHiMagenta,
		color.FgBlue,
		color.FgCyan,
		color.FgMagenta,
	}

	// https://github.com/tarantool/tarantool/blob/df4c69ec15bfa86fcb7826a2359356845ab0c64a/src/lib/core/say.c#L104
	logLevelColors = map[string]color.Attribute{
		"F": color.FgRed,
		"!": color.FgRed,
		"E": color.FgRed,
		"C": color.FgRed,
		"W": color.FgYellow,
	}

	logLineRgx *regexp.Regexp
)

func init() {
	logLineRgx = regexp.MustCompile("([F!ECWIWD])> ")
}

type ColorizedWriter struct {
	prefix string
	out    io.Writer
}

func (w *ColorizedWriter) Write(p []byte) (int, error) {
	buf := bytes.NewBuffer(p)

	n := 0
	for {
		lineBytes, err := buf.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return n, err
		}

		// prefix
		if _, err := w.out.Write([]byte(w.prefix)); err != nil {
			return n, err
		}

		// colorize line by log level
		matches := logLineRgx.FindStringSubmatch(string(lineBytes))
		if matches != nil {
			logLevel := matches[1]

			colorAttr, found := logLevelColors[logLevel]
			if found {
				color.Set(colorAttr)
			}
		}

		lineN, err := w.out.Write(lineBytes)
		n += lineN

		if err != nil {
			return n, err
		}

		color.Unset()
	}

	return n, nil
}

func (w *ColorizedWriter) Close() error {
	return nil
}

func nexPrefixColor() *color.Color {
	c := color.New(prefixColors[currentPrefixColor%len(prefixColors)])
	currentPrefixColor++

	return c
}

func newColorizedWriter(prefix string) *ColorizedWriter {
	writer := ColorizedWriter{
		out: os.Stdout,
	}

	prefixColor := nexPrefixColor()
	writer.prefix = prefixColor.Sprintf("%s | ", prefix)
	return &writer
}
