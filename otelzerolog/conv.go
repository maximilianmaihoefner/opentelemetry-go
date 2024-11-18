package otelzerolog

import (
	"fmt"
	logs "go.opentelemetry.io/otel/log"
	"math"
)

func otelAttribute(key string, value interface{}) []logs.KeyValue {
	switch value := value.(type) {
	case bool:
		return []logs.KeyValue{logs.Bool(key, value)}
		// Number information is lost when we're converting to byte to interface{}, let's recover it
	case float64:
		if _, frac := math.Modf(value); frac == 0.0 {
			return []logs.KeyValue{logs.Int64(key, int64(value))}
		} else {
			return []logs.KeyValue{logs.Float64(key, value)}
		}
	case string:
		return []logs.KeyValue{logs.String(key, value)}
	case []interface{}:
		var result []logs.KeyValue
		for _, v := range value {
			// recursively call otelAttribute to handle nested arrays
			result = append(result, otelAttribute(key, v)...)
		}
		return result
	}
	// Default case
	return []logs.KeyValue{logs.String(key, fmt.Sprintf("%v", value))}
}
