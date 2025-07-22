package wattpilot

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	w := New("localhost", "password")
	assert.NotNil(t, w)
	assert.Equal(t, "localhost", w.host)
	assert.Equal(t, "password", w.password)
	assert.NotNil(t, w.logger)
	assert.NotNil(t, w.notify)
	assert.NotNil(t, w.eventHandler)
}

func TestParseLogLevel(t *testing.T) {
	w := New("localhost", "password")
	err := w.ParseLogLevel("info")
	assert.NoError(t, err)
	assert.Equal(t, logrus.InfoLevel, w.logger.GetLevel())

	err = w.ParseLogLevel("invalid")
	assert.Error(t, err)
}

func TestGetters(t *testing.T) {
	w := New("localhost", "password")
	w.name = "test"
	w.serial = "1234"
	w.isInitialized = true

	assert.Equal(t, "test", w.GetName())
	assert.Equal(t, "1234", w.GetSerial())
	assert.Equal(t, "localhost", w.GetHost())
	assert.True(t, w.IsInitialized())
}

func TestAlias(t *testing.T) {
	w := New("localhost", "password")
	assert.NotNil(t, w.Alias())
	assert.Equal(t, "acs", w.LookupAlias("accessState"))
}
