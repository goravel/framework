package collect

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	ID   int
	Name string
	Age  int
}

func TestNew(t *testing.T) {
	assert.Equal(t, 3, Of([]int{1, 2, 3}).Count())
}

func TestCollect(t *testing.T) {
	assert.Equal(t, 5, Of([]int{1, 2, 3, 4, 5}).Count())
}

func TestAll(t *testing.T) {
	items := []int{1, 2, 3}
	assert.Equal(t, items, Of(items).All())
}

func TestCollectCount(t *testing.T) {
	assert.Equal(t, 5, Of([]int{1, 2, 3, 4, 5}).Count())
}

func TestIsEmpty(t *testing.T) {
	c := Of([]int{})
	assert.True(t, c.IsEmpty())
	c.Push(1)
	assert.False(t, c.IsEmpty())
}

func TestIsNotEmpty(t *testing.T) {
	assert.True(t, Of([]int{1, 2, 3}).IsNotEmpty())
}

func TestFirst(t *testing.T) {
	first := Of([]int{1, 2, 3}).First()
	assert.NotNil(t, first)
	assert.Equal(t, 1, *first)
}

func TestLast(t *testing.T) {
	last := Of([]int{1, 2, 3}).Last()
	assert.NotNil(t, last)
	assert.Equal(t, 3, *last)
}

func TestPush(t *testing.T) {
	c := Of([]int{1, 2})
	c.Push(3, 4)
	assert.Equal(t, 4, c.Count())
	assert.Equal(t, 4, *c.Last())
}

func TestPop(t *testing.T) {
	c := Of([]int{1, 2, 3})
	popped := c.Pop()
	assert.NotNil(t, popped)
	assert.Equal(t, 3, *popped)
	assert.Equal(t, 2, c.Count())
}

func TestShift(t *testing.T) {
	c := Of([]int{1, 2, 3})
	shifted := c.Shift()
	assert.NotNil(t, shifted)
	assert.Equal(t, 1, *shifted)
	assert.Equal(t, 2, c.Count())
}

func TestUnshift(t *testing.T) {
	c := Of([]int{2, 3})
	c.Unshift(1)
	assert.Equal(t, 3, c.Count())
	assert.Equal(t, 1, *c.First())
}

func TestCollectFilter(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	filtered := c.Filter(func(item int, index int) bool {
		return item%2 == 0
	})
	assert.Equal(t, []int{2, 4}, filtered.All())
}

func TestCollectEach(t *testing.T) {
	c := Of([]int{1, 2, 3})
	sum := 0
	c.Each(func(item int, index int) {
		sum += item
	})
	assert.Equal(t, 6, sum)
}

func TestContains(t *testing.T) {
	c := Of([]int{1, 2, 3})
	assert.True(t, c.Contains(2))
	assert.False(t, c.Contains(4))
}

func TestCollectUnique(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, Of([]int{1, 2, 2, 3, 3, 3}).Unique().All())
}

func TestCollectReverse(t *testing.T) {
	assert.Equal(t, []int{3, 2, 1}, Of([]int{1, 2, 3}).Reverse().All())
}

func TestSlice(t *testing.T) {
	assert.Equal(t, []int{2, 3, 4}, Of([]int{1, 2, 3, 4, 5}).Slice(1, 3).All())
	assert.Equal(t, []int{3, 4, 5}, Of([]int{1, 2, 3, 4, 5}).Slice(2).All())
}

func TestTake(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, Of([]int{1, 2, 3, 4, 5}).Take(3).All())
}

func TestSkip(t *testing.T) {
	assert.Equal(t, []int{3, 4, 5}, Of([]int{1, 2, 3, 4, 5}).Skip(2).All())
}

func TestChunk(t *testing.T) {
	assert.Equal(t, [][]int{{1, 2}, {3, 4}, {5}}, Of([]int{1, 2, 3, 4, 5}).Chunk(2))
}

func TestSort(t *testing.T) {
	c := Of([]int{3, 1, 4, 1, 5})
	assert.Equal(t, []int{1, 1, 3, 4, 5}, c.Sort(func(a, b int) bool { return a < b }).All())
}

func TestSortBy(t *testing.T) {
	items := []TestStruct{
		{ID: 1, Name: "Charlie", Age: 30},
		{ID: 2, Name: "Alice", Age: 25},
		{ID: 3, Name: "Bob", Age: 35},
	}
	c := Of(items)
	sorted := c.SortBy(func(item TestStruct) string {
		return item.Name
	})

	assert.Equal(t, "Alice", sorted.All()[0].Name)
}

func TestCollectSum(t *testing.T) {
	sum := Of([]int{1, 2, 3, 4, 5}).Sum(func(item int) float64 { return float64(item) })
	assert.Equal(t, 15.0, sum)
}

func TestAvg(t *testing.T) {
	avg := Of([]int{1, 2, 3, 4, 5}).Avg(func(item int) float64 { return float64(item) })
	assert.Equal(t, 3.0, avg)
}

func TestCollectMin(t *testing.T) {
	assert.Equal(t, 1.0, Of([]int{3, 1, 4, 1, 5}).Min(func(item int) float64 { return float64(item) }))
}

func TestCollectMax(t *testing.T) {
	assert.Equal(t, 5.0, Of([]int{3, 1, 4, 1, 5}).Max(func(item int) float64 { return float64(item) }))
}

func TestJoin(t *testing.T) {
	assert.Equal(t, "1,2,3", Of([]int{1, 2, 3}).Join(","))
}

func TestCollectionDiff(t *testing.T) {
	assert.Equal(t, []int{1, 2}, Of([]int{1, 2, 3, 4, 5}).Diff(Of([]int{3, 4, 5, 6, 7})).All())
}

func TestIntersect(t *testing.T) {
	assert.Equal(t, []int{3, 4, 5}, Of([]int{1, 2, 3, 4, 5}).Intersect(Of([]int{3, 4, 5, 6, 7})).All())
}

func TestWhere(t *testing.T) {
	items := []TestStruct{
		{ID: 1, Name: "Alice", Age: 25},
		{ID: 2, Name: "Bob", Age: 30},
		{ID: 3, Name: "Charlie", Age: 25},
	}
	assert.Equal(t, 2, Of(items).Where("Age", "=", 25).Count())
}

