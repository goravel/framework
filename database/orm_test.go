package database

//todo

//func (s *GormQueryTestSuite) TestTransactionSuccess() {
//	for _, db := range s.dbs {
//		user := User{Name: "transaction_success_user", Avatar: "transaction_success_avatar"}
//		user1 := User{Name: "transaction_success_user1", Avatar: "transaction_success_avatar1"}
//		s.Nil(facades.Orm.Transaction(func(tx ormcontract.Transaction) error {
//			s.Nil(tx.Create(&user))
//			s.Nil(tx.Create(&user1))
//
//			return nil
//		}))
//
//		var user2, user3 User
//		s.Nil(db.Find(&user2, user.ID))
//		s.Nil(db.Find(&user3, user1.ID))
//	}
//}
//
//func (s *GormQueryTestSuite) TestTransactionError() {
//	for _, db := range s.dbs {
//		s.NotNil(db.Transaction(func(tx ormcontract.Transaction) error {
//			user := User{Name: "transaction_error_user", Avatar: "transaction_error_avatar"}
//			s.Nil(tx.Create(&user))
//
//			user1 := User{Name: "transaction_error_user1", Avatar: "transaction_error_avatar1"}
//			s.Nil(tx.Create(&user1))
//
//			return errors.New("error")
//		}))
//
//		var users []User
//		s.Nil(db.Find(&users))
//		s.Equal(0, len(users))
//	}
//}
