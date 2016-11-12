package app

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	enableDebug()
	code := m.Run()
	os.Exit(code)
}