func TestWhereIn(t *testing.T) {
	items := []TestStruct{
		{ID: 1, Name: "Alice", Age: 25},
		{ID: 2, Name: "Bob", Age: 30},
		{ID: 3, Name: "Charlie", Age: 35},
	}
	assert.Equal(t, 2, Of(items).WhereIn("Age", []any{25, 35}).Count())
}

func TestPluck(t *testing.T) {
	items := []TestStruct{
		{ID: 1, Name: "Alice", Age: 25},
		{ID: 2, Name: "Bob", Age: 30},
		{ID: 3, Name: "Charlie", Age: 35},
	}
	assert.Equal(t, 3, Of(items).Pluck("Name").Count())
}

func TestEvery(t *testing.T) {
	allEven := Of([]int{2, 4, 6, 8}).Every(func(item int) bool { return item%2 == 0 })
	assert.True(t, allEven)
}

func TestPartition(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	even, odd := c.Partition(func(item int) bool {
		return item%2 == 0
	})

	assert.Equal(t, []int{2, 4}, even.All())
	assert.Equal(t, []int{1, 3, 5}, odd.All())
}

func TestWhen(t *testing.T) {
	c := Of([]int{1, 2, 3})
	result := c.When(true, func(col *Collection[int]) *Collection[int] {
		return col.Filter(func(item int, _ int) bool {
			return item > 1
		})
	})

	assert.Equal(t, []int{2, 3}, result.All())
}

func TestUnless(t *testing.T) {
	c := Of([]int{1, 2, 3})
	result := c.Unless(false, func(col *Collection[int]) *Collection[int] {
		return col.Filter(func(item int, _ int) bool {
			return item > 1
		})
	})

	assert.Equal(t, []int{2, 3}, result.All())
}

func TestTap(t *testing.T) {
	c := Of([]int{1, 2, 3})
	tapped := false
	result := c.Tap(func(col *Collection[int]) {
		tapped = true
	})

	assert.True(t, tapped)
	assert.Equal(t, 3, result.Count())
}

func TestPipe(t *testing.T) {
	result := Of([]int{1, 2, 3}).Pipe(func(col *Collection[int]) any {
		return col.Count()
	})

	assert.Equal(t, 3, result)
}

func TestClone(t *testing.T) {
	c := Of([]int{1, 2, 3})
	cloned := c.Clone()

	assert.Equal(t, c.All(), cloned.All())

	cloned.Push(4)
	assert.NotEqual(t, c.Count(), cloned.Count())
}

func TestSearch(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	assert.Equal(t, 2, c.Search(3))
	assert.Equal(t, -1, c.Search(6))
}

// Enhanced Where method tests
func TestWhereEnhanced(t *testing.T) {
	type User struct {
		ID        int
		Name      string
		Age       int
		Country   string
		Balance   float64
		DeletedAt *string
	}

	deletedUser := "deleted"
	users := []User{
		{ID: 1, Name: "Alice", Age: 25, Country: "FR", Balance: 150.0, DeletedAt: nil},
		{ID: 2, Name: "Bob", Age: 30, Country: "US", Balance: 80.0, DeletedAt: nil},
		{ID: 3, Name: "Charlie", Age: 25, Country: "FR", Balance: 200.0, DeletedAt: &deletedUser},
		{ID: 4, Name: "David", Age: 35, Country: "UK", Balance: 120.0, DeletedAt: nil},
		{ID: 5, Name: "Eve", Age: 40, Country: "US", Balance: 90.0, DeletedAt: nil},
	}
	c := Of(users)

	// Test 1: Two parameters (field, value) - implies '=' operator
	frenchUsers := c.Where("Country", "FR")
	assert.Equal(t, 2, frenchUsers.Count())

	youngUsers := c.Where("Age", 25)
	assert.Equal(t, 2, youngUsers.Count())

	// Test 2: Three parameters (field, operator, value)
	richUsers := c.Where("Balance", ">", 100.0)
	assert.Equal(t, 3, richUsers.Count())

	nonFrenchUsers := c.Where("Country", "!=", "FR")
	assert.Equal(t, 3, nonFrenchUsers.Count())

	seniorUsers := c.Where("Age", ">=", 35)
	assert.Equal(t, 2, seniorUsers.Count())

	youngAdults := c.Where("Age", "<", 35)
	assert.Equal(t, 3, youngAdults.Count())

	// Test 3: Single parameter (callback function)
	customFilter := c.Where(func(u User) bool {
		return u.Age > 25 && u.Country == "US"
	})
	assert.Equal(t, 2, customFilter.Count())

	// Test 4: Null comparisons
	activeUsers := c.Where("DeletedAt", "=", nil)
	assert.Equal(t, 4, activeUsers.Count())

	deletedUsers := c.Where("DeletedAt", "!=", nil)
	assert.Equal(t, 1, deletedUsers.Count())

	// Test 5: String operations (like/not like)
	nameWithA := c.Where("Name", "like", "a")
	assert.Equal(t, 3, nameWithA.Count()) // Alice, Charlie, and David

	nameNotLikeTest := c.Where("Name", "not like", "test")
	assert.Equal(t, 5, nameNotLikeTest.Count()) // All users since none have 'test' in name

	type Product struct {
		Price int
	}
	products := Of([]Product{{Price: 100}})
	assert.Equal(t, 1, products.Where("Price", "=", "100").Count())
}

func TestWhereErrorCases(t *testing.T) {
	type User struct {
		ID   int
		Name string
		Age  int
	}

	users := []User{
		{ID: 1, Name: "Alice", Age: 25},
		{ID: 2, Name: "Bob", Age: 30},
	}
	c := Of(users)

	// Test invalid parameter counts
	result := c.Where()
	assert.Equal(t, 2, result.Count())

	result = c.Where("Age", "=", 25, "extra")
	assert.Equal(t, 2, result.Count())

	// Test invalid callback type
	result = c.Where("not a callback")
	assert.Equal(t, 2, result.Count())

	// Test invalid field name (non-string)
	result = c.Where(123, 25)
	assert.Equal(t, 2, result.Count())

	// Test invalid operator (non-string)
	result = c.Where("Age", 123, 25)
	assert.Equal(t, 2, result.Count())
}

