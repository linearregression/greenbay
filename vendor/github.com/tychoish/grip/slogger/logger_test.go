package slogger

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tychoish/grip/level"
	"github.com/tychoish/grip/send"
)

func TestLoggerLogf(t *testing.T) {
	assert := assert.New(t)
	sink, err := send.NewInternalLogger("sink", send.LevelInfo{level.Info, level.Info})
	assert.NoError(err)
	defer sink.Close()
	logger := &Logger{Name: "sloggerTest", Appenders: []send.Sender{sink}}

	assert.NoError(err)

	l, errs := logger.Logf(INFO, "foo %s", "bar")
	assert.True(len(errs) == 0)
	assert.NotNil(l)
	assert.Equal(l, sink.GetMessage().Message)
}

func TestLoggerErrorf(t *testing.T) {
	assert := assert.New(t)
	sink, err := send.NewInternalLogger("sink", send.LevelInfo{level.Info, level.Info})
	assert.NoError(err)
	defer sink.Close()
	logger := &Logger{Name: "sloggerTest", Appenders: []send.Sender{sink}}

	err = logger.Errorf(INFO, "foo %s", "bar")
	assert.Error(err)
	assert.Equal("foo bar", err.Error())
	assert.True(strings.Contains(sink.GetMessage().Rendered, "foo bar"))
}

func TestLoggerStackf(t *testing.T) {
	assert := assert.New(t)
	sink, err := send.NewInternalLogger("sink", send.LevelInfo{level.Info, level.Info})
	assert.NoError(err)
	defer sink.Close()
	logger := &Logger{Name: "sloggerTest", Appenders: []send.Sender{sink}}

	assert.NoError(err)

	l, errs := logger.Stackf(INFO, errors.New("baz"), "foo %s", "bar")
	assert.True(len(errs) == 0)
	assert.NotNil(l)

	assert.True(strings.HasSuffix(sink.GetMessage().Rendered, "foo bar\nbaz\n"))

}
