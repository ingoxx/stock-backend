package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/ingoxx/stock-backend/config"
	"github.com/ingoxx/stock-backend/internal/domain"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/utils"
)

type DelReq struct {
	Id uint `json:"id" form:"id" validate:"required"`
}

type DelFileByIdReq struct {
	Id     uint `json:"id" form:"id" validate:"required"`
	FileId uint `json:"file_id" form:"file_id" validate:"required"`
}

type RegisterReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UpdateProblemCategoriesReq struct {
	Cid uint `json:"cid" form:"cid" validate:"required"` // 修复：修改 tag 为 cid
	Pid uint `json:"pid" form:"pid" validate:"required"` // 修复：修改 tag 为 pid
}

type LoginReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// ChangePasswordReq 修改密码请求结构体
type ChangePasswordReq struct {
	Username    string `json:"username" validate:"required"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password" validate:"required"`
}

// ShareCategoryReq 分类定向共享请求结构体
type ShareCategoryReq struct {
	CategoryID    uint   `json:"category_id" validate:"required"`
	TargetUserIDs []uint `json:"target_user_ids"` // 要共享的目标用户 ID 列表，传 [] 代表取消全部共享
}

// ShareProblemReq 问题/文档定向共享请求结构体
type ShareProblemReq struct {
	ProblemID     uint   `json:"problem_id" validate:"required"`
	TargetUserIDs []uint `json:"target_user_ids"` // 要共享的目标用户 ID 列表，传 [] 代表取消全部共享
}

type DocHandler struct {
	svc *service.DocService
	vd  *validator.Validate
}

func NewDocHandler(svc *service.DocService, vd *validator.Validate) *DocHandler {
	return &DocHandler{svc, vd}
}

// getUserID 辅助函数：从 Context (JWT 中间件) 或 Request Header 中提取当前登录的用户 ID
func getUserID(r *http.Request) uint {
	if uid, ok := r.Context().Value("userID").(uint); ok {
		return uid
	}

	if uidStr := r.FormValue("uid"); uidStr != "" {
		if uid, err := strconv.ParseUint(uidStr, 10, 64); err == nil {
			return uint(uid)
		}
	}

	return 0
}

// RegisterHandler 用户注册 Handler
func (dh *DocHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[RegisterReq](body, dh.vd, nil)
	if err != nil {
		writeReqError(w, err)
		return
	}

	u := &domain.User{
		Username: req.Username,
		Password: req.Password, // 此时密码已被正确赋值
	}

	user, err := dh.svc.RegisterUser(u)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "注册成功",
		Data: user,
	})
}

// LoginHandler 用户登录 Handler
func (dh *DocHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{Code: 1001, Msg: err.Error(), Data: ""})
		return
	}

	req, err := bindAndValidate[LoginReq](body, dh.vd, nil)
	if err != nil {
		writeReqError(w, err)
		return
	}

	// 1. 校验账号密码
	user, err := dh.svc.LoginUser(req.Username, req.Password)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	// 2. 登录成功，生成 JWT Token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  "生成 Token 失败: " + err.Error(),
			Data: "",
		})
		return
	}

	// 3. 返回 Token 和 用户基本信息 给前端
	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "登录成功",
		Data: map[string]interface{}{
			"token": token, // 前端保存此 token
			"user":  user,
			"uid":   user.ID,
		},
	})
}

// CreateCategoriesHandler 创建/更新分类 (绑定当前操作用户 ID)
func (dh *DocHandler) CreateCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[domain.Category](body, dh.vd, func(r *domain.Category) {})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := dh.svc.CreateCategories(userID, req)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "ok",
		Data: data,
	})
}

// CreateProblemsHandler 创建/更新问题 (绑定当前操作用户 ID 并记录编辑人)
func (dh *DocHandler) CreateProblemsHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[domain.Problem](body, dh.vd, func(r *domain.Problem) {})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := dh.svc.CreateProblems(userID, req)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "ok",
		Data: data,
	})
}

// GetProblemsHandler 分页获取问题列表 (含权限隔离)
func (dh *DocHandler) GetProblemsHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	queryParams := r.URL.Query()
	page := queryParams.Get("page")
	if page == "" {
		page = "1" // 容错处理：不传默认第 1 页
	}

	categoryId := queryParams.Get("category_id")
	if categoryId == "" {
		page = "1" // 容错处理：不传默认第 1 页
	}

	p, err := strconv.Atoi(page)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	cid, err := strconv.Atoi(categoryId)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	keyword := queryParams.Get("keyword")

	data, total, err := dh.svc.GetProblems(userID, uint(cid), keyword, p)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "ok",
		Data: utils.PageData{
			List:  data,
			Total: total,
		},
	})
}

// GetCategoriesHandler 分页获取分类列表 (含权限隔离)
func (dh *DocHandler) GetCategoriesHandler(w http.ResponseWriter, r *http.Request) {

	userID := getUserID(r)

	queryParams := r.URL.Query()
	page := queryParams.Get("page")
	if page == "" {
		page = "1"
	}

	p, err := strconv.Atoi(page)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	data, total, err := dh.svc.GetCategories(userID, p)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "ok",
		Data: utils.PageData{
			List:  data,
			Total: total,
		},
	})
}

// DeleteCategoryHandler 删除分类 (仅创建者可删除)
func (dh *DocHandler) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[DelReq](body, dh.vd, func(r *DelReq) {})
	if err != nil {
		writeReqError(w, err)
		return
	}

	if err := dh.svc.DeleteCategory(req.Id, userID); err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "ok",
		Data: "",
	})
}

