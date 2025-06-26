//go:debug x509negativeserial=1

package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/goravel/framework/contracts/database/db"
	databasedb "github.com/goravel/framework/database/db"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/convert"
	"github.com/goravel/mysql"
	"github.com/goravel/postgres"
	"github.com/goravel/sqlite"
	"github.com/goravel/sqlserver"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DBTestSuite struct {
	suite.Suite
	now     *carbon.DateTime
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

	mockApp := &mocksfoundation.Application{}
	mockApp.EXPECT().GetJson().Return(json.New())
	postgres.App = mockApp
	mysql.App = mockApp
	sqlite.App = mockApp
	sqlserver.App = mockApp
}

func (s *DBTestSuite) SetupTest() {
	for _, query := range s.queries {
		query.CreateTable(TestTableProducts)
		query.CreateTable(TestTableJsonData)
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

func (s *DBTestSuite) TestChunk() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "chunk_product1"},
				{Name: "chunk_product2"},
				{Name: "chunk_product3"},
			})

			var products []Product
			err := query.DB().Table("products").Chunk(2, func(rows []db.Row) error {
				for _, row := range rows {
					var product Product
					err := row.Scan(&product)
					s.NoError(err)
					products = append(products, product)
				}

				return nil
			})
			s.NoError(err)
			s.Equal(3, len(products))
			s.Equal("chunk_product1", products[0].Name)
			s.Equal("chunk_product2", products[1].Name)
			s.Equal("chunk_product3", products[2].Name)
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

func (s *DBTestSuite) TestCursor() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{
					Name: "cursor_product1", Weight: convert.Pointer(100), Model: Model{
						Timestamps: Timestamps{
							CreatedAt: s.now,
							UpdatedAt: s.now,
						},
					},
				},
				{
					Name: "cursor_product2", Weight: convert.Pointer(200), Model: Model{
						Timestamps: Timestamps{
							CreatedAt: s.now,
							UpdatedAt: s.now,
						},
					},
				},
			})

			s.Run("Bind Struct", func() {
				rows := query.DB().Table("products").Cursor()

				var products []Product
				for row := range rows {
					s.NoError(row.Err())

					var product Product
					err := row.Scan(&product)
					s.NoError(err)
					s.True(product.ID > 0)

					products = append(products, product)
				}

				s.Equal(2, len(products))
				s.Equal("cursor_product1", products[0].Name)
				s.Equal(100, *products[0].Weight)
				s.Equal(s.now, products[0].CreatedAt)
				s.Equal(s.now, products[0].UpdatedAt)
				s.Equal("cursor_product2", products[1].Name)
				s.Equal(200, *products[1].Weight)
				s.Equal(s.now, products[1].CreatedAt)
				s.Equal(s.now, products[1].UpdatedAt)
			})

			s.Run("Bind Map", func() {
				rows := query.DB().Table("products").Cursor()

				var products []map[string]any
				for row := range rows {
					s.NoError(row.Err())

					var product map[string]any
					err := row.Scan(&product)
					s.NoError(err)
					s.True(cast.ToUint(product["id"]) > 0)

					products = append(products, product)
				}

				s.Equal(2, len(products))
				s.Equal("cursor_product1", cast.ToString(products[0]["name"]))
				s.Equal(100, cast.ToInt(products[0]["weight"]))
				s.Equal(s.now, carbon.NewDateTime(carbon.FromStdTime(cast.ToTime(products[0]["created_at"]))))
				s.Equal(s.now, carbon.NewDateTime(carbon.FromStdTime(cast.ToTime(products[0]["updated_at"]))))
				s.Equal("cursor_product2", cast.ToString(products[1]["name"]))
				s.Equal(200, cast.ToInt(products[1]["weight"]))
				s.Equal(s.now, carbon.NewDateTime(carbon.FromStdTime(cast.ToTime(products[1]["created_at"]))))
				s.Equal(s.now, carbon.NewDateTime(carbon.FromStdTime(cast.ToTime(products[1]["updated_at"]))))
			})

			s.Run("Cursor error", func() {
				for row := range query.DB().Table("not_exist").Cursor() {
					err1 := row.Err()
					s.Error(err1)

					err2 := row.Scan(map[string]any{})
					s.Error(err2)

					s.Equal(err1, err2)
				}
			})
		})
	}
}

