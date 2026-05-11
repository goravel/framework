package image

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/foundation"
	mocksai "github.com/goravel/framework/mocks/ai"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

func TestOf(t *testing.T) {
	mockApp := mocksfoundation.NewApplication(t)
	mockAI := mocksai.NewAI(t)
	mockRequest := mocksai.NewImageRequest(t)
	previousApp := foundation.App
	foundation.App = mockApp
	t.Cleanup(func() {
		foundation.App = previousApp
	})

	mockApp.EXPECT().MakeAI().Return(mockAI).Once()
	mockAI.EXPECT().Image("draw a cat").Return(mockRequest).Once()

	request := Of("draw a cat")

	assert.Same(t, mockRequest, request)
	assert.Implements(t, (*contractsai.ImageRequest)(nil), request)
}