func TestMapMethod(t *testing.T) {
	// Test with integers
	numbers := Of([]int{1, 2, 3, 4, 5})

	// Test basic transformation
	doubled := numbers.Map(func(n int, _ int) any {
		return n * 2
	})

	assert.Equal(t, []any{2, 4, 6, 8, 10}, doubled.All())

	// Test with index
	withIndex := numbers.Map(func(n int, i int) any {
		return fmt.Sprintf("item_%d_%d", i, n)
	})

	assert.Equal(t, []any{"item_0_1", "item_1_2", "item_2_3", "item_3_4", "item_4_5"}, withIndex.All())

	// Test with structs
	type User struct {
		ID   int
		Name string
		Age  int
	}

	users := Of([]User{
		{ID: 1, Name: "Alice", Age: 25},
		{ID: 2, Name: "Bob", Age: 30},
		{ID: 3, Name: "Charlie", Age: 35},
	})

	// Extract names
	names := users.Map(func(u User, _ int) any {
		return u.Name
	})

	assert.Equal(t, []any{"Alice", "Bob", "Charlie"}, names.All())

	// Complex transformation
	summaries := users.Map(func(u User, i int) any {
		return map[string]any{
			"id":      u.ID,
			"summary": fmt.Sprintf("%s (%d years)", u.Name, u.Age),
			"index":   i,
		}
	})

	assert.Equal(t, 3, summaries.Count())

	// Test chaining with other methods
	result := numbers.
		Map(func(n int, _ int) any {
			return n * 2
		}).
		Filter(func(item any, _ int) bool {
			return item.(int) > 5
		})

	assert.Equal(t, []any{6, 8, 10}, result.All())
}

func TestMapMethodTypes(t *testing.T) {
	// Test different return types
	numbers := Of([]int{1, 2, 3})

	// Map to strings
	strings := numbers.Map(func(n int, _ int) any {
		return fmt.Sprintf("number_%d", n)
	})

	assert.Equal(t, []any{"number_1", "number_2", "number_3"}, strings.All())

	// Map to booleans
	booleans := numbers.Map(func(n int, _ int) any {
		return n%2 == 0
	})

	assert.Equal(t, []any{false, true, false}, booleans.All())

	// Map to maps
	maps := numbers.Map(func(n int, i int) any {
		return map[string]int{
			"value": n,
			"index": i,
		}
	})

	assert.Equal(t, 3, maps.Count())

	// Test empty collection
	empty := Of([]int{})
	emptyMapped := empty.Map(func(n int, _ int) any {
		return n * 2
	})

	assert.Equal(t, 0, emptyMapped.Count())
}

func TestReduceMethod(t *testing.T) {
	// Test with integers
	numbers := Of([]int{1, 2, 3, 4, 5})

	// Test basic sum reduction
	sum := numbers.Reduce(func(acc any, item int, index int) any {
		accValue, _ := acc.(int)
		return accValue + item
	}, 0)

	assert.Equal(t, 15, sum)

	// Test with string concatenation
	strings := Of([]string{"hello", "world", "test"})
	concat := strings.Reduce(func(acc any, item string, index int) any {
		accValue, _ := acc.(string)
		if index > 0 {
			return accValue + "-" + item
		}
		return accValue + item
	}, "")

	assert.Equal(t, "hello-world-test", concat)

	// Test with custom accumulator type
	type Accumulator struct {
		Sum   int
		Count int
	}

	avgAcc := numbers.Reduce(func(acc any, item int, _ int) any {
		accValue, _ := acc.(Accumulator)
		return Accumulator{
			Sum:   accValue.Sum + item,
			Count: accValue.Count + 1,
		}
	}, Accumulator{Sum: 0, Count: 0})

	expectedAcc := Accumulator{Sum: 15, Count: 5}
	assert.Equal(t, expectedAcc.Sum, avgAcc.(Accumulator).Sum)
	assert.Equal(t, expectedAcc.Count, avgAcc.(Accumulator).Count)

	// Test with empty collection
	empty := Of([]int{})
	emptyResult := empty.Reduce(func(acc any, item int, _ int) any {
		accValue, _ := acc.(int)
		return accValue + item
	}, 0)

	assert.Equal(t, 0, emptyResult)

	// Test using index in reducer
	indexSum := numbers.Reduce(func(acc any, item int, index int) any {
		accValue, _ := acc.(int)
		return accValue + (item * index)
	}, 0)

	// 0*1 + 1*2 + 2*3 + 3*4 + 4*5 = 40
	assert.Equal(t, 40, indexSum)
}

func TestMapIntoMethod(t *testing.T) {
	// Test converting ints to floats (compatible types)
	numbers := Of([]int{1, 2, 3, 4, 5})
	floatTarget := float64(0) // Target type hint

	floatColl := numbers.MapInto(floatTarget)
	assert.Equal(t, numbers.Count(), floatColl.Count())

	// Verify all items were converted to float64
	for i, item := range floatColl.All() {
		val, ok := item.(float64)
		assert.True(t, ok, "Item at index %d should be float64, got %T", i, item)
		assert.Equal(t, float64(i+1), val)
	}

	// Test with incompatible types (string to int)
	strings := Of([]string{"1", "2", "3", "foo", "bar"})
	intTarget := 0 // Target type hint

	convColl := strings.MapInto(intTarget)
	assert.Equal(t, strings.Count(), convColl.Count())

	// Verify incompatible items remain as original strings
	for _, item := range convColl.All() {
		_, ok := item.(int)
		assert.False(t, ok, "String should not convert to int, got %T: %v", item, item)
		_, isString := item.(string)
		assert.True(t, isString, "Item should remain as string, got %T: %v", item, item)
	}

	// Test with empty collection
	empty := Of([]int{})
	emptyResult := empty.MapInto(floatTarget)
	assert.Equal(t, 0, emptyResult.Count())

	// Test with custom struct types
	type SimpleStruct struct {
		Value int
	}

	type ComplexStruct struct {
		Value int
		Extra string
	}

	simple := Of([]SimpleStruct{SimpleStruct{Value: 1}, SimpleStruct{Value: 2}, SimpleStruct{Value: 3}})

	// SimpleStruct is not convertible to ComplexStruct
	complexTarget := ComplexStruct{}
	complexColl := simple.MapInto(complexTarget)

	for _, item := range complexColl.All() {
		_, ok := item.(ComplexStruct)
		assert.False(t, ok, "SimpleStruct should not be convertible to ComplexStruct")

		// Items should remain as SimpleStruct
		_, isSimple := item.(SimpleStruct)
		assert.True(t, isSimple, "Item should remain as SimpleStruct, got %T", item)
	}
}

func TestAfter(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})

	result := c.After(3)
	assert.NotNil(t, result)
	assert.Equal(t, 4, *result)
	assert.Nil(t, c.After(5))
	assert.Nil(t, c.After(99))
	assert.Nil(t, Of([]int{}).After(1))
}

