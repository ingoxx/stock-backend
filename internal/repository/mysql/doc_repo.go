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

// GetProblems 分页获取问题列表（支持：分类筛选 + 关键字搜索 + 权限隔离 + 返回总条数）
func (dr *DocRepo) GetProblems(userID uint, categoryID uint, keyword string, page int) ([]*domain.Problem, int64, error) {
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// 1. 基础查询条件：权限隔离 (只能看自己创建的 OR 别人共享的)
	query := dr.db.Model(&domain.Problem{}).Where("(creator_id = ? OR is_shared = ?)", userID, true)

	// 2. 动态拼接条件一：如果传了 categoryID（大于 0），追加分类筛选
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	// 3. 动态拼接条件二：如果传了关键字（非空），同时在标题(title)或解答(solution)中模糊匹配
	keyword = strings.TrimSpace(keyword) // 去除首尾空格
	if keyword != "" {
		pattern := "%" + keyword + "%"
		// 加括号确保优先级：AND (title LIKE %kw% OR solution LIKE %kw%)
		query = query.Where("(title LIKE ? OR solution LIKE ?)", pattern, pattern)
	}

	// 4. 统计满足上述组合条件下的总条数 total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 5. 分页查出当前页的数据列表
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

// DeleteCategory 删除分类（使用事务：仅创建者可删除，并同步清理关联问题、中间表关联和物理文件）
func (dr *DocRepo) DeleteCategory(categoryID uint, userID uint) error {
	return dr.db.Transaction(func(tx *gorm.DB) error {
		// 1. 权限校验：检查分类是否存在且是否为当前用户创建
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

		// 3. 如果分类下存在问题，执行级联清理
		if len(problemIDs) > 0 {
			// 3.1 查出文件记录并清理物理磁盘文件
			var files []domain.FileItem
			_ = tx.Where("problem_id IN ?", problemIDs).Find(&files).Error

			if err := tx.Where("problem_id IN ?", problemIDs).Delete(&domain.FileItem{}).Error; err != nil {
				return err
			}

			// 清理磁盘上的物理文件
			for _, file := range files {
				uniqueFilename := filepath.Base(file.URL)
				localPath := filepath.Join(saveDir, uniqueFilename)
				_ = os.Remove(localPath)
			}

			// 3.2 关键修复：先删除多对多中间表 problem_editors 中的关联外键记录
			if err := tx.Exec("DELETE FROM problem_editors WHERE problem_id IN ?", problemIDs).Error; err != nil {
				return err
			}

			// 3.3 再删除问题 Problem 记录
			if err := tx.Where("category_id = ?", categoryID).Delete(&domain.Problem{}).Error; err != nil {
				return err
			}
		}

		// 4. 最后删除分类 Category 本身
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

		// 4.1 先删除附件记录 (file_items)
		if err := tx.Where("problem_id = ?", problemID).Delete(&domain.FileItem{}).Error; err != nil {
			return err
		}

		// 4.2 显式清理多对多中间表 (problem_editors)，彻底杜绝 MySQL 1451 外键约束报错
		if err := tx.Exec("DELETE FROM problem_editors WHERE problem_id = ?", problemID).Error; err != nil {
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
func (dr *DocRepo) DeleteFilesByProblemID(problemID, fileID uint) error {
	// 1. 根据 fileID 和 problemID 双重条件精准查询该文件记录
	var file domain.FileItem
	err := dr.db.Where("id = ? AND problem_id = ?", fileID, problemID).First(&file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文件记录不存在或不属于当前文档")
		}
		return err
	}

	// 2. 从 MySQL 数据库中删除这一条记录
	if err := dr.db.Delete(&file).Error; err != nil {
		return err
	}

	// 3. 从服务器磁盘上删除对应的物理文件
	uniqueFilename := filepath.Base(file.URL)
	localPath := filepath.Join(saveDir, uniqueFilename)
	_ = os.Remove(localPath) // 忽略文件可能已被手动删除的错误

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

// ChangePassword 修改用户密码
func (dr *DocRepo) ChangePassword(username string, oldPassword, newPassword string) error {
	var user domain.User
	// 1. 查找当前用户（按用户名）
	if err := dr.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	//// 2. 校验旧密码是否正确
	//if !utils.CheckPassword(user.Password, oldPassword) {
	//	return errors.New("旧密码不正确")
	//}

	// 3. 对新密码进行哈希加密
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("新密码加密失败: " + err.Error())
	}

	// 4. 更新数据库中的密码与修改时间
	return dr.db.Model(&user).Updates(map[string]interface{}{
		"password":   hashedPassword,
		"updated_at": time.Now(),
	}).Error
}
