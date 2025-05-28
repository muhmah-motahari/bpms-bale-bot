package repository

import (
	"bbb/internal/models"

	"gorm.io/gorm"
)

type (
	GroupRepository interface {
		Save(req models.Group) error
		GetAll() ([]models.Group, error)
		GetByJoinKey(joinKey string) (*models.Group, error)
		GetMembers(groupID uint) ([]models.User, error)
		GetByID(id uint) (*models.Group, error)
		AddMember(groupID uint, userID uint) error
		RemoveMember(groupID uint, userID uint) error
		SaveUserGroup(req models.UserGroups) error
	}

	groupRepository struct {
		db *gorm.DB
	}
)

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{
		db: db,
	}
}

func (r *groupRepository) Save(req models.Group) error {
	if req.ID == 0 {
		if err := r.db.Create(&req).Error; err != nil {
			return err
		}
	} else {
		if err := r.db.Save(&req).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *groupRepository) GetAll() ([]models.Group, error) {
	var groups []models.Group
	if err := r.db.Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *groupRepository) GetByJoinKey(joinKey string) (*models.Group, error) {
	var group models.Group
	if err := r.db.Where("join_key = ?", joinKey).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) GetMembers(groupID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Joins("JOIN user_groups ON user_groups.user_id = users.id").
		Where("user_groups.group_id = ?", groupID).
		Find(&users).Error
	return users, err
}

func (r *groupRepository) GetByID(id uint) (*models.Group, error) {
	var group models.Group
	if err := r.db.First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) AddMember(groupID uint, userID uint) error {
	userGroup := models.UserGroups{
		UserID:  userID,
		GroupID: groupID,
	}
	return r.db.Create(&userGroup).Error
}

func (r *groupRepository) RemoveMember(groupID uint, userID uint) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.UserGroups{}).Error
}

func (r *groupRepository) SaveUserGroup(req models.UserGroups) error {
	return r.db.Create(&req).Error
}
