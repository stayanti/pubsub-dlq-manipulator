package main

import "testing"

func TestExtractProjectID(t *testing.T) {
	// Arrange
	tests := []struct {
		name     string
		path     string // input
		expected string // expected output
	}{
		{
			name:     "valid subscription path",
			path:     "projects/my-project/subscriptions/my-sub",
			expected: "my-project",
		},
		{
			name:     "valid topic path",
			path:     "projects/my-project/topics/my-topic",
			expected: "my-project",
		},
		{
			name:     "invalid path",
			path:     "invalid/path",
			expected: "",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "missing project ID",
			path:     "projects//subscriptions/my-sub",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := extractProjectID(tt.path)

			// Assert
			if result != tt.expected {
				t.Errorf("extractProjectID(%s) = %s; want %s", tt.path, result, tt.expected)
			}
		})
	}
}

func TestExtractSubscriptionID(t *testing.T) {
	// Arrange
	tests := []struct {
		name     string
		path     string // input
		expected string // expected output
	}{
		{
			name:     "valid path",
			path:     "projects/my-project/subscriptions/my-sub",
			expected: "my-sub",
		},
		{
			name:     "invalid path",
			path:     "invalid/path",
			expected: "",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "missing subscription ID",
			path:     "projects/my-project/subscriptions/",
			expected: "",
		},
		{
			name:     "wrong resource type",
			path:     "projects/my-project/topics/my-topic",
			expected: "my-topic", // Current behavior - might want to change this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := extractSubscriptionID(tt.path)

			// Assert
			if result != tt.expected {
				t.Errorf("extractSubscriptionID(%s) = %s; want %s", tt.path, result, tt.expected)
			}
		})
	}
}

func TestExtractTopicID(t *testing.T) {
	// Arrange
	tests := []struct {
		name     string
		path     string // input
		expected string // expected output
	}{
		{
			name:     "valid path",
			path:     "projects/my-project/topics/my-topic",
			expected: "my-topic",
		},
		{
			name:     "invalid path",
			path:     "invalid/path",
			expected: "",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "missing topic ID",
			path:     "projects/my-project/topics/",
			expected: "",
		},
		{
			name:     "wrong resource type",
			path:     "projects/my-project/subscriptions/my-sub",
			expected: "my-sub", // Current behavior - might want to change this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := extractTopicID(tt.path)

			// Assert
			if result != tt.expected {
				t.Errorf("extractTopicID(%s) = %s; want %s", tt.path, result, tt.expected)
			}
		})
	}
}
