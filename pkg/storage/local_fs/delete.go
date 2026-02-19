package local_fs

import (
	"os"

	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
)

func (p *LocalFS) Delete(fileKey string) error {
	dstFileKey := p.getSavePath() + fileKey
	if fileurl.IsExist(dstFileKey) {
		return os.Remove(dstFileKey)
	}
	return nil
}
