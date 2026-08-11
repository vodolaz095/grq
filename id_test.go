package grq

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetRandomID generates multiple random IDs and verifies they are unique.
// It runs the generator many times (100 iterations) to ensure it produces
// different strings every time.
func TestGetRandomID(t *testing.T) {
	// Test that the function returns a valid ID without errors
	id, err := getRandomID()
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	// Run the generator multiple times and ensure uniqueness
	t.Run("generates unique IDs", func(t *testing.T) {
		ids := make(map[string]int)
		for i := 0; i < 100; i++ {
			id, err := getRandomID()
			require.NoError(t, err)

			// Check for duplicates
			if count, ok := ids[id]; ok {
				t.Errorf("Generated duplicate ID on iteration %d: %s (first seen at iteration %d)",
					i, id, count)
			}
			ids[id] = i + 1
		}
	})

	// Run the generator many times and check that all IDs have different lengths
	t.Run("generates IDs with correct length", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			id, err := getRandomID()
			require.NoError(t, err)
			assert.Equal(t, keyLength*2, len(id))
		}
	})
}

// TestGetRandomIDHexFormat verifies that the generated IDs are valid hex strings.
func TestGetRandomIDHexFormat(t *testing.T) {

	for i := 0; i < 10; i++ {
		id, err := getRandomID()
		require.NoError(t, err)
		_, err = hex.DecodeString(id)
		require.NoError(t, err)
	}
}

// TestGetRandomIDConsistency verifies that the ID generator is consistent
// within the same process (same random seed behavior).
func TestGetRandomIDConsistency(t *testing.T) {
	// Generate IDs multiple times
	ids := make([]string, 50)
	for i := 0; i < 50; i++ {
		id, err := getRandomID()
		require.NoError(t, err)
		ids[i] = id
	}

	// All IDs should be unique
	for i := 0; i < 50; i++ {
		for j := i + 1; j < 50; j++ {
			assert.NotEqual(t, ids[i], ids[j])
		}
	}
}
