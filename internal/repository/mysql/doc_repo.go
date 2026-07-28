package mysql

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ingoxx/stock-backend/internal/domain"
	"github.com/ingoxx/stock-backend/utils"
	"gorm.io/gorm"
)

const (
	pageSize = 5
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

// CreateCategories 保存或更新分类（仅创建者可修改分类名称等核心属性）
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

	// 更新分类：权限校验（仅创建者 CreatorID == userID 允许修改）
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

// CreateProblems 保存或更新问题（创建者、全员公开、定向共享用户均允许编辑，并自动记录编辑者列表）
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

	// 更新文档：权限校验（必须是创建者 OR 开启了全员共享 OR 被定向单独共享）
	var oldProblem domain.Problem
	if err := dr.db.First(&oldProblem, data.ID).Error; err != nil {
		return nil, fmt.Errorf("文档不存在: %w", err)
	}

	// 检查当前用户是否被定向共享
	var isTargetShared int64
	_ = dr.db.Table("problem_shares").Where("problem_id = ? AND user_id = ?", data.ID, userID).Count(&isTargetShared)

	if oldProblem.CreatorID != userID && !oldProblem.IsShared && isTargetShared == 0 {
		return nil, errors.New("无权编辑此文档：您未获得此文档的编辑权限")
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

// GetProblems 分页获取问题列表（支持：分类筛选 + 关键字搜索 + [创建者/全员公开/定向共享]权限隔离 + 返回总条数）
func (dr *DocRepo) GetProblems(userID uint, categoryID uint, keyword string, page int) ([]*domain.Problem, int64, error) {
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// 1. 查出单独定向共享给当前用户的 Problem ID 列表
	var sharedProblemIDs []uint
	_ = dr.db.Table("problem_shares").Where("user_id = ?", userID).Pluck("problem_id", &sharedProblemIDs)

	// 2. 基础查询条件：我是创建者 OR 对全员公开 OR 单独共享给了我
	query := dr.db.Model(&domain.Problem{}).Where(
		"creator_id = ? OR is_shared = ? OR id IN ?",
		userID, true, sharedProblemIDs,
	)

	// 3. 动态拼接：分类筛选
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	// 4. 动态拼接：关键字模糊匹配（标题或内容）
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		pattern := "%" + keyword + "%"
		query = query.Where("(title LIKE ? OR solution LIKE ?)", pattern, pattern)
	}

	// 5. 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 6. 分页获取数据列表
	var dp []*domain.Problem
	err := query.
		Preload("Category").
		Preload("Files").
		Preload("Files.Uploader").
		Preload("Creator").
		Preload("UpdatedBy").
		Preload("Editors").
		Preload("SharedUsers"). // 预加载被单独共享的用户列表
		Limit(pageSize).
		Offset(offset).
		Find(&dp).Error

	if err != nil {
		return nil, 0, err
	}

	return dp, total, nil
}

// GetCategories 获取分类列表（支持：[创建者/全员公开/定向共享]权限隔离 + 返回总记录数 total）
func (dr *DocRepo) GetCategories(userID uint, page int) ([]domain.Category, int64, error) {
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// 1. 查出单独定向共享给当前用户的 Category ID 列表
	var sharedCategoryIDs []uint
	_ = dr.db.Table("category_shares").Where("user_id = ?", userID).Pluck("category_id", &sharedCategoryIDs)

	// 2. 查询条件：我是创建者 OR 对全员公开 OR 单独共享给了我
	query := dr.db.Model(&domain.Category{}).Where(
		"creator_id = ? OR is_shared = ? OR id IN ?",
		userID, true, sharedCategoryIDs,
	)

	// 3. 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 4. 查出定向共享给当前用户的 Problem ID 列表（用于嵌套预加载分类下的问题）
	var sharedProblemIDs []uint
	_ = dr.db.Table("problem_shares").Where("user_id = ?", userID).Pluck("problem_id", &sharedProblemIDs)

	// 5. 分页查询分类
	var ds []domain.Category
	err := query.
		Preload("Creator").
		Preload("UpdatedBy").
		Preload("SharedUsers"). // 预加载分类的共享用户列表
		Preload("Problems", "creator_id = ? OR is_shared = ? OR id IN ?", userID, true, sharedProblemIDs).
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

// DeleteCategory 删除分类（使用事务：仅创建者可删除，同步清理中间表、关联问题与磁盘物理文件）
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

		// 2. 清理分类自身的共享关联中间表 category_shares
		if err := tx.Exec("DELETE FROM category_shares WHERE category_id = ?", categoryID).Error; err != nil {
			return err
		}

		// 3. 查出该分类下的所有 Problem ID
		var problemIDs []uint
		if err := tx.Model(&domain.Problem{}).Where("category_id = ?", categoryID).Pluck("id", &problemIDs).Error; err != nil {
			return err
		}

		// 4. 如果分类下存在问题，执行关联级联清理
		if len(problemIDs) > 0 {
			// 4.1 查出附件记录并清理物理磁盘文件
			var files []domain.FileItem
			_ = tx.Where("problem_id IN ?", problemIDs).Find(&files).Error

			if err := tx.Where("problem_id IN ?", problemIDs).Delete(&domain.FileItem{}).Error; err != nil {
				return err
			}

			// 删除磁盘上的物理文件
			for _, file := range files {
				uniqueFilename := filepath.Base(file.URL)
				localPath := filepath.Join(saveDir, uniqueFilename)
				_ = os.Remove(localPath)
			}

			// 4.2 清理问题的多对多中间表：problem_editors 与 problem_shares
			if err := tx.Exec("DELETE FROM problem_editors WHERE problem_id IN ?", problemIDs).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM problem_shares WHERE problem_id IN ?", problemIDs).Error; err != nil {
				return err
			}

			// 4.3 删除问题 Problem 记录
			if err := tx.Where("category_id = ?", categoryID).Delete(&domain.Problem{}).Error; err != nil {
				return err
			}
		}

		// 5. 最后删除分类 Category 本身
		if err := tx.Delete(&domain.Category{}, categoryID).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteProblem 删除问题/文档（使用事务：仅创建者可删除，100% 安全清理中间表与物理文件）
func (dr *DocRepo) DeleteProblem(problemID uint, userID uint) error {
	return dr.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查询当前问题
		var problem domain.Problem
		if err := tx.First(&problem, problemID).Error; err != nil {
			return err
		}

		// 2. 权限校验：仅创建者可以删除
		if problem.CreatorID != userID {
			return errors.New("无权删除此文档：只有创建者才能删除")
		}

		// 3. 查出关联的附件记录，以便稍后清理本地磁盘物理文件
		var files []domain.FileItem
		_ = tx.Where("problem_id = ?", problemID).Find(&files).Error

		// 4.1 删除附件记录 (file_items)
		if err := tx.Where("problem_id = ?", problemID).Delete(&domain.FileItem{}).Error; err != nil {
			return err
		}

		// 4.2 显式清理多对多中间表 (problem_editors 和 problem_shares)
		if err := tx.Exec("DELETE FROM problem_editors WHERE problem_id = ?", problemID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM problem_shares WHERE problem_id = ?", problemID).Error; err != nil {
			return err
		}

		// 4.3 删除 Problem 主记录
		if err := tx.Delete(&problem).Error; err != nil {
			return err
		}

		// 5. 数据库清理成功后，删除磁盘上的物理文件
		for _, file := range files {
			uniqueFilename := filepath.Base(file.URL)
			localPath := filepath.Join(saveDir, uniqueFilename)
			_ = os.Remove(localPath)
		}

		return nil
	})
}

