package mysql

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ingoxx/stock-backend/internal/domain"
	"github.com/ingoxx/stock-backend/utils"
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

// CreateCategories / SaveCategory 保存或更新分类（严格限制：仅创建者可修改）
func (dr *DocRepo) CreateCategories(userID uint, data domain.Category) (domain.Category, error) {
	if data.ID == 0 {
		// 新增分类：绑定创建者与更新者
		data.CreatorID = userID
		data.UpdatedByID = userID
		if data.CreatedAt.IsZero() {
			data.CreatedAt = time.Now()
		}
		data.UpdatedAt = time.Now()

		if err := dr.db.Create(&data).Error; err != nil {
			return domain.Category{}, err
		}
		return data, nil
	}

	// 更新分类：权限校验（仅创建者 CreatorID == userID 允许更新名称）
	var oldCategory domain.Category
	if err := dr.db.First(&oldCategory, data.ID).Error; err != nil {
		return domain.Category{}, fmt.Errorf("分类不存在: %w", err)
	}

	if oldCategory.CreatorID != userID {
		return domain.Category{}, errors.New("无权修改此分类：只有分类的创建者才能修改分类")
	}

	data.CreatorID = oldCategory.CreatorID // 保持原始创建人不变
	data.UpdatedByID = userID
	data.UpdatedAt = time.Now()

	if err := dr.db.Omit("created_at").Save(&data).Error; err != nil {
		return domain.Category{}, err
	}

	return data, nil
}

// CreateProblems / SaveProblem 保存或更新问题（共享状态允许协同编辑，并自动记录编辑者列表）
func (dr *DocRepo) CreateProblems(userID uint, data domain.Problem) (*domain.Problem, error) {
	if data.ID == 0 {
		// 新增文档
		data.CreatorID = userID
		data.UpdatedByID = userID
		if data.CreatedAt.IsZero() {
			data.CreatedAt = time.Now()
		}
		data.UpdatedAt = time.Now()

		err := dr.db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&data).Error
		if err != nil {
			return nil, err
		}

		// 自动将创建者记录入多对多编辑者列表
		_ = dr.db.Model(&data).Association("Editors").Append(&domain.User{ID: userID})
		return &data, nil
	}

	// 更新文档：权限校验（必须是创建者 OR 该文档启用了共享）
	var oldProblem domain.Problem
	if err := dr.db.First(&oldProblem, data.ID).Error; err != nil {
		return nil, fmt.Errorf("文档不存在: %w", err)
	}

	if oldProblem.CreatorID != userID && !oldProblem.IsShared {
		return nil, errors.New("无权编辑此文档：该文档未开启共享")
	}

	data.CreatorID = oldProblem.CreatorID // 保持原始创建者不变
	data.UpdatedByID = userID             // 标记最新修改人
	data.UpdatedAt = time.Now()

	// 保存修改
	err := dr.db.Session(&gorm.Session{FullSaveAssociations: true}).
		Omit("created_at").
		Save(&data).Error

	if err != nil {
		return nil, err
	}

	// 自动将当前编辑用户追加到多对多编辑者列表中 (GORM 自动去重)
	_ = dr.db.Model(&data).Association("Editors").Append(&domain.User{ID: userID})

	return &data, nil
}

// GetProblems 获取问题列表（带权限隔离 + 返回总记录数 total）
func (dr *DocRepo) GetProblems(userID uint, page int) ([]*domain.Problem, int64, error) {
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// 构建基础查询条件（权限隔离：只能看自己创建的 OR 别人共享的）
	query := dr.db.Model(&domain.Problem{}).Where("creator_id = ? OR is_shared = ?", userID, true)

	// 1. 先统计满足条件的记录总数 total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 2. 再执行 Preload 和 Limit/Offset 分页查询数据列表
	var dp []*domain.Problem
	err := query.
		Preload("Category").
		Preload("Files").
		Preload("Files.Uploader").
		Preload("Creator").
		Preload("UpdatedBy").
		Preload("Editors").
		Limit(pageSize).
		Offset(offset).
		Find(&dp).Error

	if err != nil {
		return nil, 0, err
	}

	return dp, total, nil
}

// GetCategories 获取分类列表（带权限隔离 + 返回总记录数 total）
func (dr *DocRepo) GetCategories(userID uint, page int) ([]domain.Category, int64, error) {
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	query := dr.db.Model(&domain.Category{}).Where("creator_id = ? OR is_shared = ?", userID, true)

	// 1. 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 2. 分页查询
	var ds []domain.Category
	err := query.
		Preload("Creator").
		Preload("UpdatedBy").
		Preload("Problems", "creator_id = ? OR is_shared = ?", userID, true).
		Preload("Problems.Files").
		Preload("Problems.Editors").
		Limit(pageSize).
		Offset(offset).
		Find(&ds).Error

	if err != nil {
		return nil, 0, err
	}

	return ds, total, nil
}

