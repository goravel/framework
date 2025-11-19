package utils

import (
    "fmt"
    contractsnotification "github.com/goravel/framework/contracts/notification"
    "reflect"
    "strings"
)

// CallToMethod invokes a notification's channel-specific method via reflection.
// It supports both value and pointer receivers, and falls back to PayloadProvider.
// The returned value is normalized as map[string]interface{} for channel consumption.
func CallToMethod(notification interface{}, methodName string, notifiable contractsnotification.Notifiable) (map[string]interface{}, error) {
    v := reflect.ValueOf(notification)
    if !v.IsValid() {
        return nil, fmt.Errorf("invalid notification value")
    }

    // Locate method, preferring pointer receiver if available.
    method := v.MethodByName(methodName)
    if !method.IsValid() && v.CanAddr() {
        method = v.Addr().MethodByName(methodName)
    }
    if !method.IsValid() {
        // Fallback: support PayloadProvider for dynamic channel payloads.
        if provider, ok := v.Interface().(contractsnotification.PayloadProvider); ok {
            channel := strings.ToLower(strings.TrimPrefix(methodName, "To"))
            return provider.PayloadFor(channel, notifiable)
        }
        return nil, fmt.Errorf("method %s not found", methodName)
    }

    // Invoke method with the notifiable.
    results := method.Call([]reflect.Value{reflect.ValueOf(notifiable)})
    if len(results) == 0 {
        return nil, fmt.Errorf("method %s returned no values", methodName)
    }

    // Handle optional error return value.
    if len(results) >= 2 && !results[1].IsNil() {
        if err, ok := results[1].Interface().(error); ok {
            return nil, err
        }
        return nil, fmt.Errorf("second return of %s is not error", methodName)
    }

    // Convert the first return value to map[string]interface{}.
    first := results[0].Interface()
    switch data := first.(type) {
    case map[string]interface{}:
        return data, nil
    case map[string]string:
		out := make(map[string]interface{}, len(data))
		for k, v := range data {
			out[k] = v
		}
        return out, nil
    }

    // Handle struct result by exporting fields.
    if rv := reflect.ValueOf(first); rv.Kind() == reflect.Struct {
        out := make(map[string]interface{})
        rt := rv.Type()
        for i := 0; i < rv.NumField(); i++ {
            field := rt.Field(i)
            if field.PkgPath == "" { // only exported fields
                out[field.Name] = rv.Field(i).Interface()
            }
        }
        return out, nil
    }

    return nil, fmt.Errorf("unsupported return type from %s", methodName)
}
