package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessEvalTarantoolResYAML(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var resStr string
	var res *TarantoolEvalRes
	var err error

	// success
	resStr = `---
- success: true
  data: 666
...`

	res, err = processEvalTarantoolResYAML([]byte(resStr))
	assert.Nil(err)
	assert.True(res.Success)
	assert.Equal(666, res.Data)

	// error
	resStr = `---
- success: false
  err: 'Some **it happened'
...`

	res, err = processEvalTarantoolResYAML([]byte(resStr))
	assert.Nil(err)
	assert.False(res.Success)
	assert.Equal("Some **it happened", res.ErrStr)

	// syntax error
	resStr = `---
- error: '[string "wtf is it?"]:1: ''='' expected near ''is'''
...`

	res, err = processEvalTarantoolResYAML([]byte(resStr))
	assert.Equal("Syntax error: [string \"wtf is it?\"]:1: '=' expected near 'is'", err.Error())

	// multiple results
	resStr = `---
- success: false
  err: 'Some **it happened'
- success: true
  err : 'Oh my God...'
...`

	res, err = processEvalTarantoolResYAML([]byte(resStr))
	assert.Equal("Expected one result, found 2", err.Error())

	// bad result format
	resStr = `---
- 666
...`

	res, err = processEvalTarantoolResYAML([]byte(resStr))
	assert.Contains(err.Error(), "Failed to parse eval result: yaml: unmarshal errors")
}

func TestProcessEvalTarantoolResLua(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var resStr string
	var res *TarantoolEvalRes
	var err error

	// success
	resStr = `{data = 666, success = true};`

	res, err = processEvalTarantoolResLua([]byte(resStr))
	assert.Nil(err)
	assert.True(res.Success)
	assert.Equal(666, res.Data)

	// error
	resStr = `{data = 666, err = "Some **it happened"};`

	res, err = processEvalTarantoolResLua([]byte(resStr))
	assert.Nil(err)
	assert.False(res.Success)
	assert.Equal("Some **it happened", res.ErrStr)

	// syntax error
	resStr = `"[string \"wtf is it? \"]:1: '=' expected near 'is'"`
	res, err = processEvalTarantoolResLua([]byte(resStr))
	assert.Equal("Syntax error: [string \"wtf is it? \"]:1: '=' expected near 'is'", err.Error())
}
