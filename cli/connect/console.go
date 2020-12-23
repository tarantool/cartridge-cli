package connect

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/FZambia/tarantool"
	"github.com/apex/log"
	lua "github.com/yuin/gopher-lua"

	"github.com/c-bata/go-prompt"
	"github.com/tarantool/cartridge-cli/cli/common"
)

type ConsoleOutputMode string
type Protocol string

type EvalFunc func(console *Console, funcBodyFmt string, args ...interface{}) (interface{}, error)

const (
	ConsoleYAMLOutput ConsoleOutputMode = "yaml"
	ConsoleLuaOutput  ConsoleOutputMode = "lua"

	PlainTextProtocol Protocol = "plain text"
	BinaryProtocol    Protocol = "binary"

	HistoryFileName = ".tarantool_history"

	MaxLivePrefixIndent = 15
)

var (
	ControlLeftBytes  []byte
	ControlRightBytes []byte
)

func init() {
	ControlLeftBytes = []byte{0x1b, 0x62}
	ControlRightBytes = []byte{0x1b, 0x66}
}

type Console struct {
	input string

	title string

	historyFile     *os.File
	historyFilePath string
	historyLines    []string

	prefix            string
	livePrefixEnabled bool
	livePrefix        string
	livePrefixFunc    func() (string, bool)

	connOpts   *ConnOpts
	conn       net.Conn
	binaryConn *tarantool.Connection

	evalFunc  EvalFunc
	executor  func(in string)
	completer func(in prompt.Document) []prompt.Suggest

	protocol   Protocol
	outputMode ConsoleOutputMode

	luaState *lua.LState

	prompt *prompt.Prompt
}

func NewConsole(connOpts *ConnOpts, title string) (*Console, error) {
	console := &Console{
		title:      title,
		outputMode: ConsoleYAMLOutput,
		connOpts:   connOpts,
		luaState:   lua.NewState(),
	}

	var err error

	// load Tarantool console history from file
	if err := loadHistory(console); err != nil {
		log.Debugf("Failed to load Tarantool console history: %s", err)
	}

	// connect to specified address
	console.conn, err = net.Dial(connOpts.Network, connOpts.Address)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}

	// read greeting to get protocol
	// for binary protocol initialize console.binaryConn
	if err := detectProtocolAndReconnectIfRequired(console); err != nil {
		return nil, err
	}

	// initialize eval function
	console.evalFunc, err = getEvalFunc(console)
	if err != nil {
		return nil, fmt.Errorf("Failed to get eval function: %s", err)
	}

	// initialize user commands executor
	console.executor, err = getExecutor(console)
	if err != nil {
		return nil, fmt.Errorf("Failed to get executor: %s", err)
	}

	// initialize commands completer
	console.completer, err = getCompleter(console)
	if err != nil {
		return nil, fmt.Errorf("Failed to get completer: %s", err)
	}

	// set title and prompt prefix
	// <app-name>.<instance-name> for Cartridge application instances
	// <host>:<port> otherwise
	setTitle(console)
	setPrefix(console)

	return console, nil
}

func (console *Console) Run() error {
	var err error

	fmt.Printf("connected to %s\n", console.title)

	pipedInputIsFound, err := common.StdinHasUnreadData()
	if err != nil {
		return fmt.Errorf("Failed to check unread data from stdin: %s", err)
	}

	if pipedInputIsFound {
		log.Debugf("Found piped input")
		// e.g. `echo "box.info()" | cartridge enter router`
		pipedInputScanner := bufio.NewScanner(os.Stdin)
		for pipedInputScanner.Scan() {
			line := pipedInputScanner.Text()
			console.executor(line)
		}
		return nil
	}

	// get options for Prompt instance
	options := getPromptOptions(console)

	// create Prompt instance
	console.prompt = prompt.New(
		console.executor,
		console.completer,
		options...,
	)

	console.prompt.Run()

	return nil
}

func (console *Console) Close() {
	if console.historyFile != nil {
		console.historyFile.Close()
	}
}

func (console *Console) Eval(funcBody string, args ...interface{}) (interface{}, error) {
	return console.evalFunc(console, funcBody, args...)
}

