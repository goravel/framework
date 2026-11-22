package tests

import (
	"context"
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/errors"
)

func (s *QueryTestSuite) TestHas() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "has_user_with_books", Avatar: "has_avatar"},
				{Name: "has_user_without_books", Avatar: "has_avatar"},
			}
			s.Nil(query.Query().Create(&users))
			s.Nil(query.Query().Create(&Book{Name: "has_book", UserID: users[0].ID}))

			var result []User
			s.Nil(query.Query().Where("name LIKE ?", "has_user_%").Has("Books").Find(&result))
			s.Len(result, 1)
			s.Equal(users[0].ID, result[0].ID)
			// Filtering must not eager-load the relationship.
			s.Empty(result[0].Books)

			// The model is inferred from the typed destination.
			result = nil
			s.Nil(query.Query().Has("Books").Where("name LIKE ?", "has_user_%").Get(&result))
			s.Len(result, 1)

			// The model is resolved from the explicit model.
			result = nil
			s.Nil(query.Query().Model(&User{}).Has("Books").Where("name LIKE ?", "has_user_%").Get(&result))
			s.Len(result, 1)

			// A parent with several related records is returned once.
			s.Nil(query.Query().Create(&Book{Name: "has_book_2", UserID: users[0].ID}))
			result = nil
			s.Nil(query.Query().Has("Books").Where("name LIKE ?", "has_user_%").Find(&result))
			s.Len(result, 1)

			// A parent without related records is not returned.
			result = nil
			s.Nil(query.Query().Where("name", "has_user_without_books").Has("Books").Find(&result))
			s.Empty(result)
		})
	}
}

func (s *QueryTestSuite) TestHasCountOperators() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "has_count_user_0", Avatar: "has_count_avatar"},
				{Name: "has_count_user_1", Avatar: "has_count_avatar"},
				{Name: "has_count_user_2", Avatar: "has_count_avatar"},
			}
			s.Nil(query.Query().Create(&users))
			s.Nil(query.Query().Create(&Book{Name: "has_count_book_1", UserID: users[1].ID}))
			books := []Book{
				{Name: "has_count_book_2", UserID: users[2].ID},
				{Name: "has_count_book_2", UserID: users[2].ID},
				{Name: "has_count_book_2", UserID: users[2].ID},
			}
			s.Nil(query.Query().Create(&books))

			tests := []struct {
				name     string
				args     []any
				expected []string
			}{
				{name: "default", args: nil, expected: []string{"has_count_user_1", "has_count_user_2"}},
				{name: "greater or equal one", args: []any{">=", 1}, expected: []string{"has_count_user_1", "has_count_user_2"}},
				{name: "greater than zero", args: []any{">", 0}, expected: []string{"has_count_user_1", "has_count_user_2"}},
				{name: "not equal zero", args: []any{"!=", 0}, expected: []string{"has_count_user_1", "has_count_user_2"}},
				{name: "not equal zero alt", args: []any{"<>", 0}, expected: []string{"has_count_user_1", "has_count_user_2"}},
				{name: "equal zero", args: []any{"=", 0}, expected: []string{"has_count_user_0"}},
				{name: "less or equal zero", args: []any{"<=", 0}, expected: []string{"has_count_user_0"}},
				{name: "equal one", args: []any{"=", 1}, expected: []string{"has_count_user_1"}},
				{name: "equal three", args: []any{"=", 3}, expected: []string{"has_count_user_2"}},
				{name: "less than two", args: []any{"<", 2}, expected: []string{"has_count_user_0", "has_count_user_1"}},
				{name: "less or equal one", args: []any{"<=", 1}, expected: []string{"has_count_user_0", "has_count_user_1"}},
				{name: "greater than one", args: []any{">", 1}, expected: []string{"has_count_user_2"}},
				{name: "greater or equal two", args: []any{">=", 2}, expected: []string{"has_count_user_2"}},
				{name: "not equal one", args: []any{"!=", 1}, expected: []string{"has_count_user_0", "has_count_user_2"}},
			}

			for _, test := range tests {
				s.Run(test.name, func() {
					var result []User
					s.Nil(query.Query().Where("name LIKE ?", "has_count_user_%").Has("Books", test.args...).Find(&result))

					names := make([]string, 0, len(result))
					for _, user := range result {
						names = append(names, user.Name)
					}
					s.ElementsMatch(test.expected, names)
				})
			}
		})
	}
}

func (s *QueryTestSuite) TestHasOneRelationship() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "has_one_user_present", Avatar: "has_one_avatar"},
				{Name: "has_one_user_absent", Avatar: "has_one_avatar"},
				{Name: "has_one_user_several", Avatar: "has_one_avatar"},
			}
			s.Nil(query.Query().Create(&users))
			addresses := []Address{
				{Name: "has_one_address", UserID: users[0].ID},
				{Name: "has_one_address_several_1", UserID: users[2].ID},
				{Name: "has_one_address_several_2", UserID: users[2].ID},
			}
			s.Nil(query.Query().Create(&addresses))

			names := func(users []User) []string {
				result := make([]string, 0, len(users))
				for _, user := range users {
					result = append(result, user.Name)
				}
				return result
			}

			// Presence.
			var result []User
			s.Nil(query.Query().Where("name LIKE ?", "has_one_user_%").Has("Address").Find(&result))
			s.ElementsMatch([]string{"has_one_user_present", "has_one_user_several"}, names(result))

			// A parent with several related records is returned once.
			s.Len(result, 2)

			// Absence.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "has_one_user_%").Has("Address", "=", 0).Find(&result))
			s.ElementsMatch([]string{"has_one_user_absent"}, names(result))

			// Count thresholds.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "has_one_user_%").Has("Address", "=", 2).Find(&result))
			s.ElementsMatch([]string{"has_one_user_several"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "has_one_user_%").Has("Address", ">", 1).Find(&result))
			s.ElementsMatch([]string{"has_one_user_several"}, names(result))

			// A nil callback behaves like Has.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "has_one_user_%").WhereHas("Address", nil).Find(&result))
			s.ElementsMatch([]string{"has_one_user_present", "has_one_user_several"}, names(result))

			// A callback constrains the related records only.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "has_one_user_%").WhereHas("Address", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "has_one_address_several_2")
			}).Find(&result))
			s.ElementsMatch([]string{"has_one_user_several"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "has_one_user_%").WhereHas("Address", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name LIKE ?", "has_one_address_several_%")
			}, ">=", 2).Find(&result))
			s.ElementsMatch([]string{"has_one_user_several"}, names(result))
		})
	}
}

