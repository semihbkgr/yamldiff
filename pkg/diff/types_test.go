package diff

import "fmt"

func ExampleDocDiffs_Format() {
	left := []byte(`
name: Alice
city: New York
items:
    - foo
    - bar
`)

	right := []byte(`
name: Bob
age: 30
items:
    - foo
    - baz
`)

	diffs, err := Compare(left, right)
	if err != nil {
		panic(err)
	}

	s := diffs.Format(Plain)
	fmt.Println(s)

	// Output:
	// ~ .name: Alice → Bob
	// - .city: New York
	// + .age: 30
	// ~ .items[1]: bar → baz
}

func ExampleDocDiffs_Format_includeCounts() {
	left := []byte(`
name: Alice
city: New York
items:
    - foo
    - bar
`)

	right := []byte(`
name: Bob
age: 30
items:
    - foo
    - baz
`)

	diffs, err := Compare(left, right)
	if err != nil {
		panic(err)
	}

	s := diffs.Format(Plain, IncludeCounts)
	fmt.Println(s)

	// Output:
	// 1 added, 1 deleted, 2 modified
	// ~ .name: Alice → Bob
	// - .city: New York
	// + .age: 30
	// ~ .items[1]: bar → baz
}

func ExampleDocDiffs_Format_withMetadata() {
	left := []byte(`
name: Alice
city: New York
items:
    - foo
    - bar
`)

	right := []byte(`
name: Bob
age: 30
items:
    - foo
    - baz
`)

	diffs, err := Compare(left, right)
	if err != nil {
		panic(err)
	}

	s := diffs.Format(Plain, WithMetadata)
	fmt.Println(s)

	// Output:
	// ~ .name: [line:2 <String>] Alice → [line:2 <String>] Bob
	// - .city: [line:3 <String>] New York
	// + .age: [line:3 <Integer>] 30
	// ~ .items[1]: [line:6 <String>] bar → [line:6 <String>] baz
}

func ExampleDocDiffs_Format_pathsOnly() {
	left := []byte(`
name: Alice
city: New York
items:
    - foo
    - bar
`)

	right := []byte(`
name: Bob
age: 30
items:
    - foo
    - baz
`)

	diffs, err := Compare(left, right)
	if err != nil {
		panic(err)
	}

	s := diffs.Format(Plain, PathsOnly)
	fmt.Println(s)

	// Output:
	// 	~ .name
	// - .city
	// + .age
	// ~ .items[1]
}
