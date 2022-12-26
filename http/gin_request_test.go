package http

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GinRequestSuite struct {
	suite.Suite
}

func TestGinRequestSuite(t *testing.T) {
	suite.Run(t, new(GinRequestSuite))
}

func (s *GinRequestSuite) SetupTest() {

}

func (s *GinRequestSuite) TestInput() {
	//r := gin.Default()
	//r.GET("/input", func(c *gin.Context) {
	//	s.True(1 == 2)
	//})
	//
	//go func() {
	//	s.Nil(r.Run(":3000"))
	//	select {}
	//}()
	//
	//w := httptest.NewRecorder()
	//req, _ := http.NewRequest("GET", "/input", nil)
	//r.ServeHTTP(w, req)
}
