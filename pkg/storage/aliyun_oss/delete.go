package aliyun_oss

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
)

func (p *OSS) Delete(fileKey string) error {
	if p.Bucket == nil {
		err := p.GetBucket("")
		if err != nil {
			return err
		}
	}
	fileKey = fileurl.PathSuffixCheckAdd(p.Config.CustomPath, "/") + fileKey
	return p.Bucket.DeleteObject(fileKey)
}
