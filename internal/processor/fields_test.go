package processor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFields(t *testing.T) {
	t.Run("No fields provided, include all", func(t *testing.T) {
		fields, err := NewFields("")
		require.NoError(t, err)

		for _, field := range fieldNames {
			assert.True(t, fields.IncludeField(getFieldIndex(field)))
		}
	})

	t.Run("Valid fields provided", func(t *testing.T) {
		fields, err := NewFields("type,time,elb")
		require.NoError(t, err)

		assert.True(t, fields.IncludeField(getFieldIndex("type")))
		assert.True(t, fields.IncludeField(getFieldIndex("time")))
		assert.True(t, fields.IncludeField(getFieldIndex("elb")))
		assert.False(t, fields.IncludeField(getFieldIndex("client:port")))
	})

	t.Run("Invalid field provided", func(t *testing.T) {
		_, err := NewFields("invalid_field")
		require.Error(t, err)
		assert.Equal(t, "invalid field name 'invalid_field' provided", err.Error())
	})
}

func TestGetFieldNameByIndex(t *testing.T) {
	fields, err := NewFields("")
	require.NoError(t, err)

	t.Run("Valid index", func(t *testing.T) {
		fieldName, err := fields.GetFieldNameByIndex(0)
		require.NoError(t, err)
		assert.Equal(t, "type", fieldName)

		fieldName, err = fields.GetFieldNameByIndex(1)
		require.NoError(t, err)
		assert.Equal(t, "time", fieldName)
	})

	t.Run("Invalid index", func(t *testing.T) {
		_, err := fields.GetFieldNameByIndex(-1)
		require.Error(t, err)
		assert.Equal(t, "invalid field index -1", err.Error())

		_, err = fields.GetFieldNameByIndex(len(fieldNames))
		require.Error(t, err)
		assert.Equal(t, fmt.Sprintf("invalid field index %d", len(fieldNames)), err.Error())
	})
}

func TestIncludeField(t *testing.T) {
	t.Run("Include all fields", func(t *testing.T) {
		fields, err := NewFields("")
		require.NoError(t, err)

		for i := range fieldNames {
			assert.True(t, fields.IncludeField(i))
		}
	})

	t.Run("Include specific fields", func(t *testing.T) {
		fields, err := NewFields("type,time")
		require.NoError(t, err)

		assert.True(t, fields.IncludeField(getFieldIndex("type")))
		assert.True(t, fields.IncludeField(getFieldIndex("time")))
		assert.False(t, fields.IncludeField(getFieldIndex("elb")))
	})

	t.Run("Invalid index", func(t *testing.T) {
		fields, err := NewFields("")
		require.NoError(t, err)

		assert.False(t, fields.IncludeField(-1))
		assert.False(t, fields.IncludeField(len(fieldNames)))
	})
}

func getFieldIndex(fieldName string) int {
	for i, name := range fieldNames {
		if name == fieldName {
			return i
		}
	}
	return -1
}
