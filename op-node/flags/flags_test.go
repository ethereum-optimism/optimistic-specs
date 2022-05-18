package flags

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

// TestRequiredFlagsSetRequired asserts that all flags deemed required properly
// have the Required field set to true.
func TestRequiredFlagsSetRequired(t *testing.T) {
	for _, flag := range requiredFlags {
		reqFlag, ok := flag.(cli.RequiredFlag)
		require.True(t, ok)
		require.True(t, reqFlag.IsRequired())
	}
}

// TestOptionalFlagsDontSetRequired asserts that all flags deemed optional set
// the Required field to false.
func TestOptionalFlagsDontSetRequired(t *testing.T) {
	for _, flag := range optionalFlags {
		reqFlag, ok := flag.(cli.RequiredFlag)
		require.True(t, ok)
		require.False(t, reqFlag.IsRequired())
	}
}

// TestUniqueFlags asserts that all flag names are unique, to avoid accidental conflicts between the many flags.
func TestUniqueFlags(t *testing.T) {
	seenCLI := make(map[string]struct{})
	for _, flag := range Flags {
		name := flag.GetName()
		if _, ok := seenCLI[name]; ok {
			t.Errorf("duplicate flag %s", name)
			continue
		}
		seenCLI[name] = struct{}{}
	}
}