func (s *QueryTestSuite) TestBelongsToRelationship() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "belongs_to_user_a", Avatar: "belongs_to_avatar"},
				{Name: "belongs_to_user_b", Avatar: "belongs_to_avatar"},
			}
			s.Nil(query.Query().Create(&users))
			books := []Book{
				{Name: "belongs_to_book_a_1", UserID: users[0].ID},
				{Name: "belongs_to_book_a_2", UserID: users[0].ID},
				{Name: "belongs_to_book_orphan", UserID: 0},
			}
			s.Nil(query.Query().Create(&books))

			names := func(books []Book) []string {
				result := make([]string, 0, len(books))
				for _, book := range books {
					result = append(result, book.Name)
				}
				return result
			}

			// Presence.
			var result []Book
			s.Nil(query.Query().Where("name LIKE ?", "belongs_to_book_%").Has("User").Find(&result))
			s.ElementsMatch([]string{"belongs_to_book_a_1", "belongs_to_book_a_2"}, names(result))

			// Both parents are returned once, even though they share the same related record.
			s.Len(result, 2)

			// Absence.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "belongs_to_book_%").Has("User", "=", 0).Find(&result))
			s.ElementsMatch([]string{"belongs_to_book_orphan"}, names(result))

			// A parent has at most one owning record, so a count threshold above one matches nothing.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "belongs_to_book_%").Has("User", ">", 1).Find(&result))
			s.Empty(result)

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "belongs_to_book_%").Has("User", "=", 1).Find(&result))
			s.ElementsMatch([]string{"belongs_to_book_a_1", "belongs_to_book_a_2"}, names(result))

			// A nil callback behaves like Has.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "belongs_to_book_%").WhereHas("User", nil).Find(&result))
			s.ElementsMatch([]string{"belongs_to_book_a_1", "belongs_to_book_a_2"}, names(result))

			// A callback constrains the related records only.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "belongs_to_book_%").WhereHas("User", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "belongs_to_user_b")
			}).Find(&result))
			s.Empty(result)

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "belongs_to_book_%").WhereHas("User", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "belongs_to_user_a")
			}).Find(&result))
			s.ElementsMatch([]string{"belongs_to_book_a_1", "belongs_to_book_a_2"}, names(result))
		})
	}
}

func (s *QueryTestSuite) TestManyToManyHas() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "m2m_has_user_two", Avatar: "m2m_has_avatar", Roles: []*Role{{Name: "m2m_has_role_a"}, {Name: "m2m_has_role_b"}}},
				{Name: "m2m_has_user_one", Avatar: "m2m_has_avatar", Roles: []*Role{{Name: "m2m_has_role_c"}}},
				{Name: "m2m_has_user_none", Avatar: "m2m_has_avatar"},
			}
			s.Nil(query.Query().Select(orm.Associations).Create(&users))

			names := func(users []User) []string {
				result := make([]string, 0, len(users))
				for _, user := range users {
					result = append(result, user.Name)
				}
				return result
			}

			// Presence.
			var result []User
			s.Nil(query.Query().Where("name LIKE ?", "m2m_has_user_%").Has("Roles").Find(&result))
			s.ElementsMatch([]string{"m2m_has_user_two", "m2m_has_user_one"}, names(result))

			// A parent with several pivot rows is returned once.
			s.Len(result, 2)

			// Absence.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "m2m_has_user_%").Has("Roles", "=", 0).Find(&result))
			s.ElementsMatch([]string{"m2m_has_user_none"}, names(result))

			// Count thresholds count the pivot rows.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "m2m_has_user_%").Has("Roles", "=", 2).Find(&result))
			s.ElementsMatch([]string{"m2m_has_user_two"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "m2m_has_user_%").Has("Roles", ">=", 1).Find(&result))
			s.ElementsMatch([]string{"m2m_has_user_two", "m2m_has_user_one"}, names(result))

			// A nil callback behaves like Has.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "m2m_has_user_%").WhereHas("Roles", nil).Find(&result))
			s.ElementsMatch([]string{"m2m_has_user_two", "m2m_has_user_one"}, names(result))

			// A callback constrains the related records through the pivot table.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "m2m_has_user_%").WhereHas("Roles", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "m2m_has_role_c")
			}).Find(&result))
			s.ElementsMatch([]string{"m2m_has_user_one"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "m2m_has_user_%").WhereHas("Roles", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name LIKE ?", "m2m_has_role_%")
			}, "=", 2).Find(&result))
			s.ElementsMatch([]string{"m2m_has_user_two"}, names(result))

			// The inverse side of the pivot table correlates on its own columns.
			var roles []Role
			s.Nil(query.Query().Model(&Role{}).Where("name LIKE ?", "m2m_has_role_%").Has("Users", "=", 1).Find(&roles))
			roleNames := make([]string, 0, len(roles))
			for _, role := range roles {
				roleNames = append(roleNames, role.Name)
			}
			s.ElementsMatch([]string{"m2m_has_role_a", "m2m_has_role_b", "m2m_has_role_c"}, roleNames)
		})
	}
}

func (s *QueryTestSuite) TestHasCustomKeyRelationship() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			departments := []CustomKeyDepartment{
				{Name: "custom_key_dept_a", Code: "custom-key-a"},
				{Name: "custom_key_dept_b", Code: "custom-key-b"},
				{Name: "custom_key_dept_c", Code: "custom-key-c"},
			}
			s.Nil(query.Query().Create(&departments))
			staff := []CustomKeyEmployee{
				{Name: "custom_key_staff_a1", DepartmentCode: "custom-key-a"},
				{Name: "custom_key_staff_a2", DepartmentCode: "custom-key-a"},
				{Name: "custom_key_staff_b", DepartmentCode: "custom-key-b"},
			}
			s.Nil(query.Query().Create(&staff))

			names := func(departments []CustomKeyDepartment) []string {
				result := make([]string, 0, len(departments))
				for _, department := range departments {
					result = append(result, department.Name)
				}
				return result
			}

			// The correlation uses the custom foreign key and reference, not the conventional names.
			var result []CustomKeyDepartment
			s.Nil(query.Query().Where("name LIKE ?", "custom_key_dept_%").Has("Staff").Find(&result))
			s.ElementsMatch([]string{"custom_key_dept_a", "custom_key_dept_b"}, names(result))

			// The count applies to the custom reference only.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "custom_key_dept_%").Has("Staff", "=", 2).Find(&result))
			s.ElementsMatch([]string{"custom_key_dept_a"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "custom_key_dept_%").Has("Staff", "=", 0).Find(&result))
			s.ElementsMatch([]string{"custom_key_dept_c"}, names(result))

			// A callback constrains the related records reached through the custom key.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "custom_key_dept_%").WhereHas("Staff", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "custom_key_staff_b")
			}).Find(&result))
			s.ElementsMatch([]string{"custom_key_dept_b"}, names(result))
		})
	}
}

