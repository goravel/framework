package collect

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLazyCollect(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	lazy := LazyOf(items)

	assert.Equal(t, items, lazy.All())
}

func TestLazyNew(t *testing.T) {
	lazy := LazyNew(1, 2, 3, 4, 5)

	assert.Equal(t, 5, lazy.Count())
}

func TestLazyRange(t *testing.T) {
	lazy := LazyRange(1, 6)

	assert.Equal(t, []int{1, 2, 3, 4, 5}, lazy.All())
}

func TestLazyGenerate(t *testing.T) {
	lazy := LazyGenerate(func(i int) int {
		return i * 2
	}, 5)

	assert.Equal(t, []int{0, 2, 4, 6, 8}, lazy.All())
}

func TestLazyRepeat(t *testing.T) {
	lazy := LazyRepeat("hello", 3)

	assert.Equal(t, []string{"hello", "hello", "hello"}, lazy.All())
}

func TestLazyFilter(t *testing.T) {
	lazy := LazyRange(1, 11)
	filtered := lazy.Filter(func(n int, _ int) bool {
		return n%2 == 0
	})

	assert.Equal(t, []int{2, 4, 6, 8, 10}, filtered.All())
}

func TestLazyMap(t *testing.T) {
	lazy := LazyRange(1, 6)
	mapped := LazyMap(lazy, func(n int, _ int) int {
		return n * 2
	})

	assert.Equal(t, []int{2, 4, 6, 8, 10}, mapped.All())
}

func TestLazyReduce(t *testing.T) {
	lazy := LazyRange(1, 6)
	sum := LazyReduce(lazy, func(acc int, n int, _ int) int {
		return acc + n
	}, 0)

	assert.Equal(t, 15, sum)
}

func TestLazyTake(t *testing.T) {
	lazy := LazyRange(1, 100)

	assert.Equal(t, []int{1, 2, 3, 4, 5}, lazy.Take(5).All())
}

func TestLazySkip(t *testing.T) {
	lazy := LazyRange(1, 6)

	assert.Equal(t, []int{3, 4, 5}, lazy.Skip(2).All())
}

func TestLazyTakeWhile(t *testing.T) {
	lazy := LazyRange(1, 10)
	taken := lazy.TakeWhile(func(n int, _ int) bool {
		return n < 5
	})

	assert.Equal(t, []int{1, 2, 3, 4}, taken.All())
}

func TestLazyChaining(t *testing.T) {
	lazy := LazyRange(1, 21)
	result := lazy.
		Filter(func(n int, _ int) bool { return n%2 == 0 }).
		Take(5).
		All()

	assert.Equal(t, []int{2, 4, 6, 8, 10}, result)
}

func TestLazyUnique(t *testing.T) {
	lazy := LazyOf([]int{1, 2, 2, 3, 3, 3, 4, 5})

	assert.Equal(t, []int{1, 2, 3, 4, 5}, lazy.Unique().All())
}

func TestLazyReverse(t *testing.T) {
	lazy := LazyRange(1, 6)

	assert.Equal(t, []int{5, 4, 3, 2, 1}, lazy.Reverse().All())
}

func TestLazySort(t *testing.T) {
	lazy := LazyOf([]int{3, 1, 4, 1, 5, 9, 2})
	sorted := lazy.Sort(func(a, b int) bool {
		return a < b
	})

	assert.Equal(t, []int{1, 1, 2, 3, 4, 5, 9}, sorted.All())
}

func TestLazyFlatMap(t *testing.T) {
	lazy := LazyRange(1, 4)
	flattened := lazy.FlatMap(func(n int) []int {
		return []int{n, n * 10}
	})

	assert.Equal(t, []int{1, 10, 2, 20, 3, 30}, flattened.All())
}

func TestLazyFirst(t *testing.T) {
	lazy := LazyRange(5, 10)
	first := lazy.First()

	assert.NotNil(t, first)
	assert.Equal(t, 5, *first)
}

func TestLazyLast(t *testing.T) {
	lazy := LazyRange(1, 6)
	last := lazy.Last()

	assert.NotNil(t, last)
	assert.Equal(t, 5, *last)
}

func TestLazyContains(t *testing.T) {
	lazy := LazyRange(1, 6)

	assert.True(t, lazy.Contains(3))
	assert.False(t, lazy.Contains(10))
}

