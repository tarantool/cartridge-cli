package connect

import (
	"fmt"
	"io"

	"github.com/FZambia/tarantool"
	"github.com/apex/log"
	"github.com/c-bata/go-prompt"
	"github.com/tarantool/cartridge-cli/cli/common"
)

func binaryConnect(console *Console) error {
	var err error

	connectStr := fmt.Sprintf("%s://%s", console.connOpts.Network, console.connOpts.Address)
	console.binaryConn, err = tarantool.Connect(connectStr, tarantool.Opts{
		User:           console.connOpts.Username,
		Password:       console.connOpts.Password,
		RequestTimeout: common.EvalTarantoolConnTimeout,
	})
	if err != nil {
		return fmt.Errorf("Failed to connect: %s", err)
	}

	return nil
}

func binaryEval(console *Console, funcBody string, args ...interface{}) (interface{}, error) {
	if args == nil {
		args = []interface{}{}
	}

	resp, err := console.binaryConn.Exec(tarantool.Eval(funcBody, args))

	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, nil
	}

	dataRaw := resp.Data[0]
	return dataRaw, nil
}

func getBinaryCompleter(console *Console) prompt.Completer {
	getSuggestionsBinary := func(console *Console, lastWord string) interface{} {
		res, err := console.Eval(
			getSuggestionsBinaryFuncBody,
			lastWord, len(lastWord),
		)

		if err != nil {
			return nil
		}

		return res
	}

	completer := func(in prompt.Document) []prompt.Suggest {
		return getSuggestions(console, in, getSuggestionsBinary)
	}

	return completer
}

func binaryExecute(console *Console, in string) string {
	dataRaw, err := console.Eval(
		"return require('console').eval(...)",
		in,
	)

	if err == io.EOF {
		log.Fatalf("Connection was closed. Probably instance process isn't running anymore")
	}
	if err != nil {
		log.Fatalf("Failed to eval: %s", err)
	}

	data, ok := dataRaw.(string)
	if !ok {
		log.Fatalf("Failed to eval: Data received in wrong format")
	}

	return data
}

const (
	getSuggestionsBinaryFuncBody = `
local last_word, last_word_len = ...
return require('console').completion_handler(last_word, 0, last_word_len)
`
)
