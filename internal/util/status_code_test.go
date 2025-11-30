package util_test

import (
	"testing"

	"github.com/DarknessKiller/pingopher/internal/util"
)

func TestCheckStatusCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		statusCode := uint16(200)
		acceptedCodes := []string{"200", "200-299"}
		expected := true
		result, err := util.CheckStatusCode(statusCode, acceptedCodes)
		if result != expected {
			t.Errorf("CheckStatusCode(%d, %v) = %v, expected %v", statusCode, acceptedCodes, result, expected)
		}
		if err != nil {
			t.Errorf("CheckStatusCode(%d, %v) returned an error: %v", statusCode, acceptedCodes, err)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		statusCode := uint16(500)
		acceptedCodes := []string{"200", "200-299"}
		expected := false
		result, err := util.CheckStatusCode(statusCode, acceptedCodes)
		if result != expected {
			t.Errorf("CheckStatusCode(%d, %v) = %v, expected %v", statusCode, acceptedCodes, result, expected)
		}
		if err != nil {
			t.Errorf("CheckStatusCode(%d, %v) returned an error: %v", statusCode, acceptedCodes, err)
		}
	})

	t.Run("Invalid Code", func(t *testing.T) {
		statusCode := uint16(200)
		acceptedCodes := []string{"abc", "200-299"}
		_, err := util.CheckStatusCode(statusCode, acceptedCodes)
		if err == nil {
			t.Errorf("CheckStatusCode(%d, %v) expected to return an error for invalid code", statusCode, acceptedCodes)
		}
	})

	t.Run("No Accepted Codes", func(t *testing.T) {
		statusCode := uint16(200)
		acceptedCodes := []string{}
		_, err := util.CheckStatusCode(statusCode, acceptedCodes)
		if err == nil {
			t.Errorf("CheckStatusCode(%d, %v) expected to return an error for no accepted codes", statusCode, acceptedCodes)
		}
	})

	t.Run("Invalid Range", func(t *testing.T) {
		statusCode := uint16(200)
		acceptedCodes := []string{"200-299-400"}
		_, err := util.CheckStatusCode(statusCode, acceptedCodes)
		if err == nil {
			t.Errorf("CheckStatusCode(%d, %v) expected to return an error for invalid range", statusCode, acceptedCodes)
		}
	})

	t.Run("Invalid Range", func(t *testing.T) {
		statusCode := uint16(200)
		acceptedCodes := []string{"200-199"}
		result, err := util.CheckStatusCode(statusCode, acceptedCodes)
		if result != false {
			t.Errorf("CheckStatusCode(%d, %v) = %v, expected %v", statusCode, acceptedCodes, result, false)
		}
		if err == nil {
			t.Errorf("CheckStatusCode(%d, %v) expected to return an error for invalid range", statusCode, acceptedCodes)
		}
	})
}
