package parse

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/secrets"
	"github.com/sirupsen/logrus"
)

// ResolveSecrets walks the loaded DSL project and resolves any string values
// that start with "secret://<KEY>" using the provided secrets.Provider.
//
// It mutates proj in place, replacing secret references with their resolved values,
// and returns diagnostics for any failures encountered.
func ResolveSecrets(data any, prov secrets.Provider) []diag.Diagnostic {
	if data == nil || prov == nil {
		return nil
	}
	var diags []diag.Diagnostic
	v := reflect.ValueOf(data)
	logrus.Debug("🔍 Starting secret resolution from root 'project'")
	walk(v, "project", prov, &diags, 0)
	logrus.Debug("✅ Finished secret resolution")
	return diags
}

const maxWalkDepth = 64 // safety guard

// debugLog prints traversal steps with indentation based on recursion depth.
func debugLog(depth int, msg string, args ...any) {
	prefix := strings.Repeat("  ", depth)
	logrus.Debug(prefix+msg, args)
}

// walk recursively traverses any Go value, resolving secret:// strings when found.
func walk(v reflect.Value, path string, prov secrets.Provider, diags *[]diag.Diagnostic, depth int) {
	if depth > maxWalkDepth {
		*diags = append(*diags, diag.Diagnostic{
			Level:   diag.LevelWarn,
			Code:    "SECRET_WALK_DEPTH",
			Message: fmt.Sprintf("maximum walk depth reached at %s; skipping deeper fields", path),
		})
		debugLog(depth, "⚠️  Max depth reached at %s", path)
		return
	}

	// Unwrap pointers/interfaces
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			debugLog(depth, "nil pointer/interface at %s", path)
			return
		}
		debugLog(depth, "unwrapping %s -> %s", path, v.Elem().Kind())
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		debugLog(depth, "visiting string at %s = %q", path, v.String())
		if v.CanSet() {
			if resolved, ok, isSecret := tryResolveSecret(v.String(), path, prov, diags); isSecret {
				if ok {
					debugLog(depth, "✅ resolved secret at %s -> %q", path, resolved)
					v.SetString(resolved)
				} else {
					debugLog(depth, "❌ failed to resolve secret at %s", path)
				}
			}
		}

	case reflect.Struct:
		debugLog(depth, "visiting struct at %s (%s)", path, v.Type())
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			ft := t.Field(i)
			if ft.PkgPath != "" {
				continue // skip unexported
			}
			name := ft.Name
			if tag := ft.Tag.Get("yaml"); tag != "" && tag != "-" {
				parts := strings.Split(tag, ",")
				if parts[0] != "" {
					name = parts[0]
				}
			}
			walk(f, path+"."+name, prov, diags, depth+1)
		}

	case reflect.Map:
		debugLog(depth, "visiting map at %s", path)
		if v.Type().Key().Kind() != reflect.String {
			return
		}
		// Use a temporary map to store changes, as we cannot safely iterate and modify a map that
		// contains non-addressable structs.
		tmpMap := reflect.MakeMap(v.Type())
		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key()
			val := iter.Value()
			childPath := fmt.Sprintf("%s[%q]", path, k.String())

			// Create a settable pointer to the value.
			// If the value is an interface, create a pointer to what's *inside* it
			// to ensure the contents can be modified.
			var settableVal reflect.Value
			if val.Kind() == reflect.Interface && !val.IsNil() {
				elem := val.Elem()
				if !elem.IsValid() {
					tmpMap.SetMapIndex(k, val)
					continue // Skip nil interface content
				}
				settableVal = reflect.New(elem.Type())
				settableVal.Elem().Set(elem)
			} else {
				settableVal = reflect.New(val.Type())
				settableVal.Elem().Set(val)
			}

			walk(settableVal, childPath, prov, diags, depth+1)

			// After walk, put the modified value back into the temporary map.
			// If the original value was an interface, the modified value is placed
			// back directly. The map type `map[string]interface{}` will handle it.
			tmpMap.SetMapIndex(k, settableVal.Elem())
		}
		// Replace the keys in the original map with the updated values.
		for _, k := range tmpMap.MapKeys() {
			v.SetMapIndex(k, tmpMap.MapIndex(k))
		}

	case reflect.Slice, reflect.Array:
		debugLog(depth, "visiting slice/array at %s (len=%d)", path, v.Len())
		for i := 0; i < v.Len(); i++ {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			elem := v.Index(i)
			walk(elem, childPath, prov, diags, depth+1)
		}

	default:
		debugLog(depth, "skipping %s (%s)", path, v.Kind())
	}
}

// tryResolveSecret checks if a given string `s` is a secret reference.
// It returns:
// - resolvedValue, true, true: if `s` is a secret ref and was resolved successfully.
// - "", false, true: if `s` is a secret ref but resolution failed (an error diagnostic is added).
// - "", false, false: if `s` is not a secret ref.
func tryResolveSecret(s, path string, prov secrets.Provider, diags *[]diag.Diagnostic) (string, bool, bool) {
	key, isSecret := strings.CutPrefix(s, "secret://")
	if !isSecret {
		return "", false, false
	}

	logrus.Debugf("🔑 Found secret reference at %s: %q", path, s)

	if key == "" {
		*diags = append(*diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "SECRET_KEY_EMPTY",
			Message: fmt.Sprintf("empty secret reference at %s", path),
		})
		return "", false, true
	}

	resolved, err := prov.Resolve(key)
	if err != nil {
		*diags = append(*diags, diag.Diagnostic{
			Level:   diag.LevelError,
			Code:    "SECRET_RESOLVE_FAIL",
			Message: fmt.Sprintf("failed to resolve %q at %s: %v", key, path, err),
		})
		logrus.Debugf("❌ failed to resolve secret %q at %s: %v", key, path, err)
		return "", false, true
	}

	logrus.Debugf("✅ successfully resolved %q at %s -> %q", key, path, resolved)
	return resolved, true, true
}