func (s *DBTestSuite) Test_DB_Select_Insert_Update_Delete_Statement() {
	for driver, query := range s.queries {
		insertSql := "INSERT INTO products (name) VALUES (?)"
		updateSql := "UPDATE products SET name = ? WHERE id = ?"
		deleteSql := "DELETE FROM products WHERE id = ?"

		s.Run(driver, func() {
			if driver == sqlserver.Name {
				insertSql = "INSERT INTO products (name) VALUES (@p1)"
				updateSql = "UPDATE products SET name = @p1 WHERE id = @p2"
				deleteSql = "DELETE FROM products WHERE id = @p1"
			}
			if driver == postgres.Name {
				insertSql = "INSERT INTO products (name) VALUES ($1)"
				updateSql = "UPDATE products SET name = $1 WHERE id = $2"
				deleteSql = "DELETE FROM products WHERE id = $1"
			}

			result, err := query.DB().Insert(insertSql, "test_db_select_update_delete_product")
			s.NoError(err)
			s.Equal(int64(1), result.RowsAffected)

			var products []Product
			err = query.DB().Select(&products, "SELECT * FROM products")
			s.NoError(err)
			s.Equal(1, len(products))
			s.Equal("test_db_select_update_delete_product", products[0].Name)

			var product Product
			err = query.DB().Select(&product, "SELECT * FROM products")
			s.NoError(err)
			s.Equal("test_db_select_update_delete_product", product.Name)

			result, err = query.DB().Update(updateSql, "test_db_select_update_delete_product_updated", products[0].ID)
			s.NoError(err)
			s.Equal(int64(1), result.RowsAffected)

			err = query.DB().Select(&products, "SELECT * FROM products")
			s.NoError(err)
			s.Equal(1, len(products))
			s.Equal("test_db_select_update_delete_product_updated", products[0].Name)

			result, err = query.DB().Delete(deleteSql, products[0].ID)
			s.NoError(err)
			s.Equal(int64(1), result.RowsAffected)

			err = query.DB().Select(&products, "SELECT * FROM products")
			s.NoError(err)
			s.Equal(0, len(products))

			err = query.DB().Statement("drop table products")
			s.NoError(err)
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

			var products1 []Product
			err = query.DB().Table("products").Distinct("name").Get(&products1)
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

func (s *DBTestSuite) TestInRandomOrder() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "in_random_order_product1"},
				{Name: "in_random_order_product2"},
				{Name: "in_random_order_product3"},
				{Name: "in_random_order_product4"},
				{Name: "in_random_order_product5"},
				{Name: "in_random_order_product6"},
				{Name: "in_random_order_product7"},
				{Name: "in_random_order_product8"},
				{Name: "in_random_order_product9"},
				{Name: "in_random_order_product10"},
			})

			var products1 []Product
			err := query.DB().Table("products").InRandomOrder().Get(&products1)
			s.NoError(err)
			s.Equal(10, len(products1))

			var names1 []string
			for _, product := range products1 {
				names1 = append(names1, product.Name)
			}

			var products2 []Product
			err = query.DB().Table("products").InRandomOrder().Get(&products2)
			s.NoError(err)
			s.Equal(10, len(products2))

			var names2 []string
			for _, product := range products2 {
				names2 = append(names2, product.Name)
			}

			s.NotEqual(names1, names2)
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

func (s *DBTestSuite) TestInsertGetID() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			id, err := query.DB().Table("products").InsertGetID(Product{
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

func (s *DBTestSuite) TestLimit() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "limit_product1"},
				{Name: "limit_product2"},
				{Name: "limit_product3"},
			})

			var products []Product
			err := query.DB().Table("products").Limit(1).Get(&products)
			s.NoError(err)
			s.Equal(1, len(products))
			s.Equal("limit_product1", products[0].Name)
		})
	}
}

