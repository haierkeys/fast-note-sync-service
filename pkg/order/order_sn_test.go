package order

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrderGenerate(t *testing.T) {
	// Generate an order
	now := time.Now()
	order1 := Generate(now)

	// Length should be fixed at 24 chars: 14 (time) + 3 (ms) + 3 (pid) + 4 (seq)
	assert.Len(t, order1, 24)

	// Since we use global atomic numbers, generating twice immediately should yield different results
	order2 := Generate(now)
	assert.Len(t, order2, 24)
	assert.NotEqual(t, order1, order2, "Orders generated consecutively should have different sequence numbers")

	// Sup logic test (implicitly tested through structure length, but we can also test sup directly if exported, but it's unexported)
}
