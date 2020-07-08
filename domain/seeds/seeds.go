package seeds

import (
	"fmt"
	"github.com/bxcodec/faker"
	"github.com/jinzhu/gorm"
	"go-rest-skeleton/domain/entity"
	"go-rest-skeleton/infrastructure/seed"
)

func All() []seed.Seed {
	b := Prepare()
	return b
}

func Prepare() []seed.Seed {
	fakerFactories := make([]seed.Seed, 5)
	for i := 0; i < 5; i++ {
		a := entity.UserFaker{}
		err := faker.FakeData(&a)
		if err != nil {
			fmt.Println(err)
		}
		fakerFactories[i] = seed.Seed{
			Name: fmt.Sprintf("Create %s", a.FirstName),
			Run: func(db *gorm.DB) error {
				err := CreateUser(db, a.UUID, a.FirstName, a.LastName, a.Email, a.Phone, a.Password)
				return err
			},
		}
	}
	return fakerFactories
} 