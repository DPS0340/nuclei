package installer

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyncTemplateOwnershipDirectory(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, root.Close()) })
	require.NoError(t, syncTemplateOwnershipDirectory(root, "."))
}

func TestSyncTemplateOwnershipFileRestoresModeAfterOpenFailure(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, root.Close()) })
	require.NoError(t, root.WriteFile("restored.yaml", []byte("restored"), 0o444))
	originalInfo, err := root.Stat("restored.yaml")
	require.NoError(t, err)
	openErr := errors.New("open failed")

	err = syncTemplateOwnershipFileWithOpen(root, "restored.yaml", originalInfo.Mode(), func() (*os.File, error) {
		return nil, openErr
	})
	require.ErrorIs(t, err, openErr)
	info, err := root.Stat("restored.yaml")
	require.NoError(t, err)
	require.Equal(t, originalInfo.Mode().Perm(), info.Mode().Perm())
}

func TestRenameTemplateRestoreNoReplaceUsesOpenedRoot(t *testing.T) {
	parent := t.TempDir()
	rootPath := filepath.Join(parent, "templates")
	require.NoError(t, os.Mkdir(rootPath, 0o755))
	root, err := os.OpenRoot(rootPath)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, root.Close()) })

	movedPath := filepath.Join(parent, "moved-templates")
	require.NoError(t, os.Rename(rootPath, movedPath))
	require.NoError(t, os.Mkdir(rootPath, 0o755))
	require.NoError(t, root.WriteFile("temporary.yaml", []byte("restored"), 0o600))

	require.NoError(t, renameTemplateRestoreNoReplace(root, "temporary.yaml", "retired.yaml"))
	require.FileExists(t, filepath.Join(movedPath, "retired.yaml"))
	require.NoFileExists(t, filepath.Join(rootPath, "retired.yaml"))
}

func TestSyncTemplateOwnershipFile(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, root.Close()) })
	require.NoError(t, root.WriteFile("restored.yaml", []byte("restored"), 0o444))
	require.NoError(t, syncTemplateOwnershipFile(root, "restored.yaml", 0o444))
	info, err := root.Stat("restored.yaml")
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o444), info.Mode().Perm())
}
