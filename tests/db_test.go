//go:debug x509negativeserial=1

package tests

import (
	"database/sql"
	"testing"

	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/convert"
	"github.com/goravel/postgres"
	"github.com/goravel/sqlite"
	"github.com/goravel/sqlserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DBTestSuite struct {
	suite.Suite
	now     carbon.DateTime
	queries map[string]*TestQuery
}

func TestDBTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &DBTestSuite{
		queries: make(map[string]*TestQuery),
	})
}

func (s *DBTestSuite) SetupSuite() {
	s.now = carbon.NewDateTime(carbon.FromDateTime(2025, 1, 2, 3, 4, 5))
	s.queries = NewTestQueryBuilder().All("", false)
}

func (s *DBTestSuite) SetupTest() {
	for _, query := range s.queries {
		query.CreateTable(TestTableProducts)
	}
}

func (s *DBTestSuite) TearDownSuite() {
	if s.queries[sqlite.Name] != nil {
		docker, err := s.queries[sqlite.Name].Driver().Docker()
		s.NoError(err)
		s.NoError(docker.Shutdown())
	}
}

func (s *DBTestSuite) TestCount() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "count_product1"},
				{Name: "count_product2"},
			})
			count, err := query.DB().Table("products").Count()
			s.NoError(err)
			s.Equal(int64(2), count)
		})
	}
}

func (s *DBTestSuite) TestCrossJoin() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]map[string]any{
				{
					"name":   "cross_join_product1",
					"weight": 100,
					"height": 200,
				},
				{
					"name":   "cross_join_product2",
					"weight": 200,
					"height": 100,
				},
			})

			type Temp struct {
				P1Name string `db:"p1_name"`
				P2Name string `db:"p2_name"`
			}
			var temps []Temp
			err := query.DB().Table("products as p1").CrossJoin("products as p2").Select("p1.name as p1_name", "p2.name as p2_name").Get(&temps)

			s.NoError(err)
			s.Equal(4, len(temps))
			s.Contains(temps, Temp{P1Name: "cross_join_product1", P2Name: "cross_join_product1"})
			s.Contains(temps, Temp{P1Name: "cross_join_product1", P2Name: "cross_join_product2"})
			s.Contains(temps, Temp{P1Name: "cross_join_product2", P2Name: "cross_join_product1"})
			s.Contains(temps, Temp{P1Name: "cross_join_product2", P2Name: "cross_join_product2"})
		})
	}
}

func (s *DBTestSuite) TestDecrement() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert(Product{Name: "decrement_product", Weight: convert.Pointer(100)})

			s.Run("decrement", func() {
				err := query.DB().Table("products").Where("name", "decrement_product").Decrement("weight", 1)
				s.NoError(err)

				var product Product
				err = query.DB().Table("products").Where("name", "decrement_product").First(&product)
				s.NoError(err)
				s.Equal(99, *product.Weight)
			})

			s.Run("decrement with number", func() {
				err := query.DB().Table("products").Where("name", "decrement_product").Decrement("weight", 5)
				s.NoError(err)

				var product Product
				err = query.DB().Table("products").Where("name", "decrement_product").First(&product)
				s.NoError(err)
				s.Equal(94, *product.Weight)
			})
		})
	}
}

func (s *DBTestSuite) TestDistinct() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "distinct_product"},
				{Name: "distinct_product"},
			})

			var products []Product
			err := query.DB().Table("products").Distinct().Select("name").Get(&products)
			s.NoError(err)
			s.Equal(1, len(products))
			s.Equal("distinct_product", products[0].Name)
		})
	}
}

func (s *DBTestSuite) TestExists() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert(Product{Name: "exists_product"})
			exists, err := query.DB().Table("products").Where("name", "exists_product").Exists()
			s.NoError(err)
			s.True(exists)

			query.DB().Table("products").Where("name", "exists_product").Delete()
			exists, err = query.DB().Table("products").Where("name", "exists_product").Exists()
			s.NoError(err)
			s.False(exists)
		})
	}
}

