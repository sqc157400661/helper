package mysql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecutor(t *testing.T) {
	eng, err := NewMySQLEngine(ConnectInfo{
		Host:   "127.0.0.1",
		Port:   3306,
		User:   "root",
		Passwd: "",
	}, true, true)
	assert.NoError(t, err)
	executor := NewExecutorByEngine(eng)
	_, err = executor.ShowSlaveStatus()
	assert.NoError(t, err)
	_, err = executor.ShowMasterStatus()
	assert.NoError(t, err)
}
