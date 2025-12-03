//nolint:funlen,gocognit,gocyclo,unused,cyclop,maintidx,revive // This is a comprehensive test file that covers many scenarios may break some linter rules. to manage better test coverage.
package envload

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type (
	Test[T comparable] struct {
		name     string
		got      T
		expected T
	}

	Tests[T comparable] []Test[T]

	// Define a config struct to populate.
	Config struct {
		DatabaseIpPort string `env:"DatabaseIpPort"`
	}
)

const (
	sampleEnvContent = `
# Comment line
PORT=8080
GO_MODE=Release
ENABLE_COLOR=true
LOG_FILE_PATH="/var/log/app.log"
SERVER_LOGS="true"
ALLOW_CREDENTIALS=true
CORS_ORIGINS="["http://localhost:3000","https://example.com"]"
CORS_METHODS="["GET","POST","PUT","DELETE"]"
MAX_UPLOAD_SIZE=10485760
API_TIMEOUT=300s
CACHE_TTL=3600s
`

	testValue        = "value"
	testDefaultVal   = "default_val"
	testDefaultValue = "default_value"
)

var (
	config_mock Config
	once_mock   sync.Once
)

func (tests Tests[T]) runTests(t *testing.T) {
	t.Helper()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Error("Test ", tc.name, " failed: expected ", tc.expected, " got ", tc.got)
			}
		})
	}
}

func Test_String_FieldDecoding(t *testing.T) {
	// Simulated environment variables.
	envMap := map[string]string{
		"CASE_1": "value-1",
		"CASE_5": "",
	}

	// String struct with all cases.
	var config struct {
		Case1 string `                                  env:"CASE_1"` // Exists in envMap.
		Case2 string `default:"default-value"           env:"CASE_2"` // Not in envMap, fallback to default.
		Case3 string `                                  env:"CASE_3"` // Not in envMap, no default.
		Case4 string `default:""                        env:"CASE_4"` // Default is empty string.
		Case5 string `                                  env:"CASE_5"` // Exists in envMap, value is empty string.
		Case6 string `default:"default-without-env-tag"`              // No env tag., only default.
		Case7 string // No env or default tag.
	}

	// Populate struct with env.
	if err := populateStruct(envMap, &config); err != nil {
		t.Errorf("Error setting value for field %v", err)
	}

	// Now assert each case.
	tests := Tests[string]{
		{"exists in envMap", config.Case1, "value-1"},
		{"not in envMap, fallback to default", config.Case2, "default-value"},
		{"not in envMap, no default", config.Case3, ""},               // No env, no default.
		{"default is empty string", config.Case4, ""},                 // No env, empty default.
		{"exists in envMap, value is empty string", config.Case5, ""}, // Exists in env with empty string.
		{"no env tag, only default", config.Case6, ""},
		{"no env or default tag", config.Case7, ""}, // No tag at all.
	}

	// Run test.
	tests.runTests(t)
}

func Test_Int_FieldDecoding(t *testing.T) {
	envMap := map[string]string{
		"CASE_1": "42",  // Valid int.
		"CASE_4": "",    // Empty string.
		"CASE_5": "abc", // Invalid int.
	}

	var config struct {
		Case1 int `              env:"CASE_1"` // Exists in envMap.
		Case2 int `default:"100" env:"CASE_2"` // Not in envMap, fallback to default.
		Case3 int `              env:"CASE_3"` // Not in envMap, no default.
		Case4 int `              env:"CASE_4"` // Exists in envMap, but empty string.
		Case5 int `              env:"CASE_5"` // Exists in envMap, but invalid int.
		Case6 int `default:"200"`              // No env tag., only default.
		Case7 int // No env or default tag.
	}

	if err := populateStruct(envMap, &config); err == nil {
		t.Errorf("Error setting value for field %v", err)
	}

	tests := Tests[int]{
		{"exists in envMap", config.Case1, 42},
		{"not in envMap, fallback to default", config.Case2, 100},
		{"not in envMap, no default", config.Case3, 0},
		{"exists in envMap, value is empty string", config.Case4, 0},
		{"exists in envMap, invalid int", config.Case5, 0},
		{"no env tag", config.Case6, 0},
		{"no env or default tag", config.Case7, 0},
	}

	tests.runTests(t)
}