func (s *DBTestSuite) TestLockForUpdate() {
	for driver, query := range s.queries {
		if driver == sqlite.Name {
			continue
		}

		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "lock_for_update_product"},
			})

			var product Product
			err := query.DB().Table("products").Where("name", "lock_for_update_product").First(&product)
			s.NoError(err)
			s.Equal("lock_for_update_product", product.Name)
			s.True(product.ID > 0)

			for i := 0; i < 10; i++ {
				go func() {
					query.DB().Transaction(func(tx db.Tx) error {
						var product1 Product
						err := tx.Table("products").Where("id", product.ID).LockForUpdate().First(&product1)
						s.NoError(err)
						s.True(product1.ID > 0)

						_, err = tx.Table("products").Where("id", product1.ID).Update("name", product1.Name+"1")
						s.NoError(err)

						return nil
					})
				}()
			}

			time.Sleep(2 * time.Second)

			err = query.DB().Table("products").Where("id", product.ID).First(&product)
			s.NoError(err)
			s.Equal("lock_for_update_product1111111111", product.Name)
			s.True(product.ID > 0)
		})
	}
}

func (s *DBTestSuite) TestOrderBy() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "value_product"},
				{Name: "value_product1"},
			})

			var name string
			err := query.DB().Table("products").OrderBy("id").Value("name", &name)
			s.NoError(err)
			s.Equal("value_product", name)

			var name1 string
			err = query.DB().Table("products").OrderBy("id", "desc").Value("name", &name1)
			s.NoError(err)
			s.Equal("value_product1", name1)

			var name2 string
			err = query.DB().Table("products").OrderByDesc("id").Value("name", &name2)
			s.NoError(err)
			s.Equal("value_product1", name2)
		})
	}
}

func (s *DBTestSuite) TestOffset() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "offset_product1"},
				{Name: "offset_product2"},
				{Name: "offset_product3"},
			})

			var products []Product
			err := query.DB().Table("products").Offset(1).Limit(1).Get(&products)

			s.NoError(err)
			s.Equal(1, len(products))
			s.Equal("offset_product2", products[0].Name)
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

func (s *DBTestSuite) TestPaginate() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "paginate_product1"},
				{Name: "paginate_product2"},
				{Name: "paginate_product3"},
				{Name: "paginate_product4"},
				{Name: "paginate_product5"},
			})

			var products []Product
			var total int64
			err := query.DB().Table("products").WhereLike("name", "paginate_product%").Paginate(1, 2, &products, &total)
			s.NoError(err)
			s.Equal(2, len(products))
			s.Equal(int64(5), total)
			s.Equal("paginate_product1", products[0].Name)
			s.Equal("paginate_product2", products[1].Name)

			products = []Product{}
			err = query.DB().Table("products").WhereLike("name", "paginate_product%").Paginate(2, 2, &products, &total)
			s.NoError(err)
			s.Equal(2, len(products))
			s.Equal(int64(5), total)
			s.Equal("paginate_product3", products[0].Name)
			s.Equal("paginate_product4", products[1].Name)
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

func (s *DBTestSuite) TestQueryLog() {
	query := s.queries[postgres.Name]
	ctx := context.Background()
	ctx = databasedb.EnableQueryLog(ctx)

	query.DB().WithContext(ctx).Table("products").Insert([]Product{
		{Name: "query_log_product"},
	})

	var product Product
	err := query.DB().WithContext(ctx).Table("products").Where("name", "query_log_product").First(&product)
	s.NoError(err)
	s.True(product.ID > 0)

	queryLogs := databasedb.GetQueryLog(ctx)
	s.Equal(2, len(queryLogs))
	s.Equal("INSERT INTO products (name) VALUES ('query_log_product')", queryLogs[0].Query)
	s.True(queryLogs[0].Time > 0)
	s.Equal("SELECT * FROM products WHERE name = 'query_log_product'", queryLogs[1].Query)
	s.True(queryLogs[1].Time > 0)

	ctx = databasedb.DisableQueryLog(ctx)
	query.DB().WithContext(ctx).Table("products").Insert([]Product{
		{Name: "query_log_product"},
	})

	queryLogs = databasedb.GetQueryLog(ctx)
	s.Equal(0, len(queryLogs))
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

func (s *DBTestSuite) TestSharedLock() {
	for driver, query := range s.queries {
		if driver == sqlite.Name {
			continue
		}

		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "shared_lock_product"},
			})

			var product Product
			err := query.DB().Table("products").Where("name", "shared_lock_product").SharedLock().First(&product)
			s.NoError(err)
			s.Equal("shared_lock_product", product.Name)
			s.True(product.ID > 0)

			tx, err := query.DB().BeginTransaction()
			s.NoError(err)

			var product1 Product
			err = tx.Table("products").Where("id", product.ID).SharedLock().First(&product1)
			s.NoError(err)
			s.Equal("shared_lock_product", product1.Name)
			s.True(product1.ID > 0)

			var product2 Product
			err = query.DB().Table("products").Where("id", product.ID).SharedLock().First(&product2)
			s.NoError(err)
			s.Equal("shared_lock_product", product2.Name)
			s.True(product2.ID > 0)

			product1.Name += "1"
			_, err = tx.Table("products").Where("id", product1.ID).Update("name", product1.Name)
			s.NoError(err)

			s.NoError(tx.Commit())

			err = query.DB().Table("products").Where("id", product.ID).SharedLock().First(&product)
			s.NoError(err)
			s.Equal("shared_lock_product1", product.Name)
			s.True(product.ID > 0)
		})
	}
}