func TestBefore(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})

	result := c.Before(3)
	assert.NotNil(t, result)
	assert.Equal(t, 2, *result)
	assert.Nil(t, c.Before(1))
	assert.Nil(t, c.Before(99))
	assert.Nil(t, Of([]int{}).Before(1))
}

func TestGet(t *testing.T) {
	c := Of([]int{10, 20, 30, 40, 50})

	result := c.Get(2)
	assert.NotNil(t, result)
	assert.Equal(t, 30, *result)
	first := c.Get(0)
	assert.NotNil(t, first)
	assert.Equal(t, 10, *first)
	assert.Nil(t, c.Get(-1))
	assert.Nil(t, c.Get(10))
	assert.Nil(t, Of([]int{}).Get(0))
}

func TestHas(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})

	assert.True(t, c.Has(0))
	assert.True(t, c.Has(2))
	assert.True(t, c.Has(4))
	assert.False(t, c.Has(-1))
	assert.False(t, c.Has(5))
	assert.False(t, Of([]int{}).Has(0))
}

func TestHasAny(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})

	assert.True(t, c.HasAny(2, 10, 20))
	assert.True(t, c.HasAny(0, 1, 2))
	assert.False(t, c.HasAny(10, 20, 30))
	assert.False(t, c.HasAny())
	assert.False(t, Of([]int{}).HasAny(0, 1, 2))
}

func TestContainsStrict(t *testing.T) {
	c := Of([]int{1, 2, 3})
	assert.True(t, c.ContainsStrict(2))
	assert.True(t, c.ContainsStrict(1))
	assert.False(t, c.ContainsStrict(99))
	assert.False(t, Of([]int{}).ContainsStrict(1))

	type Product struct {
		ID    int
		Name  string
		Price float64
	}
	products := Of([]Product{Product{ID: 1, Name: "Book", Price: 10.99}, Product{ID: 2, Name: "Pen", Price: 2.99}})
	assert.True(t, products.ContainsStrict(Product{ID: 1, Name: "Book", Price: 10.99}))
	assert.False(t, products.ContainsStrict(Product{ID: 1, Name: "Book", Price: 11.99}))
}

func TestDoesntContain(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	assert.False(t, c.DoesntContain(3))
	assert.True(t, c.DoesntContain(99))
	assert.True(t, Of([]int{}).DoesntContain(1))

	words := Of([]string{"apple", "banana", "cherry"})
	assert.True(t, words.DoesntContain("orange"))
	assert.False(t, words.DoesntContain("banana"))
}

func TestDuplicates(t *testing.T) {
	c := Of([]int{1, 2, 2, 3, 3, 3, 4})
	assert.Equal(t, []int{2, 3}, c.Duplicates().All())
	assert.Equal(t, 0, Of([]int{1, 2, 3, 4, 5}).Duplicates().Count())
	assert.Equal(t, 1, Of([]int{5, 5, 5, 5}).Duplicates().Count())
	assert.Equal(t, 0, Of([]int{}).Duplicates().Count())
	assert.Equal(t, 2, Of([]string{"apple", "banana", "apple", "cherry", "banana", "apple"}).Duplicates().Count())
}

func TestReject(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	assert.Equal(t, []int{1, 3, 5, 7, 9}, c.Reject(func(n int, _ int) bool { return n%2 == 0 }).All())
	assert.Equal(t, 0, c.Reject(func(n int, _ int) bool { return true }).Count())
	assert.Equal(t, c.All(), c.Reject(func(n int, _ int) bool { return false }).All())

	type Item struct{ Value int }
	assert.Equal(t, 2, Of([]Item{Item{1}, Item{2}, Item{3}, Item{4}}).Reject(func(_ Item, i int) bool { return i%2 == 0 }).Count())
	assert.Equal(t, 0, Of([]int{}).Reject(func(n int, _ int) bool { return n > 5 }).Count())
}

func TestDot(t *testing.T) {
	result := Of([]string{"apple", "banana", "cherry"}).Dot()
	assert.Equal(t, "apple", result["0"])
	assert.Equal(t, "banana", result["1"])
	assert.Equal(t, "cherry", result["2"])
	assert.Equal(t, 0, len(Of([]string{}).Dot()))
	assert.Equal(t, 42, Of([]int{42}).Dot()["0"])
}

func TestKeyBy(t *testing.T) {
	type Product struct {
		ID   int
		Name string
	}

	result := Of([]Product{Product{ID: 1, Name: "Book"}, Product{ID: 2, Name: "Pen"}, Product{ID: 3, Name: "Pencil"}}).KeyBy(func(p Product) string { return p.Name })
	assert.Equal(t, 1, result["Book"].ID)
	assert.Equal(t, 2, result["Pen"].ID)

	dupResult := Of([]Product{Product{ID: 1, Name: "Item"}, Product{ID: 2, Name: "Item"}}).KeyBy(func(p Product) string { return p.Name })
	assert.Equal(t, 2, dupResult["Item"].ID)
	assert.Equal(t, 0, len(Of([]Product{}).KeyBy(func(p Product) string { return p.Name })))
}

func TestOnly(t *testing.T) {
	c := Of([]int{10, 20, 30, 40, 50})
	assert.Equal(t, []int{20, 40}, c.Only(1, 3).All())
	assert.Equal(t, []int{10, 30}, c.Only(0, 10, 2).All())
	assert.Equal(t, 0, c.Only().Count())
	assert.Equal(t, 3, c.Only(1, 1, 3).Count())
	assert.Equal(t, 0, Of([]int{}).Only(0, 1, 2).Count())
}

func TestExcept(t *testing.T) {
	c := Of([]int{10, 20, 30, 40, 50})
	assert.Equal(t, []int{10, 30, 50}, c.Except(1, 3).All())
	assert.Equal(t, []int{20, 40, 50}, c.Except(0, 10, 2).All())
	assert.Equal(t, c.All(), c.Except().All())
	assert.Equal(t, []int{10, 30, 50}, c.Except(1, 1, 3).All())
	assert.Equal(t, 0, Of([]int{}).Except(0, 1, 2).Count())
}

func TestForget(t *testing.T) {
	c := Of([]int{10, 20, 30, 40, 50})
	result := c.Forget(1, 3)
	assert.Same(t, c, result)
	assert.Equal(t, []int{10, 30, 50}, c.All())
}