func (s *QueryTestSuite) TestHasCustomKeyHasOneAndBelongsTo() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			departments := []CustomKeyDepartment{
				{Name: "custom_key_one_dept_with_profile", Code: "custom-one-key-with"},
				{Name: "custom_key_one_dept_without_profile", Code: "custom-one-key-without"},
				{Name: "custom_key_one_dept_owning", Code: "custom-one-key-owning"},
			}
			s.Nil(query.Query().Create(&departments))
			s.Nil(query.Query().Create(&CustomKeyProfile{
				Name:           "custom_key_one_profile",
				DepartmentCode: "custom-one-key-with",
			}))
			staff := []CustomKeyEmployee{
				{Name: "custom_key_one_staff_linked", DepartmentCode: "custom-one-key-owning"},
				// No department carries this code, so the record has no owning department.
				{Name: "custom_key_one_staff_orphan", DepartmentCode: "custom-one-key-missing"},
			}
			s.Nil(query.Query().Create(&staff))

			departmentNames := func(departments []CustomKeyDepartment) []string {
				result := make([]string, 0, len(departments))
				for _, department := range departments {
					result = append(result, department.Name)
				}
				return result
			}
			staffNames := func(staff []CustomKeyEmployee) []string {
				result := make([]string, 0, len(staff))
				for _, employee := range staff {
					result = append(result, employee.Name)
				}
				return result
			}

			// Has-one: the related profile matches the custom foreign key and reference.
			var result []CustomKeyDepartment
			s.Nil(query.Query().Where("name LIKE ?", "custom_key_one_dept_%").Has("Profile").Find(&result))
			s.ElementsMatch([]string{"custom_key_one_dept_with_profile"}, departmentNames(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "custom_key_one_dept_%").Has("Profile", "=", 0).Find(&result))
			s.ElementsMatch([]string{"custom_key_one_dept_without_profile", "custom_key_one_dept_owning"}, departmentNames(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "custom_key_one_dept_%").WhereHas("Profile", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "custom_key_one_profile")
			}).Find(&result))
			s.ElementsMatch([]string{"custom_key_one_dept_with_profile"}, departmentNames(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "custom_key_one_dept_%").WhereHas("Profile", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "custom_key_one_missing_profile")
			}).Find(&result))
			s.Empty(result)

			// Belongs-to: the related department matches the custom keys, and the orphan has none.
			var staffResult []CustomKeyEmployee
			s.Nil(query.Query().Where("name LIKE ?", "custom_key_one_staff_%").Has("Department").Find(&staffResult))
			s.ElementsMatch([]string{"custom_key_one_staff_linked"}, staffNames(staffResult))

			staffResult = nil
			s.Nil(query.Query().Where("name LIKE ?", "custom_key_one_staff_%").Has("Department", "=", 0).Find(&staffResult))
			s.ElementsMatch([]string{"custom_key_one_staff_orphan"}, staffNames(staffResult))

			staffResult = nil
			s.Nil(query.Query().Where("name LIKE ?", "custom_key_one_staff_%").WhereHas("Department", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "custom_key_one_dept_owning")
			}).Find(&staffResult))
			s.ElementsMatch([]string{"custom_key_one_staff_linked"}, staffNames(staffResult))

			staffResult = nil
			s.Nil(query.Query().Where("name LIKE ?", "custom_key_one_staff_%").WhereHas("Department", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "custom_key_one_dept_with_profile")
			}).Find(&staffResult))
			s.Empty(staffResult)
		})
	}
}

func (s *QueryTestSuite) TestHasCompositeReferences() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			parents := []CompositeParent{
				{Name: "composite_parent_europe_a", Region: "europe", Code: "a"},
				{Name: "composite_parent_asia_b", Region: "asia", Code: "b"},
				{Name: "composite_parent_africa_c", Region: "africa", Code: "c"},
			}
			s.Nil(query.Query().Create(&parents))
			items := []CompositeItem{
				// Matches the first parent on both columns.
				{Name: "composite_item_europe_a", ParentRegion: "europe", ParentCode: "a"},
				// Matches the first parent on the region only and the second parent on the code only.
				{Name: "composite_item_europe_b", ParentRegion: "europe", ParentCode: "b"},
				// Matches the second parent on both columns.
				{Name: "composite_item_asia_b", ParentRegion: "asia", ParentCode: "b"},
			}
			s.Nil(query.Query().Create(&items))

			names := func(parents []CompositeParent) []string {
				result := make([]string, 0, len(parents))
				for _, parent := range parents {
					result = append(result, parent.Name)
				}
				return result
			}

			var result []CompositeParent
			s.Nil(query.Query().Where("name LIKE ?", "composite_parent_%").Has("Items").Find(&result))
			s.ElementsMatch([]string{"composite_parent_europe_a", "composite_parent_asia_b"}, names(result))

			// A partial match is not an association, so each parent has exactly one item.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "composite_parent_%").Has("Items", "=", 1).Find(&result))
			s.ElementsMatch([]string{"composite_parent_europe_a", "composite_parent_asia_b"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "composite_parent_%").Has("Items", "=", 0).Find(&result))
			s.ElementsMatch([]string{"composite_parent_africa_c"}, names(result))

			// The partially matching item belongs to no parent.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "composite_parent_%").WhereHas("Items", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "composite_item_europe_b")
			}).Find(&result))
			s.Empty(result)
		})
	}
}

