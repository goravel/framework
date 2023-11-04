package gorm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TestEventModel struct {
	Name     string
	Avatar   string
	IsAdmin  bool
	IsManage int `gorm:"column:manage"`
	AdminAt  time.Time
	ManageAt time.Time
	high     int
}

var testNow = time.Now().Add(-1 * time.Second)
var testEventModel = TestEventModel{Name: "name", Avatar: "avatar", IsAdmin: true, IsManage: 0, AdminAt: testNow, ManageAt: testNow, high: 1}
var testQuery = NewQueryImplByInstance(&gorm.DB{
	Statement: &gorm.Statement{
		Selects: []string{},
		Omits:   []string{},
	},
}, &QueryImpl{
	config:        nil,
	withoutEvents: false,
})

type EventTestSuite struct {
	suite.Suite
	events []*Event
}

func TestEventTestSuite(t *testing.T) {
	suite.Run(t, new(EventTestSuite))
}

func (s *EventTestSuite) SetupTest() {
	s.events = []*Event{
		NewEvent(testQuery, &testEventModel, map[string]any{"avatar": "avatar1", "is_admin": false, "is_manage": 1, "admin_at": time.Now(), "manage": testNow}),
		NewEvent(testQuery, &testEventModel, map[string]any{"Avatar": "avatar1", "IsAdmin": false, "IsManage": 1, "AdminAt": time.Now(), "ManageAt": testNow}),
		NewEvent(testQuery, &testEventModel, TestEventModel{Avatar: "avatar1", IsAdmin: false, IsManage: 1, AdminAt: time.Now(), ManageAt: testNow}),
		NewEvent(testQuery, &testEventModel, &TestEventModel{Avatar: "avatar1", IsAdmin: false, IsManage: 1, AdminAt: time.Now(), ManageAt: testNow}),
	}
}

func (s *EventTestSuite) TestSetAttribute() {
	dest := map[string]any{"avatar": "avatar1"}
	query := NewQueryImplByInstance(&gorm.DB{
		Statement: &gorm.Statement{
			Selects: []string{},
			Omits:   []string{},
			Dest:    dest,
		},
	}, &QueryImpl{
		config:        nil,
		withoutEvents: false,
	})

	event := NewEvent(query, &testEventModel, dest)

	event.SetAttribute("Name", "name1")
	name := event.GetAttribute("Name")
	s.Equal(name, "name1")
	name = event.GetAttribute("name")
	s.Equal(name, "name1")

	event.SetAttribute("Avatar", "avatar2")
	avatar := event.GetAttribute("Avatar")
	s.Equal("avatar2", avatar)
	avatar = event.GetAttribute("avatar")
	s.Equal("avatar2", avatar)
}

func (s *EventTestSuite) TestGetAttribute() {
	events := []*Event{
		NewEvent(testQuery, &testEventModel, map[string]any{"avatar": "avatar1"}),
		NewEvent(testQuery, &testEventModel, TestEventModel{Avatar: "avatar1"}),
	}

	for _, event := range events {
		s.Equal("name", event.GetAttribute("Name"))
		s.Equal("name", event.GetAttribute("name"))
		s.Equal("avatar1", event.GetAttribute("Avatar"))
		s.Equal("avatar1", event.GetAttribute("avatar"))
	}
}

func (s *EventTestSuite) TestGetOriginal() {
	event := NewEvent(testQuery, &testEventModel, map[string]any{"avatar": "avatar1"})

	s.Equal("name", event.GetOriginal("name"))
	s.Equal("avatar", event.GetOriginal("avatar"))
	s.Equal(true, event.GetOriginal("is_admin"))
	s.Equal(0, event.GetOriginal("manage"))
	s.Equal(testNow, event.GetOriginal("admin_at"))
	s.Equal(testNow, event.GetOriginal("manage_at"))
	s.Nil(event.GetOriginal("no"))
}

func (s *EventTestSuite) TestIsDirty() {
	for _, event := range s.events {
		s.True(event.IsDirty())
		s.False(event.IsDirty("Name"))
		s.False(event.IsDirty("name"))
		s.True(event.IsDirty("Avatar"))
		s.True(event.IsDirty("avatar"))
		s.False(event.IsDirty("IsAdmin"))
		s.False(event.IsDirty("is_admin"))
		s.True(event.IsDirty("IsManage"))
		s.True(event.IsDirty("manage"))
		s.False(event.IsDirty("is_manage"))
		s.True(event.IsDirty("AdminAt"))
		s.True(event.IsDirty("admin_at"))
		s.False(event.IsDirty("ManageAt"))
		s.False(event.IsDirty("manage_at"))
		s.True(event.IsDirty("name", "avatar"))
		s.True(event.IsDirty("is_manage", "avatar"))
	}

	// Set zero value when use model update
	event := NewEvent(testQuery, &testEventModel, &TestEventModel{Avatar: "avatar1", IsAdmin: true, IsManage: 0, AdminAt: time.Now(), ManageAt: testNow})
	s.True(event.IsDirty())
	s.False(event.IsDirty("Name"))
	s.False(event.IsDirty("name"))
	s.True(event.IsDirty("Avatar"))
	s.True(event.IsDirty("avatar"))
	s.False(event.IsDirty("IsAdmin"))
	s.False(event.IsDirty("is_admin"))
	s.False(event.IsDirty("IsManage"))
	s.False(event.IsDirty("manage"))
	s.False(event.IsDirty("is_manage"))
	s.True(event.IsDirty("AdminAt"))
	s.True(event.IsDirty("admin_at"))
	s.False(event.IsDirty("ManageAt"))
	s.False(event.IsDirty("manage_at"))
	s.True(event.IsDirty("name", "avatar"))
	s.True(event.IsDirty("is_manage", "avatar"))
}

