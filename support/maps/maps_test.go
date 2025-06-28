package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	mp := map[string]any{
		"name": "Desk",
	}
	Add(mp, "price", 100)
	assert.Equal(t, map[string]any{
		"name":  "Desk",
		"price": 100,
	}, mp)

	Add(mp, "price", 200)
	assert.Equal(t, 100, Get(mp, "price"))

	sMp := map[string]string{}
	Add(sMp, "surname", "Beniwal")
	assert.Equal(t, "Beniwal", Get(sMp, "surname"))
}

func TestExists(t *testing.T) {
	mp := map[string]any{
		"foo": "bar",
	}
	assert.True(t, Exists(mp, "foo"))
	assert.False(t, Exists(mp, "bar"))
	assert.False(t, Exists(mp, "foo.bar"))
}

func TestForget(t *testing.T) {
	mp := map[string]string{
		"foo": "bar",
		"baz": "qux",
	}
	Forget(mp)
	assert.Equal(t, map[string]string{
		"foo": "bar",
		"baz": "qux",
	}, mp)

	Forget(mp, "foo")
	assert.Equal(t, map[string]string{
		"baz": "qux",
	}, mp)

	mp = map[string]string{
		"foo": "bar",
		"baz": "qux",
	}
	Forget(mp, "baz", "foo")
	assert.Equal(t, map[string]string{}, mp)

	aMp := map[string]any{
		"developers": []map[string]string{
			{
				"name": "Bowen",
			},
			{
				"name": "Krishan",
			},
		},
	}
	Forget(aMp, "developers")
	assert.Equal(t, map[string]any{}, aMp)

	// Test nil value
	mp = map[string]string{
		"foo": "bar",
		"baz": "qux",
	}
	Forget(mp, "bar")
	assert.Equal(t, map[string]string{
		"foo": "bar",
		"baz": "qux",
	}, mp)

	// Test generic type
	gMp := map[int]string{
		1: "one",
		2: "two",
	}
	Forget(gMp, 1, 3)
	assert.Equal(t, map[int]string{
		2: "two",
	}, gMp)
}

func TestFromStruct(t *testing.T) {
	type One struct {
		Name string
		Age  int
	}
	type Two struct {
		Height int
	}
	type Three struct {
		Two
		One  One
		Name string
		age  int
	}
	data := Three{
		Name: "Three",
		Two: Two{
			Height: 1,
		},
		One: One{
			Name: "One",
			Age:  18,
		},
		age: 1,
	}

	res := FromStruct(data)

	assert.Equal(t, "Three", res["Name"])
	assert.Equal(t, 1, res["Height"])

	one, ok := res["One"].(map[string]any)

	assert.True(t, ok)
	assert.Equal(t, "One", one["Name"])
	assert.Equal(t, 18, one["Age"])
}

func TestGet(t *testing.T) {
	mp := map[string]any{
		"name": "Krishan",
		"age":  21,
		"languages": []string{
			"Golang",
			"PHP",
		},
	}
	assert.Equal(t, "Krishan", Get(mp, "name"))
	assert.Equal(t, 21, Get(mp, "age"))
	assert.Equal(t, []string{"Golang", "PHP"}, Get(mp, "languages"))

	// Test nil value
	mp = map[string]any{
		"foo": nil,
		"bar": "baz",
	}
	assert.Nil(t, Get(mp, "foo", "default"))
	assert.Equal(t, "baz", Get(mp, "bar"))
	// Test missing
	assert.Nil(t, Get(mp, "baz"))

	// Test return default value
	mp = map[string]any{
		"names": []string{
			"Krishan",
			"Bowen",
		},
	}
	assert.Equal(t, "name", Get(mp, "developers", "name"))

	mp1 := map[string]int{
		"foo": 1,
		"bar": 2,
		"baz": 3,
	}
	assert.Equal(t, 1, Get(mp1, "foo"))
	assert.Equal(t, 0, Get(mp1, "qux"))
	assert.Equal(t, 10, Get(mp1, "qux", 10))
}