func (s *DBTestSuite) TestGroupBy_Having() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]map[string]any{
				{
					"name":   "group_by_having_product1",
					"weight": 25,
				},
				{
					"name":   "group_by_having_product2",
					"weight": 25,
				},
				{
					"name":   "group_by_having_product3",
					"weight": 30,
				},
			})

			var products []Product
			err := query.DB().Table("products").GroupBy("weight").Having("weight > ?", 20).OrderBy("weight").Select("weight").Get(&products)
			s.NoError(err)
			s.Equal(2, len(products))
			s.Equal(25, *products[0].Weight)
			s.Equal(30, *products[1].Weight)
		})
	}
}

func (s *DBTestSuite) TestIncrement() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert(Product{Name: "increment_product", Weight: convert.Pointer(100)})

			s.Run("increment", func() {
				err := query.DB().Table("products").Where("name", "increment_product").Increment("weight", 1)
				s.NoError(err)

				var product Product
				err = query.DB().Table("products").Where("name", "increment_product").First(&product)
				s.NoError(err)
				s.Equal(101, *product.Weight)
			})

			s.Run("increment with number", func() {
				err := query.DB().Table("products").Where("name", "increment_product").Increment("weight", 5)
				s.NoError(err)

				var product Product
				err = query.DB().Table("products").Where("name", "increment_product").First(&product)
				s.NoError(err)
				s.Equal(106, *product.Weight)
			})
		})
	}
}

func (s *DBTestSuite) TestInsert_First_Get() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			s.Run("single struct", func() {
				result, err := query.DB().Table("products").Insert(Product{
					Name: "single struct",
					Model: Model{
						Timestamps: Timestamps{
							CreatedAt: s.now,
							UpdatedAt: s.now,
						},
					},
				})

				s.NoError(err)
				s.Equal(int64(1), result.RowsAffected)

				var product Product
				err = query.DB().Table("products").Where("name", "single struct").Where("deleted_at", nil).First(&product)
				s.NoError(err)
				s.True(product.ID > 0)
				s.Equal("single struct", product.Name)
				s.Equal(s.now, product.CreatedAt)
				s.Equal(s.now, product.UpdatedAt)
				s.False(product.DeletedAt.Valid)
			})

			s.Run("multiple structs", func() {
				result, err := query.DB().Table("products").Insert([]Product{
					{
						Name: "multiple structs1",
						Model: Model{
							Timestamps: Timestamps{
								CreatedAt: s.now,
								UpdatedAt: s.now,
							},
						},
					},
					{
						Name: "multiple structs2",
					},
				})
				s.NoError(err)
				s.Equal(int64(2), result.RowsAffected)

				var products []Product
				err = query.DB().Table("products").Where("name", []string{"multiple structs1", "multiple structs2"}).Where("deleted_at", nil).Get(&products)
				s.NoError(err)
				s.Equal(2, len(products))
				s.Equal("multiple structs1", products[0].Name)
				s.Equal("multiple structs2", products[1].Name)
			})

			s.Run("single map", func() {
				result, err := query.DB().Table("products").Insert(map[string]any{
					"name":       "single map",
					"created_at": s.now,
					"updated_at": &s.now,
				})
				s.NoError(err)
				s.Equal(int64(1), result.RowsAffected)

				var product Product
				err = query.DB().Table("products").Where("name", "single map").Where("deleted_at", nil).First(&product)
				s.NoError(err)
				s.Equal("single map", product.Name)
				s.Equal(s.now, product.CreatedAt)
				s.Equal(s.now, product.UpdatedAt)
				s.False(product.DeletedAt.Valid)
			})

			s.Run("multiple map", func() {
				result, err := query.DB().Table("products").Insert([]map[string]any{
					{
						"name":       "multiple map1",
						"created_at": s.now,
						"updated_at": &s.now,
					},
					{
						"name": "multiple map2",
					},
				})
				s.NoError(err)
				s.Equal(int64(2), result.RowsAffected)

				var products []Product
				err = query.DB().Table("products").Where("name", []string{"multiple map1", "multiple map2"}).Where("deleted_at", nil).Get(&products)
				s.NoError(err)
				s.Equal(2, len(products))
				s.Equal("multiple map1", products[0].Name)
				s.Equal("multiple map2", products[1].Name)
			})
		})
	}
}

