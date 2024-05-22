package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestGetName(t *testing.T) {
	var (
		ctx     = &consolemocks.Context{}
		ttype   = "rule"
		name    = "Lowercase"
		getPath = func(name string) string {
			return "app/rules/" + name + ".go"
		}
	)

	tests := []struct {
		name    string
		setup   func()
		want    string
		wantErr error
	}{
		{
			name: "should return error when name is empty",
			setup: func() {
				name = ""
				ctx.On("Ask", "Enter the rule name", mock.Anything).Return("", errors.New("the rule name cannot be empty")).Once()
			},
			want:    "",
			wantErr: errors.New("the rule name cannot be empty"),
		},
		{
			name: "should return error when name already exists",
			setup: func() {
				name = "Uppercase"
				assert.Nil(t, file.Create(getPath(name), ""))
				ctx.On("OptionBool", "force").Return(false).Once()
			},
			want:    "",
			wantErr: errors.New("the rule already exists. Use the --force or -f flag to overwrite"),
		},
		{
			name: "should return name when name is not empty",
			setup: func() {
				name = "Lowercase"
				ctx.On("OptionBool", "force").Return(false).Once()
			},
			want:    name,
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			got, err := GetName(ctx, ttype, name, getPath)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)

			assert.Nil(t, file.Remove(getPath(name)))
			ctx.AssertExpectations(t)
		})
	}
}