func TestLazyEvery(t *testing.T) {
	lazy := LazyRange(2, 11)
	allEven := lazy.Filter(func(n int, _ int) bool {
		return n%2 == 0
	}).Every(func(n int) bool {
		return n%2 == 0
	})

	assert.True(t, allEven)
}

func TestLazySum(t *testing.T) {
	lazy := LazyRange(1, 6)
	sum := lazy.Sum(func(n int) float64 {
		return float64(n)
	})

	assert.Equal(t, 15.0, sum)
}

func TestLazyMin(t *testing.T) {
	lazy := LazyOf([]int{3, 1, 4, 1, 5})
	min := lazy.Min(func(n int) float64 {
		return float64(n)
	})

	assert.Equal(t, 1.0, min)
}

func TestLazyMax(t *testing.T) {
	lazy := LazyOf([]int{3, 1, 4, 1, 5})
	max := lazy.Max(func(n int) float64 {
		return float64(n)
	})

	assert.Equal(t, 5.0, max)
}

func TestLazyPartition(t *testing.T) {
	lazy := LazyRange(1, 11)
	evens, odds := lazy.Partition(func(n int) bool {
		return n%2 == 0
	})

	assert.Equal(t, []int{2, 4, 6, 8, 10}, evens.All())
	assert.Equal(t, []int{1, 3, 5, 7, 9}, odds.All())
}

func TestLazyIterator(t *testing.T) {
	lazy := LazyRange(1, 6)
	iter := lazy.Iterator()

	var result []int
	for {
		value, ok := iter.Next()
		if !ok {
			break
		}
		result = append(result, value)
	}

	assert.Equal(t, []int{1, 2, 3, 4, 5}, result)
}

func TestLazyIteratorClosedChannel(t *testing.T) {
	ch := make(chan int)
	close(ch)

	iter := LazyFromChannel(ch).Iterator()
	_, ok := iter.Next()
	assert.False(t, ok)
}

func TestLazyCollect_Convert(t *testing.T) {
	lazy := LazyRange(1, 6)
	collection := lazy.Collect()

	assert.Equal(t, []int{1, 2, 3, 4, 5}, collection.All())
}

func TestLazyIsEmpty(t *testing.T) {
	empty := LazyOf([]int{})
	assert.True(t, empty.IsEmpty())

	notEmpty := LazyRange(1, 3)
	assert.False(t, notEmpty.IsEmpty())
}

func TestLazyIsNotEmpty(t *testing.T) {
	notEmpty := LazyRange(1, 3)
	assert.True(t, notEmpty.IsNotEmpty())
}

func TestLazyWhen(t *testing.T) {
	lazy := LazyRange(1, 6)
	result := lazy.When(true, func(lc *LazyCollection[int]) *LazyCollection[int] {
		return lc.Filter(func(n int, _ int) bool { return n > 3 })
	})

	assert.Equal(t, []int{4, 5}, result.All())
}

func TestLazyFromChannel(t *testing.T) {
	ch := make(chan int, 5)
	ch <- 1
	ch <- 2
	ch <- 3
	ch <- 4
	ch <- 5
	close(ch)

	assert.Equal(t, []int{1, 2, 3, 4, 5}, LazyFromChannel(ch).All())
}

func TestLazyEarlyReturnMethodsDrainChannel(t *testing.T) {
	assertDrained := func(t *testing.T, call func(*LazyCollection[int])) {
		source := make(chan int)
		done := make(chan struct{})
		go func() {
			defer close(done)
			for i := 1; i <= 5; i++ {
				source <- i
			}
			close(source)
		}()

		call(LazyFromChannel(source))

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("expected source channel sender to finish")
		}
	}

	assertDrained(t, func(lc *LazyCollection[int]) { _ = lc.First() })
	assertDrained(t, func(lc *LazyCollection[int]) { _ = lc.FirstWhere(func(item int) bool { return item == 2 }) })
	assertDrained(t, func(lc *LazyCollection[int]) { _ = lc.Every(func(item int) bool { return item < 2 }) })
	assertDrained(t, func(lc *LazyCollection[int]) { _ = lc.Take(1).All() })
	assertDrained(t, func(lc *LazyCollection[int]) { _ = lc.TakeWhile(func(item int, _ int) bool { return item < 2 }).All() })
}

func TestLazyFromFunc(t *testing.T) {
	lazy := LazyFromFunc(func() <-chan int {
		ch := make(chan int)
		go func() {
			defer close(ch)
			for i := 1; i <= 5; i++ {
				ch <- i
			}
		}()
		return ch
	})

	assert.Equal(t, []int{1, 2, 3, 4, 5}, lazy.All())
}