func (s *QueryTestSuite) TestHasCompositeBelongsToAndHasOneReferences() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			parents := []CompositeParent{
				{Name: "composite_one_parent_europe_a", Region: "one_europe", Code: "a"},
				{Name: "composite_one_parent_asia_b", Region: "one_asia", Code: "b"},
				{Name: "composite_one_parent_africa_c", Region: "one_africa", Code: "c"},
			}
			s.Nil(query.Query().Create(&parents))
			items := []CompositeItem{
				// Matches the first parent on both columns.
				{Name: "composite_one_item_europe_a", ParentRegion: "one_europe", ParentCode: "a"},
				// Matches the first parent on the region only and the second parent on the code only.
				{Name: "composite_one_item_europe_b", ParentRegion: "one_europe", ParentCode: "b"},
				// No parent carries both columns, so the record has no owning parent.
				{Name: "composite_one_item_orphan", ParentRegion: "one_missing", ParentCode: "z"},
			}
			s.Nil(query.Query().Create(&items))

			parentNames := func(parents []CompositeParent) []string {
				result := make([]string, 0, len(parents))
				for _, parent := range parents {
					result = append(result, parent.Name)
				}
				return result
			}
			itemNames := func(items []CompositeItem) []string {
				result := make([]string, 0, len(items))
				for _, item := range items {
					result = append(result, item.Name)
				}
				return result
			}

			// Belongs-to presence: only the item matching both columns has an owning parent.
			var itemResult []CompositeItem
			s.Nil(query.Query().Where("name LIKE ?", "composite_one_item_%").Has("Parent").Find(&itemResult))
			s.ElementsMatch([]string{"composite_one_item_europe_a"}, itemNames(itemResult))

			// Belongs-to absence: the partial match and the orphan have no owning parent.
			itemResult = nil
			s.Nil(query.Query().Where("name LIKE ?", "composite_one_item_%").Has("Parent", "=", 0).Find(&itemResult))
			s.ElementsMatch([]string{"composite_one_item_europe_b", "composite_one_item_orphan"}, itemNames(itemResult))

			// Belongs-to partial match: the partially matching item is not associated with either parent.
			itemResult = nil
			s.Nil(query.Query().Where("name", "composite_one_item_europe_b").WhereHas("Parent", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("region", "one_europe")
			}).Find(&itemResult))
			s.Empty(itemResult)

			// Has-one presence: only the parent matching both columns has a related item.
			var result []CompositeParent
			s.Nil(query.Query().Where("name LIKE ?", "composite_one_parent_%").Has("FirstItem").Find(&result))
			s.ElementsMatch([]string{"composite_one_parent_europe_a"}, parentNames(result))

			// Has-one absence.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "composite_one_parent_%").Has("FirstItem", "=", 0).Find(&result))
			s.ElementsMatch([]string{"composite_one_parent_asia_b", "composite_one_parent_africa_c"}, parentNames(result))

			// Has-one partial match: the partially matching item belongs to no parent.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "composite_one_parent_%").WhereHas("FirstItem", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "composite_one_item_europe_b")
			}).Find(&result))
			s.Empty(result)

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "composite_one_parent_%").WhereHas("FirstItem", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "composite_one_item_europe_a")
			}).Find(&result))
			s.ElementsMatch([]string{"composite_one_parent_europe_a"}, parentNames(result))
		})
	}
}

func (s *QueryTestSuite) TestHasCompositeManyToManyReferences() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			_, err := query.DB().Table("composite_posts").Insert([]map[string]any{
				{"id": 1, "tenant_id": 1, "title": "composite_m2m_post_both"},
				// The same id in another tenant shares every other column, so a correlation that ignores the
				// tenant column would match this post against the pivot row of the first one.
				{"id": 1, "tenant_id": 2, "title": "composite_m2m_post_other_tenant"},
				{"id": 2, "tenant_id": 1, "title": "composite_m2m_post_partial"},
				{"id": 3, "tenant_id": 2, "title": "composite_m2m_post_none"},
			})
			s.Nil(err)
			_, err = query.DB().Table("composite_tags").Insert([]map[string]any{
				{"id": 1, "kind": "book", "code": "a"},
				{"id": 2, "kind": "book", "code": "b"},
			})
			s.Nil(err)
			_, err = query.DB().Table("composite_post_tag").Insert([]map[string]any{
				// Both columns match the first tag.
				{"composite_post_id": 1, "composite_post_tenant_id": 1, "tag_kind": "book", "tag_code": "a"},
				// Only the kind matches, so the row is not an association.
				{"composite_post_id": 2, "composite_post_tenant_id": 1, "tag_kind": "book", "tag_code": "x"},
			})
			s.Nil(err)

			titles := func(posts []CompositePost) []string {
				result := make([]string, 0, len(posts))
				for _, post := range posts {
					result = append(result, post.Title)
				}
				return result
			}

			// A positive match needs both reference columns on both sides.
			var result []CompositePost
			s.Nil(query.Query().Where("title LIKE ?", "composite_m2m_post_%").Has("Tags").Find(&result))
			s.ElementsMatch([]string{"composite_m2m_post_both"}, titles(result))

			// The row matching one column only is not counted.
			result = nil
			s.Nil(query.Query().Where("title LIKE ?", "composite_m2m_post_%").Has("Tags", "=", 1).Find(&result))
			s.ElementsMatch([]string{"composite_m2m_post_both"}, titles(result))

			// Absence. The post that shares the id but not the tenant has no matching pivot row.
			result = nil
			s.Nil(query.Query().Where("title LIKE ?", "composite_m2m_post_%").Has("Tags", "=", 0).Find(&result))
			s.ElementsMatch([]string{"composite_m2m_post_other_tenant", "composite_m2m_post_partial", "composite_m2m_post_none"}, titles(result))

			// A callback constrains the related tags only.
			result = nil
			s.Nil(query.Query().Where("title LIKE ?", "composite_m2m_post_%").WhereHas("Tags", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("code", "a")
			}).Find(&result))
			s.ElementsMatch([]string{"composite_m2m_post_both"}, titles(result))
		})
	}
}

func (s *QueryTestSuite) TestHasManyToManyCustomJoinKeys() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			posts := []CustomJoinPost{
				{Title: "custom_join_post_two"},
				{Title: "custom_join_post_one"},
				{Title: "custom_join_post_none"},
				{Title: "custom_join_post_shared"},
			}
			s.Nil(query.Query().Create(&posts))
			tags := []CustomJoinTag{
				{Name: "custom_join_tag_a"},
				{Name: "custom_join_tag_b"},
				{Name: "custom_join_tag_c"},
			}
			s.Nil(query.Query().Create(&tags))

			// The pivot rows go through the custom join columns, so the correlation must use them too.
			s.Nil(query.Query().Model(&posts[0]).Association("Tags").Append(&tags[0]))
			s.Nil(query.Query().Model(&posts[0]).Association("Tags").Append(&tags[1]))
			s.Nil(query.Query().Model(&posts[1]).Association("Tags").Append(&tags[2]))
			s.Nil(query.Query().Model(&posts[3]).Association("Tags").Append(&tags[0]))

			titles := func(posts []CustomJoinPost) []string {
				result := make([]string, 0, len(posts))
				for _, post := range posts {
					result = append(result, post.Title)
				}
				return result
			}

			// Presence.
			var result []CustomJoinPost
			s.Nil(query.Query().Where("title LIKE ?", "custom_join_post_%").Has("Tags").Find(&result))
			s.ElementsMatch([]string{"custom_join_post_two", "custom_join_post_one", "custom_join_post_shared"}, titles(result))

			// Absence.
			result = nil
			s.Nil(query.Query().Where("title LIKE ?", "custom_join_post_%").Has("Tags", "=", 0).Find(&result))
			s.ElementsMatch([]string{"custom_join_post_none"}, titles(result))

			// Count thresholds count the pivot rows of each post only.
			result = nil
			s.Nil(query.Query().Where("title LIKE ?", "custom_join_post_%").Has("Tags", "=", 2).Find(&result))
			s.ElementsMatch([]string{"custom_join_post_two"}, titles(result))

			result = nil
			s.Nil(query.Query().Where("title LIKE ?", "custom_join_post_%").Has("Tags", "=", 1).Find(&result))
			s.ElementsMatch([]string{"custom_join_post_one", "custom_join_post_shared"}, titles(result))

			// A callback constrains the related records through the custom pivot columns.
			result = nil
			s.Nil(query.Query().Where("title LIKE ?", "custom_join_post_%").WhereHas("Tags", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "custom_join_tag_c")
			}).Find(&result))
			s.ElementsMatch([]string{"custom_join_post_one"}, titles(result))

			result = nil
			s.Nil(query.Query().Where("title LIKE ?", "custom_join_post_%").WhereHas("Tags", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name LIKE ?", "custom_join_tag_%")
			}, ">=", 2).Find(&result))
			s.ElementsMatch([]string{"custom_join_post_two"}, titles(result))
		})
	}
}

