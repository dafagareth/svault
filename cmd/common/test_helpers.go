package common

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"svault/internal/crypto"
	"svault/internal/storage"
)

func InitTestCrypto() {
	crypto.ArgonMemory = 8 * 1024
	crypto.ArgonTime = 1
}

func SetupVault(t *testing.T) {
	t.Helper()
	InitTestCrypto()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SVAULT_NS", "default")

	storage.SetSessionFile(filepath.Join(home, ".session"))

	vpath := filepath.Join(home, ".svault", "vault.enc")
	password := []byte("testpassword")
	if err := storage.InitVault(vpath, password); err != nil {
		t.Fatalf("init vault: %v", err)
	}
	salt, err := storage.ReadSalt(vpath)
	if err != nil {
		t.Fatalf("read salt: %v", err)
	}
	key := crypto.DeriveKey(password, salt)
	if err := storage.SaveSession(key); err != nil {
		t.Fatalf("save session: %v", err)
	}
}

func Reunlock(t *testing.T) {
	t.Helper()
	vpath, err := VaultPath()
	if err != nil {
		t.Fatal(err)
	}
	salt, err := storage.ReadSalt(vpath)
	if err != nil {
		t.Fatal(err)
	}
	key := crypto.DeriveKey([]byte("testpassword"), salt)
	if err := storage.SaveSession(key); err != nil {
		t.Fatal(err)
	}
}

func Seed(t *testing.T, ns, key, value string) {
	t.Helper()
	vpath, err := VaultPath()
	if err != nil {
		t.Fatal(err)
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		t.Fatal(err)
	}
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		t.Fatal(err)
	}
	if vd.Namespaces[ns] == nil {
		vd.Namespaces[ns] = map[string]string{}
	}
	vd.Namespaces[ns][key] = value
	if err := storage.WriteVault(vpath, encKey, vd); err != nil {
		t.Fatal(err)
	}
}

func ReadSecret(t *testing.T, ns, key string) (string, bool) {
	t.Helper()
	vpath, err := VaultPath()
	if err != nil {
		t.Fatal(err)
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		t.Fatal(err)
	}
	vd, err := storage.ReadVault(vpath, encKey)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := vd.Namespaces[ns][key]
	return v, ok
}

func CaptureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()
	fn()
	_ = w.Close()
	os.Stdout = old
	return <-done
}

func NewCmd() *cobra.Command {
	return &cobra.Command{}
}
