//go:build managedtoken

package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestManagedTokenEmail_UsesTokenName(t *testing.T) {
	got := managedTokenEmail("  示例令牌  ")

	require.Equal(t, "示例令牌@tokens.local", got)
}

func TestManagedTokenEmail_TruncatesToSchemaLimit(t *testing.T) {
	got := managedTokenEmail(strings.Repeat("令", 120))

	require.Len(t, []rune(got), managedTokenUsernameMaxLen)
	require.True(t, strings.HasSuffix(got, "@tokens.local"))
}

func TestManagedTokenUsername_UsesTokenName(t *testing.T) {
	got := managedTokenUsername("  示例令牌  ")

	require.Equal(t, "示例令牌@tokens.local", got)
}

func TestManagedTokenUsername_TruncatesToSchemaLimit(t *testing.T) {
	got := managedTokenUsername(strings.Repeat("令", 120))

	require.Len(t, []rune(got), managedTokenUsernameMaxLen)
	require.True(t, strings.HasSuffix(got, "@tokens.local"))
}

func TestIsManagedTokenUser_AcceptsNewUsernameFormat(t *testing.T) {
	user := &User{
		Email:    "a1b2c3d4e5f6@tokens.local",
		Username: "示例令牌@tokens.local",
		Notes:    buildManagedTokenNotes("示例令牌", ""),
	}

	require.True(t, IsManagedTokenUser(user))
}

func TestIsManagedTokenUser_AcceptsLegacyUsernameFormat(t *testing.T) {
	user := &User{
		Email:    "a1b2c3d4e5f6@tokens.local",
		Username: "token_a1b2c3d4e5f6",
		Notes:    buildManagedTokenNotes("示例令牌", ""),
	}

	require.True(t, IsManagedTokenUser(user))
}

func TestIsManagedTokenUser_RejectsMissingMarker(t *testing.T) {
	user := &User{
		Email:    "a1b2c3d4e5f6@tokens.local",
		Username: "示例令牌@tokens.local",
		Notes:    "plain note",
	}

	require.False(t, IsManagedTokenUser(user))
}
