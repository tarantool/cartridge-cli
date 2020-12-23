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
	lua "github.com/yuin/gopher-lua"
)

const (
	endOfYAMLOutput = "\n...\n"
	endOfLuaOutput  = ";"
)

func ConnectToTarantoolSocket(socketPath string) (net.Conn, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// read greeting
	tmp := make([]byte, 1024)
	if _, err := conn.Read(tmp); err != nil && err != io.EOF {
		return nil, fmt.Errorf("Failed to read Tarantool greeting: %s", err)
	}

	return conn, nil
}

// ReadFromConn function reads plain text from Tarantool connection
// These code was inspired by Tarantool console eval
// https://github.com/tarantool/tarantool/blob/3bc4a156e937102f23e2157ef88aa6c007759005/src/box/lua/console.lua#L469
//
// By default, Tarantool sends YAML-encoded values as user command response.
// In this case the end of output value is `\n...\n`.
// What about a case when return string contains this substring?
// Everything is OK, since yaml-encoded string is indented via two spaces,
// so in fact we never have `\n...\n` in output strings.
//
// E.g.
// tarantool> return '\n...\n'
// ---
// - '
//
//   ...
//
//   '
// ...
//
// If Lua output is set, the end of input is just ";".
// And there are some problems.
// See https://github.com/tarantool/tarantool/issues/4603
//
// Code is read byte by byte to make parsing output simplier
// (in case of box.session.push() response we need to read 2 yaml-encoded values,
// it's not enough to catch end of output, we should be sure that only one
// yaml-encoded value was read).
func ReadFromConn(conn net.Conn, endOfOutput string, readTimeout time.Duration) ([]byte, error) {
	tmp := make([]byte, 1)
	data := make([]byte, 0)

	if readTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
	} else {
		conn.SetReadDeadline(time.Time{})
	}

	for {
		if n, err := conn.Read(tmp); err != nil && err != io.EOF {
			return nil, fmt.Errorf("Failed to read: %s", err)
		} else if n == 0 || err == io.EOF {
			break
		} else {
			data = append(data, tmp[:n]...)
			if strings.HasSuffix(string(data), endOfOutput) {
				break
			}
		}
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("Connection was closed")
	}

	return data, nil
}

func ReadFromConnYAML(conn net.Conn, readTimeout time.Duration) ([]byte, error) {
	return ReadFromConn(conn, endOfYAMLOutput, readTimeout)
}

func ReadFromConnLua(conn net.Conn, readTimeout time.Duration) ([]byte, error) {
	return ReadFromConn(conn, endOfLuaOutput, readTimeout)
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

func formatAndSendEvalFunc(conn net.Conn, funcBody string, evalFuncTmpl string) error {
	evalFunc, err := templates.GetTemplatedStr(&evalFuncTmpl, map[string]string{
		"FunctionBody": funcBody,
	})

	if err != nil {
		return fmt.Errorf("Failed to instantiate eval function template: %s", err)
	}

	evalFuncFormatted := strings.Join(
		strings.Split(strings.TrimSpace(evalFunc), "\n"), " ",
	)
	evalFuncFormatted = strings.Join(strings.Fields(evalFuncFormatted), " ") + "\n"

	// write to socket
	if err := WriteToConn(conn, evalFuncFormatted); err != nil {
		return fmt.Errorf("Failed to send eval function to socket: %s", err)
	}

	return nil
}

// EvalTarantoolConn calls function on Tarantool instance
// Function should return `interface{}`, `string` (res, err)
// to be correctly processed.
// Processes only YAML output.
func EvalTarantoolConn(conn net.Conn, funcBody string, readTimeout time.Duration) (interface{}, error) {
	if err := formatAndSendEvalFunc(conn, funcBody, evalFuncYAMLTmpl); err != nil {
		return nil, err
	}

	// recv from socket
	resBytes, err := ReadFromConnYAML(conn, readTimeout)
	if err != nil {
		return nil, fmt.Errorf("Failed to check returned data: %s", err)
	}

	data, err := processEvalTarantoolResYAML(resBytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse data returned from instance: %s", err)
	}

	if !data.Success {
		return nil, fmt.Errorf(data.ErrStr)
	}

	return data.Data, nil
}

func EvalTarantoolConnNoTimeout(conn net.Conn, funcBody string) (interface{}, error) {
	return EvalTarantoolConn(conn, funcBody, 0)
}

func EvalTarantoolConnLua(conn net.Conn, funcBody string, readTimeout time.Duration) (interface{}, error) {
	if err := formatAndSendEvalFunc(conn, funcBody, evalFuncLuaTmpl); err != nil {
		return nil, err
	}

	// recv from socket
	resBytes, err := ReadFromConnLua(conn, readTimeout)
	if err != nil {
		return nil, fmt.Errorf("Failed to check returned data: %s", err)
	}

	data, err := processEvalTarantoolResLua(resBytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse data returned from instance: %s", err)
	}

	if !data.Success {
		return nil, fmt.Errorf(data.ErrStr)
	}

	return data.Data, nil
}

func EvalTarantoolConnLuaNoTimeout(conn net.Conn, funcBody string) (interface{}, error) {
	return EvalTarantoolConnLua(conn, funcBody, 0)
}

func processEvalTarantoolResYAML(resBytes []byte) (*TarantoolEvalRes, error) {
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

		return nil, fmt.Errorf("Failed to parse eval result: %s", err)
	}

	if len(results) != 1 {
		return nil, fmt.Errorf("Expected one result, found %d", len(results))
	}

	res := results[0]

	return &res, nil
}

func processEvalTarantoolResLua(resBytes []byte) (*TarantoolEvalRes, error) {
	L := lua.NewState()
	defer L.Close()

	doString := fmt.Sprintf(`res = %s`, resBytes)

	if err := L.DoString(doString); err != nil {
		return nil, err
	}

	luaRes := L.Env.RawGetString("res")

	if luaRes.Type() == lua.LTString {
		return nil, fmt.Errorf("Syntax error: %s", lua.LVAsString(luaRes))
	}

	successLV := L.GetTable(luaRes, lua.LString("success"))
	messageLV := L.GetTable(luaRes, lua.LString("err"))
	// I've no idea how to get interface{} value from a map =(
	encodedDataLV := L.GetTable(luaRes, lua.LString("data"))

	success := lua.LVAsBool(successLV)
	message := lua.LVAsString(messageLV)
	encodedData := lua.LVAsString(encodedDataLV)

	res := TarantoolEvalRes{
		Success: success,
		ErrStr:  message,
	}

	if err := yaml.Unmarshal([]byte(encodedData), &res.Data); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal data: %s", err)
	}

	return &res, nil
}

const (
	evalFuncYAMLTmpl = `
local ok, res, err = pcall(function()
	require('fiber').self().storage.console = nil
	{{ .FunctionBody }}
end)

if res == nil then res = box.NULL end
if err == nil then err = box.NULL end

if not ok then
	return { success = false, err = res}
end

if err ~= nil then
	return { success = false, err = tostring(err) }
end

return { success = true, data = res }
`

	evalFuncLuaTmpl = `
local ok, res, err = pcall(function()
	require('fiber').self().storage.console = nil
	{{ .FunctionBody }}
end)

if res == nil then res = box.NULL end
if err == nil then err = box.NULL end

if not ok then
	return { success = false, err = res}
end

if err ~= nil then
	return { success = false, err = tostring(err) }
end

return { success = true, data = require('yaml').encode(res) }
`
)
