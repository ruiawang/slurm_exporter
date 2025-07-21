package collector

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseReservations(t *testing.T) {
	data, err := os.ReadFile("../test_data/sreservations.txt")
	assert.NoError(t, err)

	reservations, err := parseReservations(data)
	assert.NoError(t, err)
	assert.Len(t, reservations, 1)

	// Test the production-like reservation
	res1 := reservations[0]
	assert.Equal(t, "pre-reservation-maintenance", res1.Name)
	assert.Equal(t, "INACTIVE", res1.State)
	assert.Equal(t, "user01", res1.Users)
	assert.Equal(t, "node[001-102]", res1.Nodes)
	assert.Equal(t, "", res1.Partition) // Check for (null) parsing
	assert.Equal(t, "SPEC_NODES,ALL_NODES", res1.Flags)
	assert.Equal(t, 102.0, res1.NodeCount)
	assert.Equal(t, 25152.0, res1.CoreCount)
	expectedStartTime, _ := time.Parse(slurmTimeLayout, "2025-08-26T07:00:00")
	assert.Equal(t, expectedStartTime, res1.StartTime)
	expectedEndTime, _ := time.Parse(slurmTimeLayout, "2025-08-29T20:00:00")
	assert.Equal(t, expectedEndTime, res1.EndTime)
}
