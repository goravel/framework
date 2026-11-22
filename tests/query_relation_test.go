package tests

import (
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm"
)

func (s *QueryTestSuite) TestRelationExistence() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			withBooks := User{
				Name: "has_user_books",
				Books: []*Book{
					{Name: "has_book_published", Author: &Author{Name: "has_author"}},
					{Name: "has_book_draft"},
				},
				Roles:  []*Role{{Name: "has_role"}},
				Phones: []*Phone{{Name: "has_phone"}},
			}
			s.Nil(query.Query().Select(orm.Associations).Create(&withBooks))

			withoutBooks := User{Name: "has_user_without_books"}
			s.Nil(query.Query().Create(&withoutBooks))

			names := func(relationQuery contractsorm.Query) []string {
				var users []User
				s.Nil(relationQuery.Where("name LIKE ?", "has_user_%").Order("id").Find(&users))

				result := make([]string, 0, len(users))
				for _, user := range users {
					result = append(result, user.Name)
				}

				return result
			}
			users := func() contractsorm.Query {
				return query.Query().Model(&User{})
			}
			bookNamed := func(name string) func(contractsorm.Query) contractsorm.Query {
				return func(query contractsorm.Query) contractsorm.Query {
					return query.Where("name", name)
				}
			}

			// Has many, with and without a count.
			s.Equal([]string{"has_user_books"}, names(users().Has("Books")))
			s.Equal([]string{"has_user_books"}, names(users().Has("Books", ">=", 2)))
			s.Empty(names(users().Has("Books", ">", 2)))
			s.Equal([]string{"has_user_without_books"}, names(users().DoesntHave("Books")))

			// Constrained by a callback, and through a nested relation.
			s.Equal([]string{"has_user_books"}, names(users().WhereHas("Books", bookNamed("has_book_draft"))))
			s.Empty(names(users().WhereHas("Books", bookNamed("has_book_missing"))))
			s.Equal([]string{"has_user_books"}, names(users().WhereHas("Books", bookNamed("has_book_draft"), "=", 1)))
			s.Equal([]string{"has_user_without_books"}, names(users().WhereDoesntHave("Books", bookNamed("has_book_draft"))))
			s.Equal([]string{"has_user_books"}, names(users().WhereHas("Books.Author", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "has_author")
			})))

			// Many to many and polymorphic.
			s.Equal([]string{"has_user_books"}, names(users().Has("Roles")))
			s.Equal([]string{"has_user_books"}, names(users().Has("Phones")))
			s.Equal([]string{"has_user_without_books"}, names(users().DoesntHave("Phones")))

			// Or conditions keep their place among the other conditions.
			s.Equal([]string{"has_user_books", "has_user_without_books"}, names(users().Where(func(query contractsorm.Query) contractsorm.Query {
				return query.Has("Roles").OrDoesntHave("Books")
			})))

			// Belongs to.
			count, err := query.Query().Model(&Book{}).Where("name LIKE ?", "has_book_%").WhereHas("User", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "has_user_books")
			}).Count()
			s.Nil(err)
			s.Equal(int64(2), count)
		})
	}
}