func loadHistory(console *Console) error {
	var err error

	homeDir, err := common.GetHomeDir()
	if err != nil {
		return fmt.Errorf("Failed to get home directory: %s", err)
	}

	console.historyFilePath = filepath.Join(homeDir, HistoryFileName)

	console.historyLines, err = common.GetLastNLines(console.historyFilePath, MaxHistoryLines)
	if err != nil {
		return fmt.Errorf("Failed to read history from file: %s", err)
	}

	// open history file for appending
	// see https://unix.stackexchange.com/questions/346062/concurrent-writing-to-a-log-file-from-many-processes
	console.historyFile, err = os.OpenFile(
		console.historyFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	if err != nil {
		log.Debugf("Failed to open history file for append: %s", err)
	}

	return nil
}

func detectProtocolAndReconnectIfRequired(console *Console) error {
	greeting, err := readGreeting(console.conn)
	if err != nil {
		return fmt.Errorf("Failed to read Tarantool greeting: %s", err)
	}

	switch {
	case strings.Contains(greeting, "(Lua console)"):
		log.Debugf("Plain text protocol is detected")
		console.protocol = PlainTextProtocol
	case strings.Contains(greeting, "(Binary)"):
		log.Debugf("Binary protocol is detected")
		console.protocol = BinaryProtocol
		if err := binaryConnect(console); err != nil {
			return fmt.Errorf("Failed to connect to binary port: %s", err)
		}
	default:
		return fmt.Errorf("Unknown protocol: %s", greeting)
	}

	return nil
}

func getEvalFunc(console *Console) (EvalFunc, error) {
	switch {
	case console.protocol == PlainTextProtocol:
		return plainTextEval, nil
	case console.protocol == BinaryProtocol:
		return binaryEval, nil
	default:
		return nil, fmt.Errorf("Unknown protocol: %s", console.protocol)
	}
}

func getExecutor(console *Console) (prompt.Executor, error) {
	var executeFunc func(console *Console, in string) string
	switch {
	case console.protocol == PlainTextProtocol:
		executeFunc = plainTextExecute
	case console.protocol == BinaryProtocol:
		executeFunc = binaryExecute
	default:
		return nil, fmt.Errorf("Unknown protocol: %s", console.protocol)
	}

	executor := func(in string) {
		console.input += in + " "

		if !inputIsCompleted(console.input, console.luaState) {
			console.livePrefixEnabled = true
			return
		}

		if err := appendToHistoryFile(console, console.input); err != nil {
			log.Debugf("Failed to append command to history file: %s", err)
		}

		data := executeFunc(console, console.input)

		fmt.Printf("%s\n", data)

		console.input = ""
		console.livePrefixEnabled = false
	}

	return executor, nil
}

func inputIsCompleted(input string, luaState *lua.LState) bool {
	// see https://github.com/tarantool/tarantool/blob/b53cb2aeceedc39f356ceca30bd0087ee8de7c16/src/box/lua/console.lua#L575
	if _, err := luaState.LoadString(input); err == nil || !strings.Contains(err.Error(), "at EOF") {
		// valid Lua code or a syntax error not due to
		// an incomplete input
		return true
	}

	if _, err := luaState.LoadString(fmt.Sprintf("return %s", input)); err == nil {
		// certain obscure inputs like '(42\n)' yield the
		// same error as incomplete statement
		return true
	}

	return false
}

func getCompleter(console *Console) (prompt.Completer, error) {
	switch {
	case console.protocol == PlainTextProtocol:
		return getPlainTextCompleter(console), nil
	case console.protocol == BinaryProtocol:
		return getBinaryCompleter(console), nil
	}

	return nil, fmt.Errorf("Unknown protocol: %s", console.protocol)
}

func setTitle(console *Console) {
	if console.title != "" {
		return
	}

	titleRaw, err := console.Eval(getTitleFuncBody)
	if err != nil {
		log.Debugf("Failed to get instance title: %s", err)
	} else if err == nil {
		if title, ok := titleRaw.(string); ok {
			console.title = title
		}
	}

	if console.title == "" {
		console.title = console.connOpts.Address
	}
}

func setPrefix(console *Console) {
	console.prefix = fmt.Sprintf("%s> ", console.title)

	livePrefixIndent := len(console.title)
	if livePrefixIndent > MaxLivePrefixIndent {
		livePrefixIndent = MaxLivePrefixIndent
	}

	console.livePrefix = fmt.Sprintf("%s> ", strings.Repeat(" ", livePrefixIndent))

	console.livePrefixFunc = func() (string, bool) {
		return console.livePrefix, console.livePrefixEnabled
	}
}

func getPromptOptions(console *Console) []prompt.Option {
	options := []prompt.Option{
		prompt.OptionTitle(console.title),
		prompt.OptionPrefix(console.prefix),
		prompt.OptionLivePrefix(console.livePrefixFunc),

		prompt.OptionHistory(console.historyLines),

		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionPreviewSuggestionTextColor(prompt.DefaultColor),

		prompt.OptionCompletionWordSeparator(tarantoolWordSeparators),

		prompt.OptionAddASCIICodeBind(
			prompt.ASCIICodeBind{ // move to one word left
				ASCIICode: ControlLeftBytes,
				Fn: func(buf *prompt.Buffer) {
					d := buf.Document()
					wordLen := len([]rune(d.GetWordBeforeCursorWithSpace()))
					buf.CursorLeft(wordLen)
				},
			},
			prompt.ASCIICodeBind{ // move to one word right
				ASCIICode: ControlRightBytes,
				Fn: func(buf *prompt.Buffer) {
					d := buf.Document()
					wordLen := len([]rune(d.GetWordAfterCursorWithSpace()))
					buf.CursorRight(wordLen)
				},
			},
		),
	}

	return options
}

func readGreeting(conn net.Conn) (string, error) {
	greeting := make([]byte, 1024)
	if _, err := conn.Read(greeting); err != nil {
		return "", fmt.Errorf("Failed to read Tarantool greeting: %s", err)
	}

	return string(greeting), nil
}

func appendToHistoryFile(console *Console, in string) error {
	if console.historyFile == nil {
		return fmt.Errorf("No hostory file found")
	}

	if _, err := console.historyFile.WriteString(in + "\n"); err != nil {
		return fmt.Errorf("Failed to append to history file: %s", err)
	}

	if err := console.historyFile.Sync(); err != nil {
		return fmt.Errorf("Failed to sync history file: %s", err)
	}

	return nil
}

const (
	getTitleFuncBody = `
local ok, api_topology = pcall(require, 'cartridge.lua-api.topology')
if not ok then
	return ''
end

local self = api_topology.get_self()
if self.app_name == nil or self.instance_name == nil then
	return ''
end

return self.app_name .. '.' .. self.instance_name
`
)
