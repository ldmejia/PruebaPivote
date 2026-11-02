package shared_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ldmejia/PruebaPivote/internal/domain/shared"
)

func TestLocalDay(t *testing.T) {
	tests := map[string]string{
		"2026-11-01T12:00:00Z": "2026-11-01",
		"2026-11-02T05:59:59Z": "2026-11-01",
		"2026-11-02T06:00:00Z": "2026-11-02",
	}
	for utc, want := range tests {
		in, _ := time.Parse(time.RFC3339, utc)
		assert.Equal(t, want, shared.LocalDay(in).Format(time.DateOnly), utc)
	}
}
