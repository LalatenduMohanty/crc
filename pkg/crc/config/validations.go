package config

import (
	"fmt"
	"strconv"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/validation"
)

// ValidateBool is a fail safe in the case user
// makes a typo for boolean config values
func ValidateBool(value interface{}) (bool, string) {
	if value.(string) == "true" || value.(string) == "false" {
		return true, ""
	}
	return false, "true/false"
}

// ValidateDriver is check if driver is valid in the config
func ValidateDriver(value interface{}) (bool, string) {
	if err := validation.ValidateDriver(value.(string)); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidateCPUs is check if provided cpus count is valid in the config
func ValidateCPUs(value interface{}) (bool, string) {
	v, err := strconv.Atoi(value.(string))
	if err != nil {
		return false, fmt.Sprintf("Required integer value >=%d", constants.DefaultCPUs)
	}
	if err := validation.ValidateCPUs(v); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidateMemory is check if provided memory is valid in the config
func ValidateMemory(value interface{}) (bool, string) {
	v, err := strconv.Atoi(value.(string))
	if err != nil {
		return false, fmt.Sprintf("Required integer value in MB >=%d", constants.DefaultMemory)
	}
	if err := validation.ValidateMemory(v); err != nil {
		return false, err.Error()
	}
	return true, ""
}

// ValidateBundle is check if provided bundle path is valid
func ValidateBundle(value interface{}) (bool, string) {
	if err := validation.ValidateBundle(value.(string)); err != nil {
		return false, err.Error()
	}
	return true, ""
}