func TestHas(t *testing.T) {
	mp := map[string]any{
		"framework": map[string]any{
			"name": "Goravel",
			"lang": "Golang",
			"dev": map[string]any{
				"name": "Bowen",
			},
		},
		"developers": []map[string]any{
			{
				"name": "Krishan",
				"lang": "Golang",
			},
			{
				"name": "Bowen",
				"lang": []string{
					"Golang",
					"PHP",
				},
			},
		},
		"foo": nil,
		"bar": map[string]any{
			"baz": nil,
		},
	}
	assert.True(t, Has(mp, "developers"))

	assert.False(t, Has(mp, "developers", "qux"))

	assert.False(t, Has(mp, "qux"))

	assert.True(t, Has(mp, "foo"))

	assert.True(t, Has(mp, "developers", "foo", "framework"))

	assert.True(t, Has(map[string]any{
		"": "some",
	}, ""))

	assert.False(t, Has(map[string]any{}, ""))

	// Test Generic type
	gMp := map[int]string{
		1: "one",
		2: "two",
	}
	assert.True(t, Has(gMp, 1))
	assert.False(t, Has(gMp, 3))
	assert.False(t, Has(gMp, 1, 3))
}

func TestHasAny(t *testing.T) {
	mp := map[string]any{
		"name": "Krishan",
		"age":  "",
		"city": nil,
	}

	assert.True(t, HasAny(mp, "name"))
	assert.True(t, HasAny(mp, "age"))
	assert.True(t, HasAny(mp, "city"))
	assert.False(t, HasAny(mp, "foo"))
	assert.True(t, HasAny(mp, "name", "email"))
	assert.True(t, HasAny(mp, "email", "name"))

	mp = map[string]any{
		"name":  "Krishan",
		"email": "foo",
	}
	assert.True(t, HasAny(mp, "name", "email"))
	assert.False(t, HasAny(mp, "surname", "password"))

	iMp := map[int]string{
		1: "Krishan",
		2: "Bowen",
	}
	assert.True(t, HasAny(iMp, 1))
	assert.False(t, HasAny(iMp, 3))
	assert.True(t, HasAny(iMp, 1, 3))
}

func TestKeys(t *testing.T) {
	// Test string keys
	strMap := map[string]any{
		"name": "Krishan",
		"age":  21,
		"city": "Chandigarh",
	}
	keys := Keys(strMap)
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "name")
	assert.Contains(t, keys, "age")
	assert.Contains(t, keys, "city")

	// Test int keys
	intMap := map[int]string{
		1: "one",
		2: "two",
		3: "three",
	}
	intKeys := Keys(intMap)
	assert.Len(t, intKeys, 3)
	assert.Contains(t, intKeys, 1)
	assert.Contains(t, intKeys, 2)
	assert.Contains(t, intKeys, 3)

	// Test empty map
	emptyMap := map[string]any{}
	emptyKeys := Keys(emptyMap)
	assert.Len(t, emptyKeys, 0)
	assert.Empty(t, emptyKeys)

	// Test map with single key
	singleMap := map[string]int{
		"single": 1,
	}
	singleKeys := Keys(singleMap)
	assert.Len(t, singleKeys, 1)
	assert.Contains(t, singleKeys, "single")

	// Test map with nil values
	nilMap := map[string]any{
		"nil1": nil,
		"nil2": nil,
		"val":  "value",
	}
	nilKeys := Keys(nilMap)
	assert.Len(t, nilKeys, 3)
	assert.Contains(t, nilKeys, "nil1")
	assert.Contains(t, nilKeys, "nil2")
	assert.Contains(t, nilKeys, "val")

	// Test map with complex values
	complexMap := map[string]any{
		"slice": []string{"a", "b", "c"},
		"map":   map[string]int{"x": 1, "y": 2},
		"int":   42,
	}
	complexKeys := Keys(complexMap)
	assert.Len(t, complexKeys, 3)
	assert.Contains(t, complexKeys, "slice")
	assert.Contains(t, complexKeys, "map")
	assert.Contains(t, complexKeys, "int")

	// Test map with different key types
	boolMap := map[bool]string{
		true:  "true",
		false: "false",
	}
	boolKeys := Keys(boolMap)
	assert.Len(t, boolKeys, 2)
	assert.Contains(t, boolKeys, true)
	assert.Contains(t, boolKeys, false)
}