// UpdateProblemCategory 修改指定 Problem 所属分类（必须是创建者、公开状态或授权共享者）
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

	var isTargetShared int64
	_ = dr.db.Table("problem_shares").Where("problem_id = ? AND user_id = ?", problemID, userID).Count(&isTargetShared)

	if problem.CreatorID != userID && !problem.IsShared && isTargetShared == 0 {
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

	// 2. 生成唯一文件名
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

	// 4. 构造 FileItem 实体并存入数据库
	fileItem := domain.FileItem{
		ProblemID:  problemID,
		UploaderID: uploaderID,
		Name:       fileName,
		URL:        fmt.Sprintf("%s/%s", url, uniqueFilename),
	}

	if err := dr.db.Create(&fileItem).Error; err != nil {
		_ = os.Remove(savePath)
		return nil, fmt.Errorf("文件记录保存到数据库失败: %w", err)
	}

	return &fileItem, nil
}

// DeleteFilesByProblemID 精准删除特定问题下的某个附件文件
func (dr *DocRepo) DeleteFilesByProblemID(problemID, fileID uint) error {
	var file domain.FileItem
	err := dr.db.Where("id = ? AND problem_id = ?", fileID, problemID).First(&file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文件记录不存在或不属于当前文档")
		}
		return err
	}

	if err := dr.db.Delete(&file).Error; err != nil {
		return err
	}

	uniqueFilename := filepath.Base(file.URL)
	localPath := filepath.Join(saveDir, uniqueFilename)
	_ = os.Remove(localPath)

	return nil
}

