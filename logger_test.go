package rueidisleader

import "testing"

type testWriter struct {
	t *testing.T
}

func (t testWriter) Write(line []byte) (int, error) {
	t.t.Log(string(line))
	return len(line), nil
}
