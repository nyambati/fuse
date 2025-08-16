package parse

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/secrets"
)

// mockProvider is a simple implementation of secrets.Provider for testing.
type mockProvider struct {
	secrets map[string]string
	err     error
}

func (m *mockProvider) Resolve(key string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if val, ok := m.secrets[key]; ok {
		return val, nil
	}
	return "", fmt.Errorf("secret not found: %s", key)
}

type SimpleStruct struct {
	Value string
}

type NestedStruct struct {
	Inner SimpleStruct
	Ptr   *SimpleStruct
}

type TaggedStruct struct {
	Field string `yaml:"field_name"`
}

type unexportedStruct struct {
	value string
}

func TestResolveSecrets(t *testing.T) {
	// Prepare max depth test case
	deepInput := make(map[string]any)
	currentMap := deepInput
	// Create nesting up to maxWalkDepth. The next level will exceed it.
	for range maxWalkDepth {
		nextMap := make(map[string]any)
		currentMap["next"] = nextMap
		currentMap = nextMap
	}
	currentMap["value"] = "secret://deep" // This secret is at a depth that will trigger the limit

	// The 'want' is the same as input because it shouldn't be resolved.
	deepWant := make(map[string]any)
	currentWantMap := deepWant
	for range maxWalkDepth {
		nextMap := make(map[string]any)
		currentWantMap["next"] = nextMap
		currentWantMap = nextMap
	}
	currentWantMap["value"] = "secret://deep"

	testCases := []struct {
		name      string
		input     any
		provider  secrets.Provider
		want      any
		wantDiags []diag.Diagnostic
	}{
		{
			name:      "nil input",
			input:     nil,
			provider:  &mockProvider{},
			want:      nil,
			wantDiags: nil,
		},
		{
			name:      "nil provider",
			input:     &SimpleStruct{Value: "secret://foo"},
			provider:  nil,
			want:      &SimpleStruct{Value: "secret://foo"},
			wantDiags: nil,
		},
		{
			name:      "simple struct field",
			input:     &SimpleStruct{Value: "secret://foo"},
			provider:  &mockProvider{secrets: map[string]string{"foo": "resolved_value"}},
			want:      &SimpleStruct{Value: "resolved_value"},
			wantDiags: nil,
		},
		{
			name:      "nested struct field",
			input:     &NestedStruct{Inner: SimpleStruct{Value: "secret://bar"}},
			provider:  &mockProvider{secrets: map[string]string{"bar": "resolved_bar"}},
			want:      &NestedStruct{Inner: SimpleStruct{Value: "resolved_bar"}},
			wantDiags: nil,
		},
		{
			name:      "pointer field in struct",
			input:     &NestedStruct{Ptr: &SimpleStruct{Value: "secret://baz"}},
			provider:  &mockProvider{secrets: map[string]string{"baz": "resolved_baz"}},
			want:      &NestedStruct{Ptr: &SimpleStruct{Value: "resolved_baz"}},
			wantDiags: nil,
		},
		{
			name:      "string slice",
			input:     []string{"secret://s1", "not a secret", "secret://s2"},
			provider:  &mockProvider{secrets: map[string]string{"s1": "v1", "s2": "v2"}},
			want:      []string{"v1", "not a secret", "v2"},
			wantDiags: nil,
		},
		{
			name:      "struct slice",
			input:     []SimpleStruct{{"secret://s1"}, {"secret://s2"}},
			provider:  &mockProvider{secrets: map[string]string{"s1": "v1", "s2": "v2"}},
			want:      []SimpleStruct{{"v1"}, {"v2"}},
			wantDiags: nil,
		},
		{
			name:      "map with string values",
			input:     map[string]string{"key1": "secret://k1", "key2": "plain"},
			provider:  &mockProvider{secrets: map[string]string{"k1": "v1"}},
			want:      map[string]string{"key1": "v1", "key2": "plain"},
			wantDiags: nil,
		},
		{
			name:      "map with interface values",
			input:     map[string]any{"key1": "secret://k1", "key2": 123},
			provider:  &mockProvider{secrets: map[string]string{"k1": "v1"}},
			want:      map[string]any{"key1": "v1", "key2": 123},
			wantDiags: nil,
		},
		{
			name:      "map with struct values",
			input:     map[string]SimpleStruct{"a": {"secret://k1"}},
			provider:  &mockProvider{secrets: map[string]string{"k1": "v1"}},
			want:      map[string]SimpleStruct{"a": {"v1"}},
			wantDiags: nil,
		},
		{
			name:      "unexported field",
			input:     &unexportedStruct{value: "secret://foo"},
			provider:  &mockProvider{secrets: map[string]string{"foo": "bar"}},
			want:      &unexportedStruct{value: "secret://foo"},
			wantDiags: nil,
		},
		{
			name:      "tagged struct field",
			input:     &TaggedStruct{Field: "secret://foo"},
			provider:  &mockProvider{secrets: map[string]string{"foo": "bar"}},
			want:      &TaggedStruct{Field: "bar"},
			wantDiags: nil,
		},
		{
			name:      "secret not found",
			input:     &SimpleStruct{Value: "secret://nonexistent"},
			provider:  &mockProvider{secrets: map[string]string{}},
			want:      &SimpleStruct{Value: "secret://nonexistent"},
			wantDiags: []diag.Diagnostic{{Level: diag.LevelError, Code: "SECRET_RESOLVE_FAIL"}},
		},
		{
			name:      "empty secret key",
			input:     &SimpleStruct{Value: "secret://"},
			provider:  &mockProvider{},
			want:      &SimpleStruct{Value: "secret://"},
			wantDiags: []diag.Diagnostic{{Level: diag.LevelError, Code: "SECRET_KEY_EMPTY"}},
		},
		{
			name:      "provider returns error",
			input:     &SimpleStruct{Value: "secret://any"},
			provider:  &mockProvider{err: fmt.Errorf("provider error")},
			want:      &SimpleStruct{Value: "secret://any"},
			wantDiags: []diag.Diagnostic{{Level: diag.LevelError, Code: "SECRET_RESOLVE_FAIL"}},
		},
		{
			name: "complex nested data",
			input: map[string]any{
				"config": &NestedStruct{
					Inner: SimpleStruct{"secret://db_pass"},
				},
				"keys": []string{"secret://api_key"},
			},
			provider: &mockProvider{secrets: map[string]string{"db_pass": "secret_password", "api_key": "secret_key"}},
			want: map[string]any{
				"config": &NestedStruct{
					Inner: SimpleStruct{"secret_password"},
				},
				"keys": []string{"secret_key"},
			},
			wantDiags: nil,
		},
		{
			name:      "max depth reached",
			input:     deepInput,
			provider:  &mockProvider{secrets: map[string]string{"deep": "shhh"}},
			want:      deepWant,
			wantDiags: []diag.Diagnostic{{Level: diag.LevelWarn, Code: "SECRET_WALK_DEPTH"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			diags := ResolveSecrets(tc.input, tc.provider)

			if !reflect.DeepEqual(tc.input, tc.want) {
				t.Errorf("ResolveSecrets() got = %#v, want %#v", tc.input, tc.want)
			}

			if len(diags) != len(tc.wantDiags) {
				t.Fatalf("ResolveSecrets() returned %d diagnostics, want %d. Diags: %v", len(diags), len(tc.wantDiags), diags)
			}

			for i, d := range diags {
				if d.Code != tc.wantDiags[i].Code || d.Level != tc.wantDiags[i].Level {
					t.Errorf("Diagnostic %d mismatch: got code %s, level %s; want code %s, level %s",
						i, d.Code, d.Level, tc.wantDiags[i].Code, tc.wantDiags[i].Level)
				}
			}
		})
	}
}