func TestLazyChunk(t *testing.T) {
	lazy := LazyRange(1, 11)

	assert.Equal(t, [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}}, lazy.Chunk(3))
}

func TestLazyJoin(t *testing.T) {
	lazy := LazyRange(1, 6)

	assert.Equal(t, "1,2,3,4,5", lazy.Join(","))
}

func TestLazyToJSON(t *testing.T) {
	lazy := LazyRange(1, 4)
	json, err := lazy.ToJson()

	assert.NoError(t, err)
	assert.Equal(t, "[1,2,3]", json)
}

func TestLazyEach(t *testing.T) {
	lazy := LazyRange(1, 6)
	sum := 0

	lazy.Each(func(n int, _ int) {
		sum += n
	}).All() // Need to consume the lazy collection

	assert.Equal(t, 15, sum)
}

func TestLazyEvaluation(t *testing.T) {
	// This test demonstrates that operations are lazy
	executed := false

	lazy := LazyFromFunc(func() <-chan int {
		ch := make(chan int)
		go func() {
			defer close(ch)
			executed = true
			for i := 1; i <= 5; i++ {
				ch <- i
			}
		}()
		return ch
	})

	// Operations should not execute the generator yet
	filtered := lazy.Filter(func(n int, _ int) bool { return n > 3 })
	mapped := LazyMap(filtered, func(n int, _ int) int { return n * 2 })

	// Generator should not have executed yet
	assert.False(t, executed, "Expected generator to not execute until consumed")

	// Now consume the result
	result := mapped.All()

	// Generator should have executed
	assert.True(t, executed, "Expected generator to execute when consumed")
	assert.Equal(t, []int{8, 10}, result)
}

func TestLazyPerformance(t *testing.T) {
	// Test that lazy evaluation is more efficient for large datasets
	// when only a small portion is needed

	start := time.Now()

	// Lazy approach - only processes what's needed
	lazy := LazyRange(1, 1000000)
	result := lazy.
		Filter(func(n int, _ int) bool { return n%2 == 0 }).
		Take(5).
		All()

	duration := time.Since(start)

	assert.Equal(t, []int{2, 4, 6, 8, 10}, result)

	// Should be very fast since it stops after finding 5 items
	if duration > 100*time.Millisecond {
		t.Logf("Lazy evaluation took %v (should be fast)", duration)
	}
}

// Enhanced Where method tests for LazyCollection
func TestLazyWhereEnhanced(t *testing.T) {
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
	lc := LazyOf(users)

	// Test 1: Two parameters (field, value) - implies '=' operator
	frenchUsers := lc.Where("Country", "FR")
	assert.Equal(t, 2, frenchUsers.Count())

	youngUsers := lc.Where("Age", 25)
	assert.Equal(t, 2, youngUsers.Count())

	// Test 2: Three parameters (field, operator, value)
	richUsers := lc.Where("Balance", ">", 100.0)
	assert.Equal(t, 3, richUsers.Count())

	nonFrenchUsers := lc.Where("Country", "!=", "FR")
	assert.Equal(t, 3, nonFrenchUsers.Count())

	seniorUsers := lc.Where("Age", ">=", 35)
	assert.Equal(t, 2, seniorUsers.Count())

	youngAdults := lc.Where("Age", "<", 35)
	assert.Equal(t, 3, youngAdults.Count())

	// Test 3: Single parameter (callback function)
	customFilter := lc.Where(func(u User) bool {
		return u.Age > 25 && u.Country == "US"
	})
	assert.Equal(t, 2, customFilter.Count())

	// Test 4: Null comparisons
	activeUsers := lc.Where("DeletedAt", "=", nil)
	assert.Equal(t, 4, activeUsers.Count())

	deletedUsers := lc.Where("DeletedAt", "!=", nil)
	assert.Equal(t, 1, deletedUsers.Count())

	// Test 5: String operations (like/not like)
	nameWithA := lc.Where("Name", "like", "a")
	assert.Equal(t, 3, nameWithA.Count()) // Alice, Charlie, and David

	nameNotLikeTest := lc.Where("Name", "not like", "test")
	assert.Equal(t, 5, nameNotLikeTest.Count()) // All users since none have 'test' in name
}

