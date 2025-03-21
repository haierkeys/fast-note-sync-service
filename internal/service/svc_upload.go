package service

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"

	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/code"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/convert"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/fileurl"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/storage"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/avif"
	"github.com/pkg/errors"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
	"gorm.io/gorm"
)

type FileInfo struct {
	ImageTitle string   `json:"imageTitle"`
	ImageUrl   string   `json:"imageUrl"`
	UseStore   []string `json:"useStore"`
}

// 上传文件
type ClientUploadParams struct {
	Key    string `form:"key"`
	Type   string `form:"type"`
	Width  int    `form:"width"`
	Height int    `form:"height"`
}

// UploadFile 上传文件
func (svc *Service) UploadFile(file multipart.File, fileHeader *multipart.FileHeader, params *ClientUploadParams) (*FileInfo, error) {

	// 上传文件名
	var fileName = fileurl.GetFileNameOrRandom(fileHeader.Filename)

	// 检查文件后缀
	if !fileurl.IsContainExt(fileurl.ImageType, fileName, global.Config.App.UploadAllowExts) {
		return nil, errors.New("file suffix is not supported.")
	}
	// 检查文件大小
	if fileurl.IsFileSizeAllowed(fileurl.ImageType, file, global.Config.App.UploadMaxSize) {
		return nil, errors.New("exceeded maximum file limit.")
	}

	var fileKey = fileurl.GetDatePath(global.Config.App.UploadDatePath) + fileName
	var fileType = fileHeader.Header.Get("Content-Type")
	var dstFileKey string

	// 压缩
	writer, fileKey, fileType, err := imageResize(params, file, fileKey, fileType)
	if err != nil {
		return nil, err
	}

	var reader = bytes.NewReader(writer.Bytes())

	useStore := []string{}
	for sType := range storage.StorageTypeMap {
		config := map[string]any{}
		if sType == storage.LOCAL {
			_ = convert.StructToMap(global.Config.LocalFS, config)
		} else {
			continue
		}
		if !config["IsEnabled"].(bool) {
			continue
		}
		ins, err := storage.NewClient(sType, config)
		if err != nil {
			return nil, err
		}

		dstFileKey, err = ins.SendFile(fileKey, reader, fileType)
		if err != nil {
			return nil, err
		}
		useStore = append(useStore, sType)
	}
	accessUrl := fileurl.PathSuffixCheckAdd(global.Config.App.UploadUrlPre, "/") + fileurl.UrlEscape(dstFileKey)

	return &FileInfo{ImageTitle: fileHeader.Filename, ImageUrl: accessUrl, UseStore: useStore}, nil
}

func (svc *Service) UserUploadFile(uid int64, file multipart.File, fileHeader *multipart.FileHeader, params *ClientUploadParams) (*FileInfo, error) {

	// 上传文件名
	var fileName = fileurl.GetFileNameOrRandom(fileHeader.Filename)

	// 检查文件后缀
	if !fileurl.IsContainExt(fileurl.ImageType, fileName, global.Config.App.UploadAllowExts) {
		return nil, errors.New("file suffix is not supported.")
	}
	// 检查文件大小
	if fileurl.IsFileSizeAllowed(fileurl.ImageType, file, global.Config.App.UploadMaxSize) {
		return nil, errors.New("exceeded maximum file limit.")
	}

	var fileKey = fileurl.GetDatePath(global.Config.App.UploadDatePath) + fileName
	var fileType = fileHeader.Header.Get("Content-Type")
	var dstFileKey string

	// 压缩
	writer, fileKey, fileType, err := imageResize(params, file, fileKey, fileType)
	if err != nil {
		return nil, err
	}

	var reader = bytes.NewReader(writer.Bytes())

	var userCloudConfig = map[string]any{}
	daoCloudConfig, err := svc.dao.GetEnableByUId(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorUserNotFound
		}
		return nil, err
	}

	userCloudConfig = convert.StructToMapByReflect(daoCloudConfig)

	// 检查云存储类型是否启用
	if err := storage.IsUserEnabled(daoCloudConfig.Type); err != nil {
		return nil, err
	}

	userCloudConfig["SavePath"] = global.Config.LocalFS.SavePath

	ins, err := storage.NewClient(daoCloudConfig.Type, userCloudConfig)
	if err != nil {
		return nil, err
	}

	dstFileKey, err = ins.SendFile(fileKey, reader, fileType)
	if err != nil {
		return nil, err
	}

	useStore := []string{daoCloudConfig.Type}

	accessUrl := fileurl.PathSuffixCheckAdd(userCloudConfig["AccessUrlPrefix"].(string), "/") + fileurl.UrlEscape(dstFileKey)

	return &FileInfo{ImageTitle: fileHeader.Filename, ImageUrl: accessUrl, UseStore: useStore}, nil
}

