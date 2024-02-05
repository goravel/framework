package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	mp := map[string]any{
		"name": "Desk",
	}
	err := Add(&mp, "price", 100)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"name":  "Desk",
		"price": 100,
	}, mp)

	err = Add(&mp, "price", 200)
	assert.Nil(t, err)
	assert.Equal(t, 100, Get(mp, "price"))

	mp = map[string]any{}
	err = Add(&mp, "surname", "Beniwal")
	assert.Nil(t, err)
	assert.Equal(t, "Beniwal", Get(mp, "surname"))

	mp = map[string]any{}
	err = Add(&mp, "developer.name", "Krishan")
	assert.Nil(t, err)
	assert.Equal(t, "Krishan", Get(mp, "developer.name"))

	mp = map[string]any{
		"developer": map[string]any{
			"name": "Krishan",
			"lang": []string{},
		},
	}
	err = Add(&mp, "developer.lang.1", "Golang")
	assert.Nil(t, err)
	assert.Equal(t, "", Get(mp, "developer.lang.0"))
	assert.Equal(t, "Golang", Get(mp, "developer.lang.1"))

	mp = map[string]any{}
	err = Add(&mp, "foo", map[string]any{
		"bar": "baz",
	})
	assert.Nil(t, err)
	assert.Equal(t, "baz", Get(mp, "foo.bar"))
}

