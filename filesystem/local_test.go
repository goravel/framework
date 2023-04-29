package filesystem

import (
	"testing"

	"github.com/h2non/filetype/matchers"
	"github.com/stretchr/testify/assert"
)

func TestLocal_LastModified(t *testing.T) {
	type fields struct {
		root string
	}
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				root: ".",
			},
			args: args{
				file: "./testdata/test.png",
			},
			want: true,
		},
		{
			name: "invalid file",
			fields: fields{
				root: ".",
			},
			args: args{
				file: "./testdata/invalid.png",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Local{
				root: tt.fields.root,
			}
			got, err := r.LastModified(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Local.LastModified() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want {
				assert.NotZero(t, got)
			}
		})
	}
}

func TestLocal_MimeType(t *testing.T) {
	type fields struct {
		root string
	}
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				root: ".",
			},
			args: args{
				file: "./testdata/test.png",
			},
			want: matchers.TypePng.MIME.Value,
		},
		{
			name: "invalid file",
			fields: fields{
				root: ".",
			},
			args: args{
				file: "./testdata/invalid.png",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Local{
				root: tt.fields.root,
			}
			got, err := r.MimeType(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Local.MimeType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