func TestLazyWhereErrorCases(t *testing.T) {
	type User struct {
		ID   int
		Name string
		Age  int
	}

	users := []User{
		{ID: 1, Name: "Alice", Age: 25},
		{ID: 2, Name: "Bob", Age: 30},
	}
	lc := LazyOf(users)

	// Test invalid parameter counts
	result := lc.Where()
	assert.Equal(t, 2, result.Count())

	result = lc.Where("Age", "=", 25, "extra")
	assert.Equal(t, 2, result.Count())

	// Test invalid callback type
	result = lc.Where("not a callback")
	assert.Equal(t, 2, result.Count())

	// Test invalid field name (non-string)
	result = lc.Where(123, 25)
	assert.Equal(t, 2, result.Count())

	// Test invalid operator (non-string)
	result = lc.Where("Age", 123, 25)
	assert.Equal(t, 2, result.Count())
}

func TestLazyWhereLazyEvaluation(t *testing.T) {
	// Test that Where operations are lazy and don't execute until consumed
	executed := false

	lazy := LazyFromFunc(func() <-chan int {
		ch := make(chan int)
		go func() {
			defer close(ch)
			executed = true
			for i := 1; i <= 100; i++ {
				ch <- i
			}
		}()
		return ch
	})

	// Apply where operations - should not execute yet
	filtered := lazy.Where(func(n int) bool { return n > 50 })

	// Generator should not have executed yet
	assert.False(t, executed, "Expected generator to not execute until consumed")

	// Now consume the result
	result := filtered.Take(3).All()

	// Generator should have executed
	assert.True(t, executed, "Expected generator to execute when consumed")
	assert.Equal(t, []int{51, 52, 53}, result)
}

func TestLazyMapMethod(t *testing.T) {
	// Test with integers
	numbers := LazyRange(1, 6)

	// Test basic transformation
	doubled := numbers.Map(func(n int, _ int) interface{} {
		return n * 2
	})

	assert.Equal(t, []interface{}{2, 4, 6, 8, 10}, doubled.All())

	// Test with index
	withIndex := numbers.Map(func(n int, i int) interface{} {
		return fmt.Sprintf("item_%d_%d", i, n)
	})

	assert.Equal(t, []interface{}{"item_0_1", "item_1_2", "item_2_3", "item_3_4", "item_4_5"}, withIndex.All())

	// Test with structs
	type User struct {
		ID   int
		Name string
		Age  int
	}

	users := LazyOf([]User{
		{ID: 1, Name: "Alice", Age: 25},
		{ID: 2, Name: "Bob", Age: 30},
		{ID: 3, Name: "Charlie", Age: 35},
	})

	// Extract names
	names := users.Map(func(u User, _ int) interface{} {
		return u.Name
	})

	assert.Equal(t, []interface{}{"Alice", "Bob", "Charlie"}, names.All())

	// Complex transformation
	summaries := users.Map(func(u User, i int) interface{} {
		return map[string]interface{}{
			"id":      u.ID,
			"summary": fmt.Sprintf("%s (%d years)", u.Name, u.Age),
			"index":   i,
		}
	})

	assert.Equal(t, 3, summaries.Count())

	// Test chaining with other methods
	result := numbers.
		Map(func(n int, _ int) interface{} {
			return n * 2
		}).
		Filter(func(item interface{}, _ int) bool {
			return item.(int) > 5
		})

	assert.Equal(t, []interface{}{6, 8, 10}, result.All())
}

func TestLazyMapMethodTypes(t *testing.T) {
	// Test different return types
	numbers := LazyRange(1, 4)

	// Map to strings
	strings := numbers.Map(func(n int, _ int) interface{} {
		return fmt.Sprintf("number_%d", n)
	})

	assert.Equal(t, []interface{}{"number_1", "number_2", "number_3"}, strings.All())

	// Map to booleans
	booleans := numbers.Map(func(n int, _ int) interface{} {
		return n%2 == 0
	})

	assert.Equal(t, []interface{}{false, true, false}, booleans.All())

	// Test empty collection
	empty := LazyOf([]int{})
	emptyMapped := empty.Map(func(n int, _ int) interface{} {
		return n * 2
	})

	assert.Equal(t, 0, emptyMapped.Count())
}

