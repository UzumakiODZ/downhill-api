package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct{}

// Helper functions for pointer conversion
func strPtr(s string) *string {
	return &s
}

func int32Ptr(i uint) *int32 {
	v := int32(i)
	return &v
}

func float64Ptr(f float64) *float64 {
	return &f
}

// createToken generates a JWT token for the given user ID
func createToken(userID string) (string, error) {
	// TODO: Implement proper JWT token generation
	// For now, returning a simple token format
	return userID + "_token", nil
}
