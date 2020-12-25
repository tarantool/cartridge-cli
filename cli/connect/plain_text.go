package connect

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/c-bata/go-prompt"
	"github.com/tarantool/cartridge-cli/cli/common"
)

const (
	tagPushPrefixYAML = `%TAG !push!`
	tagPushPrefixLua  = `-- Push`

	successSetModeYAML = "---\n- true\n...\n"
	successSetModeLua  = "true;"
)

type ReadFromConnFunc func(net.Conn, common.ConnOpts) ([]byte, error)

func plainTextEval(console *Console, funcBodyFmt string, args ...interface{}) (interface{}, error) {
	var plainTextEvalFunc func(conn net.Conn, funcBody string, opts common.ConnOpts) (interface{}, error)

	switch {
	case console.outputMode == ConsoleYAMLOutput:
		plainTextEvalFunc = common.EvalTarantoolConn
	case console.outputMode == ConsoleLuaOutput:
		plainTextEvalFunc = common.EvalTarantoolConn
	default:
		return nil, fmt.Errorf("Unknown output mode: %s", console.outputMode)
	}

	return plainTextEvalFunc(console.conn, fmt.Sprintf(funcBodyFmt, args...), common.ConnOpts{})
}

func getPlainTextCompleter(console *Console) prompt.Completer {
	getSuggestionsPlainText := func(console *Console, lastWord string) interface{} {
		res, err := console.Eval(
			getSuggestionsPlainTextFuncBodyFmt,
			lastWord, len(lastWord),
		)

		if err != nil {
			return nil
		}

		return res
	}

	completer := func(in prompt.Document) []prompt.Suggest {
		return getSuggestions(console, in, getSuggestionsPlainText)
	}

	return completer
}

func plainTextExecute(console *Console, in string) string {
	var readFromConnFunc ReadFromConnFunc
	if err := common.WriteToConn(console.conn, in+"\n"); err != nil {
		log.Debugf("Failed to write to instance socket: %s", err)
		log.Fatalf("Connection was closed. Probably instance process isn't running anymore")
	}

	console.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	newOutputMode := getNewOutputMode(in)

	switch {
	case newOutputMode != "":
		readFromConnFunc = getReadFromConnSetMode(console, newOutputMode)
	case console.outputMode == ConsoleYAMLOutput:
		readFromConnFunc = common.ReadFromConnYAML
	case console.outputMode == ConsoleLuaOutput:
		readFromConnFunc = common.ReadFromConnLua
	default:
		log.Fatalf("Unknown output mode: %s", console.outputMode)
	}

	dataBytes, err := readFromConnFunc(console.conn, common.ConnOpts{
		PushCallback: func(s string) {
			fmt.Printf("%s", s)
		},
	})

	if err != nil {
		log.Debugf(err.Error())
		log.Fatalf("Connection was closed. Probably instance process isn't running anymore")
	}

	return string(dataBytes)
}

func pushTagIsReceived(console *Console, dataPortion string) bool {
	switch {
	case console.outputMode == ConsoleYAMLOutput:
		return strings.HasPrefix(dataPortion, tagPushPrefixYAML)
	case console.outputMode == ConsoleLuaOutput:
		return strings.HasPrefix(dataPortion, tagPushPrefixLua)
	}

	return false
}

func getReadFromConnSetMode(console *Console, newOutputMode ConsoleOutputMode) ReadFromConnFunc {
	readFromConnFunc := func(conn net.Conn, opts common.ConnOpts) ([]byte, error) {
		// This function handles only result of `\set output <mode>` commands.
		// So, only encoded `true` value or error can be returned and we can
		// simply read one portion of bytes to buffer of 1024 bytes.
		// Of course, it works only if all other results are handled correctly,
		// and it is so =).
		//
		if opts.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(opts.ReadTimeout))
		} else {
			conn.SetReadDeadline(time.Time{})
		}

		setModeResBytes := make([]byte, 1024)
		n, err := conn.Read(setModeResBytes)
		if err != nil {
			return nil, err
		}

		if modeIsSetSuccessfully(newOutputMode, string(setModeResBytes)) {
			console.outputMode = newOutputMode
		}

		return setModeResBytes[:n], nil
	}

	return readFromConnFunc
}

func modeIsSetSuccessfully(newOutputMode ConsoleOutputMode, setModeRes string) bool {
	switch {
	case newOutputMode == ConsoleYAMLOutput:
		return strings.HasPrefix(setModeRes, successSetModeYAML)
	case newOutputMode == ConsoleLuaOutput:
		return strings.HasPrefix(setModeRes, successSetModeLua)
	}

	return false
}

func getNewOutputMode(in string) ConsoleOutputMode {
	inWords := strings.Fields(in)

	if len(inWords) != 3 {
		return ""
	}

	if inWords[0] != "\\set" || inWords[1] != "output" {
		return ""
	}

	// handle \set output lua,line
	mode := strings.SplitN(inWords[2], ",", 2)[0]

	return ConsoleOutputMode(mode)
}

const (
	getSuggestionsPlainTextFuncBodyFmt = `
return require('console').completion_handler('%s', 0, %d)
`
)
