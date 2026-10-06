package cep

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid cep without dash",
			input:     "12345678",
			wantValue: "12345678",
			wantErr:   false,
		},
		{
			name:      "valid cep with dash",
			input:     "12345-678",
			wantValue: "12345-678",
			wantErr:   false,
		},
		{
			name:    "invalid cep with letters",
			input:   "1234567a",
			wantErr: true,
		},
		{
			name:    "invalid cep too short",
			input:   "1234567",
			wantErr: true,
		},
		{
			name:    "invalid cep too long",
			input:   "123456789",
			wantErr: true,
		},
		{
			name:    "invalid cep empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid cep with spaces",
			input:   "12345 678",
			wantErr: true,
		},
		{
			name:    "invalid cep wrong dash position",
			input:   "1234-5678",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantValue, got.Value)
		})
	}
}