func (s *QueryTestSuite) TestHasSelfRelations() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			ceo := Employee{Name: "self_ceo"}
			s.Nil(query.Query().Create(&ceo))
			leadA := Employee{Name: "self_lead_a", ManagerID: &ceo.ID}
			leadB := Employee{Name: "self_lead_b", ManagerID: &ceo.ID}
			s.Nil(query.Query().Create(&leadA))
			s.Nil(query.Query().Create(&leadB))
			worker := Employee{Name: "self_worker", ManagerID: &leadA.ID}
			none := Employee{Name: "self_none"}
			s.Nil(query.Query().Create(&worker))
			s.Nil(query.Query().Create(&none))

			names := func(employees []Employee) []string {
				result := make([]string, 0, len(employees))
				for _, employee := range employees {
					result = append(result, employee.Name)
				}
				return result
			}

			// Has-many: the reports of a record are the records whose manager is that record.
			var result []Employee
			s.Nil(query.Query().Where("name LIKE ?", "self_%").Has("Reports").Find(&result))
			s.ElementsMatch([]string{"self_ceo", "self_lead_a"}, names(result))

			// The count threshold counts the reports only.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").Has("Reports", ">=", 2).Find(&result))
			s.ElementsMatch([]string{"self_ceo"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").Has("Reports", "=", 0).Find(&result))
			s.ElementsMatch([]string{"self_lead_b", "self_worker", "self_none"}, names(result))

			// A callback constrains the reports only.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").WhereHas("Reports", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "self_lead_a")
			}).Find(&result))
			s.ElementsMatch([]string{"self_ceo"}, names(result))

			// Belongs-to: the manager of a record is the record with that primary key.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").Has("Manager").Find(&result))
			s.ElementsMatch([]string{"self_lead_a", "self_lead_b", "self_worker"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").Has("Manager", "=", 0).Find(&result))
			s.ElementsMatch([]string{"self_ceo", "self_none"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").WhereHas("Manager", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "self_ceo")
			}).Find(&result))
			s.ElementsMatch([]string{"self_lead_a", "self_lead_b"}, names(result))

			// Many-to-many: only the records that own a pivot row are parents, not the records they point to.
			s.Nil(query.Query().Model(&ceo).Association("Friends").Append(&worker))
			s.Nil(query.Query().Model(&ceo).Association("Friends").Append(&leadB))
			s.Nil(query.Query().Model(&leadA).Association("Friends").Append(&worker))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").Has("Friends").Find(&result))
			s.ElementsMatch([]string{"self_ceo", "self_lead_a"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").Has("Friends", "=", 2).Find(&result))
			s.ElementsMatch([]string{"self_ceo"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").Has("Friends", "=", 0).Find(&result))
			s.ElementsMatch([]string{"self_lead_b", "self_worker", "self_none"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").WhereHas("Friends", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "self_worker")
			}).Find(&result))
			s.ElementsMatch([]string{"self_ceo", "self_lead_a"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").WhereHas("Friends", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "self_worker")
			}, "=", 0).Find(&result))
			s.ElementsMatch([]string{"self_lead_b", "self_worker", "self_none"}, names(result))

			// The related soft delete uses the alias of the self relationship.
			_, err := query.Query().Where("name", "self_lead_b").Delete(&Employee{})
			s.Nil(err)

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").Has("Reports", ">=", 2).Find(&result))
			s.Empty(result)

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").WhereHas("Reports", func(query contractsorm.Query) contractsorm.Query {
				return query.WithTrashed()
			}, ">=", 2).Find(&result))
			s.ElementsMatch([]string{"self_ceo"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").Has("Friends", "=", 2).Find(&result))
			s.Empty(result)

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "self_%").WhereHas("Friends", func(query contractsorm.Query) contractsorm.Query {
				return query.WithTrashed()
			}, "=", 2).Find(&result))
			s.ElementsMatch([]string{"self_ceo"}, names(result))

			// Has-one on the same table: the related record is the row whose parent id is the record id. The
			// prefix stays outside "self_%", so the assertions above keep their own records.
			parent := Employee{Name: "selfone_parent"}
			s.Nil(query.Query().Create(&parent))
			firstChild := Employee{Name: "selfone_child", ParentID: &parent.ID}
			secondChild := Employee{Name: "selfone_child_two", ParentID: &parent.ID}
			s.Nil(query.Query().Create(&firstChild))
			s.Nil(query.Query().Create(&secondChild))
			s.Nil(query.Query().Create(&Employee{Name: "selfone_absent"}))

			// Presence.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "selfone_%").Has("Child").Find(&result))
			s.ElementsMatch([]string{"selfone_parent"}, names(result))

			// Absence.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "selfone_%").Has("Child", "=", 0).Find(&result))
			s.ElementsMatch([]string{"selfone_child", "selfone_child_two", "selfone_absent"}, names(result))

			// Count thresholds above one count the related records only.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "selfone_%").Has("Child", ">", 1).Find(&result))
			s.ElementsMatch([]string{"selfone_parent"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "selfone_%").Has("Child", "=", 2).Find(&result))
			s.ElementsMatch([]string{"selfone_parent"}, names(result))

			// A callback constrains the related records only.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "selfone_%").WhereHas("Child", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "selfone_child")
			}).Find(&result))
			s.ElementsMatch([]string{"selfone_parent"}, names(result))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "selfone_%").WhereHas("Child", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "selfone_parent")
			}).Find(&result))
			s.Empty(result)
		})
	}
}

