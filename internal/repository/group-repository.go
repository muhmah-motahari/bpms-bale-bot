package repository

import (
	"bbb/internal/models"

	"gorm.io/gorm"
)

type (
	TeamRepository interface {
		Save(req models.Team) error
		GetAll() ([]models.Team, error)
		GetByJoinKey(joinKey string) (*models.Team, error)
		GetMembers(groupID uint) ([]models.User, error)
		GetByID(id uint) (*models.Team, error)
		AddMember(groupID uint, userID int64) error
		RemoveMember(groupID uint, userID int64) error
		SaveUserGroup(req models.UserTeams) error
		GetGroupsByOwnerID(ownerID int64) ([]*models.Team, error)
	}

	teamRepository struct {
		db *gorm.DB
	}
)

func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{
		db: db,
	}
}

func (r *teamRepository) Save(req models.Team) error {
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

func (r *teamRepository) GetAll() ([]models.Team, error) {
	var teams []models.Team
	if err := r.db.Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *teamRepository) GetByJoinKey(joinKey string) (*models.Team, error) {
	var team models.Team
	if err := r.db.Where("join_key = ?", joinKey).First(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) GetMembers(groupID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Joins("JOIN user_groups ON user_groups.user_id = users.id").
		Where("user_groups.group_id = ?", groupID).
		Find(&users).Error
	return users, err
}

func (r *teamRepository) GetByID(id uint) (*models.Team, error) {
	var team models.Team
	if err := r.db.First(&team, id).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) AddMember(groupID uint, userID int64) error {
	userGroup := models.UserTeams{
		UserID: userID,
		TeamID: groupID,
	}
	return r.db.Create(&userGroup).Error
}

func (r *teamRepository) RemoveMember(groupID uint, userID int64) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.UserTeams{}).Error
}

func (r *teamRepository) SaveUserGroup(req models.UserTeams) error {
	return r.db.Create(&req).Error
}

func (r *teamRepository) GetGroupsByOwnerID(ownerID int64) ([]*models.Team, error) {
	var teams []*models.Team
	if err := r.db.Where("owner_id = ?", ownerID).Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}
