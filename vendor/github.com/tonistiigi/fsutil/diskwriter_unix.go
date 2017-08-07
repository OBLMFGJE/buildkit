// +build !windows

package fsutil

import (
	"os"
	"syscall"

	"github.com/pkg/errors"
	"github.com/stevvooe/continuity/sysx"
)

func rewriteMetadata(p string, stat *Stat) error {
	for key, value := range stat.Xattrs {
		sysx.Setxattr(p, key, value, 0)
	}

	if err := os.Lchown(p, int(stat.Uid), int(stat.Gid)); err != nil {
		return errors.Wrapf(err, "failed to lchown %s", p)
	}

	if os.FileMode(stat.Mode)&os.ModeSymlink == 0 {
		if err := os.Chmod(p, os.FileMode(stat.Mode)); err != nil {
			return errors.Wrapf(err, "failed to chown %s", p)
		}
	}

	if err := chtimes(p, stat.ModTime); err != nil {
		return errors.Wrapf(err, "failed to chtimes %s", p)
	}

	return nil
}

// handleTarTypeBlockCharFifo is an OS-specific helper function used by
// createTarFile to handle the following types of header: Block; Char; Fifo
func handleTarTypeBlockCharFifo(path string, stat *Stat) error {
	mode := uint32(stat.Mode & 07777)
	if os.FileMode(stat.Mode)&os.ModeCharDevice != 0 {
		mode |= syscall.S_IFCHR
	} else if os.FileMode(stat.Mode)&os.ModeNamedPipe != 0 {
		mode |= syscall.S_IFIFO
	} else {
		mode |= syscall.S_IFBLK
	}

	if err := syscall.Mknod(path, mode, int(mkdev(stat.Devmajor, stat.Devminor))); err != nil {
		return err
	}
	return nil
}
