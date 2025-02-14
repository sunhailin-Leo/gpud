package nvml

import (
	"testing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/stretchr/testify/assert"
)

func TestIsNotSupportError(t *testing.T) {
	tests := []struct {
		name     string
		ret      nvml.Return
		expected bool
	}{
		{
			name:     "Direct ERROR_NOT_SUPPORTED match",
			ret:      nvml.ERROR_NOT_SUPPORTED,
			expected: true,
		},
		{
			name:     "Success is not a not-supported error",
			ret:      nvml.SUCCESS,
			expected: false,
		},
		{
			name:     "Unknown error is not a not-supported error",
			ret:      nvml.ERROR_UNKNOWN,
			expected: false,
		},
		{
			name:     "Invalid argument is not a not-supported error",
			ret:      nvml.ERROR_INVALID_ARGUMENT,
			expected: false,
		},
		{
			name:     "Uninitialized is not a not-supported error",
			ret:      nvml.ERROR_UNINITIALIZED,
			expected: false,
		},
		{
			name:     "No permission is not a not-supported error",
			ret:      nvml.ERROR_NO_PERMISSION,
			expected: false,
		},
		{
			name:     "Already initialized is not a not-supported error",
			ret:      nvml.ERROR_ALREADY_INITIALIZED,
			expected: false,
		},
		{
			name:     "Insufficient size is not a not-supported error",
			ret:      nvml.ERROR_INSUFFICIENT_SIZE,
			expected: false,
		},
		{
			name:     "Driver not loaded is not a not-supported error",
			ret:      nvml.ERROR_DRIVER_NOT_LOADED,
			expected: false,
		},
		{
			name:     "Timeout is not a not-supported error",
			ret:      nvml.ERROR_TIMEOUT,
			expected: false,
		},
		{
			name:     "IRQ issue is not a not-supported error",
			ret:      nvml.ERROR_IRQ_ISSUE,
			expected: false,
		},
		{
			name:     "Library not found is not a not-supported error",
			ret:      nvml.ERROR_LIBRARY_NOT_FOUND,
			expected: false,
		},
		{
			name:     "Function not found is not a not-supported error",
			ret:      nvml.ERROR_FUNCTION_NOT_FOUND,
			expected: false,
		},
		{
			name:     "Corrupted inforom is not a not-supported error",
			ret:      nvml.ERROR_CORRUPTED_INFOROM,
			expected: false,
		},
		{
			name:     "GPU is lost is not a not-supported error",
			ret:      nvml.ERROR_GPU_IS_LOST,
			expected: false,
		},
		{
			name:     "Reset required is not a not-supported error",
			ret:      nvml.ERROR_RESET_REQUIRED,
			expected: false,
		},
		{
			name:     "Operating system call failed is not a not-supported error",
			ret:      nvml.ERROR_OPERATING_SYSTEM,
			expected: false,
		},
		{
			name:     "Memory is not a not-supported error",
			ret:      nvml.ERROR_MEMORY,
			expected: false,
		},
		{
			name:     "No data is not a not-supported error",
			ret:      nvml.ERROR_NO_DATA,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotSupportError(tt.ret)
			assert.Equal(t, tt.expected, result, "IsNotSupportError(%v) = %v, want %v", tt.ret, result, tt.expected)
		})
	}
}

// TestIsNotSupportErrorStringMatch tests the string-based matching of not supported errors
func TestIsNotSupportErrorStringMatch(t *testing.T) {
	// Create a custom Return type that will produce different error strings
	tests := []struct {
		name     string
		ret      nvml.Return
		expected bool
	}{
		{
			name:     "String contains 'not supported' (lowercase)",
			ret:      nvml.Return(1000), // This will produce "Unknown Error" which we'll handle in ErrorString
			expected: true,
		},
		{
			name:     "String contains 'NOT SUPPORTED' (uppercase)",
			ret:      nvml.Return(1001),
			expected: true,
		},
		{
			name:     "String contains 'Not Supported' (mixed case)",
			ret:      nvml.Return(1002),
			expected: true,
		},
		{
			name:     "String contains 'not supported' with leading/trailing spaces",
			ret:      nvml.Return(1003),
			expected: true,
		},
		{
			name:     "String contains 'not supported' within a longer message",
			ret:      nvml.Return(1004),
			expected: true,
		},
		{
			name:     "String does not contain 'not supported'",
			ret:      nvml.Return(1005),
			expected: false,
		},
		{
			name:     "Empty string",
			ret:      nvml.Return(1006),
			expected: false,
		},
		{
			name:     "String with similar but not exact match",
			ret:      nvml.Return(1007),
			expected: false,
		},
	}

	// Override nvml.ErrorString for testing
	originalErrorString := nvml.ErrorString
	defer func() {
		nvml.ErrorString = originalErrorString
	}()

	nvml.ErrorString = func(ret nvml.Return) string {
		switch ret {
		case 1000:
			return "operation is not supported on this device"
		case 1001:
			return "THIS OPERATION IS NOT SUPPORTED"
		case 1002:
			return "Feature Not Supported"
		case 1003:
			return "  not supported  "
		case 1004:
			return "The requested operation is not supported on device 0"
		case 1005:
			return "Some other error"
		case 1006:
			return ""
		case 1007:
			return "notsupported" // No space between 'not' and 'supported'
		default:
			return originalErrorString(ret)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotSupportError(tt.ret)
			assert.Equal(t, tt.expected, result, "IsNotSupportError(%v) = %v, want %v", tt.ret, result, tt.expected)
		})
	}
}
