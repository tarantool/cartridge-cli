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
		// from the doc
		// https://golang.org/pkg/bytes/#Buffer.ReadBytes
		// > ReadBytes returns err != nil if and only if the returned data does not end in delim
		// so, we read bytes until 0 bytes returned
		lineBytes, _ := buf.ReadBytes('\n')
		if len(lineBytes) == 0 {
			break
		}

		// prefix
		if nPrefix, err := w.out.Write([]byte(w.prefix)); err != nil {
			n += nPrefix
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

		nLine, err := w.out.Write(lineBytes)
		n += nLine

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
