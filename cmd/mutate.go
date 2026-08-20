// Copyright 2026 Dafa
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cmd

import "svault/internal/storage"

// mutateVault performs a read-modify-write on the vault while holding an
// exclusive lock for the whole operation. fn receives the decrypted vault,
// mutates it in place, and may return an error to abort the write. Holding the
// lock across the read and the write prevents two concurrent svault processes
// from clobbering each other's changes.
//
// The vault is only written back if fn returns nil.
func mutateVault(fn func(vd *storage.VaultData) error) error {
	vpath, err := vaultPath()
	if err != nil {
		return err
	}
	dir, err := vaultDir()
	if err != nil {
		return err
	}
	encKey, err := storage.LoadSession()
	if err != nil {
		return err
	}
	return storage.WithVaultLock(dir, func() error {
		vd, err := storage.ReadVault(vpath, encKey)
		if err != nil {
			return err
		}
		if err := fn(vd); err != nil {
			return err
		}
		return storage.WriteVault(vpath, encKey, vd)
	})
}