func TestLazyMapLazyEvaluation(t *testing.T) {
	// Test that Map operations are lazy and don't execute until consumed
	executed := false

	lazy := LazyFromFunc(func() <-chan int {
		ch := make(chan int)
		go func() {
			defer close(ch)
			executed = true
			for i := 1; i <= 100; i++ {
				ch <- i
			}
		}()
		return ch
	})

	// Apply map operations - should not execute yet
	mapped := lazy.Map(func(n int, _ int) interface{} {
		return n * 2
	})

	// Generator should not have executed yet
	assert.False(t, executed, "Expected generator to not execute until consumed")

	// Now consume the result
	result := mapped.Take(3).All()

	// Generator should have executed
	assert.True(t, executed, "Expected generator to execute when consumed")
	assert.Equal(t, []interface{}{2, 4, 6}, result)
}

func TestLazyMapPerformance(t *testing.T) {
	// Test that lazy Map is efficient for large datasets when only a portion is needed
	start := time.Now()

	lazy := LazyRange(1, 1000000)
	result := lazy.
		Map(func(n int, _ int) interface{} {
			return n * 2
		}).
		Filter(func(item interface{}, _ int) bool {
			return item.(int) > 10
		}).
		Take(5).
		All()

	duration := time.Since(start)

	assert.Equal(t, []interface{}{12, 14, 16, 18, 20}, result)

	// Should be very fast since it stops after finding 5 items
	if duration > 100*time.Millisecond {
		t.Logf("Lazy Map evaluation took %v (should be fast)", duration)
	}
}

func TestLazyZip(t *testing.T) {
	// Test zip with equal length collections
	assert.Equal(t, [][]int{{1, 4}, {2, 5}, {3, 6}},
		LazyOf([]int{1, 2, 3}).Zip(LazyOf([]int{4, 5, 6})))

	// Test zip with first collection shorter
	assert.Equal(t, [][]int{{1, 4}, {2, 5}},
		LazyOf([]int{1, 2}).Zip(LazyOf([]int{4, 5, 6, 7})))

	// Test zip with second collection shorter
	assert.Equal(t, [][]int{{1, 5}, {2, 6}, {3}, {4}},
		LazyOf([]int{1, 2, 3, 4}).Zip(LazyOf([]int{5, 6})))

	// Test with empty first collection
	assert.Empty(t, LazyOf([]int{}).Zip(LazyOf([]int{4, 5, 6})))

	// Test with empty second collection
	assert.Equal(t, [][]int{{1}, {2}, {3}},
		LazyOf([]int{1, 2, 3}).Zip(LazyOf([]int{})))

	// Test with string type
	assert.Equal(t, [][]string{{"a", "x"}, {"b", "y"}, {"c", "z"}},
		LazyOf([]string{"a", "b", "c"}).Zip(LazyOf([]string{"x", "y", "z"})))
}

func TestLazyAvg(t *testing.T) {
	assert.Equal(t, 3.0, LazyRange(1, 6).Avg(func(item int) float64 {
		return float64(item)
	}))

	assert.Equal(t, 0.0, LazyOf([]int{}).Avg(func(item int) float64 {
		return float64(item)
	}))
}

func TestLazyCountBy(t *testing.T) {
	counts := LazyOf([]string{"apple", "banana", "apricot", "blueberry", "avocado"}).CountBy(func(item string) string {
		return item[:1]
	})

	assert.Equal(t, map[string]int{"a": 3, "b": 2}, counts)
}

func TestLazyFirstWhereNotFound(t *testing.T) {
	first := LazyRange(1, 5).FirstWhere(func(item int) bool {
		return item > 100
	})

	assert.Nil(t, first)
}

func TestLazyFirstOnEmpty(t *testing.T) {
	assert.Nil(t, LazyOf([]int{}).First())
}

func TestLazyLastOnEmpty(t *testing.T) {
	assert.Nil(t, LazyOf([]int{}).Last())
}

func TestLazySkipWhile(t *testing.T) {
	assert.Equal(t, []int{4, 5, 6}, LazyRange(1, 7).SkipWhile(func(item int, _ int) bool {
		return item < 4
	}).All())

	assert.Empty(t, LazyRange(1, 4).SkipWhile(func(item int, _ int) bool {
		return item < 10
	}).All())
}

func TestLazySortBy(t *testing.T) {
	type item struct {
		Name string
	}

	sorted := LazyOf([]item{{Name: "charlie"}, {Name: "alice"}, {Name: "bob"}}).SortBy(func(v item) string {
		return v.Name
	}).All()

	assert.Equal(t, []item{{Name: "alice"}, {Name: "bob"}, {Name: "charlie"}}, sorted)
}

