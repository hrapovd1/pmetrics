package handlers

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_gzipWriter_Write(t *testing.T) {
	type fields struct {
		ResponseWriter http.ResponseWriter
		Writer         io.Writer
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "Positive",
			fields:  fields{Writer: &bytes.Buffer{}},
			args:    args{b: []byte(`Test string`)},
			want:    11,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := gzipWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				Writer:         tt.fields.Writer,
			}
			got, err := w.Write(tt.args.b)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