func (s *QueryTestSuite) TestWhereHas() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "where_has_user_a", Avatar: "where_has_avatar"},
				{Name: "where_has_user_b", Avatar: "where_has_avatar"},
				{Name: "where_has_user_c", Avatar: "where_has_avatar"},
				{Name: "where_has_user_d", Avatar: "where_has_avatar"},
			}
			s.Nil(query.Query().Create(&users))
			books := []Book{
				{Name: "where_has_match", UserID: users[0].ID},
				{Name: "where_has_match", UserID: users[0].ID},
				{Name: "where_has_other", UserID: users[0].ID},
				{Name: "where_has_other", UserID: users[1].ID},
				{Name: "where_has_match", UserID: users[2].ID},
			}
			s.Nil(query.Query().Create(&books))

			names := func(users []User) []string {
				result := make([]string, 0, len(users))
				for _, user := range users {
					result = append(result, user.Name)
				}
				return result
			}

			// A nil callback behaves like Has.
			var result []User
			s.Nil(query.Query().Where("name LIKE ?", "where_has_user_%").WhereHas("Books", nil).Find(&result))
			s.ElementsMatch([]string{"where_has_user_a", "where_has_user_b", "where_has_user_c"}, names(result))

			// The callback constrains the related records only.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "where_has_user_%").WhereHas("Books", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "where_has_match")
			}).Find(&result))
			s.ElementsMatch([]string{"where_has_user_a", "where_has_user_c"}, names(result))

			// Only related records matching the callback are counted.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "where_has_user_%").WhereHas("Books", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "where_has_match")
			}, ">=", 2).Find(&result))
			s.Len(result, 1)
			s.Equal(users[0].ID, result[0].ID)

			// The absence of matching related records.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "where_has_user_%").WhereHas("Books", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "where_has_match")
			}, "=", 0).Find(&result))
			s.ElementsMatch([]string{"where_has_user_b", "where_has_user_d"}, names(result))

			// The absence of any related records.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "where_has_user_%").WhereHas("Books", nil, "=", 0).Find(&result))
			s.Len(result, 1)
			s.Equal(users[3].ID, result[0].ID)
		})
	}
}

func (s *QueryTestSuite) TestWhereHasComposition() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "where_has_compose_a", Avatar: "where_has_compose_avatar_a"},
				{Name: "where_has_compose_b", Avatar: "where_has_compose_avatar_b"},
				{Name: "where_has_compose_c", Avatar: "where_has_compose_avatar_c"},
				{Name: "where_has_compose_d", Avatar: "where_has_compose_avatar_d"},
				// This user has no books, so it can only be returned by the left operand of the OR below.
				{Name: "where_has_compose_e", Avatar: "where_has_compose_avatar_e"},
			}
			s.Nil(query.Query().Create(&users))
			books := []Book{
				{Name: "where_has_compose_book_a", UserID: users[0].ID},
				{Name: "where_has_compose_book_b", UserID: users[1].ID},
				{Name: "where_has_compose_book_d", UserID: users[3].ID},
			}
			s.Nil(query.Query().Create(&books))

			// The new condition composes with existing where conditions as an AND group. AND binds tighter than
			// OR, so the relationship filters only the right operand: the user without books is still returned.
			var result []User
			s.Nil(query.Query().Where("name", "where_has_compose_e").OrWhere("name", "where_has_compose_b").Has("Books").Find(&result))
			ids := make([]uint, 0, len(result))
			for _, user := range result {
				ids = append(ids, user.ID)
			}
			s.ElementsMatch([]uint{users[4].ID, users[1].ID}, ids)

			// The new condition composes inside a where group.
			result = nil
			s.Nil(query.Query().Where(func(query contractsorm.Query) contractsorm.Query {
				return query.Has("Books").Where("name", "where_has_compose_a")
			}).Where("name LIKE ?", "where_has_compose_%").Find(&result))
			s.Len(result, 1)
			s.Equal(users[0].ID, result[0].ID)

			// Grouping works with the new condition.
			var grouped []User
			s.Nil(query.Query().Model(&User{}).Select("avatar").Where("name LIKE ?", "where_has_compose_%").GroupBy("avatar").Has("Books").Find(&grouped))
			avatars := make([]string, 0, len(grouped))
			for _, user := range grouped {
				avatars = append(avatars, user.Avatar)
			}
			s.ElementsMatch([]string{"where_has_compose_avatar_a", "where_has_compose_avatar_b", "where_has_compose_avatar_d"}, avatars)

			// Count only counts parents matching the relationship.
			count, err := query.Query().Model(&User{}).Where("name LIKE ?", "where_has_compose_%").Has("Books").Count()
			s.Nil(err)
			s.Equal(int64(3), count)

			// Pagination keeps the relationship filter for the total and the page.
			var total int64
			result = nil
			s.Nil(query.Query().Model(&User{}).Where("name LIKE ?", "where_has_compose_%").Has("Books").OrderBy("id").Paginate(1, 2, &result, &total))
			s.Equal(int64(3), total)
			s.Len(result, 2)

			result = nil
			s.Nil(query.Query().Model(&User{}).Where("name LIKE ?", "where_has_compose_%").Has("Books").OrderBy("id").Paginate(2, 2, &result, &total))
			s.Equal(int64(3), total)
			s.Len(result, 1)
		})
	}
}

func (s *QueryTestSuite) TestWhereHasNestedWhereGroups() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "where_has_nested_group_a", Avatar: "where_has_nested_group_avatar"},
				{Name: "where_has_nested_group_b", Avatar: "where_has_nested_group_avatar"},
				{Name: "where_has_nested_group_c", Avatar: "where_has_nested_group_avatar"},
			}
			s.Nil(query.Query().Create(&users))
			s.Nil(query.Query().Create(&Book{Name: "where_has_nested_group_book_a", UserID: users[0].ID}))
			s.Nil(query.Query().Create(&Book{Name: "where_has_nested_group_book_b", UserID: users[1].ID}))

			// Two nested where groups resolve the explicit model of the owning query, so the innermost
			// relationship filter builds instead of reporting a missing model.
			var result []User
			s.Nil(query.Query().Model(&User{}).Where(func(outer contractsorm.Query) contractsorm.Query {
				return outer.Where(func(inner contractsorm.Query) contractsorm.Query {
					return inner.Has("Books")
				})
			}).Where("name LIKE ?", "where_has_nested_group_%").Find(&result))
			s.Len(result, 2)
			names := make([]string, 0, len(result))
			for _, user := range result {
				names = append(names, user.Name)
			}
			s.ElementsMatch([]string{"where_has_nested_group_a", "where_has_nested_group_b"}, names)

			// The innermost group also keeps its own conditions.
			result = nil
			s.Nil(query.Query().Model(&User{}).Where(func(outer contractsorm.Query) contractsorm.Query {
				return outer.Where(func(inner contractsorm.Query) contractsorm.Query {
					return inner.Has("Books").Where("name", "where_has_nested_group_a")
				})
			}).Find(&result))
			s.Len(result, 1)
			s.Equal(users[0].ID, result[0].ID)

			// A parent without related records is not returned.
			result = nil
			s.Nil(query.Query().Model(&User{}).Where(func(outer contractsorm.Query) contractsorm.Query {
				return outer.Where(func(inner contractsorm.Query) contractsorm.Query {
					return inner.Has("Books").Where("name", "where_has_nested_group_c")
				})
			}).Find(&result))
			s.Empty(result)
		})
	}
}