func TestPrepend(t *testing.T) {
	c := Of([]int{3, 4, 5})
	result := c.Prepend(1, 2)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, c.All())
	assert.Same(t, c, result)

	empty := Of([]int{})
	empty.Prepend(10)
	assert.Equal(t, 1, empty.Count())
	assert.Equal(t, 10, *empty.First())
}

func TestPull(t *testing.T) {
	c := Of([]int{10, 20, 30, 40, 50})
	pulled := c.Pull(2)
	assert.NotNil(t, pulled)
	assert.Equal(t, 30, *pulled)
	assert.Equal(t, []int{10, 20, 40, 50}, c.All())
	assert.Nil(t, c.Pull(10))
	assert.Nil(t, c.Pull(-1))
}

func TestPut(t *testing.T) {
	c := Of([]int{10, 20, 30, 40, 50})
	result := c.Put(2, 99)
	expected := []int{10, 20, 99, 40, 50}
	assert.Equal(t, expected, c.All())
	assert.Same(t, c, result)
	c.Put(10, 100)
	c.Put(-1, 200)
	assert.Equal(t, expected, c.All())
}

func TestFirstWhere(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	result := c.FirstWhere(func(n int) bool { return n%2 == 0 })
	assert.NotNil(t, result)
	assert.Equal(t, 2, *result)
	assert.Nil(t, c.FirstWhere(func(n int) bool { return n > 10 }))
	assert.Nil(t, Of([]int{}).FirstWhere(func(n int) bool { return true }))

	type User struct {
		Name string
		Age  int
	}
	oldUser := Of([]User{User{Name: "Alice", Age: 25}, User{Name: "Bob", Age: 30}, User{Name: "Charlie", Age: 35}}).FirstWhere(func(u User) bool { return u.Age >= 30 })
	assert.NotNil(t, oldUser)
	assert.Equal(t, "Bob", oldUser.Name)
}

func TestSearchBy(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	assert.Equal(t, 1, c.SearchBy(func(n int) bool { return n%2 == 0 }))
	assert.Equal(t, -1, c.SearchBy(func(n int) bool { return n > 10 }))
	assert.Equal(t, -1, Of([]int{}).SearchBy(func(n int) bool { return true }))

	type Product struct {
		Name  string
		Price float64
	}
	assert.Equal(t, 0, Of([]Product{Product{Name: "Book", Price: 10.99}, Product{Name: "Pen", Price: 2.99}, Product{Name: "Notebook", Price: 5.99}}).SearchBy(func(p Product) bool { return p.Price > 5.00 }))
}

func TestSortByDesc(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	users := Of([]User{User{Name: "Alice", Age: 25}, User{Name: "Bob", Age: 35}, User{Name: "Charlie", Age: 30}})
	sorted := users.SortByDesc(func(u User) string { return fmt.Sprintf("%02d", u.Age) })

	assert.Equal(t, "Bob", sorted.All()[0].Name)
	assert.Equal(t, "Charlie", sorted.All()[1].Name)
	assert.Equal(t, "Alice", sorted.All()[2].Name)
	assert.Equal(t, "Alice", users.All()[0].Name)
	assert.Equal(t, 0, Of([]User{}).SortByDesc(func(u User) string { return u.Name }).Count())
}

func TestChunkWhile(t *testing.T) {
	c := Of([]int{1, 2, 2, 3, 3, 3, 4})
	chunks := c.ChunkWhile(func(item int, _ int, chunk []int) bool { return len(chunk) == 0 || item == chunk[0] })
	assert.Equal(t, 4, len(chunks))
	assert.Equal(t, 1, len(chunks[0]))
	assert.Equal(t, 2, len(chunks[1]))
	assert.Equal(t, 3, len(chunks[2]))
	assert.Equal(t, 0, len(Of([]int{}).ChunkWhile(func(int, int, []int) bool { return true })))
}

func TestCollectionCountBy(t *testing.T) {
	type Product struct{ Type string }
	products := Of([]Product{Product{"fruit"}, Product{"veg"}, Product{"fruit"}, Product{"meat"}})
	counts := products.CountBy(func(p Product) string { return p.Type })
	assert.Equal(t, 2, counts["fruit"])
	assert.Equal(t, 1, counts["veg"])
	assert.Equal(t, 1, counts["meat"])
	assert.Equal(t, 0, len(Of([]Product{}).CountBy(func(p Product) string { return p.Type })))
}

func TestCrossJoin(t *testing.T) {
	c1 := Of([]int{1, 2})
	c2 := Of([]int{3, 4})
	result := c1.CrossJoin(c2)
	assert.Equal(t, 4, len(result))
	assert.Equal(t, []int{1, 3}, result[0])
	assert.Equal(t, []int{2, 4}, result[3])
	assert.Equal(t, 0, len(Of([]int{}).CrossJoin(c2)))
	assert.Equal(t, 0, len(c1.CrossJoin(Of([]int{}))))
}

func TestDiffAssoc(t *testing.T) {
	c1 := Of([]int{1, 2, 3, 4})
	c2 := Of([]int{1, 2, 5, 6})
	diff := c1.DiffAssoc(c2)
	assert.Equal(t, []int{3, 4}, diff.All())
	assert.Equal(t, 0, Of([]int{}).DiffAssoc(c2).Count())
	assert.Equal(t, 4, c1.DiffAssoc(Of([]int{})).Count())
}

func TestEachSpread(t *testing.T) {
	count := 0
	c := Of([]int{1, 2, 3})
	c.EachSpread(func(items ...int) {
		count += len(items)
	})
	assert.Equal(t, 3, count)
}

func TestFlatMap(t *testing.T) {
	c := Of([]int{1, 2, 3})
	result := c.FlatMap(func(n int) []int { return []int{n, n * 2} })
	assert.Equal(t, []int{1, 2, 2, 4, 3, 6}, result.All())
	assert.Equal(t, 0, Of([]int{}).FlatMap(func(n int) []int { return []int{n} }).Count())
}

func TestForPage(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	assert.Equal(t, []int{1, 2, 3}, c.ForPage(1, 3).All())
	assert.Equal(t, []int{4, 5, 6}, c.ForPage(2, 3).All())
	assert.Equal(t, 0, c.ForPage(10, 3).Count())
	assert.Equal(t, 0, Of([]int{}).ForPage(1, 3).Count())
}

