package validator

import (
	"reflect"
	"testing"
	"time"

	"github.com/haierkeys/fast-note-sync-service/pkg/timex"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name string `binding:"required"`
	Age  int    `binding:"min=18"`
}

func TestCustomValidator_ValidateStruct(t *testing.T) {
	validator := NewCustomValidator()

	// Test valid struct
	validObj := TestStruct{Name: "Alice", Age: 20}
	err := validator.ValidateStruct(&validObj)
	assert.NoError(t, err)

	// Test invalid struct (missing required)
	invalidObj1 := TestStruct{Age: 20}
	err = validator.ValidateStruct(&invalidObj1)
	assert.Error(t, err)

	// Test invalid struct (min condition)
	invalidObj2 := TestStruct{Name: "Bob", Age: 15}
	err = validator.ValidateStruct(&invalidObj2)
	assert.Error(t, err)

	// Non-struct should return nil because kindOfData check ignores non-structs
	notStruct := "just a string"
	err = validator.ValidateStruct(notStruct)
	assert.NoError(t, err)
}

func TestCustomValidator_Engine(t *testing.T) {
	validator := NewCustomValidator()
	engine := validator.Engine()
	assert.NotNil(t, engine)

	// Should be the same engine upon second call due to sync.Once
	engine2 := validator.Engine()
	assert.Equal(t, engine, engine2)
}

func TestValidateJSONDateType(t *testing.T) {
	// Setup standard time
	standardTime := time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)
	txTime := timex.Time(standardTime)

	val := reflect.ValueOf(txTime)
	res := ValidateJSONDateType(val)
	assert.NotNil(t, res)
	assert.IsType(t, "", res)
	assert.Equal(t, "2023-10-01 12:00:00", res)

	// Setup zero time
	zeroTime := time.Time{}
	txZeroTime := timex.Time(zeroTime)
	valZero := reflect.ValueOf(txZeroTime)
	resZero := ValidateJSONDateType(valZero)
	assert.Nil(t, resZero, "Zero time should return nil to fail 'required' binding")
}