func (s *QueryTestSuite) TestWhereHasTransactionAndContext() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "where_has_tx_user", Avatar: "where_has_tx_avatar"}
			tx, err := query.Query().Begin()
			s.Nil(err)
			s.Nil(tx.Create(&user))
			s.Nil(tx.Create(&Book{Name: "where_has_tx_book", UserID: user.ID}))

			// The subquery runs on the transaction connection.
			var result []User
			s.Nil(tx.Model(&User{}).Has("Books").Where("name", "where_has_tx_user").Find(&result))
			s.Len(result, 1)

			count, err := tx.Model(&User{}).Has("Books").Where("name", "where_has_tx_user").Count()
			s.Nil(err)
			s.Equal(int64(1), count)

			s.Nil(tx.Rollback())

			result = nil
			s.Nil(query.Query().Has("Books").Where("name", "where_has_tx_user").Find(&result))
			s.Empty(result)

			// The query context is preserved through the relationship filter.
			ctxUser := User{Name: "where_has_tx_user", Avatar: "where_has_tx_avatar"}
			s.Nil(query.Query().Create(&ctxUser))
			s.Nil(query.Query().Create(&Book{Name: "where_has_ctx_book", UserID: ctxUser.ID}))
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result = nil
			s.Nil(query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Model(&User{}).Has("Books").
				Where("name", "where_has_tx_user").Find(&result))
			s.Len(result, 1)
		})
	}
}

func (s *QueryTestSuite) TestHasSoftDeletesAndScopes() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			// The parent soft delete applies to the parent query.
			trashedParent := User{Name: "has_trashed_parent", Avatar: "has_scope_avatar"}
			s.Nil(query.Query().Create(&trashedParent))
			s.Nil(query.Query().Create(&Book{Name: "has_trashed_parent_book", UserID: trashedParent.ID}))
			_, err := query.Query().Where("name", "has_trashed_parent").Delete(&User{})
			s.Nil(err)

			var result []User
			s.Nil(query.Query().Has("Books").Where("name", "has_trashed_parent").Find(&result))
			s.Empty(result)

			result = nil
			s.Nil(query.Query().WithTrashed().Has("Books").Where("name", "has_trashed_parent").Find(&result))
			s.Len(result, 1)

			// The related soft delete applies to the related query.
			trashedRelated := User{Name: "has_trashed_related", Avatar: "has_scope_avatar"}
			s.Nil(query.Query().Create(&trashedRelated))
			s.Nil(query.Query().Create(&Task{Name: "has_trashed_task", Status: "active", UserID: trashedRelated.ID}))
			_, err = query.Query().Where("name", "has_trashed_task").Delete(&Task{})
			s.Nil(err)

			result = nil
			s.Nil(query.Query().Where("name", "has_trashed_related").Has("Tasks").Find(&result))
			s.Empty(result)

			result = nil
			s.Nil(query.Query().Where("name", "has_trashed_related").WhereHas("Tasks", func(query contractsorm.Query) contractsorm.Query {
				return query.WithTrashed()
			}).Find(&result))
			s.Len(result, 1)

			// The parent WithTrashed does not disable the related soft delete.
			result = nil
			s.Nil(query.Query().WithTrashed().Where("name", "has_trashed_related").Has("Tasks").Find(&result))
			s.Empty(result)

			result = nil
			s.Nil(query.Query().WithTrashed().Where("name", "has_trashed_related").WhereHas("Tasks", func(query contractsorm.Query) contractsorm.Query {
				return query.WithTrashed()
			}).Find(&result))
			s.Len(result, 1)

			// A soft-deleted parent stays excluded when the related query includes trashed records.
			_, err = query.Query().Where("name", "has_trashed_related").Delete(&User{})
			s.Nil(err)

			result = nil
			s.Nil(query.Query().Where("name", "has_trashed_related").WhereHas("Tasks", func(query contractsorm.Query) contractsorm.Query {
				return query.WithTrashed()
			}).Find(&result))
			s.Empty(result)

			// The related global scope applies to the related query.
			active := User{Name: "has_scope_active", Avatar: "has_scope_avatar"}
			archived := User{Name: "has_scope_archived", Avatar: "has_scope_avatar"}
			s.Nil(query.Query().Create(&active))
			s.Nil(query.Query().Create(&archived))
			s.Nil(query.Query().Create(&Task{Name: "has_scope_active_task", Status: "active", UserID: active.ID}))
			s.Nil(query.Query().Create(&Task{Name: "has_scope_archived_task", Status: "archived", UserID: archived.ID}))

			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "has_scope_%").Has("Tasks").Find(&result))
			s.Len(result, 1)
			s.Equal(active.ID, result[0].ID)

			// The callback can disable the related scope, and the parent scope is still untouched.
			result = nil
			s.Nil(query.Query().Where("name LIKE ?", "has_scope_%").WhereHas("Tasks", func(query contractsorm.Query) contractsorm.Query {
				return query.WithoutGlobalScopes()
			}).Find(&result))
			s.Len(result, 2)
			s.ElementsMatch([]uint{active.ID, archived.ID}, []uint{result[0].ID, result[1].ID})

			// Disabling the parent scopes does not disable the related scope.
			result = nil
			s.Nil(query.Query().WithoutGlobalScopes().Where("name LIKE ?", "has_scope_%").Has("Tasks").Find(&result))
			s.Len(result, 1)
			s.Equal(active.ID, result[0].ID)
		})
	}
}

