package common

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/tarantool/cartridge-cli/cli/templates"
)

func ReadFromConn(conn net.Conn) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	tmp := make([]byte, 1024)
	data := make([]byte, 0)

	for {
		if n, err := conn.Read(tmp); n == 0 || err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("Failed to read: %s", err)
		} else {
			data = append(data, tmp[:n]...)
		}
	}

	return data, nil
}

type TarantoolEvalRes struct {
	Success bool        `yaml:"success"`
	Data    interface{} `yaml:"data"`
	ErrStr  string      `yaml:"err"`
}

// EvalTarantoolConn calls function on Tarantool instance
// Function should return `interface{}`, `string` (res, err)
// to be correctly processed
func EvalTarantoolConn(conn net.Conn, funcBody string) (interface{}, error) {
	evalFuncTmpl := `
	local ok, res, err = pcall(function()
		local function f()
			require('fiber').self().storage.console = nil
			{{ .FunctionBody }}
		end
		return f()
	end)

	if not ok then
		return { success = false, err = res or box.NULL }
	end

	if err ~= nil then
		return { success = false, err = err or box.NULL }
	end

	return { success = true, data = res or box.NULL }
	`

	evalFunc, err := templates.GetTemplatedStr(&evalFuncTmpl, map[string]string{
		"FunctionBody": funcBody,
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to instantiate eval function template: %s", err)
	}

	evalFuncFormatted := strings.Join(
		strings.Split(strings.TrimSpace(evalFunc), "\n"), " ",
	)
	evalFuncFormatted = strings.Join(strings.Fields(evalFuncFormatted), " ") + "\n"

	// write to socket
	writer := bufio.NewWriter(conn)
	if _, err := writer.WriteString(evalFuncFormatted); err != nil {
		return nil, fmt.Errorf("Failed to send to socket: %s", err)
	}

	writer.Flush()

	// recv from socket
	res, err := ReadFromConn(conn)
	if err != nil {
		return nil, fmt.Errorf("Failed to check returned data: %s", err)
	}

	results := []TarantoolEvalRes{}
	if err := yaml.UnmarshalStrict(res, &results); err != nil {
		errorStrings := make([]map[string]string, 0)
		if err := yaml.UnmarshalStrict(res, &errorStrings); err == nil {
			if len(errorStrings) > 0 {
				err, found := errorStrings[0]["error"]
				if found {
					return nil, fmt.Errorf("Failed to eval: %s", err)
				}
			}

		}

		return nil, fmt.Errorf("Failed to unmarshal results: %s", err)
	}

	data := results[0]

	if !data.Success {
		return nil, fmt.Errorf("Failed to eval: %s", data.ErrStr)
	}

	return data, nil
}