func timeinsecond(value int) time.Duration {
	return time.Duration(value) * time.Second
}

func Test_timeduration_FieldDecoding(t *testing.T) {
	envMap := map[string]string{
		"CASE_1": "5s",  // Valid time duration.
		"CASE_4": "",    // Empty string.
		"CASE_5": "abc", // Invalid time duration.
	}

	var config struct {
		Case1 time.Duration `              env:"CASE_1"` // Exists in envMap.
		Case2 time.Duration `default:"10s" env:"CASE_2"` // Not in envMap, fallback to default.
		Case3 time.Duration `              env:"CASE_3"` // Not in envMap, no default.
		Case4 time.Duration `              env:"CASE_4"` // Exists in envMap, but empty string.
		Case5 time.Duration `              env:"CASE_5"` // Exists in envMap, but invalid time duration.
		Case6 time.Duration `default:"10s"`              // No env tag., only default.
		Case7 time.Duration // No env or default tag.
	}

	if err := populateStruct(envMap, &config); err == nil {
		t.Errorf("Error setting value for field %v", err)
	}

	tests := Tests[time.Duration]{
		{"exists in envMap", config.Case1, timeinsecond(5)},
		{"not in envMap, fallback to default", config.Case2, timeinsecond(10)},
		{"not in envMap, no default", config.Case3, timeinsecond(0)},
		{"exists in envMap, value is empty string", config.Case4, timeinsecond(0)},
		{"exists in envMap, invalid time duration", config.Case5, timeinsecond(0)},
		{"no env tag", config.Case6, timeinsecond(0)},
		{"no env or default tag", config.Case7, timeinsecond(0)},
	}

	tests.runTests(t)
}

func Test_bool_FieldDecoding(t *testing.T) {
	envMap := map[string]string{
		"CASE_1": "true", // Valid time duration.
		"CASE_4": "",     // Empty string.
		"CASE_5": "abc",  // Invalid time duration.
	}

	var config struct {
		Case1 bool `               env:"CASE_1"` // Exists in envMap.
		Case2 bool `default:"true" env:"CASE_2"` // Not in envMap, fallback to default.
		Case3 bool `               env:"CASE_3"` // Not in envMap, no default.
		Case4 bool `               env:"CASE_4"` // Exists in envMap, but empty bool.
		Case5 bool `               env:"CASE_5"` // Exists in envMap, but invalid bool.
		Case6 bool `default:"true"`              // No env tag., only default.
		Case7 bool // No env or default tag.
	}

	if err := populateStruct(envMap, &config); err == nil {
		t.Errorf("Error setting value for field %v", err)
	}

	tests := Tests[bool]{
		{"exists in envMap", config.Case1, true},
		{"not in envMap, fallback to default", config.Case2, true},
		{"not in envMap, no default", config.Case3, false},
		{"exists in envMap, value is empty string", config.Case4, false},
		{"exists in envMap, invalid bool", config.Case5, false},
		{"no env tag", config.Case6, false},
		{"no env or default tag", config.Case7, false},
	}

	tests.runTests(t)
}

func Test_float_FieldDecoding(t *testing.T) {
	envMap := map[string]string{
		"CASE_1": "1.5", // Valid time duration.
		"CASE_4": "",    // Empty string.
		"CASE_5": "abc", // Invalid time duration.
	}

	var config struct {
		Case1 float32 `              env:"CASE_1"` // Exists in envMap.
		Case2 float32 `default:"2.5" env:"CASE_2"` // Not in envMap, fallback to default.
		Case3 float32 `              env:"CASE_3"` // Not in envMap, no default.
		Case4 float32 `              env:"CASE_4"` // Exists in envMap, but empty float32.
		Case5 float32 `              env:"CASE_5"` // Exists in envMap, but invalid float32.
		Case6 float32 `default:"2.5"`              // No env tag., only default.
		Case7 float32 // No env or default tag.
	}

	if err := populateStruct(envMap, &config); err == nil {
		t.Errorf("Error setting value for field %v", err)
	}

	tests := Tests[float32]{
		{"exists in envMap", config.Case1, 1.5},
		{"not in envMap, fallback to default", config.Case2, 2.5},
		{"not in envMap, no default", config.Case3, 0},
		{"exists in envMap, value is empty string", config.Case4, 0},
		{"exists in envMap, invalid float32", config.Case5, 0},
		{"no env tag", config.Case6, 0},
		{"no env or default tag", config.Case7, 0},
	}

	tests.runTests(t)
}

