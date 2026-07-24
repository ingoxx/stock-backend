package mysql

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ingoxx/stock-backend/internal/domain"
	"gorm.io/gorm"
)

const (
	pageSize = 10
	saveDir  = "/opt/uploads/profile3"
	url      = "https://ai.anythingai.online/static/profile3"
)

type DocRepo struct {
	db *gorm.DB
}

func NewDocRepo(db *gorm.DB) domain.DocRepository {
	return &DocRepo{
		db: db,
	}
}

// CreateCategories 保存分类（ID 为 0 时新增，ID > 0 时更新）
func (dr *DocRepo) CreateCategories(data domain.Category) (domain.Category, error) {
	if data.ID == 0 && data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	data.UpdatedAt = time.Now()

	db := dr.db
	if data.ID > 0 {
		db = db.Omit("created_at")
	}

	if err := db.Save(&data).Error; err != nil {
		return domain.Category{}, err
	}
	return data, nil
}

// CreateProblems 保存问题（支持新增与更新，并修复时间零值报错）
func (dr *DocRepo) CreateProblems(data domain.Problem) (*domain.Problem, error) {
	// 1. 如果是新增（ID == 0），且前端没有传创建时间，手动补全为当前系统时间
	if data.ID == 0 && data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	// 每次保存都更新修改时间
	data.UpdatedAt = time.Now()

	// 2. 开启 Session 准备保存
	db := dr.db.Session(&gorm.Session{FullSaveAssociations: true})

	// 3. 如果是更新操作（ID > 0），跳过 created_at 字段，防止被 0000-00-00 覆盖
	if data.ID > 0 {
		db = db.Omit("created_at")
	}

	// 4. 执行保存
	if err := db.Save(&data).Error; err != nil {
		return nil, err
	}

	return &data, nil
}

// GetProblems 获取问题列表（带分页 + 关联分类 + 关联文件列表）
func (dr *DocRepo) GetProblems(page int) ([]*domain.Problem, error) {
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var dp []*domain.Problem
	err := dr.db.
		Preload("Category"). // 预加载分类
		Preload("Files"). // 预加载当前问题的附件列表
		Limit(pageSize).
		Offset(offset).
		Find(&dp).Error

	if err != nil {
		return nil, err
	}

	return dp, nil
}

// GetCategories 获取分类列表（带分页 + 关联问题 + 关联问题的附件）
func (dr *DocRepo) GetCategories(page int) ([]domain.Category, error) {
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var ds []domain.Category
	err := dr.db.
		Preload("Problems"). // 预加载分类下的问题
		Preload("Problems.Files"). // 嵌套预加载：同时把问题的附件列表也查出来
		Limit(pageSize).
		Offset(offset).
		Find(&ds).Error

	if err != nil {
		return nil, err
	}

	return ds, nil
}

// DeleteCategory 删除分类（使用事务：同时删除分类下的所有 Problem 及其关联的 FileItem 附件）
func (dr *DocRepo) DeleteCategory(id uint) error {
	// 使用事务保证全成功或全失败（原子性）
	return dr.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查询该分类下所有 Problem 的 ID
		var problemIDs []uint
		if err := tx.Model(&domain.Problem{}).Where("category_id = ?", id).Pluck("id", &problemIDs).Error; err != nil {
			return err
		}

		// 2. 如果存在问题，先清理这些问题绑定的文件附件 FileItem
		if len(problemIDs) > 0 {
			if err := tx.Where("problem_id IN ?", problemIDs).Delete(&domain.FileItem{}).Error; err != nil {
				return err
			}

			// 3. 删除该分类下的所有 Problem
			if err := tx.Where("category_id = ?", id).Delete(&domain.Problem{}).Error; err != nil {
				return err
			}
		}

		// 4. 最后删除分类 Category 本身
		if err := tx.Delete(&domain.Category{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteProblem 删除问题（同时级联删除该问题下的所有 FileItem 附件）
func (dr *DocRepo) DeleteProblem(id uint) error {
	// 通过 Select("Files")，GORM 会自动在删除 Problem 前先把关联的 FileItem 记录全删掉
	return dr.db.Select("Files").Delete(&domain.Problem{ID: id}).Error
}

// UpdateProblemCategory 修改指定 Problem 所属的分类
func (dr *DocRepo) UpdateProblemCategory(problemID uint, newCategoryID uint) error {
	// 1. (可选校验) 检查目标分类是否存在
	var categoryCount int64
	if err := dr.db.Model(&domain.Category{}).Where("id = ?", newCategoryID).Count(&categoryCount).Error; err != nil {
		return err
	}
	if categoryCount == 0 {
		return errors.New("目标分类不存在")
	}

	// 2. 更新 Problem 的 category_id
	res := dr.db.Model(&domain.Problem{}).
		Where("id = ?", problemID).
		Update("category_id", newCategoryID)

	if res.Error != nil {
		return res.Error
	}

	// 3. 检查是否有记录被更新
	if res.RowsAffected == 0 {
		return errors.New("修改失败：问题不存在或分类未发生改变")
	}

	return nil
}

// UploadFile 将文件保存至磁盘并写入 MySQL 数据库
func (dr *DocRepo) UploadFile(problemID uint, fileName string, src io.Reader) (*domain.FileItem, error) {
	// 1. 创建本地保存目录
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("创建文件保存目录失败: %w", err)
	}

	// 2. 生成唯一文件名，防止重名覆盖
	uniqueFilename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileName)
	savePath := filepath.Join(saveDir, uniqueFilename)

	// 3. 创建磁盘文件并写入数据
	dst, err := os.Create(savePath)
	if err != nil {
		return nil, fmt.Errorf("创建磁盘文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("写入磁盘文件失败: %w", err)
	}

	// 4. 构造 FileItem 实体并存入 MySQL 数据库
	fileItem := domain.FileItem{
		ProblemID: problemID,
		Name:      fileName,
		URL:       fmt.Sprintf("%s/%s", url, uniqueFilename),
	}

	if err := dr.db.Create(&fileItem).Error; err != nil {
		// 容错处理：如果写入数据库失败，删除刚刚存入磁盘的文件，防止产生无用文件
		_ = os.Remove(savePath)
		return nil, fmt.Errorf("文件记录保存到数据库失败: %w", err)
	}

	return &fileItem, nil
}

// DeleteFilesByProblemID 删除指定问题下的所有附件文件（同步清理数据库记录与磁盘文件）
func (dr *DocRepo) DeleteFilesByProblemID(problemID uint) error {
	// 1. 先查询出该问题关联的所有 FileItem，获取物理文件 URL
	var files []domain.FileItem
	if err := dr.db.Where("problem_id = ?", problemID).Find(&files).Error; err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	// 2. 从 MySQL 中删除对应记录
	if err := dr.db.Where("problem_id = ?", problemID).Delete(&domain.FileItem{}).Error; err != nil {
		return err
	}

	// 3. 异步/同步删除服务器磁盘上的物理文件
	for _, file := range files {
		// file.URL 形如 "/uploads/1710500000_test.png"
		// 拼成本地文件路径： "./uploads/1710500000_test.png"
		localPath := filepath.Join(saveDir, file.Name)
		_ = os.Remove(localPath) // 忽略可能出现的文件已被手动删掉的错误
	}

	return nil
}
