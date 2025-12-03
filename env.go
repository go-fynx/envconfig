package envload

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type (
	fieldResolver struct {
		field    reflect.StructField
		value    reflect.Value
		rawValue string
	}
)

const (
	// [keyValueSeparatorLimit] is the maximum number of parts when splitting key:value pairs.
	keyValueSeparatorLimit = 2
)

var (
	errTargetMustBePointer         = errors.New("target must be a pointer")
	errTargetMustBePointerToStruct = errors.New("target must be a pointer to struct")
	errInvalidMapFormat            = errors.New("invalid map format for field")
	errUnsupportedMapValueType     = errors.New("unsupported map value type")
	errMissingRequiredField        = errors.New("missing required field")
)

// LoadAndParse reads a .env file and maps its values to a struct.
// It supports env, default, and required struct tags.
// If the env file cannot be read, it logs a warning and continues with default values only.
func LoadAndParse(filePath string, target any) error {
	envMap, err := godotenv.Read(filePath)
	if err != nil {
		// Log warning and continue with defaults only - allows graceful degradation.
		log.Printf("\033[33m[Warning]:\033[0m Could not read env file [%s: %v]. Using defaults only.", filePath, err)

		envMap = make(map[string]string)
	}

	return populateStruct(envMap, target)
}

// validateStruct validates that the target is a pointer to a struct.
func validateStruct(target any) error {
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr {
		return errTargetMustBePointer
	}

	if value.Elem().Kind() != reflect.Struct {
		return errTargetMustBePointerToStruct
	}

	return nil
}

// populateStruct sets values from envMap into the target struct.
// Uses struct tags: `env` for key, `default` for fallback value, `required` for validation.
func populateStruct(envMap map[string]string, target any) error {
	if err := validateStruct(target); err != nil {
		return err
	}

	value := reflect.ValueOf(target)

	value = value.Elem()
	typ := value.Type()

	var resolver fieldResolver
	for i := range value.NumField() {
		resolver.field = typ.Field(i)
		resolver.value = value.Field(i)

		resolver.resolveValue(envMap)

		if resolver.rawValue == "" && resolver.isRequired() {
			return fmt.Errorf("%w: field=%s env=%s",
				errMissingRequiredField,
				resolver.field.Name,
				resolver.field.Tag.Get("env"),
			)
		}

		if resolver.rawValue == "" {
			// Skip fields without env tag or that can't be set.
			continue
		}

		if err := resolver.setValue(); err != nil {
			return err
		}
	}

	return nil
}

func (resolver *fieldResolver) resolveValue(envMap map[string]string) {
	resolver.rawValue = ""
	envKey := resolver.field.Tag.Get("env")

	if envKey == "" || !resolver.value.CanSet() {
		return // Skip fields without env tag or that can't be set.
	}

	rawValue, ok := envMap[envKey]
	if !ok {
		rawValue = resolver.field.Tag.Get("default")
	}

	resolver.rawValue = rawValue
}

// isRequired checks if a field has the required tag set to true.
func (resolver *fieldResolver) isRequired() bool {
	return resolver.field.Tag.Get("required") == "true"
}

// setValue sets rawValue into the given fieldVal based on its kind and type.
// Supported types: string, int, uint, float, bool, time.Duration,
// slices ([]string, []int, []float64, []bool), maps (map[string]string, map[string]int, etc.).
//
//nolint:exhaustive,revive,cyclop // note: This function is used to set values into the given fieldVal based on its kind and type. so we need to ignore some linters.
func (resolver *fieldResolver) setValue() error {
	switch resolver.field.Type.Kind() {
	case reflect.String:
		return resolver.setString()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return resolver.setIntOrDuration()

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return resolver.setUint()

	case reflect.Float32, reflect.Float64:
		return resolver.setFloat()

	case reflect.Bool:
		return resolver.setBool()

	case reflect.Slice:
		return resolver.setSlice()

	case reflect.Map:
		return resolver.setMap()
	default:
	}

	// Skip unsupported types silently.
	return nil
}

// setString sets a plain string value.
func (resolver *fieldResolver) setString() error {
	resolver.value.SetString(resolver.rawValue)
	return nil
}

// setIntOrDuration sets an integer or time.Duration.
// Example: TIMEOUT=5s -> time.Duration(5 * time.Second).
func (resolver *fieldResolver) setIntOrDuration() error {
	if resolver.value.Type().PkgPath() == "time" && resolver.value.Type().Name() == "Duration" {
		return resolver.setDuration()
	}
	return resolver.setInt()
}