func BenchmarkLoadAndParse(b *testing.B) {
	filePath := createTempEnvFile(b)

	defer os.Remove(filePath)

	b.ResetTimer()

	for range b.N {
		err := LoadEnv_test(filePath)
		if err != nil {
			b.Fatalf("LoadAndParse failed: %v", err)
		}
	}
}

// LoadEnv loads a .env file and decodes values into a struct using `env` tags.
func LoadEnv_test(filePath string) error {
	var (
		source Config
		err    error
	)

	once_mock.Do(func() {
		err = LoadAndParse(filePath, &source)
		if err == nil {
			config_mock = source
		}
	})

	return err
}

// Func Benchmark_LoadAndParse(b *testing.B) {
// 	tmpFile, err := CreateTempEnv()
// 	if err != nil {
// 		b.Fatalf("failed to load file : %v", err)
// 	}
// 	defer os.Remove(tmpFile.Name()) // cleanup
// 	defer tmpFile.Close()

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		err := bkconfig.InitConfig(tmpFile.Name())
// 		if err != nil {
// 			b.Fatalf("LoadAndParse failed: %v", err)
// 		}
// 	}
// }.

func createTempEnvFile(b *testing.B) string {
	b.Helper()

	tempFile, err := os.CreateTemp(b.TempDir(), "benchmark_env_*.env")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}

	_, err = tempFile.WriteString(sampleEnvContent)
	if err != nil {
		b.Fatalf("failed to write to temp file: %v", err)
	}

	if err := tempFile.Close(); err != nil {
		b.Fatalf("failed to close temp file: %v", err)
	}

	return tempFile.Name()
}

func createHeavyTempEnvFile(b *testing.B) string {
	b.Helper()

	tempFile, err := os.CreateTemp(b.TempDir(), "benchmark_env_heavy_*.env")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}

	// Write ~10,000 key=value lines to simulate a heavy .env file.
	writer := bufio.NewWriter(tempFile)
	_, _ = writer.WriteString("# Simulated heavy .env file\n")

	for i := range 10000 {
		fmt.Fprintf(writer, "KEY_%d=value_%d\n", i, i)
	}

	if err := writer.Flush(); err != nil {
		b.Fatalf("failed to flush writer: %v", err)
	}

	if err := tempFile.Close(); err != nil {
		b.Fatalf("failed to close temp file: %v", err)
	}

	return tempFile.Name()
}

