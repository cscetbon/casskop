package storageupsize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeDataCapacityChange(t *testing.T) {
	tests := []struct {
		name         string
		oldCapacity  string
		newCapacity  string
		expectedDiff CapacityChange
	}{
		{name: "valid increase", oldCapacity: "10Gi", newCapacity: "20Gi", expectedDiff: CapacityUpsize},
		{name: "no change", oldCapacity: "10Gi", newCapacity: "10Gi", expectedDiff: CapacityNoChange},
		{name: "syntactic change", oldCapacity: "1Gi", newCapacity: "1024Mi", expectedDiff: CapacitySyntacticChange},
		{name: "syntactic change", oldCapacity: "1T", newCapacity: "1000G", expectedDiff: CapacitySyntacticChange},
		{name: "capacity decrease", oldCapacity: "20Gi", newCapacity: "10Gi", expectedDiff: CapacityDownsize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := AnalyzeDataCapacityChange(tt.oldCapacity, tt.newCapacity)
			assert.Equal(t, tt.expectedDiff, diff)
		})
	}
}