// RegisterUser 用户注册
func (dr *DocRepo) RegisterUser(user *domain.User) (*domain.User, error) {
	var count int64
	if err := dr.db.Model(&domain.User{}).Where("username = ?", user.Username).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, errors.New("密码加密失败: " + err.Error())
	}
	user.Password = hashedPassword

	if err := dr.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// LoginUser 用户登录校验
func (dr *DocRepo) LoginUser(username, password string) (*domain.User, error) {
	var user domain.User

	if err := dr.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	if !utils.CheckPassword(user.Password, password) {
		return nil, errors.New("用户名或密码错误")
	}

	return &user, nil
}

// ChangePassword 修改用户密码
func (dr *DocRepo) ChangePassword(username string, oldPassword, newPassword string) error {
	var user domain.User
	if err := dr.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("新密码加密失败: " + err.Error())
	}

	return dr.db.Model(&user).Updates(map[string]interface{}{
		"password":   hashedPassword,
		"updated_at": time.Now(),
	}).Error
}

// ShareCategoryToUsers 设置分类的定向共享用户列表（若传入空切片 targetUserIDs，则代表取消全部共享）
func (dr *DocRepo) ShareCategoryToUsers(categoryID uint, operatorID uint, targetUserIDs []uint) error {
	var category domain.Category
	if err := dr.db.First(&category, categoryID).Error; err != nil {
		return fmt.Errorf("分类不存在: %w", err)
	}

	if category.CreatorID != operatorID {
		return errors.New("无权修改此分类的共享权限：只有创建者可以配置")
	}

	var targetUsers []domain.User
	for _, uid := range targetUserIDs {
		targetUsers = append(targetUsers, domain.User{ID: uid})
	}

	// GORM Replace 会用新的用户列表替换原有的映射。若 targetUsers 为空切片，将清理全量关联（实现取消共享）
	return dr.db.Model(&category).Association("SharedUsers").Replace(targetUsers)
}

// ShareProblemToUsers 设置文档的定向共享用户列表（若传入空切片 targetUserIDs，则代表取消全部共享）
func (dr *DocRepo) ShareProblemToUsers(problemID uint, operatorID uint, targetUserIDs []uint) error {
	var problem domain.Problem
	if err := dr.db.First(&problem, problemID).Error; err != nil {
		return fmt.Errorf("文档不存在: %w", err)
	}

	if problem.CreatorID != operatorID {
		return errors.New("无权修改此文档的共享权限：只有创建者可以配置")
	}

	var targetUsers []domain.User
	for _, uid := range targetUserIDs {
		targetUsers = append(targetUsers, domain.User{ID: uid})
	}

	// GORM Replace 自动更新/替换/清空中间表 problem_shares 的映射记录
	return dr.db.Model(&problem).Association("SharedUsers").Replace(targetUsers)
}
