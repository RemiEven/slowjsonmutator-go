package slowjsonmutator

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestModify(t *testing.T) {
	tests := map[string]struct {
		input          string
		modifications  []JsonModification
		expectedOutput string
		expectedError  error
	}{
		"identity": {
			input:          `{}`,
			expectedOutput: `{}`,
		},
		"remove one first level string attribute": {
			input: `{"name": "Perceval"}`,
			modifications: []JsonModification{
				Remove("name"),
			},
			expectedOutput: `{}`,
		},
		"remove one first level object attribute": {
			input: `{
				"name": "Perceval",
				"manager": {
					"name": "Arthur"
				}
			}`,
			modifications: []JsonModification{
				Remove("manager"),
			},
			expectedOutput: `{"name": "Perceval"}`,
		},
		"remove several first level attributes": {
			input: `{
				"name": "Perceval",
				"aka": "Provençal le Gaulois",
				"questsAchieved": 0,
				"title": "Knight"
			}`,
			modifications: []JsonModification{
				Remove("name"),
				Remove("aka"),
				Remove("questsAchieved"),
			},
			expectedOutput: `{"title": "Knight"}`,
		},
		"add a first level string attribute": {
			input: `{"name": "Perceval"}`,
			modifications: []JsonModification{
				Set("surname", "de Galles"),
			},
			expectedOutput: `{
				"name": "Perceval",
				"surname": "de Galles"
			}`,
		},
		"add a first level number attribute": {
			input: `{"name": "Perceval"}`,
			modifications: []JsonModification{
				Set("questsAchieved", 0),
			},
			expectedOutput: `{
				"name": "Perceval",
				"questsAchieved": 0
			}`,
		},
		"add a first level object attribute": {
			input: `{"name": "Perceval"}`,
			modifications: []JsonModification{
				Set("manager", json.RawMessage(`{
					"name": "Arthur"
				}`)),
			},
			expectedOutput: `{
				"name": "Perceval",
				"manager": {
					"name": "Arthur"
				}
			}`,
		},
		"remove a nested attribute": {
			input: `{
				"name": "Perceval",
				"manager": {
					"name": "Arthur"
				}
			}`,
			modifications: []JsonModification{
				Remove("manager.name"),
			},
			expectedOutput: `{
				"name": "Perceval",
				"manager": {}
			}`,
		},
		"remove a deeply nested attribute": {
			input: `{
				"name": "Perceval",
				"manager": {
					"name": "Arthur",
					"home": {
						"type": "Castle"
					}
				}
			}`,
			modifications: []JsonModification{
				Remove("manager.home.type"),
			},
			expectedOutput: `{
				"name": "Perceval",
				"manager": {
					"name": "Arthur",
					"home": {}
				}
			}`,
		},
		"remove an attribute that is already not there": {
			input: `{"name": "Perceval"}`,
			modifications: []JsonModification{
				Remove("manager.home.type"),
			},
			expectedOutput: `{"name": "Perceval"}`,
		},
		"remove a deeply nested attribute which a higher-level segment that is null": {
			input: `{
				"name": "Perceval",
				"manager": null
			}`,
			modifications: []JsonModification{
				Remove("manager.home.type"),
			},
			expectedOutput: `{
				"name": "Perceval",
				"manager": null
			}`,
		},
		"remove an attribute with an invalid path (middle segment is not an object)": {
			input: `{"name": "Perceval"}`,
			modifications: []JsonModification{
				Remove("name.complete"),
			},
			expectedError: errors.New("invalid path"),
		},
		"input is not valid json": {
			input:         `{`,
			expectedError: errors.New(`unexpected end of JSON input`),
		},
		"set a cyclic structure so that marshalling fails": {
			input: `{}`,
			modifications: []JsonModification{
				Set("attr", func() interface{} {
					type dummyStructWithCycle struct {
						Attr *dummyStructWithCycle
					}
					a, b := dummyStructWithCycle{}, dummyStructWithCycle{}
					a.Attr = &b
					b.Attr = &a
					return a
				}()),
			},
			expectedError: errors.New(`json: unsupported value: encountered a cycle via *slowjsonmutator.dummyStructWithCycle`),
		},
		"remove one nested attribute inside an array": {
			input: `{
				"knights": [
					{
						"name": "Perceval",
						"aka": "Provençal le Gaulois"
					}
				]
			}`,
			modifications: []JsonModification{
				Remove("knights[0].name"),
			},
			expectedOutput: `{
				"knights": [
					{
						"aka": "Provençal le Gaulois"
					}
				]
			}`,
		},
		"remove element from array at an index higher than the array length": {
			input: `{
				"knights": [
					{ "name": "Perceval" }
				]
			}`,
			modifications: []JsonModification{
				Remove("knights[2]"),
			},
			expectedOutput: `{
				"knights": [
					{ "name": "Perceval" }
				]
			}`,
		},
		"remove element from array": {
			input: `{
				"knights": [
					{ "name": "Lancelot" },
					{ "name": "Perceval" },
					{ "name": "Karadoc" }
				]
			}`,
			modifications: []JsonModification{
				Remove("knights[1]"),
			},
			expectedOutput: `{
				"knights": [
					{ "name": "Lancelot" },
					{ "name": "Karadoc" }
				]
			}`,
		},
		"remove last element from array": {
			input: `{
				"knights": [
					{ "name": "Lancelot" },
					{ "name": "Perceval" },
					{ "name": "Karadoc" }
				]
			}`,
			modifications: []JsonModification{
				Remove("knights[2]"),
			},
			expectedOutput: `{
				"knights": [
					{ "name": "Lancelot" },
					{ "name": "Perceval" }
				]
			}`,
		},
		"try to remove element from array with an index too high": {
			input: `{
				"knights": [
					{ "name": "Lancelot" },
					{ "name": "Perceval" }
				]
			}`,
			modifications: []JsonModification{
				Remove("knights[2]"),
			},
			expectedOutput: `{
				"knights": [
					{ "name": "Lancelot" },
					{ "name": "Perceval" }
				]
			}`,
		},
		"remove element from top-level array": {
			input: `[
				{ "name": "Lancelot" },
				{ "name": "Perceval" },
				{ "name": "Karadoc" }
			]`,
			modifications: []JsonModification{
				Remove("[1]"),
			},
			expectedOutput: `[
				{ "name": "Lancelot" },
				{ "name": "Karadoc" }
			]`,
		},
		"wrongfully address content of json array by attribute when removing": {
			input: `[]`,
			modifications: []JsonModification{
				Remove("knight"),
			},
			expectedError: errors.New("cannot address content of JSON array by attribute"),
		},
		"wrongfully address content of json object by index when removing": {
			input: `{}`,
			modifications: []JsonModification{
				Remove("[0]"),
			},
			expectedError: errors.New("cannot address content of JSON object by index"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := Modify(test.input, test.modifications...)
			if !ErrorEqual(err, test.expectedError) {
				t.Errorf("unexpected error: wanted [%v], got [%v]", test.expectedError, err)
				return
			}
			if output == test.expectedOutput {
				return
			}
			if message, ok := JSONEqual(output, test.expectedOutput); !ok {
				t.Error("unexpected output: " + message)
			}
		})
	}
}

