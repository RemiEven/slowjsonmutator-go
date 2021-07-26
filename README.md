# slowjsonmutator-go

[![Go Reference](https://pkg.go.dev/badge/github.com/remieven/slowjsonmutator-go.svg)](https://pkg.go.dev/github.com/remieven/slowjsonmutator-go)

This is a small library in Go that can change (possibly deeply nested) JSON data without needing you to write a go struct.
Under the hood it deals with `map[string]interface{}`, `[]interface{}` and type casting, so it is quite slow.
It is mainly intended to be used in tests and to help with reducing boilerplate.

## Install

`go get -u github.com/remieven/slowjsonmutator-go`

## Examples

### Remove several first level attributes

```go
import sjm "github.com/remieven/slowjsonmutator-go"

input := `{
    "name": "Perceval",
    "questsAchieved": 0,
    "title": "Knight"
}`
output, _ := sjm.Modify(input, sjm.Remove("name"), sjm.Remove("questsAchieved"))
fmt.Println(output)
// {"title":"Knight"}
```

### Add a deeply nested string attribute, with missing array and object

```go
import sjm "github.com/remieven/slowjsonmutator-go"

input := `{ "name": "Perceval" }`
output, _ := sjm.Modify(input, sjm.Set("manager.titles[0].fr", "Suzerain"))
fmt.Println(output)
// {"manager":{"titles":[{"fr":"Suzerain"}]},"name":"Perceval"}
```

## License

MIT licensed. See the LICENSE file for details.