// setDuration parses and sets a time.Duration value from a string.
// It expects strings like "5s", "2m", "1h30m" etc., and sets the duration into the fieldVal.
// Example: TIMEOUT="5s" -> fieldVal.Set(time.Duration(5 * time.Second)).
func (resolver *fieldResolver) setDuration() error {
	dur, err := time.ParseDuration(resolver.rawValue)
	if err != nil {
		return fmt.Errorf("invalid duration for field '%s': %w", resolver.field.Name, err)
	}

	resolver.value.SetInt(int64(dur)) // Duration is an alias of int64.
	return nil
}

// setInt parses and sets an integer value from a string into the reflect.Value.
// It supports all integer kinds (int, int8, int16, int32, int64).
// Example: RETRIES="3" -> fieldVal.SetInt(3).
func (resolver *fieldResolver) setInt() error {
	intVal, err := strconv.ParseInt(resolver.rawValue, 10, resolver.value.Type().Bits())
	if err != nil {
		return fmt.Errorf("invalid int for field '%s': %w", resolver.field.Name, err)
	}

	resolver.value.SetInt(intVal)
	return nil
}

// setUint sets an unsigned integer value.
func (resolver *fieldResolver) setUint() error {
	uintVal, err := strconv.ParseUint(resolver.rawValue, 10, resolver.value.Type().Bits())
	if err != nil {
		return fmt.Errorf("invalid uint for field '%s': %w", resolver.field.Name, err)
	}

	resolver.value.SetUint(uintVal)
	return nil
}

// setFloat sets a float value (float32 or float64).
func (resolver *fieldResolver) setFloat() error {
	floatVal, err := strconv.ParseFloat(resolver.rawValue, resolver.value.Type().Bits())
	if err != nil {
		return fmt.Errorf("invalid float for field '%s': %w", resolver.field.Name, err)
	}

	resolver.value.SetFloat(floatVal)
	return nil
}

// setBool sets a boolean value.
// Accepts: "true", "false", "1", "0".
func (resolver *fieldResolver) setBool() error {
	boolVal, err := strconv.ParseBool(resolver.rawValue)
	if err != nil {
		return fmt.Errorf("invalid bool for field '%s': %w", resolver.field.Name, err)
	}

	resolver.value.SetBool(boolVal)
	return nil
}

// setSlice sets a slice by splitting on commas and converting to appropriate types.
// Supports: []string, []int, []int64, []float64, []bool
// Example: TAGS=dev,prod,test -> []string{"dev", "prod", "test"}
//
//	PORTS=8080,9090,3000 -> []int{8080, 9090, 3000}.
//
//nolint:exhaustive // note: This function is used to set values into the given fieldVal based on its kind and type.
func (resolver *fieldResolver) setSlice() error {
	elemKind := resolver.value.Type().Elem().Kind()
	parts := strings.Split(resolver.rawValue, ",")

	// Trim spaces from all parts.
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	switch elemKind {
	case reflect.String:
		resolver.value.Set(reflect.ValueOf(parts))
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return resolver.setIntSlice(parts)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return resolver.setUintSlice(parts)

	case reflect.Float32, reflect.Float64:
		return resolver.setFloatSlice(parts)

	case reflect.Bool:
		return resolver.setBoolSlice(parts)

	default:
		return nil // Skip unsupported slice types.
	}
}

// [setIntSlice] converts string parts to integers and sets the slice.
func (resolver *fieldResolver) setIntSlice(parts []string) error {
	elemType := resolver.value.Type().Elem()

	// Filter out empty parts first to get correct slice size.
	validParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			validParts = append(validParts, part)
		}
	}

	// Create slice with correct size (number of valid parts).
	slice := reflect.MakeSlice(resolver.value.Type(), len(validParts), len(validParts))

	// Process only valid parts with sequential indexes.
	for i, part := range validParts {
		intVal, err := strconv.ParseInt(part, 10, elemType.Bits())
		if err != nil {
			return fmt.Errorf("invalid int in slice for field '%s' at index %d: %w", resolver.field.Name, i, err)
		}

		slice.Index(i).SetInt(intVal)
	}

	resolver.value.Set(slice)

	return nil
}