// Test cases specifically for populateStruct function.
func Test_populateStruct(t *testing.T) {
	t.Run("Input Validation", func(t *testing.T) {
		envMap := map[string]string{"TEST": testValue}

		t.Run("target is not a pointer", func(t *testing.T) {
			var config struct {
				Field string `env:"TEST"`
			}

			err := populateStruct(envMap, config) // Pass struct instead of pointer.
			if err == nil {
				t.Error("Expected error when target is not a pointer")
			}

			expectedMsg := "target must be a pointer"
			if err.Error() != expectedMsg {
				t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
			}
		})

		t.Run("target is not a pointer to struct", func(t *testing.T) {
			var notStruct string

			err := populateStruct(envMap, &notStruct) // Pointer to string instead of struct.
			if err == nil {
				t.Error("Expected error when target is not a pointer to struct")
			}

			expectedMsg := "target must be a pointer to struct"
			if err.Error() != expectedMsg {
				t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
			}
		})

		t.Run("valid pointer to struct", func(t *testing.T) {
			var config struct {
				Field string `env:"TEST"`
			}

			err := populateStruct(envMap, &config)
			if err != nil {
				t.Errorf("Unexpected error for valid struct: %v", err)
			}

			if config.Field != testValue {
				t.Errorf("Expected 'value', got '%s'", config.Field)
			}
		})
	})

	t.Run("Required Field Validation", func(t *testing.T) {
		t.Run("required field present in envMap", func(t *testing.T) {
			envMap := map[string]string{"REQUIRED_FIELD": "present"}

			var config struct {
				RequiredField string `env:"REQUIRED_FIELD" required:"true"`
			}

			err := populateStruct(envMap, &config)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if config.RequiredField != "present" {
				t.Errorf("Expected 'present', got '%s'", config.RequiredField)
			}
		})

		t.Run("required field missing from envMap with no default", func(t *testing.T) {
			envMap := map[string]string{} // Empty envMap.

			var config struct {
				RequiredField string `env:"MISSING_FIELD" required:"true"`
			}

			err := populateStruct(envMap, &config)
			if err == nil {
				t.Error("Expected error for missing required field")
			}

			expectedSubstring := "missing required field: field=RequiredField env=MISSING_FIELD"
			if !strings.Contains(err.Error(), expectedSubstring) {
				t.Errorf("Expected error to contain '%s', got '%s'", expectedSubstring, err.Error())
			}
		})

		t.Run("required field missing but has default value", func(t *testing.T) {
			envMap := map[string]string{} // Empty envMap.

			var config struct {
				RequiredField string `default:"default_val" env:"MISSING_FIELD" required:"true"`
			}

			err := populateStruct(envMap, &config)
			if err != nil {
				t.Errorf("Unexpected error when required field has default: %v", err)
			}

			if config.RequiredField != testDefaultVal {
				t.Errorf("Expected 'default_val', got '%s'", config.RequiredField)
			}
		})

		t.Run("optional field missing", func(t *testing.T) {
			envMap := map[string]string{} // Empty envMap.

			var config struct {
				OptionalField string `env:"MISSING_FIELD"`
			}

			err := populateStruct(envMap, &config)
			if err != nil {
				t.Errorf("Unexpected error for missing optional field: %v", err)
			}

			if config.OptionalField != "" {
				t.Errorf("Expected empty string, got '%s'", config.OptionalField)
			}
		})
	})

	t.Run("Mixed Required and Optional Fields", func(t *testing.T) {
		t.Run("all fields valid", func(t *testing.T) {
			envMap := map[string]string{
				"REQ_FIELD": "req_value",
				"OPT_FIELD": "opt_value",
			}

			var config struct {
				RequiredField string `                      env:"REQ_FIELD" required:"true"`
				OptionalField string `                      env:"OPT_FIELD"`
				DefaultField  string `default:"default_val" env:"MISSING"`
			}

			err := populateStruct(envMap, &config)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if config.RequiredField != "req_value" {
				t.Errorf("Expected 'req_value', got '%s'", config.RequiredField)
			}

			if config.OptionalField != "opt_value" {
				t.Errorf("Expected 'opt_value', got '%s'", config.OptionalField)
			}

			if config.DefaultField != "default_val" {
				t.Errorf("Expected 'default_val', got '%s'", config.DefaultField)
			}
		})

		t.Run("required field missing in mixed scenario", func(t *testing.T) {
			envMap := map[string]string{
				"OPT_FIELD": "opt_value",
			}

			var config struct {
				RequiredField string `env:"MISSING_REQ" required:"true"`
				OptionalField string `env:"OPT_FIELD"`
			}

			err := populateStruct(envMap, &config)
			if !errors.Is(err, errMissingRequiredField) {
				t.Errorf("expected ErrMissingRequiredField, got %v", err)
			}

			if config.OptionalField != "" {
				t.Errorf("expected OptionalField to remain unset, got %q", config.OptionalField)
			}
		})
	})

	t.Run("Error Handling from setValue", func(t *testing.T) {
		t.Run("invalid int conversion", func(t *testing.T) {
			envMap := map[string]string{"INVALID_INT": "not_a_number"}

			var config struct {
				InvalidInt int `env:"INVALID_INT"`
			}

			err := populateStruct(envMap, &config)
			if err == nil {
				t.Error("Expected error for invalid int conversion")
			}

			expectedSubstring := "invalid int for field"
			if !strings.Contains(err.Error(), expectedSubstring) {
				t.Errorf("Expected error to contain '%s', got '%s'", expectedSubstring, err.Error())
			}
		})

		t.Run("invalid bool conversion", func(t *testing.T) {
			envMap := map[string]string{"INVALID_BOOL": "not_a_bool"}

			var config struct {
				InvalidBool bool `env:"INVALID_BOOL"`
			}

			err := populateStruct(envMap, &config)
			if err == nil {
				t.Error("Expected error for invalid bool conversion")
			}

			expectedSubstring := "invalid bool for field"
			if !strings.Contains(err.Error(), expectedSubstring) {
				t.Errorf("Expected error to contain '%s', got '%s'", expectedSubstring, err.Error())
			}
		})

		t.Run("invalid duration conversion", func(t *testing.T) {
			envMap := map[string]string{"INVALID_DURATION": "not_a_duration"}

			var config struct {
				InvalidDuration time.Duration `env:"INVALID_DURATION"`
			}

			err := populateStruct(envMap, &config)
			if err == nil {
				t.Error("Expected error for invalid duration conversion")
			}

			expectedSubstring := "invalid duration for field"
			if !strings.Contains(err.Error(), expectedSubstring) {
				t.Errorf("Expected error to contain '%s', got '%s'", expectedSubstring, err.Error())
			}
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("empty struct", func(t *testing.T) {
			envMap := map[string]string{"KEY": "value"}

			var config struct{}

			err := populateStruct(envMap, &config)
			if err != nil {
				t.Errorf("Unexpected error for empty struct: %v", err)
			}
		})

		t.Run("struct with no env tags", func(t *testing.T) {
			envMap := map[string]string{"KEY": "value"}

			var config struct {
				Field1 string
				Field2 int
				Field3 bool
			}

			err := populateStruct(envMap, &config)
			if err != nil {
				t.Errorf("Unexpected error for struct with no env tags: %v", err)
			}
			// All fields should remain at zero values.
			if config.Field1 != "" || config.Field2 != 0 || config.Field3 != false {
				t.Error("Fields without env tags should remain at zero values")
			}
		})

		t.Run("struct with unexported fields", func(t *testing.T) {
			envMap := map[string]string{"PUBLIC": "value"}

			var config struct {
				PublicField  string `env:"PUBLIC"`  // Exported field (capital P).
				privateField string `env:"PRIVATE"` // Unexported field.
			}

			err := populateStruct(envMap, &config)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if config.PublicField != "value" {
				t.Errorf("Expected 'value', got '%s'", config.PublicField)
			}
			// PrivateField should remain empty since it can't be set.
		})

		t.Run("nil envMap", func(t *testing.T) {
			var config struct {
				Field string `default:"default_value" env:"TEST"`
			}

			err := populateStruct(nil, &config)
			if err != nil {
				t.Errorf("Unexpected error for nil envMap: %v", err)
			}

			if config.Field != testDefaultValue {
				t.Errorf("Expected 'default_value', got '%s'", config.Field)
			}
		})

		t.Run("empty envMap", func(t *testing.T) {
			envMap := make(map[string]string)

			var config struct {
				Field string `default:"default_value" env:"TEST"`
			}

			err := populateStruct(envMap, &config)
			if err != nil {
				t.Errorf("Unexpected error for empty envMap: %v", err)
			}

			if config.Field != "default_value" {
				t.Errorf("Expected 'default_value', got '%s'", config.Field)
			}
		})
	})

	t.Run("Complex Scenarios", func(t *testing.T) {
		t.Run("multiple types with mixed validation", func(t *testing.T) {
			envMap := map[string]string{
				"STRING_VAL":   "test_string",
				"INT_VAL":      "42",
				"BOOL_VAL":     "true",
				"DURATION_VAL": "5s",
			}

			var config struct {
				RequiredString   string        `                env:"STRING_VAL"   required:"true"`
				OptionalInt      int           `                env:"INT_VAL"`
				DefaultBool      bool          `default:"false" env:"MISSING_BOOL"`
				RequiredDuration time.Duration `                env:"DURATION_VAL" required:"true"`
				SkippedField     string        // No env tag.
			}

			err := populateStruct(envMap, &config)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if config.RequiredString != "test_string" {
				t.Errorf("Expected 'test_string', got '%s'", config.RequiredString)
			}

			if config.OptionalInt != 42 {
				t.Errorf("Expected 42, got %d", config.OptionalInt)
			}

			if config.DefaultBool {
				t.Errorf("Expected false, got %t", config.DefaultBool)
			}

			if config.RequiredDuration != 5*time.Second {
				t.Errorf("Expected 5s, got %v", config.RequiredDuration)
			}

			if config.SkippedField != "" {
				t.Errorf("Expected empty string for skipped field, got '%s'", config.SkippedField)
			}
		})
	})
}

// Test to demonstrate the critical slice processing bug.
func Test_SliceProcessingBug(t *testing.T) {
	t.Run("slice with empty values causes panic", func(t *testing.T) {
		envMap := map[string]string{
			"INT_SLICE_WITH_EMPTY":  "1,,3,4", // Empty value in middle.
			"INT_SLICE_WITH_SPACES": "1, ,3",  // Space-only value.
		}

		var config struct {
			IntSliceWithEmpty  []int `env:"INT_SLICE_WITH_EMPTY"`
			IntSliceWithSpaces []int `env:"INT_SLICE_WITH_SPACES"`
		}

		// This should handle empty values gracefully but currently might panic.
		err := populateStruct(envMap, &config)
		if err != nil {
			t.Errorf("Should handle empty slice values gracefully, got error: %v", err)
		}

		// Expected behavior: skip empty values.
		expectedEmpty := []int{1, 3, 4} // Should skip the empty middle value.
		expectedSpaces := []int{1, 3}   // Should skip the space-only value.

		if len(config.IntSliceWithEmpty) != len(expectedEmpty) {
			t.Errorf("Expected slice length %d, got %d for IntSliceWithEmpty",
				len(expectedEmpty), len(config.IntSliceWithEmpty))
		}

		if len(config.IntSliceWithSpaces) != len(expectedSpaces) {
			t.Errorf("Expected slice length %d, got %d for IntSliceWithSpaces",
				len(expectedSpaces), len(config.IntSliceWithSpaces))
		}
	})
}

// Test_IntegerOverflowDetection tests integer overflow handling.
func Test_IntegerOverflowDetection(t *testing.T) {
	envMap := map[string]string{
		"INT8_OVERFLOW":  "256", // Max int8 is 127.
		"UINT8_OVERFLOW": "256", // Max uint8 is 255.
	}

	var config struct {
		Int8Field  int8  `env:"INT8_OVERFLOW"`
		Uint8Field uint8 `env:"UINT8_OVERFLOW"`
	}

	err := populateStruct(envMap, &config)
	// Should detect overflow and return error.
	if err == nil {
		t.Error("Expected error for integer overflow, but got none")
	}
}

// Test_MapDuplicateKeyHandling tests map duplicate key handling behavior.
func Test_MapDuplicateKeyHandling(t *testing.T) {
	envMap := map[string]string{
		"SETTINGS_WITH_DUPLICATES": "key1:value1,key1:value2,key2:value3",
	}

	var config struct {
		Settings map[string]string `env:"SETTINGS_WITH_DUPLICATES"`
	}

	err := populateStruct(envMap, &config)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Current behavior: last value wins (value2)
	// But this should ideally be detected and either error or warn.
	if config.Settings["key1"] != "value2" {
		t.Errorf("Expected duplicate key to have last value 'value2', got '%s'",
			config.Settings["key1"])
	}

	// This demonstrates the issue - silent overwrite.
	t.Logf("Map with duplicates: %+v", config.Settings)
}