func (s *DBTestSuite) TestInsertGetId() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			id, err := query.DB().Table("products").InsertGetId(Product{
				Name: "insert get id",
			})

			if driver == sqlserver.Name || driver == postgres.Name {
				s.Error(err)
				s.Equal(int64(0), id)
			} else {
				s.NoError(err)
				s.True(id > 0)

				var product Product
				err = query.DB().Table("products").Where("id", id).First(&product)
				s.NoError(err)
				s.Equal("insert get id", product.Name)
			}
		})
	}
}

func (s *DBTestSuite) TestJoin() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]map[string]any{
				{
					"name":   "join_product1",
					"weight": 100,
					"height": 200,
				},
				{
					"name":   "join_product2",
					"weight": 50,
					"height": 100,
				},
				{
					"name":   "join_product3",
					"weight": 50,
					"height": 300,
				},
			})

			type Temp struct {
				P1Name   string `db:"p1_name"`
				P1Weight int    `db:"p1_weight"`
				P2Name   string `db:"p2_name"`
				P2Weight int    `db:"p2_weight"`
			}
			var temps []Temp
			err := query.DB().Table("products as p1").Join("products as p2 ON p1.weight = p2.height").Select("p1.name as p1_name", "p1.weight as p1_weight", "p2.name as p2_name", "p2.weight as p2_weight").Get(&temps)

			s.NoError(err)
			s.Equal(1, len(temps))
			s.Equal("join_product1", temps[0].P1Name)
			s.Equal(100, temps[0].P1Weight)
			s.Equal("join_product2", temps[0].P2Name)
			s.Equal(50, temps[0].P2Weight)
		})
	}
}

func (s *DBTestSuite) TestLeftJoin() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]map[string]any{
				{
					"name":   "join_product1",
					"weight": 100,
					"height": 200,
				},
				{
					"name":   "join_product2",
					"weight": 50,
					"height": 100,
				},
				{
					"name":   "join_product3",
					"weight": 50,
					"height": 300,
				},
			})

			type Temp struct {
				P1Name   *string `db:"p1_name"`
				P1Weight *int    `db:"p1_weight"`
				P2Name   *string `db:"p2_name"`
				P2Weight *int    `db:"p2_weight"`
			}
			var temps []Temp
			err := query.DB().Table("products as p1").LeftJoin("products as p2 ON p1.weight = p2.height").Select("p1.name as p1_name", "p1.weight as p1_weight", "p2.name as p2_name", "p2.weight as p2_weight").Get(&temps)

			s.NoError(err)
			s.Equal(3, len(temps))
			s.Contains(temps, Temp{P1Name: convert.Pointer("join_product1"), P1Weight: convert.Pointer(100), P2Name: convert.Pointer("join_product2"), P2Weight: convert.Pointer(50)})
			s.Contains(temps, Temp{P1Name: convert.Pointer("join_product2"), P1Weight: convert.Pointer(50)})
			s.Contains(temps, Temp{P1Name: convert.Pointer("join_product3"), P1Weight: convert.Pointer(50)})
		})
	}
}

func (s *DBTestSuite) TestOrWhere() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{
					Name: "or where model",
				},
				{
					Name: "or where model1",
				},
			})

			s.Run("simple where condition", func() {
				var products []Product
				err := query.DB().Table("products").Where("name", "or where model").OrWhere("name", "or where model1").Get(&products)
				s.NoError(err)
				s.Equal(2, len(products))
				s.Equal("or where model", products[0].Name)
				s.Equal("or where model1", products[1].Name)
			})

			s.Run("nested condition", func() {
				var products []Product
				err := query.DB().Table("products").Where("name", "or where model").OrWhere(func(query db.Query) db.Query {
					return query.Where("name", "or where model1")
				}).Get(&products)
				s.NoError(err)
				s.Equal(2, len(products))
				s.Equal("or where model", products[0].Name)
				s.Equal("or where model1", products[1].Name)
			})
		})
	}
}