func (s *EventTestSuite) TestValidColumn() {
	for _, event := range s.events {
		s.True(event.validColumn("name"))
		s.True(event.validColumn("Name"))
		s.True(event.validColumn("IsAdmin"))
		s.True(event.validColumn("is_admin"))
		s.True(event.validColumn("IsManage"))
		s.False(event.validColumn("is_manage"))
		s.True(event.validColumn("manage"))
		s.False(event.validColumn("age"))

		event.query = NewQueryImplByInstance(&gorm.DB{
			Statement: &gorm.Statement{
				Selects: []string{"name"},
				Omits:   []string{},
			},
		}, &QueryImpl{
			config:        nil,
			withoutEvents: false,
		})
		s.True(event.validColumn("Name"))
		s.True(event.validColumn("name"))
		s.False(event.validColumn("avatar"))
		s.False(event.validColumn("Avatar"))

		event.query = NewQueryImplByInstance(&gorm.DB{
			Statement: &gorm.Statement{
				Selects: []string{},
				Omits:   []string{"name"},
			},
		}, &QueryImpl{
			config:        nil,
			withoutEvents: false,
		})
		s.False(event.validColumn("Name"))
		s.False(event.validColumn("name"))
		s.True(event.validColumn("avatar"))
		s.True(event.validColumn("Avatar"))
	}
}

func (s *EventTestSuite) TestDirty() {
	for _, event := range s.events {
		s.False(event.dirty("Name", "name"))
		s.True(event.dirty("Name", "name1"))
		s.False(event.dirty("name", "name"))
		s.True(event.dirty("name", "name1"))
		s.False(event.dirty("IsAdmin", true))
		s.True(event.dirty("IsAdmin", false))
		s.False(event.dirty("is_admin", true))
		s.True(event.dirty("is_admin", false))
		s.False(event.dirty("IsManage", 0))
		s.True(event.dirty("IsManage", 1))
		s.False(event.dirty("manage", 0))
		s.True(event.dirty("is_manage", 0))
		s.True(event.dirty("is_manage", 1))
		s.True(event.dirty("manage", 1))
		s.False(event.dirty("AdminAt", testNow))
		s.True(event.dirty("AdminAt", time.Now()))
		s.False(event.dirty("admin_at", testNow))
		s.True(event.dirty("admin_at", time.Now()))
	}
}

func (s *EventTestSuite) TestCompareColumnName() {
	for _, event := range s.events {
		s.False(event.equalColumnName("address", "name"))
		s.False(event.equalColumnName("address", "address"))
		s.True(event.equalColumnName("Name", "name"))
		s.True(event.equalColumnName("is_admin", "IsAdmin"))
	}
}

func (s *EventTestSuite) TestToDBColumnName() {
	for _, event := range s.events {
		s.Equal("", event.toDBColumnName("address"))
		s.Equal("name", event.toDBColumnName("Name"))
		s.Equal("name", event.toDBColumnName("name"))
		s.Equal("is_admin", event.toDBColumnName("IsAdmin"))
		s.Equal("is_admin", event.toDBColumnName("is_admin"))
	}
}

func (s *EventTestSuite) TestColumnNamesWithDbColumnNames() {
	for _, event := range s.events {
		s.Equal(map[string]string{
			"Name":      "name",
			"name":      "name",
			"Avatar":    "avatar",
			"avatar":    "avatar",
			"IsAdmin":   "is_admin",
			"is_admin":  "is_admin",
			"IsManage":  "manage",
			"manage":    "manage",
			"AdminAt":   "admin_at",
			"admin_at":  "admin_at",
			"ManageAt":  "manage_at",
			"manage_at": "manage_at",
		}, event.ColumnNamesWithDbColumnNames())
	}
}

func TestStructToMap(t *testing.T) {
	assert.Equal(t, map[string]any{
		"name":      testEventModel.Name,
		"avatar":    testEventModel.Avatar,
		"is_admin":  testEventModel.IsAdmin,
		"manage":    testEventModel.IsManage,
		"admin_at":  testEventModel.AdminAt,
		"manage_at": testEventModel.ManageAt,
	}, structToMap(testEventModel))
}

func TestStructNameToDbColumnName(t *testing.T) {
	assert.Equal(t, "is_admin", structNameToDbColumnName("IsAdmin", ""))
	assert.Equal(t, "admin", structNameToDbColumnName("IsAdmin", "column:admin"))
	assert.Equal(t, "admin", structNameToDbColumnName("IsAdmin", "column: admin"))
}
