package postgres

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/vele/temp_test_repo/internal/domain"
	"github.com/vele/temp_test_repo/internal/repository"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := db.AutoMigrate(&UserModel{}, &FileModel{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	return &Repository{db: db}, nil
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (r *Repository) List(ctx context.Context) ([]domain.User, error) {
	var models []UserModel
	if err := r.db.WithContext(ctx).Preload("Files").Find(&models).Error; err != nil {
		return nil, err
	}
	users := make([]domain.User, len(models))
	for i := range models {
		users[i] = models[i].toDomain()
	}
	return users, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Preload("Files").First(&model, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	user := model.toDomain()
	return &user, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	user := model.toDomain()
	return &user, nil
}

func (r *Repository) Create(ctx context.Context, user *domain.User) error {
	model := fromDomain(user)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	*user = model.toDomain()
	return nil
}

func (r *Repository) Update(ctx context.Context, user *domain.User) error {
	model := fromDomain(user)
	model.ID = user.ID
	target := &UserModel{Model: gorm.Model{ID: user.ID}}
	if err := r.db.WithContext(ctx).Model(target).Updates(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&UserModel{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uint) ([]domain.File, error) {
	var models []FileModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, err
	}
	files := make([]domain.File, len(models))
	for i := range models {
		files[i] = models[i].toDomain()
	}
	return files, nil
}

func (r *Repository) Add(ctx context.Context, file *domain.File) error {
	model := fileModelFromDomain(file)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	*file = model.toDomain()
	return nil
}

func (r *Repository) DeleteByUser(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&FileModel{}).Error
}

func (r *Repository) Truncate(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Exec("TRUNCATE TABLE file_models RESTART IDENTITY CASCADE").Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Exec("TRUNCATE TABLE user_models RESTART IDENTITY CASCADE").Error
}

type UserModel struct {
	gorm.Model
	Name  string
	Email string `gorm:"uniqueIndex"`
	Age   int
	Files []FileModel `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
}

func (u UserModel) toDomain() domain.User {
	files := make([]domain.File, len(u.Files))
	for i := range u.Files {
		files[i] = u.Files[i].toDomain()
	}
	return domain.User{
		ID:        uint(u.ID),
		Name:      u.Name,
		Email:     u.Email,
		Age:       u.Age,
		Files:     files,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func fromDomain(u *domain.User) UserModel {
	return UserModel{
		Model: gorm.Model{
			ID:        uint(u.ID),
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		},
		Name:  u.Name,
		Email: u.Email,
		Age:   u.Age,
	}
}

type FileModel struct {
	gorm.Model
	UserID uint
	Name   string
	Path   string
}

func (f FileModel) toDomain() domain.File {
	return domain.File{
		ID:        uint(f.ID),
		UserID:    f.UserID,
		Name:      f.Name,
		Path:      f.Path,
		CreatedAt: f.CreatedAt,
	}
}

func fileModelFromDomain(f *domain.File) FileModel {
	return FileModel{
		Model: gorm.Model{
			ID:        uint(f.ID),
			CreatedAt: f.CreatedAt,
		},
		UserID: f.UserID,
		Name:   f.Name,
		Path:   f.Path,
	}
}

var _ repository.UserRepository = (*Repository)(nil)
var _ repository.FileRepository = (*Repository)(nil)
