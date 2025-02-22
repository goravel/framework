package auth

type Driver interface {
    Login(userID string, data map[string]interface{}) (string, error)    // User login
    Logout(sessionID string) error                                      // User logout
    Authenticate(sessionID string) error                                // Check authentication
}
