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

type ReadFromConnFunc func(net.Conn, time.Duration) ([]byte, error)

func plainTextEval(console *Console, funcBodyFmt string, args ...interface{}) (interface{}, error) {
	var plainTextEvalFunc func(conn net.Conn, funcBody string) (interface{}, error)

	switch {
	case console.outputMode == ConsoleYAMLOutput:
		plainTextEvalFunc = common.EvalTarantoolConnNoTimeout
	case console.outputMode == ConsoleLuaOutput:
		plainTextEvalFunc = common.EvalTarantoolConnLuaNoTimeout
	default:
		return nil, fmt.Errorf("Unknown output mode: %s", console.outputMode)
	}

	return plainTextEvalFunc(console.conn, fmt.Sprintf(funcBodyFmt, args...))
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

	var data string

	for {
		// data is read in cycle because of `box.session.push` command
		// it prints a tag and returns pushed value, and then true is returned additionally
		// e.g.
		// myapp.router> box.session.push('xx')
		// %TAG !push! tag:tarantool.io/push,2018
		// --- xx
		// ...
		// ---
		// - true
		// ...
		//
		// So, when data portion starts with a tag prefix, we have to read one more value
		//
		dataPortionBytes, err := readFromConnFunc(console.conn, readFromConnExecTimeout)
		if err != nil {
			log.Debugf("Failed to read from instance socket: %s", err)
			log.Fatalf("Connection was closed. Probably instance process isn't running anymore")
		}

		dataPortion := string(dataPortionBytes)
		data += dataPortion

		if !pushTagIsReceived(console, dataPortion) {
			break
		}
	}

	return data
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
	readFromConnFunc := func(conn net.Conn, readTimeout time.Duration) ([]byte, error) {
		// This function handles only result of `\set output <mode>` commands.
		// So, only encoded `true` value or error can be returned and we can
		// simply read one portion of bytes to buffer of 1024 bytes.
		// Of course, it works only if all other results are handled correctly,
		// and it is so =).
		//
		if readTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(readTimeout))
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