func (s *DBTestSuite) TestOrWhereColumn() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{
					Name:   "or where column model",
					Height: convert.Pointer(100),
					Weight: convert.Pointer(100),
				},
				{
					Name:   "or where column model1",
					Height: convert.Pointer(100),
					Weight: convert.Pointer(110),
				},
			})

			s.Run("simple condition", func() {
				var products []Product
				err := query.DB().Table("products").Where("name", "or where column model1").OrWhereColumn("height", "weight").Get(&products)
				s.NoError(err)
				s.Equal(2, len(products))
				s.Equal("or where column model", products[0].Name)
				s.Equal("or where column model1", products[1].Name)
			})

			s.Run("with operator", func() {
				var products []Product
				err := query.DB().Table("products").Where("name", "or where column model").OrWhereColumn("height", "<", "weight").Get(&products)
				s.NoError(err)
				s.Equal(2, len(products))
				s.Equal("or where column model", products[0].Name)
				s.Equal("or where column model1", products[1].Name)
			})

			s.Run("with multiple columns", func() {
				var product Product
				err := query.DB().Table("products").OrWhereColumn("name", ">", "age", "name").First(&product)
				s.Equal(errors.DatabaseInvalidArgumentNumber.Args(3, "1 or 2"), err)
			})

			s.Run("with not enough arguments", func() {
				var product Product
				err := query.DB().Table("products").OrWhereColumn("name").First(&product)
				s.Equal(errors.DatabaseInvalidArgumentNumber.Args(2, "1 or 2"), err)
			})
		})
	}
}

func (s *DBTestSuite) TestOrWhereNot() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{
					Name: "or where not model1",
				},
				{
					Name: "or where not model2",
				},
			})

			s.Run("simple condition", func() {
				var product Product
				err := query.DB().Table("products").Where("name", "or where not model1").OrWhereNot("name", "or where not model2").First(&product)
				s.NoError(err)
				s.Equal("or where not model1", product.Name)
			})

			s.Run("raw query", func() {
				var product Product
				err := query.DB().Table("products").Where("name", "or where not model1").OrWhereNot("name = ?", "or where not model2").First(&product)
				s.NoError(err)
				s.Equal("or where not model1", product.Name)
			})

			s.Run("nested condition", func() {
				var product Product
				err := query.DB().Table("products").Where("name", "or where not model1").OrWhereNot(func(query db.Query) db.Query {
					return query.Where("name", "or where not model2")
				}).First(&product)
				s.NoError(err)
				s.Equal("or where not model1", product.Name)
			})
		})
	}
}

func (s *DBTestSuite) TestPluck() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "pluck_product1"},
				{Name: "pluck_product2"},
			})

			var names []string
			err := query.DB().Table("products").WhereLike("name", "pluck_product%").Pluck("name", &names)

			s.NoError(err)
			s.Equal([]string{"pluck_product1", "pluck_product2"}, names)
		})
	}
}

func (s *DBTestSuite) TestRightJoin() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]map[string]any{
				{
					"name":   "join_product1",
					"weight": 100,
					"height": 200,
				},
				{
					"name":   "join_product2",
					"weight": 50,
					"height": 100,
				},
				{
					"name":   "join_product3",
					"weight": 50,
					"height": 300,
				},
			})

			type Temp struct {
				P1Name   *string `db:"p1_name"`
				P1Weight *int    `db:"p1_weight"`
				P2Name   *string `db:"p2_name"`
				P2Weight *int    `db:"p2_weight"`
			}
			var temps []Temp
			err := query.DB().Table("products as p1").RightJoin("products as p2 ON p1.weight = p2.height").Select("p1.name as p1_name", "p1.weight as p1_weight", "p2.name as p2_name", "p2.weight as p2_weight").Get(&temps)

			s.NoError(err)
			s.Equal(3, len(temps))
			s.Contains(temps, Temp{P1Name: convert.Pointer("join_product1"), P1Weight: convert.Pointer(100), P2Name: convert.Pointer("join_product2"), P2Weight: convert.Pointer(50)})
			s.Contains(temps, Temp{P2Name: convert.Pointer("join_product1"), P2Weight: convert.Pointer(100)})
			s.Contains(temps, Temp{P2Name: convert.Pointer("join_product3"), P2Weight: convert.Pointer(50)})
		})
	}
}

