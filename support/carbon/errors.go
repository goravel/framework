package carbon

import "fmt"

// returns a failed scan error.
// 失败的扫描错误
var failedScanError = func(src interface{}) error {
	return fmt.Errorf("failed to scan value: %v", src)
}

// returns a invalid timestamp error.
// 无效的时间戳错误
var invalidTimestampError = func(value string) error {
	return fmt.Errorf("invalid timestamp %s, please make sure the timestamp is valid", value)
}
