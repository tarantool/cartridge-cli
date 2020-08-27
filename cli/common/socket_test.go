package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessEvalTarantoolRes(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var resStr string
	var resData interface{}
	var err error

	// success
	resStr = `---
- success: true
  data: 666
...`

	resData, err = processEvalTarantoolRes([]byte(resStr))
	assert.Nil(err)
	assert.Equal(666, resData)

	// error
	resStr = `---
- success: false
  err: 'Some **it happened'
...`

	resData, err = processEvalTarantoolRes([]byte(resStr))
	assert.Equal("Failed to eval: Some **it happened", err.Error())

	// syntax error
	resStr = `---
- error: '[string "wtf is it?"]:1: ''='' expected near ''is'''
...`

	resData, err = processEvalTarantoolRes([]byte(resStr))
	assert.Equal("Syntax error: [string \"wtf is it?\"]:1: '=' expected near 'is'", err.Error())

	// multiple results
	resStr = `---
- success: false
  err: 'Some **it happened'
- success: true
  err : 'Oh my God...'
...`

	resData, err = processEvalTarantoolRes([]byte(resStr))
	assert.Equal("Expected one result, found 2", err.Error())

	// bad result format
	resStr = `---
- 666
...`

	resData, err = processEvalTarantoolRes([]byte(resStr))
	assert.Equal("Function should return { success = ..., err = ..., data = .... }", err.Error())
}