func TestIntersectByKeys(t *testing.T) {
	c1 := Of([]int{10, 20, 30, 40, 50})
	c2 := Of([]int{1, 2, 3})
	result := c1.IntersectByKeys(c2)
	assert.Equal(t, []int{10, 20, 30}, result.All())
	assert.Equal(t, 0, Of([]int{}).IntersectByKeys(c2).Count())
	assert.Equal(t, 0, c1.IntersectByKeys(Of([]int{})).Count())
}

func TestCollectionKeys(t *testing.T) {
	c := Of([]int{10, 20, 30})
	keys := c.Keys()
	assert.Equal(t, []int{0, 1, 2}, keys)
	assert.Equal(t, 0, len(Of([]int{}).Keys()))
}

func TestMapSpread(t *testing.T) {
	c := Of([]int{1, 2, 3})
	result := c.MapSpread(func(items ...int) int {
		sum := 0
		for _, v := range items {
			sum += v
		}
		return sum
	})
	assert.Equal(t, 3, result.Count())
	assert.Equal(t, 1, result.All()[0])
	assert.Equal(t, 2, result.All()[1])
	assert.Equal(t, 3, result.All()[2])
}

func TestMapToDictionary(t *testing.T) {
	type Product struct {
		Name string
		Type string
	}
	products := Of([]Product{Product{"Apple", "fruit"}, Product{"Carrot", "veg"}, Product{"Banana", "fruit"}})
	dict := products.MapToDictionary(func(p Product) string { return p.Type })
	assert.Equal(t, 2, len(dict["fruit"]))
	assert.Equal(t, 1, len(dict["veg"]))
	assert.Equal(t, 0, len(Of([]Product{}).MapToDictionary(func(p Product) string { return p.Type })))
}

func TestMedian(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	assert.Equal(t, 3.0, c.Median(func(n int) float64 { return float64(n) }))
	c2 := Of([]int{1, 2, 3, 4})
	assert.Equal(t, 2.5, c2.Median(func(n int) float64 { return float64(n) }))
	assert.Equal(t, 0.0, Of([]int{}).Median(func(n int) float64 { return float64(n) }))
}

func TestMode(t *testing.T) {
	type Item struct{ Category string }
	items := Of([]Item{Item{"A"}, Item{"B"}, Item{"A"}, Item{"C"}, Item{"A"}})
	modes := items.Mode(func(i Item) string { return i.Category })
	assert.Equal(t, 1, len(modes))
	assert.Equal(t, "A", modes[0])

	assert.Equal(t, 0, len(Of([]Item{}).Mode(func(i Item) string { return i.Category })))

	items = Of([]Item{Item{"A"}, Item{"B"}, Item{"A"}, Item{"B"}})
	modes = items.Mode(func(i Item) string { return i.Category })
	assert.Equal(t, 2, len(modes))
	assert.Contains(t, modes, "A")
	assert.Contains(t, modes, "B")
}

func TestNth(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	assert.Equal(t, []int{1, 3, 5, 7, 9}, c.Nth(2).All())
	assert.Equal(t, 0, c.Nth(0).Count())
	assert.Equal(t, 0, c.Nth(-1).Count())
	assert.Equal(t, 0, Of([]int{}).Nth(2).Count())
}

func TestPad(t *testing.T) {
	c := Of([]int{1, 2, 3})
	result := c.Pad(5, 0)
	assert.Equal(t, []int{1, 2, 3, 0, 0}, result.All())
	assert.Equal(t, 3, c.Pad(2, 0).Count())
}

func TestReplace(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	result := c.Replace(map[int]int{0: 10, 2: 30, 10: 100})
	assert.Equal(t, []int{10, 2, 30, 4, 5, 100}, result.All())
	assert.Equal(t, []int{1, 2, 3, 4, 5}, c.All())
}

func TestCollectionShuffle(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	result := c.Shuffle()
	assert.Equal(t, 5, result.Count())
	assert.Equal(t, []int{1, 2, 3, 4, 5}, c.All())
	assert.Equal(t, 0, Of([]int{}).Shuffle().Count())
}

func TestSkipUntil(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	result := c.SkipUntil(func(n int) bool { return n >= 3 })
	assert.Equal(t, []int{3, 4, 5}, result.All())
	assert.Equal(t, 0, c.SkipUntil(func(n int) bool { return n > 10 }).Count())
}

func TestSkipWhile(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	result := c.SkipWhile(func(n int) bool { return n < 3 })
	assert.Equal(t, []int{3, 4, 5}, result.All())
	assert.Equal(t, 0, c.SkipWhile(func(n int) bool { return n < 10 }).Count())
}

func TestSortDesc(t *testing.T) {
	c := Of([]int{3, 1, 4, 1, 5, 9, 2, 6})
	result := c.SortDesc(func(a, b int) bool { return a < b })
	assert.Equal(t, []int{9, 6, 5, 4, 3, 2, 1, 1}, result.All())
	assert.Equal(t, []int{3, 1, 4, 1, 5, 9, 2, 6}, c.All())
}

func TestSplice(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	removed := c.Splice(1, 2, 10, 20)
	assert.Equal(t, []int{2, 3}, removed.All())
	assert.Equal(t, []int{1, 10, 20, 4, 5}, c.All())

	removed = c.Splice(-2, 1, 99)
	assert.Equal(t, []int{4}, removed.All())
	assert.Equal(t, []int{1, 10, 20, 99, 5}, c.All())

	removed = c.Splice(2, -1, 8)
	assert.Equal(t, []int{}, removed.All())
	assert.Equal(t, []int{1, 10, 8, 20, 99, 5}, c.All())
}

func TestZip(t *testing.T) {
	// Equal length collections
	assert.Equal(t, [][]int{{1, 3}, {2, 4}}, Of([]int{1, 2}).Zip(Of([]int{3, 4})))

	// First collection is longer — trailing element forms a partial pair
	assert.Equal(t, [][]int{{1, 3}, {2}}, Of([]int{1, 2}).Zip(Of([]int{3})))

	// Second collection is longer — trailing element forms a partial pair
	assert.Equal(t, [][]int{{1, 2}, {3}}, Of([]int{1}).Zip(Of([]int{2, 3})))

	// Empty collections
	assert.Equal(t, 0, len(Of([]int{}).Zip(Of([]int{}))))
}

func TestCollectionSplit(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5, 6, 7})
	groups := c.Split(3)
	assert.Equal(t, 3, len(groups))
	assert.Equal(t, 3, len(groups[0]))
	assert.Equal(t, 1, len(groups[2]))
	assert.Equal(t, 0, len(Of([]int{}).Split(3)))
	assert.Equal(t, 0, len(c.Split(0)))
}