// [setUintSlice] converts string parts to unsigned integers and sets the slice.
func (resolver *fieldResolver) setUintSlice(parts []string) error {
	elemType := resolver.value.Type().Elem()

	// Filter out empty parts first to get correct slice size.
	validParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			validParts = append(validParts, part)
		}
	}

	// Create slice with correct size (number of valid parts).
	slice := reflect.MakeSlice(resolver.value.Type(), len(validParts), len(validParts))

	// Process only valid parts with sequential indexes.
	for i, part := range validParts {
		uintVal, err := strconv.ParseUint(part, 10, elemType.Bits())
		if err != nil {
			return fmt.Errorf("invalid uint in slice for field '%s' at index %d: %w", resolver.field.Name, i, err)
		}

		slice.Index(i).SetUint(uintVal)
	}

	resolver.value.Set(slice)

	return nil
}

// [setFloatSlice] converts string parts to floats and sets the slice.
func (resolver *fieldResolver) setFloatSlice(parts []string) error {
	elemType := resolver.value.Type().Elem()

	// Filter out empty parts first to get correct slice size.
	validParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			validParts = append(validParts, part)
		}
	}

	// Create slice with correct size (number of valid parts).
	slice := reflect.MakeSlice(resolver.value.Type(), len(validParts), len(validParts))

	// Process only valid parts with sequential indexes.
	for i, part := range validParts {
		floatVal, err := strconv.ParseFloat(part, elemType.Bits())
		if err != nil {
			return fmt.Errorf("invalid float in slice for field '%s' at index %d: %w", resolver.field.Name, i, err)
		}

		slice.Index(i).SetFloat(floatVal)
	}

	resolver.value.Set(slice)

	return nil
}

// setBoolSlice converts string parts to booleans and sets the slice.
func (resolver *fieldResolver) setBoolSlice(parts []string) error {
	// Filter out empty parts first to get correct slice size.
	validParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			validParts = append(validParts, part)
		}
	}

	// Create slice with correct size (number of valid parts).
	slice := reflect.MakeSlice(resolver.value.Type(), len(validParts), len(validParts))

	// Process only valid parts with sequential indexes.
	for i, part := range validParts {
		boolVal, err := strconv.ParseBool(part)
		if err != nil {
			return fmt.Errorf("invalid bool in slice for field '%s' at index %d: %w", resolver.field.Name, i, err)
		}

		slice.Index(i).SetBool(boolVal)
	}

	resolver.value.Set(slice)

	return nil
}

// setMap sets a map by parsing comma-separated key:value pairs.
// Supports: map[string]string, map[string]int, map[string]float64, map[string]bool
// Example: SETTINGS=debug:true,theme:dark -> map[string]string{"debug":"true", "theme":"dark"}
//
//	PORTS=api:8080,db:5432 -> map[string]int{"api":8080, "db":5432}
func (resolver *fieldResolver) setMap() error {
	keyKind := resolver.value.Type().Key().Kind()
	valueKind := resolver.value.Type().Elem().Kind()

	// Only support string keys for now.
	if keyKind != reflect.String {
		return nil // Unsupported map key type.
	}

	pairs := strings.Split(resolver.rawValue, ",")
	mapType := resolver.value.Type()
	result := reflect.MakeMap(mapType)

	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), ":", keyValueSeparatorLimit)
		if len(kv) != keyValueSeparatorLimit {
			return fmt.Errorf("%w for field '%s': '%s'", errInvalidMapFormat, resolver.field.Name, pair)
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		// Convert value based on map's value type.
		convertedValue, err := resolver.convertMapValue(value, valueKind)
		if err != nil {
			return fmt.Errorf("invalid map value for field '%s' key '%s': %w", resolver.field.Name, key, err)
		}

		keyVal := reflect.ValueOf(key)
		result.SetMapIndex(keyVal, convertedValue)
	}

	resolver.value.Set(result)

	return nil
}

// convertMapValue converts a string value to the appropriate type for map values.
//
//nolint:exhaustive,gocyclo,cyclop,revive // note: This function is used to set values into the given fieldVal based on its kind and type. so we need to ignore some linters.
func (resolver *fieldResolver) convertMapValue(value string, valueKind reflect.Kind) (reflect.Value, error) {
	switch valueKind {
	case reflect.String:
		return reflect.ValueOf(value), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(intVal).Convert(resolver.value.Type().Elem()), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(uintVal).Convert(resolver.value.Type().Elem()), nil

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(floatVal).Convert(resolver.value.Type().Elem()), nil

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(boolVal), nil

	default:
		return reflect.Value{}, fmt.Errorf("%w: %v", errUnsupportedMapValueType, valueKind)
	}
}
