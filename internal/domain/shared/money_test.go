package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ldmejia/PruebaPivote/internal/domain/shared"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"100", 10000, false},
		{"100.5", 10050, false},
		{"100.50", 10050, false},
		{"0.01", 1, false},
		{" 25.00 ", 2500, false},
		{"", 0, true},
		{"-10", 0, true},
		{"10.123", 0, true},
		{"10.", 0, true},
		{".50", 0, true},
		{"1e3", 0, true},
		{"abc", 0, true},
	}
	for _, tt := range tests {
		got, err := shared.ParseAmount(tt.in)
		if tt.wantErr {
			assert.ErrorIs(t, err, shared.ErrInvalidAmount, tt.in)
			continue
		}
		assert.NoError(t, err, tt.in)
		assert.Equal(t, tt.want, got, tt.in)
	}
}

func TestFormatAmount(t *testing.T) {
	assert.Equal(t, "0.00", shared.FormatAmount(0))
	assert.Equal(t, "0.01", shared.FormatAmount(1))
	assert.Equal(t, "100.50", shared.FormatAmount(10050))
	assert.Equal(t, "-2.50", shared.FormatAmount(-250))
}
