package yaml

import (
	"fmt"
	"strconv"
)

// Build represents a build element in compose file.
// It can take multiple form in the compose file, hence this special type
type Build struct {
	Context    string
	Dockerfile string
	Args       map[string]string
}

// MarshalYAML implements the Marshaller interface.
func (b Build) MarshalYAML() (tag string, value interface{}, err error) {
	m := map[string]interface{}{}
	if b.Context != "" {
		m["context"] = b.Context
	}
	if b.Dockerfile != "" {
		m["dockerfile"] = b.Dockerfile
	}
	if len(b.Args) > 0 {
		m["args"] = b.Args
	}
	return "", m, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (b *Build) UnmarshalYAML(tag string, value interface{}) error {
	switch v := value.(type) {
	case string:
		b.Context = v
	case map[interface{}]interface{}:
		for mapKey, mapValue := range v {
			switch mapKey {
			case "context":
				b.Context = mapValue.(string)
			case "dockerfile":
				b.Dockerfile = mapValue.(string)
			case "args":
				args, err := handleBuildArgs(mapValue)
				if err != nil {
					return err
				}
				b.Args = args
			default:
				// Ignore unknown keys
				continue
			}
		}
	default:
		return fmt.Errorf("Failed to unmarshal Build: %#v", value)
	}
	return nil
}

func handleBuildArgs(value interface{}) (map[string]string, error) {
	var args map[string]string
	switch v := value.(type) {
	case map[interface{}]interface{}:
		return handleBuildArgMap(v)
	case []interface{}:
		args = map[string]string{}
		for _, elt := range v {
			uniqArgs, err := handleBuildArgs(elt)
			if err != nil {
				return args, nil
			}
			for mapKey, mapValue := range uniqArgs {
				args[mapKey] = mapValue
			}
		}
		return args, nil
	default:
		return args, fmt.Errorf("Failed to unmarshal Build args: %#v", value)
	}
}

func handleBuildArgMap(m map[interface{}]interface{}) (map[string]string, error) {
	args := map[string]string{}
	for mapKey, mapValue := range m {
		var argValue string
		name, ok := mapKey.(string)
		if !ok {
			return args, fmt.Errorf("Cannot unmarshal '%v' to type %T into a string value", name, name)
		}
		switch a := mapValue.(type) {
		case string:
			argValue = a
		case int64:
			argValue = strconv.Itoa(int(a))
		default:
			return args, fmt.Errorf("Cannot unmarshal '%v' to type %T into a string value", mapValue, mapValue)
		}
		args[name] = argValue
	}
	return args, nil
}