// DeleteCategory 删除分类（使用事务：仅创建者可删除，并同步清理关联问题和物理文件）
func (dr *DocRepo) DeleteCategory(categoryID uint, userID uint) error {
	return dr.db.Transaction(func(tx *gorm.DB) error {
		// 1. 权限校验
		var category domain.Category
		if err := tx.First(&category, categoryID).Error; err != nil {
			return err
		}
		if category.CreatorID != userID {
			return errors.New("无权删除此分类：只有分类创建者才能删除")
		}

		// 2. 查出该分类下的所有 Problem ID
		var problemIDs []uint
		if err := tx.Model(&domain.Problem{}).Where("category_id = ?", categoryID).Pluck("id", &problemIDs).Error; err != nil {
			return err
		}

		// 3. 删除分类下问题的附件记录与磁盘物理文件
		if len(problemIDs) > 0 {
			var files []domain.FileItem
			_ = tx.Where("problem_id IN ?", problemIDs).Find(&files).Error

			if err := tx.Where("problem_id IN ?", problemIDs).Delete(&domain.FileItem{}).Error; err != nil {
				return err
			}

			// 清理磁盘上的附件物理文件
			for _, file := range files {
				uniqueFilename := filepath.Base(file.URL)
				localPath := filepath.Join(saveDir, uniqueFilename)
				_ = os.Remove(localPath)
			}

			// 删除问题记录
			if err := tx.Where("category_id = ?", categoryID).Delete(&domain.Problem{}).Error; err != nil {
				return err
			}
		}

		// 4. 删除分类本身
		if err := tx.Delete(&domain.Category{}, categoryID).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteProblem 删除问题（仅创建者可删除，同步级联清理附件及磁盘物理文件）
func (dr *DocRepo) DeleteProblem(problemID uint, userID uint) error {
	var problem domain.Problem
	if err := dr.db.First(&problem, problemID).Error; err != nil {
		return err
	}

	if problem.CreatorID != userID {
		return errors.New("无权删除此文档：只有创建者才能删除")
	}

	// 查出该问题的所有附件，准备清理磁盘文件
	var files []domain.FileItem
	_ = dr.db.Where("problem_id = ?", problemID).Find(&files).Error

	// 删除数据库中的 Problem 记录以及关联的中间表关联
	err := dr.db.Select("Files", "Editors").Delete(&problem).Error
	if err != nil {
		return err
	}

	// 清理磁盘上的物理文件
	for _, file := range files {
		uniqueFilename := filepath.Base(file.URL)
		localPath := filepath.Join(saveDir, uniqueFilename)
		_ = os.Remove(localPath)
	}

	return nil
}

// UpdateProblemCategory 修改指定 Problem 所属分类（要求必须是创建者或开启了共享）
func (dr *DocRepo) UpdateProblemCategory(problemID uint, newCategoryID uint, userID uint) error {
	// 1. 检查目标分类是否存在
	var category domain.Category
	if err := dr.db.First(&category, newCategoryID).Error; err != nil {
		return errors.New("目标分类不存在")
	}

	// 2. 检查 Problem 权限
	var problem domain.Problem
	if err := dr.db.First(&problem, problemID).Error; err != nil {
		return errors.New("文档不存在")
	}

	if problem.CreatorID != userID && !problem.IsShared {
		return errors.New("无权修改此文档的分类")
	}

	// 3. 执行分类变更并记录最新修改人
	res := dr.db.Model(&domain.Problem{}).
		Where("id = ?", problemID).
		Updates(map[string]interface{}{
			"category_id":   newCategoryID,
			"updated_by_id": userID,
		})

	if res.Error != nil {
		return res.Error
	}

	// 4. 追加当前用户到编辑者关联
	_ = dr.db.Model(&problem).Association("Editors").Append(&domain.User{ID: userID})

	return nil
}

// UploadFile 将文件保存至磁盘并写入 MySQL 数据库（记录上传者 uploaderID）
func (dr *DocRepo) UploadFile(problemID uint, uploaderID uint, fileName string, src io.Reader) (*domain.FileItem, error) {
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

	// 4. 构造 FileItem 实体并存入 MySQL 数据库（记录上传人 UploaderID）
	fileItem := domain.FileItem{
		ProblemID:  problemID,
		UploaderID: uploaderID,
		Name:       fileName,
		URL:        fmt.Sprintf("%s/%s", url, uniqueFilename),
	}

	if err := dr.db.Create(&fileItem).Error; err != nil {
		// 容错处理：写入数据库失败时删除对应的磁盘文件
		_ = os.Remove(savePath)
		return nil, fmt.Errorf("文件记录保存到数据库失败: %w", err)
	}

	return &fileItem, nil
}

// DeleteFilesByProblemID 删除指定问题下的所有附件文件（已修正磁盘物理文件路径匹配 Bug）
func (dr *DocRepo) DeleteFilesByProblemID(problemID uint) error {
	// 1. 先查询出关联的所有 FileItem 记录
	var files []domain.FileItem
	if err := dr.db.Where("problem_id = ?", problemID).Find(&files).Error; err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	// 2. 从 MySQL 中删除记录
	if err := dr.db.Where("problem_id = ?", problemID).Delete(&domain.FileItem{}).Error; err != nil {
		return err
	}

	// 3. 删除磁盘上的物理文件 (从 URL 中提取唯一的磁盘文件名)
	for _, file := range files {
		uniqueFilename := filepath.Base(file.URL)
		localPath := filepath.Join(saveDir, uniqueFilename)
		_ = os.Remove(localPath)
	}

	return nil
}

// RegisterUser 用户注册（密码加盐哈希后落库）
func (dr *DocRepo) RegisterUser(user *domain.User) (*domain.User, error) {
	// 1. 检查用户名是否已存在
	var count int64
	if err := dr.db.Model(&domain.User{}).Where("username = ?", user.Username).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 2. 对明文密码进行不可逆加密（哈希散列）
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, errors.New("密码加密失败: " + err.Error())
	}
	user.Password = hashedPassword

	// 3. 写入数据库
	if err := dr.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// LoginUser 用户登录校验
func (dr *DocRepo) LoginUser(username, password string) (*domain.User, error) {
	var user domain.User

	// 1. 根据用户名查找用户
	if err := dr.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	// 2. 比对明文密码与加密哈希
	if !utils.CheckPassword(user.Password, password) {
		return nil, errors.New("用户名或密码错误")
	}

	return &user, nil
}