func TestDot(t *testing.T) {
	mp := Dot(map[string]any{
		"foo": map[string]any{
			"bar": "baz",
		},
	})
	assert.Equal(t, map[string]any{
		"foo.bar": "baz",
	}, mp)

	mp = Dot(map[string]any{
		"foo": map[int]int{
			10: 100,
		},
	})
	assert.Equal(t, map[string]any{
		"foo.10": 100,
	}, mp)

	mp = Dot(map[string]any{})
	assert.Equal(t, map[string]any{}, mp)

	mp = Dot(map[string]any{
		"foo": []string{},
	})
	assert.Equal(t, map[string]any{}, mp)

	mp = Dot(map[string]any{
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
	}, mp)

	mp = Dot(map[string]any{
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
	}, mp)
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
	mp := map[string]any{
		"products": map[string]any{
			"desk": map[string]int{
				"price": 100,
			},
		},
	}
	Forget(mp)
	assert.Equal(t, map[string]any{
		"products": map[string]any{
			"desk": map[string]int{
				"price": 100,
			},
		},
	}, mp)

	Forget(mp, "products.desk")
	assert.Equal(t, map[string]any{
		"products": map[string]any{},
	}, mp)

	mp = map[string]any{
		"products": map[string]any{
			"desk": map[string]int{
				"price": 100,
			},
		},
	}
	Forget(mp, "products.desk.price")
	Forget(mp, "products.desk.quantity")
	assert.Equal(t, map[string]any{
		"products": map[string]any{
			"desk": map[string]int{},
		},
	}, mp)

	mp = map[string]any{
		"products": map[string]any{
			"desk": map[string]int{
				"price": 100,
			},
		},
	}
	Forget(mp, "products.chair.price")
	assert.Equal(t, map[string]any{
		"products": map[string]any{
			"desk": map[string]int{
				"price": 100,
			},
		},
	}, mp)

	mp = map[string]any{
		"products": map[string]any{
			"desk": map[string]any{
				"price": map[string]int{
					"original": 100,
					"taxes":    120,
				},
			},
		},
	}
	Forget(mp, "products.desk.price.taxes")
	assert.Equal(t, map[string]any{
		"products": map[string]any{
			"desk": map[string]any{
				"price": map[string]int{
					"original": 100,
				},
			},
		},
	}, mp)

	mp = map[string]any{
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
	Forget(mp, "developers.*.lang")
	assert.Equal(t, map[string]any{
		"developers": []map[string]any{
			{
				"name": "Bowen",
			},
			{
				"name": "Krishan",
			},
		},
	}, mp)

	mp = map[string]any{
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
				"lang": []string{
					"C",
					"Golang",
				},
			},
		},
	}
	Forget(mp, "developers.*.lang.1", "developers.*.name")
	assert.Equal(t, map[string]any{
		"developers": []map[string]any{
			{
				"lang": []string{
					"Golang",
				},
			},
			{
				"lang": []string{
					"C",
				},
			},
		},
	}, mp)

	// Only works on first level keys
	mp = map[string]any{
		"joe@example.com": "Joe",
		"jane@localhost":  "Jane",
	}
	Forget(mp, "joe@example.com")
	assert.Equal(t, map[string]any{
		"jane@localhost": "Jane",
	}, mp)

	// Doesn't remove nested keys
	mp = map[string]any{
		"emails": map[string]string{
			"joe@example.com": "Joe",
			"jane@localhost":  "Jane",
		},
	}
	Forget(mp, "emails.joe@example.com")
	assert.Equal(t, map[string]any{
		"emails": map[string]string{
			"joe@example.com": "Joe",
			"jane@localhost":  "Jane",
		},
	}, mp)

	mp = map[string]any{
		"developers": []map[string]string{
			{
				"name": "Bowen",
			},
			{
				"name": "Krishan",
			},
		},
	}
	Forget(mp, "developers.*.name")
	assert.Equal(t, map[string]any{
		"developers": []map[string]string{},
	}, mp)

	// Test nil value
	mp = map[string]any{
		"shop": map[string]any{
			"cart": map[any]any{
				150:   0,
				"foo": "bar",
			},
		},
	}
	Forget(mp, "shop.cart.150")
	Forget(mp, "shop.cart.100")
	Forget(mp, "shop.cart.foo")
	assert.Equal(t, map[string]any{
		"shop": map[string]any{
			"cart": map[any]any{},
		},
	}, mp)

	mp = map[string]any{
		"developers": []map[string]any{
			{
				"lang": []string{
					"Golang",
					"PHP",
				},
			},
		},
	}
	Forget(mp, "developers.0.lang.0")
	assert.Equal(t, map[string]any{
		"developers": []map[string]any{
			{
				"lang": []string{
					"PHP",
				},
			},
		},
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

func TestGet(t *testing.T) {
	mp := map[string]any{
		"products": map[string]any{
			"desk": map[string]int{
				"price": 100,
			},
		},
	}
	assert.Equal(t, map[string]int{"price": 100}, Get(mp, "products.desk"))

	// Test nil value
	mp = map[string]any{
		"foo": nil,
		"bar": map[string]any{
			"baz": nil,
		},
	}
	assert.Nil(t, Get(mp, "foo", "default"))
	assert.Nil(t, Get(mp, "bar.baz", "default"))
	// Test missing
	assert.Nil(t, Get(mp, "foo.bar"))

	// Test numeric keys
	mp = map[string]any{
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
	assert.Equal(t, "Krishan", Get(mp, "developers.1.name"))
	assert.Equal(t, "Bowen", Get(mp, "developers.0.name"))
	assert.Equal(t, "Golang", Get(mp, "developers.0.lang.0"))

	// Test return default value
	mp = map[string]any{
		"names": map[string]any{
			"developer": "Krishan",
		},
	}
	assert.Equal(t, "name", Get(mp, "names.designer", "name"))
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

	assert.True(t, Has(mp, "framework.lang"))

	assert.True(t, Has(mp, "framework.dev.name"))

	assert.False(t, Has(mp, "framework.foo"))

	assert.False(t, Has(mp, "framework.dev.foo"))

	assert.True(t, Has(mp, "foo"))

	assert.True(t, Has(mp, "bar.baz"))

	assert.True(t, Has(mp, "framework.name", "framework.dev.name"))

	assert.False(t, Has(mp, "framework.name", "framework.dev.foo"))

	assert.True(t, Has(mp, "developers.0.name"))

	assert.False(t, Has(mp, "product.developers.0.foo"))

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

	mp = map[string]any{
		"foo": map[string]any{
			"bar": nil,
			"baz": "",
		},
	}
	assert.True(t, HasAny(mp, "foo.bar"))
	assert.True(t, HasAny(mp, "foo.baz"))
	assert.False(t, HasAny(mp, "foo.bax"))
	assert.True(t, HasAny(mp, "foo.bax", "foo.baz"))
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

	// Only works on first level keys
	mp = map[string]any{
		"joe@example.com": "Joe",
		"jane@localhost":  "Jane",
	}
	assert.Equal(t, "Joe", Pull(mp, "joe@example.com"))
	assert.Equal(t, map[string]any{"jane@localhost": "Jane"}, mp)

	// Doesn't remove nested keys
	mp = map[string]any{
		"emails": map[string]any{
			"joe@example.com": "Joe",
			"jane@localhost":  "Jane",
		},
	}
	assert.Nil(t, Pull(mp, "emails.joe@example.com"))
	assert.Equal(t, map[string]any{
		"emails": map[string]any{
			"joe@example.com": "Joe",
			"jane@localhost":  "Jane",
		},
	}, mp)

	// work with slices
	mp = map[string]any{
		"names": []string{"Bowen", "Krishan"},
	}
	assert.Equal(t, "Bowen", Pull(mp, "names.0"))
	assert.Equal(t, map[string]any{"names": []string{"Krishan"}}, mp)

	mp = map[string]any{
		"names": []string{"Bowen", "Krishan"},
	}
	assert.Equal(t, []string{"Bowen", "Krishan"}, Pull(mp, "names.*"))
	assert.Equal(t, map[string]any{"names": []string{}}, mp)

	// default value
	mp = map[string]any{
		"name": "Krishan",
	}
	assert.Equal(t, "default", Pull(mp, "age", "default"))

	// Test generic type
	gMp := map[int]string{
		1: "one",
		2: "two",
	}
	assert.Equal(t, "one", Pull(gMp, 1))
	assert.Equal(t, map[int]string{2: "two"}, gMp)
	assert.Equal(t, nil, Pull(gMp, 3))
}

func TestSet(t *testing.T) {
	mp := map[string]any{
		"products": map[string]any{
			"desk": map[string]int{
				"price": 100,
			},
		},
	}
	err := Set(&mp, "products.desk.price", 200)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"products": map[string]any{
			"desk": map[string]int{
				"price": 200,
			},
		},
	}, mp)

	// key does not exist
	mp = map[string]any{
		"products": map[string]any{},
	}
	err = Set(&mp, "products.desk.price", 200)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"products": map[string]any{
			"desk": map[string]any{
				"price": 200,
			},
		},
	}, mp)
}

func TestDeleteByPathKeys(t *testing.T) {
	mp := map[string]any{
		"products": map[string]any{
			"desk": map[string]int{
				"price": 100,
			},
		},
	}
	val, ok := deleteByPathKeys(mp, mp, []string{})
	assert.False(t, ok)
	assert.Nil(t, val)

	mp = map[string]any{
		"names": []string{"Bowen", "Krishan"},
	}
	val, ok = deleteByPathKeys(mp, mp, []string{"names", "*", "foo"})
	assert.False(t, ok)
	assert.Nil(t, val)

	val, ok = deleteByPathKeys(mp, mp, []string{"names", "3"})
	assert.False(t, ok)
	assert.Nil(t, val)
}
