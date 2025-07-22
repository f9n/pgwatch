package cmdopts

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaintenanceCleanupCommand_Execute(t *testing.T) {
	t.Run("missing sink configuration", func(t *testing.T) {
		os.Args = []string{0: "config_test", "maintenance", "cleanup"}
		_, err := New(nil)
		assert.Error(t, err, "should error when no sink configured")
	})

	t.Run("help output includes maintenance commands", func(t *testing.T) {
		w := &strings.Builder{}
		os.Args = []string{0: "config_test", "maintenance", "--help"}
		_, err := New(w)
		assert.Error(t, err) // help command returns error
		assert.Contains(t, w.String(), "cleanup")
		assert.Contains(t, w.String(), "maintain-unique-sources")
	})
}

func TestMaintenanceMaintainUniqueSourcesCommand_Execute(t *testing.T) {
	t.Run("missing sink configuration", func(t *testing.T) {
		os.Args = []string{0: "config_test", "maintenance", "maintain-unique-sources"}
		_, err := New(nil)
		assert.Error(t, err, "should error when no sink configured")
	})
}
