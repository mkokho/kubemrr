package pkg

import (
	"testing"
	"reflect"
	"encoding/json"
)

func TestGetPods(t *testing.T) {
	var p Pod
	json.Unmarshal([]byte(mockPod), &p)

	c := Cache {}

	tests := []struct {
    actual []Pod
    expected []Pod
  }{
		{
			actual: c.getPods(),
			expected: []Pod{p},
		},
	}

	for i, test := range tests {
		if !reflect.DeepEqual(test.actual, test.expected) {
			t.Errorf("Test %d: expected pod %v, found %v", i+1, test.expected, test.actual)
		}
	}
}
