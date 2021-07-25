package slowjsonmutator

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
)

// JSONEqual checks whether two strings are valid JSON and that their contents are semantically equivalent
func JSONEqual(actual, expected string) (string, bool) {
	var parsedActual, parsedExpected interface{}

	if err := json.Unmarshal([]byte(actual), &parsedActual); err != nil {
		return fmt.Sprintf("failed to parse actual json: %v", err), false
	}
	if err := json.Unmarshal([]byte(expected), &parsedExpected); err != nil {
		return fmt.Sprintf("failed to parse expected json: %v", err), false
	}

	if diff := DeepEqual(parsedActual, parsedExpected); diff != "" {
		return "JSON contents do not match: " + diff, false
	}

	return "", true
}

func TestJsonEqual(t *testing.T) {
	tests := map[string]struct {
		actualJSON          string
		expectedJSON        string
		shouldReturnMessage string
		shouldReturnMatch   bool
	}{
		"invalid json in actual": {
			actualJSON:          `invalid`,
			expectedJSON:        `null`,
			shouldReturnMessage: "failed to parse actual json: invalid character 'i' looking for beginning of value",
		},
		"invalid json in expected": {
			actualJSON:          `null`,
			expectedJSON:        `invalid`,
			shouldReturnMessage: "failed to parse expected json: invalid character 'i' looking for beginning of value",
		},
		"equal primitive value": {
			actualJSON:        `null`,
			expectedJSON:      `null`,
			shouldReturnMatch: true,
		},
		"different primitive values": {
			actualJSON:          `2`,
			expectedJSON:        `null`,
			shouldReturnMessage: `JSON contents do not match: difference(s) found between actual and expected: 2 != <nil pointer>`,
		},
		"equal arrays": {
			actualJSON:        `[3, 4]`,
			expectedJSON:      `[3, 4]`,
			shouldReturnMatch: true,
		},
		"arrays with different orders": {
			actualJSON:          `[3, 4]`,
			expectedJSON:        `[4, 3]`,
			shouldReturnMessage: "JSON contents do not match: difference(s) found between actual and expected:\n- slice[0]: 3 != 4\n- slice[1]: 4 != 3",
		},
		"equal objects": {
			actualJSON:        `{"a": 3, "b": 4}`,
			expectedJSON:      `{"a": 3, "b": 4}`,
			shouldReturnMatch: true,
		},
		"different objects": {
			actualJSON:          `{"a": 3, "b": 4}`,
			expectedJSON:        `{"a": 4, "b": 4}`,
			shouldReturnMessage: `JSON contents do not match: difference(s) found between actual and expected: map[a]: 3 != 4`,
		},
		"equal complex objects": {
			actualJSON:        `{"a": [3, 4], "b": {"c": 5, "d": "some message"}}`,
			expectedJSON:      `{"a": [3, 4], "b": {"c": 5, "d": "some message"}}`,
			shouldReturnMatch: true,
		},
		"different complex objects": {
			actualJSON:          `{"a": [3, 4], "b": {"c": 5, "d": "some message"}}`,
			expectedJSON:        `{"a": [3, 4], "b": {"c": 5, "d": "another message"}}`,
			shouldReturnMessage: `JSON contents do not match: difference(s) found between actual and expected: map[b].map[d]: some message != another message`,
		},
		"different object attributes order": {
			actualJSON:        `{"a": 3, "b": 4}`,
			expectedJSON:      `{"b": 4, "a": 3}`,
			shouldReturnMatch: true,
		},
		"different whitespacing": {
			actualJSON: `{
				"a": 3,
				"b": 4
			}`,
			expectedJSON:      `{"b": 4, "a": 3}`,
			shouldReturnMatch: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			message, match := JSONEqual(test.actualJSON, test.expectedJSON)
			if message != test.shouldReturnMessage {
				t.Errorf("got message [%q], wanted [%q]", message, test.shouldReturnMessage)
			}
			if match != test.shouldReturnMatch {
				t.Errorf("got match [%v], wanted [%v]", match, test.shouldReturnMatch)
			}
		})
	}
}

// DeepEqual compares a and b and returns a formatted string explaining the differences if there are any
func DeepEqual(actual, expected interface{}) string {
	diff := deep.Equal(actual, expected)
	diffHeader := "difference(s) found between actual and expected:"
	switch len(diff) {
	case 0:
		return ""
	case 1:
		return diffHeader + " " + diff[0]
	default:
		return diffHeader + "\n- " + strings.Join(diff, "\n- ")
	}
}