func TestTakeUntil(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	result := c.TakeUntil(func(n int) bool { return n >= 3 })
	assert.Equal(t, []int{1, 2}, result.All())
	assert.Equal(t, 5, c.TakeUntil(func(n int) bool { return n > 10 }).Count())
}

func TestTakeWhile(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	result := c.TakeWhile(func(n int) bool { return n < 3 })
	assert.Equal(t, []int{1, 2}, result.All())
	assert.Equal(t, 0, c.TakeWhile(func(n int) bool { return n > 10 }).Count())
}

func TestTransform(t *testing.T) {
	c := Of([]int{1, 2, 3, 4, 5})
	result := c.Transform(func(n int, _ int) int { return n * 2 })
	assert.Equal(t, []int{2, 4, 6, 8, 10}, result.All())
	assert.Same(t, c, result)
}

func TestCollectionNewFunction(t *testing.T) {
	c := New(1, 2, 3)
	assert.Equal(t, []int{1, 2, 3}, c.All())
}

func TestCollectionCollapse(t *testing.T) {
	items := Of([]any{1, []any{2, 3}, "x"}).Collapse().All()
	assert.Equal(t, []any{1, 2, 3, "x"}, items)
}

func TestCollectionConcat(t *testing.T) {
	left := Of([]int{1, 2})
	right := Of([]int{3, 4})
	assert.Equal(t, []int{1, 2, 3, 4}, left.Concat(right).All())
}

func TestCollectionDiffKeys(t *testing.T) {
	result := Of([]int{10, 20, 30, 40, 50}).DiffKeys(Of([]int{1, 2, 3})).All()
	assert.Equal(t, []int{40, 50}, result)
}

func TestCollectionEachUntil(t *testing.T) {
	visited := 0
	Of([]int{1, 2, 3, 4}).EachUntil(func(item int, _ int) bool {
		visited += item
		return item < 3
	})
	assert.Equal(t, 6, visited)
}

func TestCollectionMapToGroups(t *testing.T) {
	groups := Of([]int{1, 2, 3, 4}).MapToGroups(func(item int, index int) string {
		if (item+index)%2 == 0 {
			return "even"
		}
		return "odd"
	})
	assert.Equal(t, []int{1, 2, 3, 4}, groups["odd"].All())
}

func TestCollectionMapWithKeys(t *testing.T) {
	mapped := Of([]int{1, 2, 3}).MapWithKeys(func(item int) (string, int) {
		return string(rune('a' + item - 1)), item * 10
	})
	assert.Equal(t, map[string]int{"a": 10, "b": 20, "c": 30}, mapped)
}

func TestCollectionRandom(t *testing.T) {
	empty := Of([]int{}).Random()
	assert.Nil(t, empty)

	got := Of([]int{1, 2, 3}).Random()
	assert.NotNil(t, got)
	assert.Contains(t, []int{1, 2, 3}, *got)
}

func TestCollectionToJsonError(t *testing.T) {
	json, err := Of([]int{1, 2, 3}).ToJson()
	assert.NoError(t, err)
	assert.Equal(t, "[1,2,3]", json)

	type bad struct {
		Fn func()
	}
	_, err = Of([]bad{{Fn: func() {}}}).ToJson()
	assert.Error(t, err)
}

func TestCollectionUnion(t *testing.T) {
	result := Of([]int{1, 2, 2}).Union(Of([]int{2, 3, 4})).All()
	assert.Equal(t, []int{1, 2, 2, 3, 4}, result)
}

func TestCollectionUniqueByAdditional(t *testing.T) {
	type user struct {
		Name    string
		Country string
	}

	result := Of([]user{
		{Name: "alice", Country: "US"},
		{Name: "bob", Country: "US"},
		{Name: "charlie", Country: "UK"},
	}).UniqueBy(func(item user) string {
		return item.Country
	}).All()

	assert.Equal(t, []user{{Name: "alice", Country: "US"}, {Name: "charlie", Country: "UK"}}, result)
}

func TestCollectionUnlessAndWhenVariants(t *testing.T) {
	emptyCalled := false
	notEmptyCalled := false

	empty := Of([]int{})
	nonEmpty := Of([]int{1, 2})

	empty.UnlessEmpty(func(c *Collection[int]) *Collection[int] {
		emptyCalled = true
		return c.Push(10)
	})
	assert.False(t, emptyCalled)

	nonEmpty.UnlessEmpty(func(c *Collection[int]) *Collection[int] {
		notEmptyCalled = true
		return c.Push(3)
	})
	assert.True(t, notEmptyCalled)
	assert.Equal(t, []int{1, 2, 3}, nonEmpty.All())

	notEmptyCalled = false
	nonEmpty.UnlessNotEmpty(func(c *Collection[int]) *Collection[int] {
		notEmptyCalled = true
		return c
	})
	assert.False(t, notEmptyCalled)

	emptyCalled = false
	empty.WhenEmpty(func(c *Collection[int]) *Collection[int] {
		emptyCalled = true
		return c.Push(20)
	})
	assert.True(t, emptyCalled)

	notEmptyCalled = false
	nonEmpty.WhenNotEmpty(func(c *Collection[int]) *Collection[int] {
		notEmptyCalled = true
		return c.Push(4)
	})
	assert.True(t, notEmptyCalled)
	assert.Equal(t, []int{1, 2, 3, 4}, nonEmpty.All())
}

func TestCollectionWhereNullVariants(t *testing.T) {
	type user struct {
		Name      string
		Age       int
		DeletedAt *string
	}

	deleted := "yes"
	users := Of([]user{
		{Name: "alice", Age: 20, DeletedAt: nil},
		{Name: "bob", Age: 30, DeletedAt: &deleted},
		{Name: "charlie", Age: 40, DeletedAt: nil},
	})

	assert.Equal(t, []user{{Name: "bob", Age: 30, DeletedAt: &deleted}}, users.WhereNotNull("DeletedAt").All())
	assert.Equal(t, 2, users.WhereNull("DeletedAt").Count())
	assert.Equal(t, []user{{Name: "bob", Age: 30, DeletedAt: &deleted}}, users.WhereNotIn("Age", []any{20, 40}).All())
}