func (s *DBTestSuite) TestSum() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "sum_product1", Weight: convert.Pointer(100)},
				{Name: "sum_product2", Weight: convert.Pointer(200)},
			})

			sum, err := query.DB().Table("products").Sum("weight")
			s.NoError(err)
			s.Equal(int64(300), sum)
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
				err := query.DB().Transaction(func(tx db.Tx) error {
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

				err = query.DB().Transaction(func(tx db.Tx) error {
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

func (s *DBTestSuite) TestUpdateOrInsert() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			s.Run("map", func() {
				result, err := query.DB().Table("products").Where("height", 100).UpdateOrInsert(map[string]any{
					"name": "update or insert product1",
				}, map[string]any{
					"weight": 200,
				})
				s.NoError(err)
				s.Equal(int64(1), result.RowsAffected)

				var product1 Product
				err = query.DB().Table("products").Where("name", "update or insert product1").First(&product1)
				s.NoError(err)
				s.True(product1.ID > 0)
				s.Equal("update or insert product1", product1.Name)
				s.Equal(200, *product1.Weight)
				s.Nil(product1.Height)

				result, err = query.DB().Table("products").UpdateOrInsert(map[string]any{
					"name": "update or insert product1",
				}, map[string]any{
					"weight": 300,
				})
				s.NoError(err)
				s.Equal(int64(1), result.RowsAffected)

				var product2 Product
				err = query.DB().Table("products").Where("name", "update or insert product1").First(&product2)
				s.NoError(err)
				s.Equal("update or insert product1", product2.Name)
				s.Equal(300, *product2.Weight)
			})

			s.Run("struct", func() {
				result, err := query.DB().Table("products").Where("height", 100).UpdateOrInsert(Product{
					Name: "update or insert product2",
				}, Product{
					Weight: convert.Pointer(200),
				})
				s.NoError(err)
				s.Equal(int64(1), result.RowsAffected)

				var product1 Product
				err = query.DB().Table("products").Where("name", "update or insert product2").First(&product1)
				s.NoError(err)
				s.True(product1.ID > 0)
				s.Equal("update or insert product2", product1.Name)
				s.Equal(200, *product1.Weight)
				s.Nil(product1.Height)

				result, err = query.DB().Table("products").UpdateOrInsert(Product{
					Name: "update or insert product2",
				}, Product{
					Weight: convert.Pointer(300),
				})
				s.NoError(err)
				s.Equal(int64(1), result.RowsAffected)

				var product2 Product
				err = query.DB().Table("products").Where("name", "update or insert product2").First(&product2)
				s.NoError(err)
				s.Equal("update or insert product2", product2.Name)
				s.Equal(300, *product2.Weight)
			})
		})
	}
}

func (s *DBTestSuite) TestValue() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			query.DB().Table("products").Insert([]Product{
				{Name: "value_product"},
				{Name: "value_product1"},
			})

			var name string
			err := query.DB().Table("products").OrderByDesc("id").Value("name", &name)

			s.NoError(err)
			s.Equal("value_product1", name)
		})
	}
}

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