func TestLazyReject(t *testing.T) {
	rejected := LazyRange(1, 7).Reject(func(item int, _ int) bool {
		return item%2 == 0
	})

	assert.Equal(t, []int{1, 3, 5}, rejected.All())
}

func TestLazyPluck(t *testing.T) {
	type profile struct {
		Name string
		Age  int
	}

	users := []profile{{Name: "alice", Age: 20}, {Name: "bob", Age: 30}}
	assert.Equal(t, []any{"alice", "bob"}, LazyOf(users).Pluck("Name").All())
	assert.Equal(t, []any{20, 30}, LazyOf([]*profile{&users[0], &users[1]}).Pluck("Age").All())
	assert.Empty(t, LazyOf(users).Pluck("Unknown").All())
}

func TestLazyUniqueBy(t *testing.T) {
	type user struct {
		Name    string
		Country string
	}

	result := LazyOf([]user{
		{Name: "alice", Country: "US"},
		{Name: "bob", Country: "US"},
		{Name: "charlie", Country: "UK"},
	}).UniqueBy(func(item user) string {
		return item.Country
	}).All()

	assert.Equal(t, []user{{Name: "alice", Country: "US"}, {Name: "charlie", Country: "UK"}}, result)
}

func TestLazyWhenFalse(t *testing.T) {
	called := false
	result := LazyRange(1, 4).When(false, func(collection *LazyCollection[int]) *LazyCollection[int] {
		called = true
		return collection.Filter(func(item int, _ int) bool {
			return item > 1
		})
	})

	assert.False(t, called)
	assert.Equal(t, []int{1, 2, 3}, result.All())
}

func TestLazyWhereIn(t *testing.T) {
	type user struct {
		Name string
		Age  int
	}

	result := LazyOf([]user{{Name: "alice", Age: 20}, {Name: "bob", Age: 30}, {Name: "charlie", Age: 40}}).
		WhereIn("Age", []any{20, 40}).
		All()

	assert.Equal(t, []user{{Name: "alice", Age: 20}, {Name: "charlie", Age: 40}}, result)
}

func TestLazyWhereNotIn(t *testing.T) {
	type user struct {
		Name string
		Age  int
	}

	result := LazyOf([]user{{Name: "alice", Age: 20}, {Name: "bob", Age: 30}, {Name: "charlie", Age: 40}}).
		WhereNotIn("Age", []any{20, 40}).
		All()

	assert.Equal(t, []user{{Name: "bob", Age: 30}}, result)
}

func TestLazyChunkInvalidSize(t *testing.T) {
	assert.Equal(t, [][]int{}, LazyRange(1, 4).Chunk(0))
	assert.Equal(t, [][]int{}, LazyRange(1, 4).Chunk(-1))
}

func TestLazyIteratorReset(t *testing.T) {
	iter := LazyRange(1, 3).Iterator()

	first, ok := iter.Next()
	assert.True(t, ok)
	assert.Equal(t, 1, first)

	_, ok = iter.Next()
	assert.True(t, ok)

	_, ok = iter.Next()
	assert.False(t, ok)

	iter.Reset()
	reset, ok := iter.Next()
	assert.True(t, ok)
	assert.Equal(t, 1, reset)
}

func TestLazyTap(t *testing.T) {
	called := false
	result := LazyRange(1, 4).Tap(func(c *LazyCollection[int]) {
		called = true
		assert.Equal(t, []int{1, 2, 3}, c.All())
	})

	assert.True(t, called)
	assert.Equal(t, []int{1, 2, 3}, result.All())
}

func TestLazyToJsonError(t *testing.T) {
	type bad struct {
		Fn func()
	}

	_, err := LazyOf([]bad{{Fn: func() {}}}).ToJson()
	assert.Error(t, err)
}

func TestLazyIteratorNextDoneBranch(t *testing.T) {
	iter := LazyOf([]int{}).Iterator()

	_, ok := iter.Next()
	assert.False(t, ok)

	_, ok = iter.Next()
	assert.False(t, ok)
}

func TestLazyWhereInAndWhereNotInMissingField(t *testing.T) {
	type user struct {
		Name string
		Age  int
	}

	users := LazyOf([]user{{Name: "alice", Age: 20}, {Name: "bob", Age: 30}})
	assert.Empty(t, users.WhereIn("Unknown", []any{20}).All())
	assert.Equal(t, 2, users.WhereNotIn("Unknown", []any{20}).Count())
}