func TestParseJsonPath(t *testing.T) {
	tests := map[string]struct {
		inputPath        string
		expectedSegments []jsonPathSegment
		expectedError    error
	}{
		"invalid path (empty)": {
			inputPath:     "",
			expectedError: errors.New(`cannot parse json path [""], it doesn't seem valid`),
		},
		"invalid path (missing closing bracket)": {
			inputPath:     `[`,
			expectedError: errors.New(`cannot parse json path ["["], it doesn't seem valid`),
		},
		"invalid path (non digit character inside bracked)": {
			inputPath:     `[a]`,
			expectedError: errors.New(`cannot parse json path ["[a]"], it doesn't seem valid`),
		},
		"invalid path (too many closing brackets)": {
			inputPath:     `[]]`,
			expectedError: errors.New(`cannot parse json path ["[]]"], it doesn't seem valid`),
		},
		"invalid path (just a dot)": {
			inputPath:     `.`,
			expectedError: errors.New(`cannot parse json path ["."], it doesn't seem valid`),
		},
		"invalid path (two dots together)": {
			inputPath:     `a..b`,
			expectedError: errors.New(`cannot parse json path ["a..b"], it doesn't seem valid`),
		},
		"invalid path (empty brackets)": {
			inputPath:     `[]`,
			expectedError: errors.New(`cannot parse json path ["[]"], it doesn't seem valid`),
		},
		"invalid path (trailing dot)": {
			inputPath:     `rrrr.`,
			expectedError: errors.New(`cannot parse json path ["rrrr."], it doesn't seem valid`),
		},
		"invalid path (dot before bracket)": {
			inputPath:     `a.[1]`,
			expectedError: errors.New(`cannot parse json path ["a.[1]"], it doesn't seem valid`),
		},
		"invalid path (missing opening bracket)": {
			inputPath:     `3]`,
			expectedError: errors.New(`cannot parse json path ["3]"], it doesn't seem valid`),
		},
		"single segment, attribute": {
			inputPath: `name`,
			expectedSegments: []jsonPathSegment{
				stringSegment("name"),
			},
		},
		"multiple segments, all attributes": {
			inputPath: `manager.home.type`,
			expectedSegments: []jsonPathSegment{
				stringSegment("manager"),
				stringSegment("home"),
				stringSegment("type"),
			},
		},
		"single segment, index": {
			inputPath: `[1]`,
			expectedSegments: []jsonPathSegment{
				indexSegment(1),
			},
		},
		"two segments, both indices": {
			inputPath: `[34][9]`,
			expectedSegments: []jsonPathSegment{
				indexSegment(34),
				indexSegment(9),
			},
		},
		"several segments, mixed attributes and indices": {
			inputPath: `knights[0].quests[2]`,
			expectedSegments: []jsonPathSegment{
				stringSegment("knights"),
				indexSegment(0),
				stringSegment("quests"),
				indexSegment(2),
			},
		},
		"several segments, mixed attributes and indices (2)": {
			inputPath: `knights[0].name`,
			expectedSegments: []jsonPathSegment{
				stringSegment("knights"),
				indexSegment(0),
				stringSegment("name"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			segments, err := parseJsonPath(test.inputPath)
			if !ErrorEqual(err, test.expectedError) {
				t.Errorf("unexpected error: wanted [%v], got [%v]", test.expectedError, err)
			}
			if diff := DeepEqual(segments, test.expectedSegments); diff != "" {
				t.Errorf("unexpected segments: " + diff)
			}
		})
	}
}