func TestGetFieldValue(t *testing.T) {
	type item struct {
		Name string
		Age  int
		Ptr  *string
	}

	s := "hello"
	obj := item{Name: "Alice", Age: 30, Ptr: &s}

	// existing field
	v := getFieldValue(obj, "Name")
	assert.NotNil(t, v)
	assert.Equal(t, "Alice", *v)

	// numeric field
	v = getFieldValue(obj, "Age")
	assert.NotNil(t, v)
	assert.Equal(t, 30, *v)

	// pointer field that is non-nil
	v = getFieldValue(obj, "Ptr")
	assert.NotNil(t, v)

	// pointer field that is nil
	obj.Ptr = nil
	v = getFieldValue(obj, "Ptr")
	assert.Nil(t, v)

	// pointer to struct
	v = getFieldValue(&obj, "Name")
	assert.NotNil(t, v)
	assert.Equal(t, "Alice", *v)

	// non-existent field
	v = getFieldValue(obj, "Missing")
	assert.Nil(t, v)

	// non-struct value
	v = getFieldValue(42, "whatever")
	assert.Nil(t, v)
}

func TestIsSimpleComparable(t *testing.T) {
	assert.True(t, isSimpleComparable(42))
	assert.True(t, isSimpleComparable(int8(1)))
	assert.True(t, isSimpleComparable(int16(1)))
	assert.True(t, isSimpleComparable(int32(1)))
	assert.True(t, isSimpleComparable(int64(1)))
	assert.True(t, isSimpleComparable(uint(1)))
	assert.True(t, isSimpleComparable(uint8(1)))
	assert.True(t, isSimpleComparable(uint16(1)))
	assert.True(t, isSimpleComparable(uint32(1)))
	assert.True(t, isSimpleComparable(uint64(1)))
	assert.True(t, isSimpleComparable(float32(1.0)))
	assert.True(t, isSimpleComparable(float64(1.0)))
	assert.True(t, isSimpleComparable("hello"))
	assert.True(t, isSimpleComparable(true))

	// nil is not simple comparable
	assert.False(t, isSimpleComparable(nil))

	// slice is not simple comparable
	assert.False(t, isSimpleComparable([]int{1, 2}))

	// struct is not simple comparable
	assert.False(t, isSimpleComparable(struct{ X int }{1}))
}

func TestCompareValues(t *testing.T) {
	// numeric comparisons
	assert.Equal(t, -1, compareValues(1, 2))
	assert.Equal(t, 0, compareValues(2, 2))
	assert.Equal(t, 1, compareValues(3, 2))

	// float numeric comparisons
	assert.Equal(t, -1, compareValues(1.5, 2.5))
	assert.Equal(t, 0, compareValues(2.5, 2.5))
	assert.Equal(t, 1, compareValues(3.5, 2.5))

	// mixed int / float
	assert.Equal(t, 0, compareValues(2, 2.0))
	assert.Equal(t, -1, compareValues(1, 1.5))

	// string comparisons (non-numeric)
	assert.Equal(t, -1, compareValues("apple", "banana"))
	assert.Equal(t, 0, compareValues("apple", "apple"))
	assert.Equal(t, 1, compareValues("banana", "apple"))
}

func TestValuesEqual(t *testing.T) {
	// identical types
	assert.True(t, valuesEqual(1, 1))
	assert.True(t, valuesEqual("hello", "hello"))
	assert.True(t, valuesEqual(1.5, 1.5))

	// cross-type numeric equality via string conversion
	assert.True(t, valuesEqual(int(2), float64(2.0)))
	assert.True(t, valuesEqual(int64(3), int(3)))

	// unequal values
	assert.False(t, valuesEqual(1, 2))
	assert.False(t, valuesEqual("a", "b"))

	// non-simple comparable (slice) falls back to DeepEqual
	assert.True(t, valuesEqual([]int{1, 2}, []int{1, 2}))
	assert.False(t, valuesEqual([]int{1, 2}, []int{1, 3}))

	// mixed simple and non-simple
	assert.False(t, valuesEqual(1, []int{1}))
}

func TestCompareFieldValue(t *testing.T) {
	type product struct {
		Name  string
		Price float64
		Tag   *string
	}

	tag := "sale"
	p := product{Name: "Widget", Price: 9.99, Tag: &tag}

	// equality operators
	assert.True(t, compareFieldValue(p, "Name", "=", "Widget"))
	assert.True(t, compareFieldValue(p, "Name", "==", "Widget"))
	assert.False(t, compareFieldValue(p, "Name", "=", "Gadget"))

	// inequality
	assert.True(t, compareFieldValue(p, "Name", "!=", "Gadget"))
	assert.False(t, compareFieldValue(p, "Name", "!=", "Widget"))

	// numeric comparison operators
	assert.True(t, compareFieldValue(p, "Price", ">", 5.0))
	assert.False(t, compareFieldValue(p, "Price", ">", 10.0))
	assert.True(t, compareFieldValue(p, "Price", ">=", 9.99))
	assert.True(t, compareFieldValue(p, "Price", "<", 20.0))
	assert.False(t, compareFieldValue(p, "Price", "<", 5.0))
	assert.True(t, compareFieldValue(p, "Price", "<=", 9.99))

	// like / not like
	assert.True(t, compareFieldValue(p, "Name", "like", "wid"))
	assert.True(t, compareFieldValue(p, "Name", "like", "WIDGET"))
	assert.False(t, compareFieldValue(p, "Name", "like", "xyz"))
	assert.True(t, compareFieldValue(p, "Name", "not like", "xyz"))
	assert.False(t, compareFieldValue(p, "Name", "not like", "widget"))

	// unknown operator
	assert.False(t, compareFieldValue(p, "Name", "~=", "Widget"))

	// value is nil: field is non-nil
	assert.False(t, compareFieldValue(p, "Name", "=", nil))
	assert.True(t, compareFieldValue(p, "Name", "!=", nil))
	assert.False(t, compareFieldValue(p, "Name", ">", nil))

	// fieldValue is nil (nil pointer field), value is non-nil
	p.Tag = nil
	assert.False(t, compareFieldValue(p, "Tag", "=", "sale"))
	assert.True(t, compareFieldValue(p, "Tag", "!=", "sale"))
	assert.False(t, compareFieldValue(p, "Tag", ">", "sale"))

	// both nil
	assert.True(t, compareFieldValue(p, "Tag", "=", nil))
	assert.False(t, compareFieldValue(p, "Tag", "!=", nil))
	assert.False(t, compareFieldValue(p, "Tag", ">", nil))
}