func (s *QueryTestSuite) TestHasTerminalOperations() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "has_ops_user_with", Avatar: "has_ops_avatar"},
				{Name: "has_ops_user_without", Avatar: "has_ops_avatar"},
			}
			s.Nil(query.Query().Create(&users))
			s.Nil(query.Query().Create(&Book{Name: "has_ops_book", UserID: users[0].ID}))

			// First returns the first parent with a related record.
			var first User
			s.Nil(query.Query().Model(&User{}).Where("name LIKE ?", "has_ops_user_%").Has("Books").First(&first))
			s.Equal(users[0].ID, first.ID)

			// Exists reports whether any parent matches.
			exists, err := query.Query().Model(&User{}).Where("name", "has_ops_user_with").Has("Books").Exists()
			s.Nil(err)
			s.True(exists)

			exists, err = query.Query().Model(&User{}).Where("name", "has_ops_user_without").Has("Books").Exists()
			s.Nil(err)
			s.False(exists)

			// DoesntExist is the inverse.
			doesntExist, err := query.Query().Model(&User{}).Where("name", "has_ops_user_with").Has("Books").DoesntExist()
			s.Nil(err)
			s.False(doesntExist)

			doesntExist, err = query.Query().Model(&User{}).Where("name", "has_ops_user_without").Has("Books").DoesntExist()
			s.Nil(err)
			s.True(doesntExist)

			// Count counts only parents that match the relationship filter.
			count, err := query.Query().Model(&User{}).Where("name LIKE ?", "has_ops_user_%").Has("Books").Count()
			s.Nil(err)
			s.Equal(int64(1), count)

			count, err = query.Query().Model(&User{}).Where("name LIKE ?", "has_ops_user_%").WhereHas("Books", func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name", "has_ops_book")
			}).Count()
			s.Nil(err)
			s.Equal(int64(1), count)

			// Paginate infers the model from the typed destination for the total and the page.
			var result []User
			var total int64
			s.Nil(query.Query().Where("name LIKE ?", "has_ops_user_%").Has("Books").Paginate(1, 10, &result, &total))
			s.Equal(int64(1), total)
			s.Len(result, 1)
			s.Equal(users[0].ID, result[0].ID)
		})
	}
}

func (s *QueryTestSuite) TestHasInvalidArguments() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "has_invalid_user", Avatar: "has_invalid_avatar"}
			s.Nil(query.Query().Create(&user))
			s.Nil(query.Query().Create(&Book{Name: "has_invalid_book", UserID: user.ID}))

			tests := []struct {
				name     string
				args     []any
				expected error
			}{
				{name: "operator without count", args: []any{">"}, expected: errors.OrmQueryRelationshipArguments},
				{name: "count without operator", args: []any{2}, expected: errors.OrmQueryRelationshipArguments},
				{name: "excess arguments", args: []any{">", 2, 3}, expected: errors.OrmQueryRelationshipArguments},
				{name: "negative count", args: []any{">", -1}, expected: errors.OrmQueryRelationshipCount},
				{name: "float count", args: []any{">", 1.5}, expected: errors.OrmQueryRelationshipCount},
				{name: "string count", args: []any{">", "1"}, expected: errors.OrmQueryRelationshipCount},
				{name: "unsupported operator", args: []any{"==", 1}, expected: errors.OrmQueryRelationshipOperator},
			}

			for _, test := range tests {
				s.Run(test.name, func() {
					var result []User
					err := query.Query().Model(&User{}).Has("Books", test.args...).Find(&result)
					s.ErrorIs(err, test.expected)
					s.Empty(result)

					err = query.Query().Model(&User{}).WhereHas("Books", nil, test.args...).Find(&result)
					s.ErrorIs(err, test.expected)
					s.Empty(result)
				})
			}
		})
	}
}

func (s *QueryTestSuite) TestHasInvalidRelationships() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			tests := []struct {
				name         string
				model        any
				relationship string
				expected     error
			}{
				{name: "unknown relationship", model: &User{}, relationship: "Missing", expected: errors.OrmQueryRelationshipNotFound},
				{name: "nested relationship", model: &User{}, relationship: "Books.Author", expected: errors.OrmQueryRelationshipUnsupported},
				{name: "polymorphic has one relationship", model: &User{}, relationship: "House", expected: errors.OrmQueryRelationshipUnsupported},
				{name: "polymorphic has many relationship", model: &User{}, relationship: "Phones", expected: errors.OrmQueryRelationshipUnsupported},
			}

			for _, test := range tests {
				s.Run(test.name, func() {
					var result []User
					err := query.Query().Model(test.model).Has(test.relationship).Find(&result)
					s.ErrorIs(err, test.expected)
					s.Empty(result)
				})
			}

			// The model cannot be identified without an explicit model or a typed destination.
			var result []User
			s.ErrorIs(query.Query().Has("Books").Find(nil), errors.OrmQueryRelationshipModel)
			s.ErrorIs(query.Query().Has("Books").Get(nil), errors.OrmQueryRelationshipModel)
			s.ErrorIs(query.Query().Has("Books").First(nil), errors.OrmQueryRelationshipModel)

			count, err := query.Query().Has("Books").Count()
			s.ErrorIs(err, errors.OrmQueryRelationshipModel)
			s.NotContains(err.Error(), "Table not set")
			s.Equal(int64(0), count)

			exists, err := query.Query().Has("Books").Exists()
			s.ErrorIs(err, errors.OrmQueryRelationshipModel)
			s.False(exists)

			doesntExist, err := query.Query().Has("Books").DoesntExist()
			s.ErrorIs(err, errors.OrmQueryRelationshipModel)
			s.False(doesntExist)

			// A Row-based aggregate returns the collected build error instead of panicking on the nil row.
			var sum float64
			var sumErr error
			s.NotPanics(func() {
				sumErr = query.Query().Model(&User{}).Has("Missing").Sum("id", &sum)
			})
			s.ErrorIs(sumErr, errors.OrmQueryRelationshipNotFound)

			var total int64
			s.ErrorIs(query.Query().Has("Books").Paginate(1, 10, nil, &total), errors.OrmQueryRelationshipModel)

			// The callback must return a query created by the ORM.
			err = query.Query().Model(&User{}).WhereHas("Books", func(query contractsorm.Query) contractsorm.Query {
				return nil
			}).Find(&result)
			s.ErrorIs(err, errors.OrmQueryInvalidSubQuery)
			s.Empty(result)

			// A callback that builds an invalid query fails the terminal operation with its error.
			err = query.Query().Model(&User{}).WhereHas("Books", func(query contractsorm.Query) contractsorm.Query {
				return query.WhereColumn("name", "=", "title", "extra")
			}).Find(&result)
			s.ErrorIs(err, errors.DatabaseInvalidArgumentNumber)
			s.Empty(result)

			_, err = query.Query().Model(&User{}).WhereHas("Books", func(query contractsorm.Query) contractsorm.Query {
				return query.WhereColumn("name", "=", "title", "extra")
			}).Count()
			s.ErrorIs(err, errors.DatabaseInvalidArgumentNumber)
		})
	}
}