func (s *DBTestSuite) TestTransaction() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			s.Run("separate transaction", func() {
				tx, err := query.DB().BeginTransaction()
				s.NoError(err)
				s.NotNil(tx)

				result, err := tx.Table("products").Insert(Product{Name: "transaction product"})
				s.NoError(err)
				s.Equal(int64(1), result.RowsAffected)

				s.NoError(tx.Commit())

				var product Product
				err = query.DB().Table("products").Where("name", "transaction product").First(&product)
				s.NoError(err)
				s.Equal("transaction product", product.Name)

				tx, err = query.DB().BeginTransaction()
				s.NoError(err)
				s.NotNil(tx)

				result, err = tx.Table("products").Where("name", "transaction product").Update("name", "transaction product updated")
				s.NoError(err)
				s.Equal(int64(1), result.RowsAffected)
				s.NoError(tx.Rollback())

				var product1 Product
				err = query.DB().Table("products").Where("name", "transaction product").First(&product1)
				s.NoError(err)
				s.Equal("transaction product", product1.Name)

				var product2 Product
				err = query.DB().Table("products").Where("name", "transaction product updated").FirstOrFail(&product2)
				s.Equal(sql.ErrNoRows, err)
			})

			s.Run("transaction", func() {
				err := query.DB().Transaction(func(tx db.DB) error {
					_, err := tx.Table("products").Insert(Product{Name: "transaction product1"})
					if err != nil {
						return err
					}

					_, err = tx.Table("products").Where("name", "transaction product1").Update("name", "transaction product1 updated")
					if err != nil {
						return err
					}

					return nil
				})

				s.NoError(err)

				var product Product
				err = query.DB().Table("products").Where("name", "transaction product1 updated").First(&product)
				s.NoError(err)
				s.Equal("transaction product1 updated", product.Name)

				err = query.DB().Transaction(func(tx db.DB) error {
					_, err := tx.Table("products").Where("name", "transaction product1 updated").Delete()
					if err != nil {
						return err
					}

					return assert.AnError
				})

				s.Equal(assert.AnError, err)

				var product1 Product
				err = query.DB().Table("products").Where("name", "transaction product1 updated").First(&product1)
				s.NoError(err)
				s.Equal("transaction product1 updated", product1.Name)
			})
		})
	}
}

func (s *DBTestSuite) TestUpdate_Delete() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			result, err := query.DB().Table("products").Insert([]Product{
				{
					Name: "update structs1",
					Model: Model{
						Timestamps: Timestamps{
							CreatedAt: s.now,
							UpdatedAt: s.now,
						},
					},
				},
				{
					Name: "update structs2",
				},
			})
			s.NoError(err)
			s.Equal(int64(2), result.RowsAffected)

			// Create success
			var products1 []Product
			err = query.DB().Table("products").Where("name", []string{"update structs1", "update structs2"}).Where("deleted_at", nil).Get(&products1)
			s.NoError(err)
			s.Equal(2, len(products1))
			s.Equal("update structs1", products1[0].Name)
			s.Equal("update structs2", products1[1].Name)

			// Update success via map
			result, err = query.DB().Table("products").Where("name", "update structs1").Update(map[string]any{
				"name": "update structs1 updated",
			})
			s.NoError(err)
			s.Equal(int64(1), result.RowsAffected)

			var product1 Product
			err = query.DB().Table("products").Where("name", "update structs1 updated").Where("deleted_at", nil).First(&product1)
			s.NoError(err)
			s.Equal("update structs1 updated", product1.Name)

			// Update success via struct
			result, err = query.DB().Table("products").Where("name", "update structs2").Update(Product{
				Name: "update structs2 updated",
			})
			s.NoError(err)
			s.Equal(int64(1), result.RowsAffected)

			var product2 Product
			err = query.DB().Table("products").Where("name", "update structs2 updated").Where("deleted_at", nil).First(&product2)
			s.NoError(err)
			s.Equal("update structs2 updated", product2.Name)

			// Delete success
			result, err = query.DB().Table("products").Where("name like ?", "update structs%").Delete()
			s.NoError(err)
			s.Equal(int64(2), result.RowsAffected)

			var products2 []Product
			err = query.DB().Table("products").Where("name", []string{"update structs1 updated", "update structs2 updated"}).Where("deleted_at", nil).Get(&products2)
			s.NoError(err)
			s.Equal(0, len(products2))
		})
	}
}

