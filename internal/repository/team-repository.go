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
		GetMembers(teamID uint) ([]models.User, error)
		GetByID(id uint) (*models.Team, error)
		AddMember(teamID uint, userID int64) error
		RemoveMember(teamID uint, userID int64) error
		SaveUserTeam(req models.UserTeams) error
		GetTeamsByOwnerID(ownerID int64) ([]*models.Team, error)
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

func (r *teamRepository) GetMembers(teamID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Joins("JOIN user_teams ON user_teams.user_id = users.id").
		Where("user_teams.team_id = ?", teamID).
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

func (r *teamRepository) AddMember(teamID uint, userID int64) error {
	userTeam := models.UserTeams{
		UserID: userID,
		TeamID: teamID,
	}
	return r.db.Create(&userTeam).Error
}

func (r *teamRepository) RemoveMember(teamID uint, userID int64) error {
	return r.db.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&models.UserTeams{}).Error
}

func (r *teamRepository) SaveUserTeam(req models.UserTeams) error {
	return r.db.Create(&req).Error
}

func (r *teamRepository) GetTeamsByOwnerID(ownerID int64) ([]*models.Team, error) {
	var teams []*models.Team
	if err := r.db.Where("owner_id = ?", ownerID).Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}
