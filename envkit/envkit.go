package envkit

import (
	"fmt"
	"os"
	"strconv"

	"github.com/half-ogre/go-kit/kit"
)

func GetenvBoolWithDefault(key string, defaultValue bool) (bool, error) {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue, nil
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, kit.WrapError(err, "failed to parse %s as bool", value)
	}

	return boolValue, nil
}

func GetenvIntWithDefault(key string, defaultValue int) (int, error) {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue, nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, kit.WrapError(err, "failed to parse %s as int", value)
	}

	return intValue, nil
}

func GetenvWithDefault(key string, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}

func MustGetenv(key string) string {
	value := os.Getenv(key)

	if value == "" {
		panic(fmt.Sprintf("environment variable %s not set", key))
	}

	return value
}
