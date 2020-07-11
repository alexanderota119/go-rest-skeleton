package persistence

import (
	"fmt"
	"go-rest-skeleton/domain/entity"
	"go-rest-skeleton/domain/repository"
	"go-rest-skeleton/domain/seeds"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"    // for mysql driver (optional)
	_ "github.com/jinzhu/gorm/dialects/postgres" // for postgres driver (optional)
)

// Repositories represent it self.
type Repositories struct {
	Permission repository.PermissionRepository
	Role       repository.RoleRepository
	User       repository.UserRepository
	db         *gorm.DB
}

// NewRepositories will initialize db connection and return repositories.
func NewRepositories(dbDriver, dbUser, dbPassword, dbHost, dbName, dbPort string) (*Repositories, error) {
	dbURL := ""
	switch dbDriver {
	case "postgres":
		dbURL = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
			dbHost,
			dbPort,
			dbUser,
			dbName,
			dbPassword,
		)
	case "mysql":
		dbURL = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
			dbUser,
			dbPassword,
			dbHost,
			dbPort,
			dbName,
		)
	}

	db, err := gorm.Open(dbDriver, dbURL)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)

	return &Repositories{
		Permission: NewPermissionRepository(db),
		Role:       NewRoleRepository(db),
		User:       NewUserRepository(db),
		db:         db,
	}, nil
}

// Close will closes the database connection.
func (s *Repositories) Close() error {
	return s.db.Close()
}

// AutoMigrate will migrate all tables.
func (s *Repositories) AutoMigrate() error {
	return s.db.AutoMigrate(
		&entity.Permission{},
		&entity.Role{},
		&entity.RolePermission{},
		&entity.User{},
		&entity.UserRole{},
	).Error
}

// Seeds all seeders.
func (s *Repositories) Seeds() error {
	db := s.db
	var err error
	for _, seed := range seeds.All() {
		errSeed := seed.Run(db)
		if errSeed != nil {
			err = errSeed
			log.Fatalf("Running seed '%s', failed with error: %s", seed.Name, err)
		}
	}

	return err
}
