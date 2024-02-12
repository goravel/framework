package database

import "unicode"

// Helper function to convert camel case to snake case
func ToSnakeCase(s string) string {
    var result string
    for i, c := range s {
        if unicode.IsUpper(c) {
            if i > 0 {
                result += "_"
            }
            result += string(unicode.ToLower(c))
        } else {
            result += string(c)
        }
    }
    return result
}