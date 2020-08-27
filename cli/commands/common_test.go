package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDuration(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var duration time.Duration
	var err error

	duration, err = getDuration("72h1m0.5s")
	assert.Nil(err)
	assert.Equal("72h1m0.5s", duration.String())

	duration, err = getDuration("100")
	assert.Nil(err)
	assert.Equal("1m40s", duration.String())

	_, err = getDuration("forever")
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), `invalid duration "forever"`), err.Error())

	_, err = getDuration("-1")
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), `Negative duration is specified`), err.Error())

	_, err = getDuration("-10m")
	assert.NotNil(err)
	assert.True(strings.Contains(err.Error(), `Negative duration is specified`), err.Error())
}
