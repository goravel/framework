package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	obj := map[string]any{
		"name": "Desk",
	}
	err := Add(&obj, "price", 100)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"name":  "Desk",
		"price": 100,
	}, obj)

	err = Add(&obj, "price", 200)
	assert.Nil(t, err)
	assert.Equal(t, 100, Get(obj, "price"))

	obj = map[string]any{}
	err = Add(&obj, "surname", "Beniwal")
	assert.Nil(t, err)
	assert.Equal(t, "Beniwal", Get(obj, "surname"))

	obj = map[string]any{}
	err = Add(&obj, "developer.name", "Krishan")
	assert.Nil(t, err)
	assert.Equal(t, "Krishan", Get(obj, "developer.name"))

	obj = map[string]any{
		"developer": map[string]any{
			"name": "Krishan",
			"lang": []string{},
		},
	}
	err = Add(&obj, "developer.lang.1", "Golang")
	assert.Nil(t, err)
	assert.Equal(t, "", Get(obj, "developer.lang.0"))
	assert.Equal(t, "Golang", Get(obj, "developer.lang.1"))

	obj = map[string]any{}
	err = Add(&obj, "foo", map[string]any{
		"bar": "baz",
	})
	assert.Nil(t, err)
	assert.Equal(t, "baz", Get(obj, "foo.bar"))
}

func TestDot(t *testing.T) {
	obj := Dot(map[string]any{
		"foo": map[string]any{
			"bar": "baz",
		},
	})
	assert.Equal(t, map[string]any{
		"foo.bar": "baz",
	}, obj)

	obj = Dot(map[string]any{
		"foo": map[int]int{
			10: 100,
		},
	})
	assert.Equal(t, map[string]any{
		"foo.10": 100,
	}, obj)

	obj = Dot(map[string]any{})
	assert.Equal(t, map[string]any{}, obj)

	obj = Dot(map[string]any{
		"foo": []string{},
	})
	assert.Equal(t, map[string]any{}, obj)

	obj = Dot(map[string]any{
		"user": map[string]any{
			"name": "Krishan",
			"age":  21,
			"languages": []string{
				"Golang",
				"PHP",
			},
		},
	})
	assert.Equal(t, map[string]any{
		"user.name":         "Krishan",
		"user.age":          21,
		"user.languages[0]": "Golang",
		"user.languages[1]": "PHP",
	}, obj)

	obj = Dot(map[string]any{
		"user": map[string]any{
			"name": "Krishan",
		},
		"empty_slice": []string{},
		"key":         "value",
		"zero":        0,
	})
	assert.Equal(t, map[string]any{
		"user.name": "Krishan",
		"key":       "value",
		"zero":      0,
	}, obj)
}

func TestExists(t *testing.T) {
	obj := map[string]any{
		"foo": "bar",
	}
	assert.True(t, Exists(obj, "foo"))
	assert.False(t, Exists(obj, "bar"))
	assert.False(t, Exists(obj, "foo.bar"))
}

func TestForget(t *testing.T) {
	//
}

func TestGet(t *testing.T) {
	obj := map[string]any{
		"products": map[string]any{
			"desk": map[string]int{
				"price": 100,
			},
		},
	}
	assert.Equal(t, map[string]int{"price": 100}, Get(obj, "products.desk"))

	// Test nil value
	obj = map[string]any{
		"foo": nil,
		"bar": map[string]any{
			"baz": nil,
		},
	}
	assert.Nil(t, Get(obj, "foo", "default"))
	assert.Nil(t, Get(obj, "bar.baz", "default"))
	// Test missing
	assert.Nil(t, Get(obj, "foo.bar"))

	// Test numeric keys
	obj = map[string]any{
		"developers": []map[string]any{
			{
				"name": "Bowen",
				"lang": []string{
					"Golang",
					"PHP",
				},
			},
			{
				"name": "Krishan",
				"lang": "Golang",
			},
		},
	}
	assert.Equal(t, "Krishan", Get(obj, "developers.1.name"))
	assert.Equal(t, "Bowen", Get(obj, "developers.0.name"))
	assert.Equal(t, "Golang", Get(obj, "developers.0.lang.0"))

	// Test return default value
	obj = map[string]any{
		"names": map[string]any{
			"developer": "Krishan",
		},
	}
	assert.Equal(t, "name", Get(obj, "names.designer", "name"))
}

func TestHas(t *testing.T) {
	obj := map[string]any{
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
	assert.True(t, Has(obj, "developers"))

	assert.True(t, Has(obj, "framework.lang"))

	assert.True(t, Has(obj, "framework.dev.name"))

	assert.False(t, Has(obj, "framework.foo"))

	assert.False(t, Has(obj, "framework.dev.foo"))

	assert.True(t, Has(obj, "foo"))

	assert.True(t, Has(obj, "bar.baz"))

	assert.False(t, Has(nil, ""))

	assert.True(t, Has(obj, "framework.name", "framework.dev.name"))

	assert.False(t, Has(obj, "framework.name", "framework.dev.foo"))

	assert.True(t, Has(obj, "developers.0.name"))

	assert.False(t, Has(obj, "product.developers.0.foo"))

	assert.True(t, Has(map[string]any{
		"": "some",
	}, ""))

	assert.False(t, Has(map[string]any{}, ""))
}

func TestHasAny(t *testing.T) {
	obj := map[string]any{
		"name": "Krishan",
		"age":  "",
		"city": nil,
	}

	assert.True(t, HasAny(obj, "name"))
	assert.True(t, HasAny(obj, "age"))
	assert.True(t, HasAny(obj, "city"))
	assert.False(t, HasAny(obj, "foo"))
	assert.True(t, HasAny(obj, "name", "email"))
	assert.True(t, HasAny(obj, "email", "name"))

	obj = map[string]any{
		"name":  "Krishan",
		"email": "foo",
	}
	assert.True(t, HasAny(obj, "name", "email"))
	assert.False(t, HasAny(obj, "surname", "password"))

	obj = map[string]any{
		"foo": map[string]any{
			"bar": nil,
			"baz": "",
		},
	}
	assert.True(t, HasAny(obj, "foo.bar"))
	assert.True(t, HasAny(obj, "foo.baz"))
	assert.False(t, HasAny(obj, "foo.bax"))
	assert.True(t, HasAny(obj, "foo.bax", "foo.baz"))
}

func TestOnly(t *testing.T) {
	obj := map[string]any{
		"name": "Krishan",
		"age":  21,
		"foo":  "bar",
	}
	assert.Equal(t, map[string]any{
		"name": "Krishan",
		"age":  21,
	}, Only(obj, "name", "age"))

	// empty
	assert.Equal(t, map[string]any{}, Only(obj))

	// not found
	assert.Equal(t, map[string]any{}, Only(obj, "notfound"))
}

func TestPull(t *testing.T) {
	//
}

func TestSet(t *testing.T) {
	obj := map[string]any{
		"products": map[string]any{
			"desk": map[string]int{
				"price": 100,
			},
		},
	}
	err := Set(&obj, "products.desk.price", 200)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"products": map[string]any{
			"desk": map[string]int{
				"price": 200,
			},
		},
	}, obj)

	// key does not exist
	obj = map[string]any{
		"products": map[string]any{},
	}
	err = Set(&obj, "products.desk.price", 200)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"products": map[string]any{
			"desk": map[string]any{
				"price": 200,
			},
		},
	}, obj)
}