// func (s *DBTestSuite) TestValue() {
// 	for driver, query := range s.queries {
// 		s.Run(driver, func() {
// 			query.DB().Table("products").Insert(Product{Name: "value_product"})

// 			var name string
// 			err := query.DB().Table("products").Where("name", "value_product").Value("name", &name)

// 			s.NoError(err)
// 			s.Equal("value_product", name)
// 		})
// 	}
// }

func (s *DBTestSuite) TestWhere() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert(Product{
				Name: "where model",
				Model: Model{
					Timestamps: Timestamps{
						CreatedAt: s.now,
						UpdatedAt: s.now,
					},
				},
			})

			s.Run("simple condition", func() {
				var product Product
				err := query.DB().Table("products").Where("name", "where model").First(&product)
				s.NoError(err)
				s.Equal("where model", product.Name)
			})

			s.Run("multiple arguments", func() {
				var products []Product
				err := query.DB().Table("products").Where("name", []string{"where model", "where model1"}).Get(&products)
				s.NoError(err)
				s.Equal(1, len(products))
				s.Equal("where model", products[0].Name)
			})

			s.Run("raw query", func() {
				var product Product
				err := query.DB().Table("products").Where("name = ?", "where model").First(&product)
				s.NoError(err)
				s.Equal("where model", product.Name)
			})

			s.Run("nested condition", func() {
				var product Product
				err := query.DB().Table("products").Where(func(query db.Query) db.Query {
					return query.Where("name", "where model")
				}).First(&product)
				s.NoError(err)
				s.Equal("where model", product.Name)
			})
		})
	}
}

func (s *DBTestSuite) TestWhereColumn() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{
					Name:   "where column model",
					Weight: convert.Pointer(100),
					Height: convert.Pointer(100),
				},
				{
					Name:   "where column model1",
					Weight: convert.Pointer(100),
					Height: convert.Pointer(110),
				},
			})

			s.Run("simple condition", func() {
				var products []Product
				err := query.DB().Table("products").WhereColumn("height", "weight").Get(&products)
				s.NoError(err)
				s.Equal(1, len(products))
				s.Equal("where column model", products[0].Name)
			})

			s.Run("with operator", func() {
				var products []Product
				err := query.DB().Table("products").WhereColumn("height", ">", "weight").Get(&products)
				s.NoError(err)
				s.Equal(1, len(products))
				s.Equal("where column model1", products[0].Name)
			})

			s.Run("with multiple columns", func() {
				var product Product
				err := query.DB().Table("products").WhereColumn("name", ">", "age", "name").First(&product)
				s.Equal(errors.DatabaseInvalidArgumentNumber.Args(3, "1 or 2"), err)
			})

			s.Run("with not enough arguments", func() {
				var product Product
				err := query.DB().Table("products").WhereColumn("name").First(&product)
				s.Equal(errors.DatabaseInvalidArgumentNumber.Args(2, "1 or 2"), err)
			})
		})
	}
}

