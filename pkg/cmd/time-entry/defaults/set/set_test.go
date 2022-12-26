package set_test

import (
	"testing"

	"github.com/lucassabreu/clockify-cli/internal/mocks"
	"github.com/lucassabreu/clockify-cli/pkg/cmd/time-entry/defaults/set"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdSet_ShouldCreateAndUpdate_DefaultsFile(t *testing.T) {
	tts := []struct{ name string }{}
	for i := range tts {
		tt := &tts[i]
		t.Run(tt.name, func(t *testing.T) {

			cmd := set.NewCmdSet(
				mocks.NewMockFactory(t),
				nil,
			)
			_, err := cmd.ExecuteC()
			assert.Error(t, err)
		})
	}
}
