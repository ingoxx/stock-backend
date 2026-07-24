package handler

import (
	"io"
	"net/http"
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

type UpdateProblemCategoriesReq struct {
	Cid uint `json:"cid" form:"id" validate:"required"`
	Pid uint `json:"pid" form:"id" validate:"required"`
}

type DocHandler struct {
	svc *service.DocService
	vd  *validator.Validate
}

func NewDocHandler(svc *service.DocService, vd *validator.Validate) *DocHandler {
	return &DocHandler{svc, vd}
}

func (dh *DocHandler) CreateCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[domain.Category](body, dh.vd, func(r *domain.Category) {
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := dh.svc.CreateCategories(req)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: data,
	})

}

func (dh *DocHandler) CreateProblemsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[domain.Problem](body, dh.vd, func(r *domain.Problem) {
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	data, err := dh.svc.CreateProblems(req)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: data,
	})
}

func (dh *DocHandler) GetProblemsHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	page := queryParams.Get("page")
	if page == "" {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "required parameter 'page' is missing or empty.",
			Data: "",
		})
		return
	}

	p, err := strconv.Atoi(page)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	data, err := dh.svc.GetProblems(p)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: data,
	})

}

func (dh *DocHandler) GetCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	page := queryParams.Get("page")
	if page == "" {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "required parameter 'page' is missing or empty.",
			Data: "",
		})
		return
	}

	p, err := strconv.Atoi(page)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	data, err := dh.svc.GetCategories(p)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: data,
	})
}

func (dh *DocHandler) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[DelReq](body, dh.vd, func(r *DelReq) {
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	if err := dh.svc.DeleteCategory(req.Id); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: "",
	})
}

func (dh *DocHandler) DeleteProblemHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[DelReq](body, dh.vd, func(r *DelReq) {
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	if err := dh.svc.DeleteProblem(req.Id); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: "",
	})
}

func (dh *DocHandler) UpdateProblemCategoryHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[UpdateProblemCategoriesReq](body, dh.vd, func(r *UpdateProblemCategoriesReq) {
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	if err := dh.svc.UpdateProblemCategory(req.Pid, req.Cid); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: "",
	})
}

func (dh *DocHandler) UploadFileHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(config.MaxMemory); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "表单解析失败: " + err.Error(),
			Data: "",
		})
		return
	}

	// 2. 获取上传的文件句柄
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "获取上传文件失败: " + err.Error(),
			Data: "",
		})
		return
	}
	defer file.Close()

	// 3. 解析 problem_id 参数
	problemIDStr := r.FormValue("problem_id")
	problemID, err := strconv.ParseUint(problemIDStr, 10, 64)
	if err != nil || problemID == 0 {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "缺少或无效的 problem_id",
			Data: "",
		})
		return
	}

	// 4. 调用底层保存文件并写入 MySQL
	data, err := dh.svc.UploadFile(uint(problemID), fileHeader.Filename, file)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	// 5. 响应成功
	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: data,
	})
}

// DeleteFileHandler 删除单张附件 Handler
func (dh *DocHandler) DeleteFileHandler(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[DelReq](body, dh.vd, func(r *DelReq) {
	})
	if err != nil {
		writeReqError(w, err)
		return
	}

	// 调用 repo/service 删除文件
	if err := dh.svc.DeleteFilesByProblemID(req.Id); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  "删除文件失败: " + err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: nil,
	})
}
