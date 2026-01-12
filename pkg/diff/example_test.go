package diff

func ExampleCompare() {
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

	result, err := Compare(left, right)
	if err != nil {
		panic(err)
	}

	_ = result
}
