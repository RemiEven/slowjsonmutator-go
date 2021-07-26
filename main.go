package slowjsonmutator

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// JSONModification is a function that can modify parsed untyped json data
type JSONModification func(interface{}) (interface{}, error)

// Remove removes the element at the given path
func Remove(path string) JSONModification {
	return func(toModify interface{}) (interface{}, error) {
		pathSegments, err := parseJSONPath(path)
		if err != nil {
			return nil, err
		}
		return remove(toModify, pathSegments)
	}
}

type jsonPathSegment struct {
	attribute *string
	index     *int
}

func stringSegment(attribute string) jsonPathSegment {
	return jsonPathSegment{
		attribute: &attribute,
	}
}

func indexSegment(index int) jsonPathSegment {
	return jsonPathSegment{
		index: &index,
	}
}

func remove(toModify interface{}, parsedPath []jsonPathSegment) (interface{}, error) {
	switch toModify := toModify.(type) {
	case map[string]interface{}:
		if parsedPath[0].attribute == nil {
			return nil, errors.New("cannot address content of JSON object by index")
		}
		if len(parsedPath) == 1 {
			delete(toModify, *parsedPath[0].attribute)
			return toModify, nil
		}

		deeper, ok := toModify[*parsedPath[0].attribute]
		if !ok {
			return toModify, nil
		}

		modifiedDeeper, err := remove(deeper, parsedPath[1:])
		if err != nil {
			return nil, err
		}
		toModify[*parsedPath[0].attribute] = modifiedDeeper
		return toModify, nil
	case []interface{}:
		if parsedPath[0].index == nil {
			return nil, errors.New("cannot address content of JSON array by attribute")
		}
		index := *parsedPath[0].index
		if len(toModify) <= index {
			return toModify, nil
		}

		if len(parsedPath) == 1 {
			return removeFromSlice(toModify, index), nil
		}

		deeper := toModify[index]
		modifiedDeeper, err := remove(deeper, parsedPath[1:])
		if err != nil {
			return nil, err
		}
		toModify[index] = modifiedDeeper
		return toModify, nil
	case nil:
		return toModify, nil
	default:
		return nil, errors.New("invalid path")
	}
}

func removeFromSlice(slice []interface{}, index int) []interface{} {
	switch {
	case index < 0 || len(slice) <= index:
		return slice
	case index == len(slice)-1:
		return slice[:len(slice)-1]
	default:
		return append(slice[:index], slice[index+1:]...)
	}
}

// Set sets the element at the given path to value
func Set(path string, value interface{}) JSONModification {
	return func(toModify interface{}) (interface{}, error) {
		pathSegments, err := parseJSONPath(path)
		if err != nil {
			return nil, err
		}
		return set(toModify, pathSegments, value)
	}
}

func set(toModify interface{}, parsedPath []jsonPathSegment, value interface{}) (interface{}, error) {
	if len(parsedPath) == 0 {
		return value, nil
	}
	switch toModify := toModify.(type) {
	case map[string]interface{}:
		if parsedPath[0].attribute == nil {
			return nil, errors.New("cannot address content of JSON object by index")
		}

		deeper := toModify[*parsedPath[0].attribute]
		modifiedDeeper, err := set(deeper, parsedPath[1:], value)
		if err != nil {
			return nil, err
		}
		toModify[*parsedPath[0].attribute] = modifiedDeeper

		return toModify, nil
	case []interface{}:
		if parsedPath[0].index == nil {
			return nil, errors.New("cannot address content of JSON array by attribute")
		}
		index := *parsedPath[0].index
		if index < 0 || len(toModify) < index {
			return nil, errors.New("out of bounds insertion index")
		}

		var deeper interface{} = nil
		if index < len(toModify) {
			deeper = toModify[index]
		}

		if modifiedDeeper, err := set(deeper, parsedPath[1:], value); err != nil {
			return nil, err
		} else if index == len(toModify) {
			toModify = append(toModify, modifiedDeeper)
		} else {
			toModify[index] = modifiedDeeper
		}

		return toModify, nil
	case nil:
		var deeper interface{} = make(map[string]interface{}, 1)
		if parsedPath[0].attribute == nil {
			deeper = make([]interface{}, 0, 1)
		}
		return set(deeper, parsedPath, value)
	default:
		return nil, errors.New("invalid path")
	}
}

// Modify applies modifications to a json string
func Modify(input string, modifications ...JSONModification) (string, error) {
	var untypedParsed interface{}
	if err := json.Unmarshal([]byte(input), &untypedParsed); err != nil {
		return "", err
	}

	for _, modification := range modifications {
		var err error
		if untypedParsed, err = modification(untypedParsed); err != nil {
			return "", err
		}
	}

	result, err := json.Marshal(untypedParsed)
	if err != nil {
		return "", err
	}
	return string(result), err
}

var naiveJSONPathRegexp = regexp.MustCompile(`^(?:[a-zA-Z0-9_\-]+|\[[0-9]+\])(?:(?:\.[a-zA-Z0-9_\-]+)|\[[0-9]+\])*$`)

func parseJSONPath(path string) ([]jsonPathSegment, error) {
	if !naiveJSONPathRegexp.Match([]byte(path)) {
		return nil, fmt.Errorf("cannot parse json path [%q], it doesn't seem valid", path)
	}

	numberOfSegments := 1 + strings.Count(path, ".") + strings.Count(path, "[")
	pathSegments := make([]jsonPathSegment, 0, numberOfSegments)

	toParse := path

	for toParse != "" {
		var segment jsonPathSegment
		var err error
		segment, toParse, err = parseFirstSegment(toParse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse json path: %w", err)
		}
		pathSegments = append(pathSegments, segment)
	}

	return pathSegments, nil
}

func parseFirstSegment(path string) (jsonPathSegment, string, error) {
	if path[0] == '.' {
		path = path[1:]
	}
	if path[0] == '[' {
		nextClosingSquareBracketIndex := strings.Index(path, "]")
		index, err := strconv.Atoi(path[1:nextClosingSquareBracketIndex])
		if err != nil {
			return jsonPathSegment{}, "", fmt.Errorf("failed to parse index: %w", err)
		}
		return indexSegment(index), path[nextClosingSquareBracketIndex+1:], nil
	}

	nextDotIndex, nextSquareBracketIndex := strings.Index(path, "."), strings.Index(path, "[")
	if nextDotIndex == -1 && nextSquareBracketIndex == -1 {
		return stringSegment(path), "", nil
	}
	if nextDotIndex == -1 {
		return stringSegment(path[:nextSquareBracketIndex]), path[nextSquareBracketIndex:], nil
	}
	nextIndex := nextDotIndex
	if 0 < nextSquareBracketIndex && nextSquareBracketIndex < nextDotIndex {
		nextIndex = nextSquareBracketIndex
	}
	return stringSegment(path[:nextIndex]), path[nextIndex:], nil
}
