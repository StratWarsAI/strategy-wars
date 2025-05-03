// pkg/utils/base.go

package utils

func GetStringOrDefault(config map[string]interface{}, key string, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

func GetIntOrDefault(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key].(float64); ok {
		return int(val)
	}
	if val, ok := config[key].(int); ok {
		return val
	}
	return defaultValue
}

func GetFloat64OrDefault(config map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := config[key].(float64); ok {
		return val
	}
	if val, ok := config[key].(int); ok {
		return float64(val)
	}
	return defaultValue
}
