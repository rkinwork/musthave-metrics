package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONFileSaver_Save(t *testing.T) {
	testCases := []struct {
		Name     string
		Metrics  []Metrics
		FilePath string
		IsError  bool
	}{
		{
			Name:     "successful save",
			Metrics:  []Metrics{NewCounterMetrics("test", 1), NewCounterMetrics("test", 2)},
			FilePath: "testfile.json",
		},
		{
			Name:     "empty metrics",
			Metrics:  []Metrics{},
			FilePath: "testfile.json",
		},
		{
			Name:     "invalid file path",
			Metrics:  []Metrics{NewCounterMetrics("test", 1), NewCounterMetrics("test", 2)},
			FilePath: "/not/existent/path.json",
			IsError:  true,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			defer func(name string) {
				_ = os.Remove(name)

			}(test.FilePath) // clean up

			saver := &JSONFileSaver{FilePath: test.FilePath}

			err := saver.Save(test.Metrics)
			if test.IsError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				restoredMetrics, loadErr := saver.Load()
				require.NoError(t, loadErr)

				// If we load the same data we previously saved, they should match
				require.Equal(t, test.Metrics, restoredMetrics)
			}
		})
	}
}
