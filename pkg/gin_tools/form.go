// // Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// // Use of this source code is governed by a MIT style
// // license that can be found in the LICENSE file.
package gin_tools

//
//import (
//	"errors"
//	"net/httpclient"
//)
//
//const defaultMemory = 32 << 20
//
//type formBinding struct{}
//type formMultipartBinding struct{}
//
//func (formBinding) Name() string {
//	return "form"
//}
//
//func (formBinding) Parse(req *httpclient.Request, params *map[string]string) error {
//	if err := req.ParseForm(); err != nil {
//		return err
//	}
//	if err := req.ParseMultipartForm(defaultMemory); err != nil && !errors.Is(err, httpclient.ErrNotMultipart) {
//		return err
//	}
//
//	if err := req.ParseForm(); err != nil {
//		return err
//	}
//	if err := req.ParseMultipartForm(defaultMemory); err != nil {
//		if err != httpclient.ErrNotMultipart {
//			return err
//		}
//	}
//	postMap := *params
//	for k, v := range req.PostForm {
//		if len(v) > 1 {
//			postMap[k] = v
//		} else if len(v) == 1 {
//			postMap[k] = v[0]
//		}
//	}
//	//if err := mapForm(obj, req.Form); err != nil {
//	//	return nil, err
//	//}
//	return nil
//}
//
//func (formMultipartBinding) Name() string {
//	return "multipart/form-data"
//}
//
//func (formMultipartBinding) Parse(req *httpclient.Request, params *map[string]string) error {
//	if err := req.ParseMultipartForm(defaultMemory); err != nil {
//		return err
//	}
//	//if err := mappingByPtr(obj, (*multipartRequest)(req), "form"); err != nil {
//	//	return nil, err
//	//}
//
//	return nil
//
//}
