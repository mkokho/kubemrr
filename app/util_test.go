package app

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParseKubeConfigFailures(t *testing.T) {
	tests := []struct {
		filename string
		complain string
	}{
		{
			filename: "test_data/kubeconfig_missing",
			complain: "not read",
		},
		{
			filename: "test_data/kubeconfig_invalid",
			complain: "invalid",
		},
	}

	for _, test := range tests {
		_, err := parseKubeConfig(test.filename)
		if err == nil {
			t.Errorf("Expected an error for file %s", test.filename)
			continue
		}

		if !strings.Contains(err.Error(), test.complain) {
			t.Errorf("Error [%s] does not contain [%s]", err, test.complain)
		}
	}
}

func TestParseKubeConfig(t *testing.T) {
	actual, err := parseKubeConfig("test_data/kubeconfig_valid")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	expected := Config{
		CurrentContext: "prod",
		Contexts: []ContextWrap{
			{"dev", Context{"cluster_2", "red", "user_2"}},
			{"prod", Context{"cluster_1", "blue", "user_1"}},
		},
		Clusters: []ClusterWrap{
			{"cluster_1", Cluster{"https://foo.com", "ca1"}},
			{"cluster_2", Cluster{"https://bar.com", "ca2"}},
		},
		Users: []UserWrap{
			{"user_1", User{"cert1", "key1"}},
			{"user_2", User{"cert2", "key2"}},
		},
	}

	assert.Equal(t, expected, actual)
}
