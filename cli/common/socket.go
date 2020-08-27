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

func WriteToConn(conn net.Conn, data string) error {
	writer := bufio.NewWriter(conn)
	if _, err := writer.WriteString(data); err != nil {
		return fmt.Errorf("Failed to send to socket: %s", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Failed to flush: %s", err)
	}

	return nil
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

	if res == nil then res = box.NULL end
	if err == nil then err = box.NULL end

	if not ok then
		return { success = false, err = res}
	end

	if err ~= nil then
		return { success = false, err = err }
	end

	return { success = true, data = res }
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
	if err := WriteToConn(conn, evalFuncFormatted); err != nil {
		return nil, fmt.Errorf("Failed to send eval function to socket: %s", err)
	}

	// recv from socket
	resBytes, err := ReadFromConn(conn)
	if err != nil {
		return nil, fmt.Errorf("Failed to check returned data: %s", err)
	}

	data, err := processEvalTarantoolRes(resBytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to get result data: %s", err)
	}

	return data, nil
}

func processEvalTarantoolRes(resBytes []byte) (interface{}, error) {
	results := []TarantoolEvalRes{}
	if err := yaml.UnmarshalStrict(resBytes, &results); err != nil {
		errorStrings := make([]map[string]string, 0)
		if err := yaml.UnmarshalStrict(resBytes, &errorStrings); err == nil {
			if len(errorStrings) > 0 {
				err, found := errorStrings[0]["error"]
				if found {
					return nil, fmt.Errorf("Syntax error: %s", err)
				}
			}

		}

		return nil, fmt.Errorf("Function should return { success = ..., err = ..., data = .... }")
	}

	if len(results) != 1 {
		return nil, fmt.Errorf("Expected one result, found %d", len(results))
	}

	data := results[0]

	if !data.Success {
		return nil, fmt.Errorf("Failed to eval: %s", data.ErrStr)
	}

	return data.Data, nil
}