// imageResize 压缩图片
// 默认裁剪 | 居中裁剪 | 固定尺寸拉伸 | 固定尺寸等比缩放不裁切 | 不处理
// type: "fill-topleft" | "fill-center" | "resize" | "fit" | "none";
func imageResize(params *ClientUploadParams, file multipart.File, fileKey string, fileType string) (*bytes.Buffer, string, string, error) {

	var writer = &bytes.Buffer{}
	// 压缩
	_, err := file.Seek(0, 0)
	if err != nil {
		return nil, fileKey, fileType, err
	}

	img, fileRealType, err := image.Decode(file)

	if err != nil {
		return nil, fileKey, fileType, err
	}

	var imgSize = img.Bounds().Size()

	// 服务器强制限制图片的宽度和高度
	var imageMaxWidth = global.Config.App.ImageMaxSizeWidth
	var imageMaxHeight = global.Config.App.ImageMaxSizeHeight
	var newWidth, newHeight int
	var newImage image.Image
	var isNewImage bool

	if params.Type == "none" || params.Type == "" {
		newWidth = imageMaxWidth
		newHeight = imageMaxHeight
		if (imgSize.X != newWidth || imgSize.Y != newHeight) && (newWidth != 0 || newHeight != 0) {
			if newWidth == 0 || newHeight == 0 {
				newImage = imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
			} else {
				newImage = imaging.Fit(img, newWidth, newHeight, imaging.Lanczos)
			}
			isNewImage = true
		}
	} else if params.Type == "fill-topleft" {
		if params.Width < imageMaxWidth || imageMaxWidth == 0 {
			newWidth = params.Width
		} else {
			newWidth = imageMaxWidth
		}
		if params.Height < imageMaxHeight || imageMaxHeight == 0 {
			newHeight = params.Height
		} else {
			newHeight = imageMaxHeight
		}
		newImage = imaging.Fill(img, newWidth, newHeight, imaging.TopLeft, imaging.Lanczos)
		isNewImage = true
	} else if params.Type == "fill-center" {
		if params.Width < imageMaxWidth || imageMaxWidth == 0 {
			newWidth = params.Width
		} else {
			newWidth = imageMaxWidth
		}
		if params.Height < imageMaxHeight || imageMaxHeight == 0 {
			newHeight = params.Height
		} else {
			newHeight = imageMaxHeight
		}
		// newImage = imaging.Fit(img, newWidth, newHeight, imaging.Lanczos)
		newImage = imaging.Fill(img, newWidth, newHeight, imaging.Center, imaging.Lanczos)
		isNewImage = true
	} else if params.Type == "resize" {
		if params.Width < imageMaxWidth || imageMaxWidth == 0 {
			newWidth = params.Width
		} else {
			newWidth = imageMaxWidth
		}
		if params.Height < imageMaxHeight || imageMaxHeight == 0 {
			newHeight = params.Height
		} else {
			newHeight = imageMaxHeight
		}
		if params.Width != 0 && params.Height != 0 && (imgSize.X != newWidth || imgSize.Y != newHeight) {
			newImage = imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
			isNewImage = true
		}
	} else if params.Type == "fit" {
		if params.Width < imageMaxWidth || imageMaxWidth == 0 {
			newWidth = params.Width
		} else {
			newWidth = imageMaxWidth
		}
		if params.Height < imageMaxHeight || imageMaxHeight == 0 {
			newHeight = params.Height
		} else {
			newHeight = imageMaxHeight
		}
		if (imgSize.X != newWidth || imgSize.Y != newHeight) && (newWidth != 0 || newHeight != 0) {
			if newWidth == 0 || newHeight == 0 {
				newImage = imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
			} else {
				newImage = imaging.Fit(img, newWidth, newHeight, imaging.Lanczos)
			}
			isNewImage = true
		}
	}
	if isNewImage {
		// 调整图片大小
		switch fileRealType {
		case "png":
			err = png.Encode(writer, newImage)
		case "gif":
			err = gif.Encode(writer, newImage, &gif.Options{NumColors: 256})
		case "jpeg", "jpg":
			err = jpeg.Encode(writer, newImage, &jpeg.Options{Quality: global.Config.App.ImageQuality})
		case "bmp":
			err = bmp.Encode(writer, newImage)
		case "tif", "tiff":
			err = tiff.Encode(writer, newImage, nil)
		case "webp":
			// 暂时不支持
			fileType = "image/jpg"
			ext := fileurl.GetFileExt(fileKey)
			fileKey = fileKey[0:len(fileKey)-len(ext)] + ".jpg"
			err = jpeg.Encode(writer, newImage, &jpeg.Options{Quality: global.Config.App.ImageQuality})
		case "avif":
			err = avif.Encode(writer, newImage, avif.Options{Quality: global.Config.App.ImageQuality})
		default:
			return nil, fileKey, fileType, errors.New("Unknown image type:" + fileRealType)
		}
		if err != nil {
			return nil, fileKey, fileType, err
		}
	} else {
		_, err = file.Seek(0, 0)
		if err != nil {
			return nil, fileKey, fileType, err
		}
		_, err = io.Copy(writer, file)
		if err != nil {
			return nil, fileKey, fileType, err
		}
	}
	return writer, fileKey, fileType, nil
}
