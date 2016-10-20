package app

import "testing"

func TestMain(m *testing.M) {
	enableDebug()
	m.Run()
}
