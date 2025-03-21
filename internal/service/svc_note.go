package service

/**
* FileCreateRequestParams
* @Description        文件创建请求参数
* @Create             HaierKeys 2025-03-01 17:30
* @Param              Credentials  string  表单字段，凭证，必填
* @Param              Password     string  表单字段，密码，必填
 */
type FileCreateRequestParams struct {
	Credentials string `form:"credentials" binding:"required"`
	Password    string `form:"password" binding:"required"`
}

/**
* FileModifyRequestParams
* @Description        文件修改请求参数
* @Create             HaierKeys 2025-03-01 17:30
* @Param              Credentials  string  表单字段，凭证，必填
* @Param              Password     string  表单字段，密码，必填
 */
type FileModifyRequestParams struct {
	Credentials string `form:"credentials" binding:"required"`
	Password    string `form:"password" binding:"required"`
}

/**
* ContentModifyRequestParams
* @Description        文件内容修改请求参数
* @Create             HaierKeys 2025-03-01 17:30
* @Param              Credentials  string  表单字段，凭证，必填
* @Param              Password     string  表单字段，密码，必填
 */
type ContentModifyRequestParams struct {
	Credentials string `form:"credentials" binding:"required"`
	Password    string `form:"password" binding:"required"`
}

/**
* FileDeleteRequestParams
* @Description        文件删除请求参数
* @Create             HaierKeys 2025-03-01 17:30
* @Param              Credentials  string  表单字段，凭证，必填
* @Param              Password     string  表单字段，密码，必填
 */
type FileDeleteRequestParams struct {
	Credentials string `form:"credentials" binding:"required"`
	Password    string `form:"password" binding:"required"`
}

/**
* FileCreate
* @Description        创建文件
* @Create             HaierKeys 2025-03-01 17:30
* @Param              params  *FileCreateRequestParams  文件创建请求参数
* @Return             int64  文件ID
* @Return             error  错误信息
 */
func (svc *Service) FileCreate(params *FileCreateRequestParams) (int64, error) {
	return 0, nil
}

/**
* FileModify
* @Description        修改文件
* @Create             HaierKeys 2025-03-01 17:30
* @Param              params  *FileModifyRequestParams  文件修改请求参数
* @Return             int64  文件ID
* @Return             error  错误信息
 */
func (svc *Service) FileModify(params *FileModifyRequestParams) (int64, error) {
	return 0, nil
}

/**
* ContentModify
* @Description        修改文件内容
* @Create             HaierKeys 2025-03-01 17:30
* @Param              params  *ContentModifyRequestParams  文件内容修改请求参数
* @Return             int64  文件ID
* @Return             error  错误信息
 */
func (svc *Service) ContentModify(params *ContentModifyRequestParams) (int64, error) {
	return 0, nil
}

/**
* FileDelete
* @Description        删除文件
* @Create             HaierKeys 2025-03-01 17:30
* @Param              params  *FileDeleteRequestParams  文件删除请求参数
* @Return             int64  文件ID
* @Return             error  错误信息
 */
func (svc *Service) FileDelete(params *FileDeleteRequestParams) (int64, error) {
	return 0, nil
}