// DeleteProblemHandler 删除问题/文档 (仅创建者可删除)
func (dh *DocHandler) DeleteProblemHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[DelReq](body, dh.vd, func(r *DelReq) {})
	if err != nil {
		writeReqError(w, err)
		return
	}

	if err := dh.svc.DeleteProblem(req.Id, userID); err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "ok",
		Data: "",
	})
}

// UpdateProblemCategoryHandler 修改指定问题归属的分类
func (dh *DocHandler) UpdateProblemCategoryHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[UpdateProblemCategoriesReq](body, dh.vd, func(r *UpdateProblemCategoriesReq) {})
	if err != nil {
		writeReqError(w, err)
		return
	}

	if err := dh.svc.UpdateProblemCategory(req.Pid, req.Cid, userID); err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "ok",
		Data: "",
	})
}

// UploadFileHandler 上传文件附件 (传入上传者 userID)
func (dh *DocHandler) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	if err := r.ParseMultipartForm(config.MaxMemory); err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  "表单解析失败: " + err.Error(),
			Data: "",
		})
		return
	}

	// 获取上传的文件句柄
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  "获取上传文件失败: " + err.Error(),
			Data: "",
		})
		return
	}
	defer file.Close()

	// 解析 problem_id 参数
	problemIDStr := r.FormValue("problem_id")
	problemID, err := strconv.ParseUint(problemIDStr, 10, 64)
	if err != nil || problemID == 0 {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  "缺少或无效的 problem_id",
			Data: "",
		})
		return
	}

	// 调用 service 层保存文件并写入数据库
	data, err := dh.svc.UploadFile(uint(problemID), userID, fileHeader.Filename, file)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "ok",
		Data: data,
	})
}

// DeleteFileHandler 删除指定问题下的所有附件
func (dh *DocHandler) DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[DelFileByIdReq](body, dh.vd, func(r *DelFileByIdReq) {})
	if err != nil {
		writeReqError(w, err)
		return
	}

	if err := dh.svc.DeleteFilesByProblemID(req.Id, req.FileId); err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  "删除文件失败: " + err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "ok",
		Data: nil,
	})
}

// ChangePasswordHandler 修改密码 Handler
func (dh *DocHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {

	//userID := getUserID(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{Code: 1001, Msg: err.Error(), Data: ""})
		return
	}

	req, err := bindAndValidate[ChangePasswordReq](body, dh.vd, nil)
	if err != nil {
		writeReqError(w, err)
		return
	}

	// 调用 service 修改密码
	if err := dh.svc.ChangePassword(req.Username, req.OldPassword, req.NewPassword); err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "密码修改成功，请重新登录",
		Data: nil,
	})
}

// GetUserListHandler 分页获取用户列表 Handler
func (dh *DocHandler) GetUserListHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	pageStr := queryParams.Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  "页码格式错误",
			Data: "",
		})
		return
	}

	// 调用 service 层
	list, total, err := dh.svc.GetUserList(page)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	// 返回标准分页响应结构
	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "ok",
		Data: utils.PageData{
			List:  list,
			Total: total,
		},
	})
}

// ShareCategoryToUsersHandler 定向共享分类给指定用户 (或取消共享)
func (dh *DocHandler) ShareCategoryToUsersHandler(w http.ResponseWriter, r *http.Request) {
	operatorID := getUserID(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  "读取请求体失败: " + err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[ShareCategoryReq](body, dh.vd, nil)
	if err != nil {
		writeReqError(w, err)
		return
	}

	// 容错处理：如果前端传 nil，转为空切片 []uint{} 以清空共享
	if req.TargetUserIDs == nil {
		req.TargetUserIDs = []uint{}
	}

	// 调用 service/repo 层设置定向共享
	if err := dh.svc.ShareCategoryToUsers(req.CategoryID, operatorID, req.TargetUserIDs); err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "分类共享设置成功",
		Data: nil,
	})
}

// ShareProblemToUsersHandler 定向共享文档给指定用户 (或取消共享)
func (dh *DocHandler) ShareProblemToUsersHandler(w http.ResponseWriter, r *http.Request) {
	operatorID := getUserID(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  "读取请求体失败: " + err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[ShareProblemReq](body, dh.vd, nil)
	if err != nil {
		writeReqError(w, err)
		return
	}

	if req.TargetUserIDs == nil {
		req.TargetUserIDs = []uint{}
	}

	// 调用 service/repo 层设置定向共享
	if err := dh.svc.ShareProblemToUsers(req.ProblemID, operatorID, req.TargetUserIDs); err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, utils.Response{
		Code: 1000,
		Msg:  "文档共享设置成功",
		Data: nil,
	})
}

// DownloadFileHandler 文件下载 Handler
func (dh *DocHandler) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	// 1. 获取 URL Query 参数中的 file_id (如 GET /v1/download-file?file_id=1001)
	fileIDStr := r.URL.Query().Get("file_id")
	if fileIDStr == "" {
		fileIDStr = r.URL.Query().Get("id")
	}

	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil || fileID == 0 {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  "缺少或无效的 file_id 参数",
			Data: "",
		})
		return
	}

	// 2. 调用 Service 获取文件本地物理路径和原文件名（内部已做权限与存在性校验）
	localPath, originalName, err := dh.svc.GetFileForDownload(uint(fileID), userID)
	if err != nil {
		utils.ResponseJSON(w, utils.Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	// 3. 对中文文件名做 URL Encode 编码，防止 Header 中的中文导致乱码或浏览器报错
	encodedFilename := url.QueryEscape(originalName)

	// 4. 设置 HTTP 响应头，指示浏览器触发弹窗下载（Content-Disposition: attachment）
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", encodedFilename, encodedFilename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// 5. 使用 Go 原生 http.ServeFile 将文件以高效二进制流的形式返回给客户端
	http.ServeFile(w, r, localPath)
}