func (s *DBTestSuite) TestWhereExists() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			result, err := query.DB().Table("products").Insert([]Product{
				{
					Name:   "where exists model1",
					Height: convert.Pointer(100),
					Weight: convert.Pointer(100),
				},
				{
					Name:   "where exists model2",
					Height: convert.Pointer(100),
					Weight: convert.Pointer(110),
				},
			})
			s.NoError(err)
			s.Equal(int64(2), result.RowsAffected)

			s.Run("simple condition", func() {
				var product Product
				err = query.DB().Table("products").WhereExists(func() db.Query {
					return query.DB().Table("products").Where("name", "where exists model1")
				}).First(&product)
				s.NoError(err)
				s.Equal("where exists model1", product.Name)
			})

			s.Run("with WhereColumn", func() {
				var product Product
				err = query.DB().Table("products").WhereExists(func() db.Query {
					return query.DB().Table("products").WhereColumn("height", "weight")
				}).First(&product)
				s.NoError(err)
				s.Equal("where exists model1", product.Name)
			})
		})
	}
}

func (s *DBTestSuite) TestWhereNot() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{
					Name: "where not model1",
				},
				{
					Name: "where not model2",
				},
			})

			s.Run("simple condition", func() {
				var product Product
				err := query.DB().Table("products").Where("name", "where not model1").WhereNot("name", "where not model2").First(&product)
				s.NoError(err)
				s.Equal("where not model1", product.Name)
			})

			s.Run("raw query", func() {
				var product Product
				err := query.DB().Table("products").Where("name", "where not model1").WhereNot("name = ?", "where not model2").First(&product)
				s.NoError(err)
				s.Equal("where not model1", product.Name)
			})

			s.Run("nested condition", func() {
				var product Product
				err := query.DB().Table("products").Where("name", "where not model1").WhereNot(func(query db.Query) db.Query {
					return query.Where("name", "where not model2")
				}).First(&product)
				s.NoError(err)
				s.Equal("where not model1", product.Name)
			})
		})
	}
}

func TestDB_Connection(t *testing.T) {
	t.Parallel()
	postgresTestQuery := NewTestQueryBuilder().Postgres("", false)
	postgresTestQuery.CreateTable(TestTableProducts)

	sqliteTestQuery := NewTestQueryBuilder().Sqlite("", false)
	sqliteTestQuery.CreateTable(TestTableProducts)
	defer func() {
		docker, err := sqliteTestQuery.Driver().Docker()
		assert.NoError(t, err)
		assert.NoError(t, docker.Shutdown())
	}()

	sqliteConnection := sqliteTestQuery.Driver().Config().Connection
	mockDatabaseConfig(postgresTestQuery.MockConfig(), sqliteTestQuery.Driver().Config(), sqliteConnection, "", false)

	result, err := postgresTestQuery.DB().Table("products").Insert(Product{
		Name: "connection",
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.RowsAffected)

	var product Product
	err = postgresTestQuery.DB().Table("products").Where("name", "connection").First(&product)
	assert.NoError(t, err)
	assert.True(t, product.ID > 0)
	assert.Equal(t, "connection", product.Name)

	var product1 Product
	err = postgresTestQuery.DB().Connection(sqliteConnection).Table("products").Where("name", "connection").First(&product1)
	assert.NoError(t, err)
	assert.True(t, product1.ID == 0)

	result, err = postgresTestQuery.DB().Connection(sqliteConnection).Table("products").Insert(Product{
		Name: "sqlite connection",
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.RowsAffected)

	var product2 Product
	err = postgresTestQuery.DB().Connection(sqliteConnection).Table("products").Where("name", "sqlite connection").First(&product2)
	assert.NoError(t, err)
	assert.True(t, product2.ID > 0)
	assert.Equal(t, "sqlite connection", product2.Name)

	var product3 Product
	err = postgresTestQuery.DB().Table("products").Where("name", "sqlite connection").First(&product3)
	assert.NoError(t, err)
	assert.True(t, product3.ID == 0)
}