func (s *DBTestSuite) TestJsonWhereClauses() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			data := []JsonData{
				{
					Model: Model{
						Timestamps: Timestamps{
							CreatedAt: s.now,
							UpdatedAt: s.now,
						},
					},
					Data: `{"string":"first","int":123,"float":123.456,"bool":true,"array":["abc","def","ghi"],"nested":{"string":"first","int":456},"objects":[{"level":"first","value":"abc"},{"level":"second","value":"def"}]}`,
				},
				{
					Model: Model{
						Timestamps: Timestamps{
							CreatedAt: s.now,
							UpdatedAt: s.now,
						},
					},
					Data: `{"string":"second","int":123,"float":789.123,"bool":false,"array":["jkl","def","abc"]}`,
				},
			}
			result, err := query.DB().Table("json_data").Insert(&data)
			s.Equal(int64(2), result.RowsAffected)
			s.NoError(err)

			tests := []struct {
				name   string
				find   func(any, ...any) error
				assert func([]JsonData)
			}{
				{
					name: "string key",
					find: query.DB().Table("json_data").Where("data->string", "first").Find,
					assert: func(items []JsonData) {
						s.Len(items, 1)
						s.JSONEq(data[0].Data, items[0].Data)
					},
				},
				{
					name: "int key",
					find: query.DB().Table("json_data").Where("data->int", 123).Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
				{
					name: "float key(multiple values)",
					find: query.DB().Table("json_data").WhereIn("data->float", []any{123.456, 789.123}).Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
				{
					name: "bool key(pointer)",
					find: query.DB().Table("json_data").Where("data->bool", convert.Pointer(false)).Find,
					assert: func(items []JsonData) {
						s.Len(items, 1)
						s.JSONEq(data[1].Data, items[0].Data)
					},
				},
				{
					name: "nested key",
					find: query.DB().Table("json_data").Where("data->nested->int", 456).Find,
					assert: func(items []JsonData) {
						s.Len(items, 1)
						s.JSONEq(data[0].Data, items[0].Data)
					},
				},
				{
					name: "nested key with array",
					find: query.DB().Table("json_data").Where("data->objects[0]->level", "first").Find,
					assert: func(items []JsonData) {
						s.Len(items, 1)
						s.JSONEq(data[0].Data, items[0].Data)
					},
				},
				{
					name: "key exists",
					find: query.DB().Table("json_data").WhereJsonContainsKey("data->nested->string").Find,
					assert: func(items []JsonData) {
						s.Len(items, 1)
						s.JSONEq(data[0].Data, items[0].Data)
					},
				},
				{
					name: "key does not exist",
					find: query.DB().Table("json_data").WhereJsonDoesntContainKey("data->nested->string").Find,
					assert: func(items []JsonData) {
						s.Len(items, 1)
						s.JSONEq(data[1].Data, items[0].Data)
					},
				},
				{
					name: "array contains",
					find: query.DB().Table("json_data").WhereJsonContains("data->array", "abc").Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
				{
					name: "array does not contain",
					find: query.DB().Table("json_data").WhereJsonDoesntContain("data->array", "abc").Find,
					assert: func(items []JsonData) {
						s.Len(items, 0)
					},
				},
				{
					name: "array contains multiple values",
					find: query.DB().Table("json_data").WhereJsonContains("data->array", []string{"abc", "def"}).Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
				{
					name: "array length",
					find: query.DB().Table("json_data").WhereJsonLength("data->array", 2).Find,
					assert: func(items []JsonData) {
						s.Len(items, 0)
					},
				},
				{
					name: "array length greater than",
					find: query.DB().Table("json_data").WhereJsonLength("data->array > ?", 2).Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
				{
					name: "string or float key",
					find: query.DB().Table("json_data").Where("data->string", "first").OrWhere("data->float", 789.123).Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
				{
					name: "contains or key does not exist",
					find: query.DB().Table("json_data").WhereJsonContains("data->array", "ghi").OrWhereJsonDoesntContainKey("data->nested->string").Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
			}

			for _, tt := range tests {
				s.Run(tt.name, func() {
					var items []JsonData
					s.Nil(tt.find(&items))
					tt.assert(items)
				})
			}
		})
	}
}

func (s *DBTestSuite) TestJsonColumnsUpdate() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			data := []JsonData{
				{
					Data: `{"string":"first","int":123,"float":123.456,"bool":true,"array":["abc","def","ghi"],"nested":{"string":"first","int":456},"objects":[{"level":"first","value":"abc"},{"level":"second","value":"def"}]}`,
				},
			}
			result, err := query.DB().Table("json_data").Insert(&data)
			s.Equal(int64(1), result.RowsAffected)
			s.NoError(err)

			tests := []struct {
				name   string
				update map[string]any
				assert func(before JsonData, after JsonData)
			}{
				{
					name:   "update string",
					update: map[string]any{"data->string": "updated_first"},
					assert: func(before JsonData, after JsonData) {
						s.NotContains(before.Data, "updated_first")
						s.Contains(after.Data, "updated_first")
					},
				},
				{
					name:   "update int",
					update: map[string]any{"data->int": 789},
					assert: func(before JsonData, after JsonData) {
						s.NotContains(before.Data, "789")
						s.Contains(after.Data, "789")
					},
				},
				{
					name:   "update float(pointer)",
					update: map[string]any{"data->float": convert.Pointer(456.789)},
					assert: func(before JsonData, after JsonData) {
						s.NotContains(before.Data, "456.789")
						s.Contains(after.Data, "456.789")
					},
				},
				{
					name:   "update array",
					update: map[string]any{"data->array": []string{"uvw", "xyz"}},
					assert: func(before JsonData, after JsonData) {
						s.NotContains(before.Data, "uvw")
						s.Contains(after.Data, "uvw")

						s.NotContains(before.Data, "xyz")
						s.Contains(after.Data, "xyz")
					},
				},
				{
					name: "update multiple keys",
					update: map[string]any{
						"data->bool":              false,
						"data->objects[0]->level": "first_changed",
						"data->nested->string":    "updated_nested_string",
					},
					assert: func(before JsonData, after JsonData) {
						s.NotContains(before.Data, "false")
						s.Contains(after.Data, "false")

						s.NotContains(before.Data, "first_changed")
						s.Contains(after.Data, "first_changed")

						s.NotContains(before.Data, "updated_nested_string")
						s.Contains(after.Data, "updated_nested_string")

					},
				},
			}

			for _, tt := range tests {
				s.Run(tt.name, func() {
					var before, after JsonData
					s.NoError(query.DB().Table("json_data").First(&before))
					res, err := query.DB().Table("json_data").Where("id", before.ID).Update(tt.update)
					s.NoError(err)
					s.Equal(int64(1), res.RowsAffected)
					s.NoError(query.DB().Table("json_data").Where("id", before.ID).First(&after))
					s.NotEqual(before.Data, after.Data)
					tt.assert(before, after)
				})
			}
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

	dbConfig := sqliteTestQuery.Driver().Pool().Writers[0]
	sqliteConnection := dbConfig.Connection
	mockDatabaseConfig(postgresTestQuery.MockConfig(), dbConfig)

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

func TestDbReadWriteSeparate(t *testing.T) {
	dbs := NewTestQueryBuilder().AllWithReadWrite()

	for drive, db := range dbs {
		t.Run(drive, func(t *testing.T) {
			db["read"].CreateTable(TestTableProducts)
			db["write"].CreateTable(TestTableProducts)

			product1 := Product{Name: "read write separate product"}
			result, err := db["mix"].DB().Table("products").Insert(&product1)
			assert.Nil(t, err)
			assert.Equal(t, int64(1), result.RowsAffected)

			var product2 Product
			assert.Nil(t, db["mix"].DB().Table("products").Where("name", product1.Name).First(&product2))
			assert.True(t, product2.ID == 0)

			var product3 Product
			assert.Nil(t, db["read"].DB().Table("products").Where("name", product1.Name).First(&product3))
			assert.True(t, product3.ID == 0)

			var product4 Product
			assert.Nil(t, db["write"].DB().Table("products").Where("name", product1.Name).First(&product4))
			assert.True(t, product4.ID > 0)
		})
	}

	docker, err := dbs[sqlite.Name]["read"].Driver().Docker()
	assert.NoError(t, err)
	assert.NoError(t, docker.Shutdown())

	docker, err = dbs[sqlite.Name]["write"].Driver().Docker()
	assert.NoError(t, err)
	assert.NoError(t, docker.Shutdown())
}

func Benchmark_DB(b *testing.B) {
	query := NewTestQueryBuilder().Postgres("", false)
	query.CreateTable(TestTableAuthors)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := query.DB().Table("authors").Insert(Author{
			BookID: 1,
			Name:   "benchmark",
		})
		if err != nil {
			b.Error(err)
		}

		var authors []Author
		err = query.DB().Table("authors").Limit(50).Find(&authors)
		if err != nil {
			b.Error(err)
		}
	}
}
