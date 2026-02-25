package aliyun_oss

import (
	"path"
)

func (p *OSS) Delete(fileKey string) error {
	if p.Bucket == nil {
		err := p.GetBucket("")
		if err != nil {
			return err
		}
	}
	fileKey = path.Join(p.Config.CustomPath, fileKey)
	return p.Bucket.DeleteObject(fileKey)
}