type exampleStruct struct {
	FirstName, LastName string
	unexportedName      string
}

func TestDeepEqual(t *testing.T) {
	tests := map[string]struct {
		a, b           interface{}
		expectedResult string
	}{
		"Equal objects": {
			a: exampleStruct{
				FirstName:      "FirstName",
				LastName:       "LastName",
				unexportedName: "Name",
			},
			b: exampleStruct{
				FirstName:      "FirstName",
				LastName:       "LastName",
				unexportedName: "Name",
			},
			expectedResult: ``,
		},
		"Different object types": {
			a: exampleStruct{
				FirstName:      "FirstName",
				LastName:       "LastName",
				unexportedName: "Name",
			},
			b:              time.Time{},
			expectedResult: `difference(s) found between actual and expected: slowjsonmutator.exampleStruct != time.Time`,
		},
		"Objects with one difference": {
			a: exampleStruct{
				FirstName:      "FirstName",
				LastName:       "LastName",
				unexportedName: "Name",
			},
			b: exampleStruct{
				FirstName:      "Other FirstName",
				LastName:       "LastName",
				unexportedName: "Name",
			},
			expectedResult: `difference(s) found between actual and expected: FirstName: FirstName != Other FirstName`,
		},
		"Objects with several differences": {
			a: exampleStruct{
				FirstName:      "FirstName",
				LastName:       "LastName",
				unexportedName: "Name",
			},
			b: exampleStruct{
				FirstName:      "Other FirstName",
				LastName:       "Other LastName",
				unexportedName: "Name",
			},
			expectedResult: `difference(s) found between actual and expected:
- FirstName: FirstName != Other FirstName
- LastName: LastName != Other LastName`,
		},
		"Objects with difference but it is in an unexported field": {
			a: exampleStruct{
				FirstName:      "FirstName",
				LastName:       "LastName",
				unexportedName: "Name",
			},
			b: exampleStruct{
				FirstName:      "FirstName",
				LastName:       "LastName",
				unexportedName: "Other Secret Name",
			},
			expectedResult: ``,
		},
		"Nil with empty slice": {
			a:              nil,
			b:              []string{},
			expectedResult: `difference(s) found between actual and expected: <nil pointer> != []`,
		},
		"Int with solid float": {
			a:              int(1),
			b:              float64(1.0),
			expectedResult: `difference(s) found between actual and expected: int != float64`,
		},
		"Slices containing different types": {
			a:              []string{},
			b:              []interface{}{},
			expectedResult: `difference(s) found between actual and expected: []string != []interface {}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualResult := DeepEqual(test.a, test.b)

			if actualResult != test.expectedResult {
				t.Errorf("got result [%v], wanted [%v]", actualResult, test.expectedResult)
			}
		})
	}
}

// ErrorEqual checks whether two errors have the same message (or are both nil)
func ErrorEqual(actual, expected error) bool {
	if actual == nil && expected == nil {
		return true
	}
	if (actual != nil && expected == nil) || (actual == nil && expected != nil) {
		return false
	}
	if actual.Error() != expected.Error() {
		return false
	}
	return true
}

func TestErrorEqual(t *testing.T) {
	testCases := map[string]struct {
		actualErr    error
		expectedErr  error
		shouldReturn bool
	}{
		"Both nil": {
			actualErr:    nil,
			expectedErr:  nil,
			shouldReturn: true,
		},
		"Both same messages": {
			actualErr:    errors.New("message"),
			expectedErr:  errors.New("message"),
			shouldReturn: true,
		},
		"Actual is nil": {
			actualErr:    nil,
			expectedErr:  errors.New("message"),
			shouldReturn: false,
		},
		"Expected is nil": {
			actualErr:    errors.New("message"),
			expectedErr:  nil,
			shouldReturn: false,
		},
		"Messages differ": {
			actualErr:    errors.New("a message"),
			expectedErr:  errors.New("another message"),
			shouldReturn: false,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			actualReturn := ErrorEqual(test.actualErr, test.expectedErr)
			if actualReturn != test.shouldReturn {
				t.Errorf("Got result [%v], expected [%v]", actualReturn, test.shouldReturn)
			}
		})
	}
}

func (s jsonPathSegment) Equal(other jsonPathSegment) bool {
	if s.attribute != nil && other.attribute != nil {
		return *s.attribute == *other.attribute
	}
	if s.index != nil && other.index != nil {
		return *s.index == *other.index
	}
	return s.attribute == other.attribute && s.index == other.index
}