func TestOnly(t *testing.T) {
	mp := map[string]any{
		"name": "Krishan",
		"age":  21,
		"foo":  "bar",
	}
	assert.Equal(t, map[string]any{
		"name": "Krishan",
		"age":  21,
	}, Only(mp, "name", "age"))

	// empty
	assert.Equal(t, map[string]any{}, Only(mp))

	// not found
	assert.Equal(t, map[string]any{}, Only(mp, "notfound"))
}

func TestPull(t *testing.T) {
	mp := map[string]any{
		"name": "Krishan",
		"age":  21,
	}
	assert.Equal(t, "Krishan", Pull(mp, "name"))
	assert.Equal(t, map[string]any{"age": 21}, mp)

	// work with slices
	mp = map[string]any{
		"names": []string{"Bowen", "Krishan"},
	}
	assert.Equal(t, []string{"Bowen", "Krishan"}, Pull(mp, "names"))
	assert.Equal(t, map[string]any{}, mp)

	// default value
	mp = map[string]any{
		"name": "Krishan",
	}
	assert.Equal(t, "default", Pull(mp, "age", "default"))
	assert.Equal(t, map[string]any{"name": "Krishan"}, mp)

	// Test generic type
	gMp := map[int]string{
		1: "one",
		2: "two",
	}
	assert.Equal(t, "one", Pull(gMp, 1))
	assert.Equal(t, map[int]string{2: "two"}, gMp)
	assert.Equal(t, "", Pull(gMp, 3))
	assert.Equal(t, "default", Pull(gMp, 3, "default"))

	mp1 := map[string]int{}
	assert.Equal(t, 0, Pull(mp1, "foo"))
	assert.Equal(t, 10, Pull(mp1, "foo", 10))
}

func TestSet(t *testing.T) {
	mp := map[string]any{
		"name": "Krishan",
		"age":  21,
	}
	Set(mp, "name", "Bowen")
	assert.Equal(t, map[string]any{
		"name": "Bowen",
		"age":  21,
	}, mp)

	Set(mp, "city", "Chandigarh")
	assert.Equal(t, map[string]any{
		"name": "Bowen",
		"age":  21,
		"city": "Chandigarh",
	}, mp)

	// Test nil value
	mp = map[string]any{
		"foo": nil,
		"bar": "baz",
	}
	Set(mp, "foo", "bar")
	assert.Equal(t, map[string]any{
		"foo": "bar",
		"bar": "baz",
	}, mp)

	// Test generic type
	gMp := map[int]string{
		1: "one",
		2: "two",
	}
	Set(gMp, 1, "1")
	assert.Equal(t, map[int]string{
		1: "1",
		2: "two",
	}, gMp)
}

func TestWhere(t *testing.T) {
	mp := map[string]any{
		"name": "Krishan",
		"age":  21,
		"city": "Chandigarh",
	}
	assert.Equal(t, map[string]any{
		"name": "Krishan",
		"age":  21,
	}, Where(mp, func(key string, value any) bool {
		return key != "city"
	}))

	// empty
	assert.Equal(t, map[string]any{}, Where(mp, func(key string, value any) bool {
		return false
	}))

	iMp := map[string]int{
		"foo": 1,
		"bar": 2,
		"baz": 3,
	}
	assert.Equal(t, map[string]int{
		"foo": 1,
	}, Where(iMp, func(key string, value int) bool {
		return value < 2
	}))
}
