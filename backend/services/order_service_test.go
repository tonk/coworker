package services

import (
	"testing"
	"math"

	"github.com/stretchr/testify/assert"
)

func TestMidPosition(t *testing.T) {
	t.Run("both zero returns 1000", func(t *testing.T) {
		assert.Equal(t, 1000.0, MidPosition(0, 0))
	})

	t.Run("midpoint between two values", func(t *testing.T) {
		assert.Equal(t, 1500.0, MidPosition(1000, 2000))
	})

	t.Run("midpoint with negative values", func(t *testing.T) {
		assert.Equal(t, -500.0, MidPosition(-1000, 0))
	})

	t.Run("gap too small returns -1 (rebalance)", func(t *testing.T) {
		assert.Equal(t, -1.0, MidPosition(1000, 1000+minGap/2))
	})

	t.Run("gap at boundary returns midpoint", func(t *testing.T) {
		// At exactly minGap the abs diff is not < minGap
		mid := MidPosition(0, minGap)
		assert.Equal(t, minGap/2, mid)
	})

	t.Run("a == b but non-zero returns -1", func(t *testing.T) {
		assert.Equal(t, -1.0, MidPosition(500, 500))
	})
}

func TestPositionAfter(t *testing.T) {
	t.Run("zero returns 1000", func(t *testing.T) {
		assert.Equal(t, 1000.0, PositionAfter(0))
	})

	t.Run("adds 1000 to non-zero", func(t *testing.T) {
		assert.Equal(t, 2000.0, PositionAfter(1000))
	})

	t.Run("handles negative", func(t *testing.T) {
		assert.Equal(t, 900.0, PositionAfter(-100))
	})

	t.Run("handles large floats", func(t *testing.T) {
		assert.Equal(t, 1e12+1000, PositionAfter(1e12))
	})
}

func TestMidPositionEdgeCases(t *testing.T) {
	t.Run("very small positive numbers below minGap", func(t *testing.T) {
		// Gap of 1e-12 is below minGap (1e-9), so returns -1 for rebalance
		assert.Equal(t, -1.0, MidPosition(1e-12, 2e-12))
	})

	t.Run("minGap boundary at large values", func(t *testing.T) {
		large := 1e15
		assert.Equal(t, -1.0, MidPosition(large, large+minGap/100))
	})

	t.Run("infinity", func(t *testing.T) {
		// Inf - Inf is NaN, NaN < minGap is false, so it returns Inf
		assert.True(t, math.IsInf(MidPosition(math.Inf(1), math.Inf(1)), 1))
	})
}
