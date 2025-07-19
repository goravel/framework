package collect

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestLazyCollect(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	lazy := LazyCollect(items)

	result := lazy.All()
	if !reflect.DeepEqual(result, items) {
		t.Errorf("Expected %v, got %v", items, result)
	}
}

func TestLazyNew(t *testing.T) {
	lazy := LazyNew(1, 2, 3, 4, 5)

	if lazy.Count() != 5 {
		t.Errorf("Expected count 5, got %d", lazy.Count())
	}
}

func TestLazyRange(t *testing.T) {
	lazy := LazyRange(1, 6)
	expected := []int{1, 2, 3, 4, 5}

	result := lazy.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyGenerate(t *testing.T) {
	lazy := LazyGenerate(func(i int) int {
		return i * 2
	}, 5)

	expected := []int{0, 2, 4, 6, 8}
	result := lazy.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyRepeat(t *testing.T) {
	lazy := LazyRepeat("hello", 3)
	expected := []string{"hello", "hello", "hello"}

	result := lazy.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyFilter(t *testing.T) {
	lazy := LazyRange(1, 11)
	filtered := lazy.Filter(func(n int, _ int) bool {
		return n%2 == 0
	})

	expected := []int{2, 4, 6, 8, 10}
	result := filtered.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyMap(t *testing.T) {
	lazy := LazyRange(1, 6)
	mapped := LazyMap(lazy, func(n int, _ int) int {
		return n * 2
	})

	expected := []int{2, 4, 6, 8, 10}
	result := mapped.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyReduce(t *testing.T) {
	lazy := LazyRange(1, 6)
	sum := LazyReduce(lazy, func(acc int, n int, _ int) int {
		return acc + n
	}, 0)

	expected := 15
	if sum != expected {
		t.Errorf("Expected %d, got %d", expected, sum)
	}
}

func TestLazyTake(t *testing.T) {
	lazy := LazyRange(1, 100)
	taken := lazy.Take(5)

	expected := []int{1, 2, 3, 4, 5}
	result := taken.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazySkip(t *testing.T) {
	lazy := LazyRange(1, 6)
	skipped := lazy.Skip(2)

	expected := []int{3, 4, 5}
	result := skipped.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyTakeWhile(t *testing.T) {
	lazy := LazyRange(1, 10)
	taken := lazy.TakeWhile(func(n int) bool {
		return n < 5
	})

	expected := []int{1, 2, 3, 4}
	result := taken.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyDropWhile(t *testing.T) {
	lazy := LazyRange(1, 10)
	dropped := lazy.DropWhile(func(n int) bool {
		return n < 5
	})

	expected := []int{5, 6, 7, 8, 9}
	result := dropped.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyChaining(t *testing.T) {
	lazy := LazyRange(1, 21)
	result := lazy.
		Filter(func(n int, _ int) bool { return n%2 == 0 }).
		Take(5).
		All()

	expected := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyUnique(t *testing.T) {
	lazy := LazyCollect([]int{1, 2, 2, 3, 3, 3, 4, 5})
	unique := lazy.Unique()

	expected := []int{1, 2, 3, 4, 5}
	result := unique.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyReverse(t *testing.T) {
	lazy := LazyRange(1, 6)
	reversed := lazy.Reverse()

	expected := []int{5, 4, 3, 2, 1}
	result := reversed.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazySort(t *testing.T) {
	lazy := LazyCollect([]int{3, 1, 4, 1, 5, 9, 2})
	sorted := lazy.Sort(func(a, b int) bool {
		return a < b
	})

	expected := []int{1, 1, 2, 3, 4, 5, 9}
	result := sorted.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyFlatMap(t *testing.T) {
	lazy := LazyRange(1, 4)
	flattened := lazy.FlatMap(func(n int) []int {
		return []int{n, n * 10}
	})

	expected := []int{1, 10, 2, 20, 3, 30}
	result := flattened.All()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyFirst(t *testing.T) {
	lazy := LazyRange(5, 10)
	first := lazy.First()

	if first == nil || *first != 5 {
		t.Errorf("Expected first element to be 5, got %v", first)
	}
}

func TestLazyLast(t *testing.T) {
	lazy := LazyRange(1, 6)
	last := lazy.Last()

	if last == nil || *last != 5 {
		t.Errorf("Expected last element to be 5, got %v", last)
	}
}

func TestLazyContains(t *testing.T) {
	lazy := LazyRange(1, 6)

	if !lazy.Contains(3) {
		t.Error("Expected collection to contain 3")
	}

	if lazy.Contains(10) {
		t.Error("Expected collection to not contain 10")
	}
}

func TestLazyEvery(t *testing.T) {
	lazy := LazyRange(2, 11)
	allEven := lazy.Filter(func(n int, _ int) bool {
		return n%2 == 0
	}).Every(func(n int) bool {
		return n%2 == 0
	})

	if !allEven {
		t.Error("Expected all filtered items to be even")
	}
}

func TestLazySome(t *testing.T) {
	lazy := LazyRange(1, 6)
	hasEven := lazy.Some(func(n int) bool {
		return n%2 == 0
	})

	if !hasEven {
		t.Error("Expected at least one item to be even")
	}
}

func TestLazySum(t *testing.T) {
	lazy := LazyRange(1, 6)
	sum := lazy.Sum(func(n int) float64 {
		return float64(n)
	})

	expected := 15.0
	if sum != expected {
		t.Errorf("Expected sum %f, got %f", expected, sum)
	}
}

func TestLazyAverage(t *testing.T) {
	lazy := LazyRange(1, 6)
	avg := lazy.Average(func(n int) float64 {
		return float64(n)
	})

	expected := 3.0
	if avg != expected {
		t.Errorf("Expected average %f, got %f", expected, avg)
	}
}

func TestLazyMin(t *testing.T) {
	lazy := LazyCollect([]int{3, 1, 4, 1, 5})
	min := lazy.Min(func(n int) float64 {
		return float64(n)
	})

	expected := 1.0
	if min != expected {
		t.Errorf("Expected min %f, got %f", expected, min)
	}
}

func TestLazyMax(t *testing.T) {
	lazy := LazyCollect([]int{3, 1, 4, 1, 5})
	max := lazy.Max(func(n int) float64 {
		return float64(n)
	})

	expected := 5.0
	if max != expected {
		t.Errorf("Expected max %f, got %f", expected, max)
	}
}

func TestLazyGroupBy(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	users := []User{
		{Name: "Alice", Age: 25},
		{Name: "Bob", Age: 30},
		{Name: "Charlie", Age: 25},
	}

	lazy := LazyCollect(users)
	groups := lazy.GroupBy(func(u User) string {
		return fmt.Sprintf("%d", u.Age)
	})

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	if groups["25"].Count() != 2 {
		t.Errorf("Expected 2 users aged 25, got %d", groups["25"].Count())
	}
}

func TestLazyPartition(t *testing.T) {
	lazy := LazyRange(1, 11)
	evens, odds := lazy.Partition(func(n int) bool {
		return n%2 == 0
	})

	expectedEvens := []int{2, 4, 6, 8, 10}
	expectedOdds := []int{1, 3, 5, 7, 9}

	if !reflect.DeepEqual(evens.All(), expectedEvens) {
		t.Errorf("Expected evens %v, got %v", expectedEvens, evens.All())
	}

	if !reflect.DeepEqual(odds.All(), expectedOdds) {
		t.Errorf("Expected odds %v, got %v", expectedOdds, odds.All())
	}
}

func TestLazyIterator(t *testing.T) {
	lazy := LazyRange(1, 6)
	iter := lazy.Iterator()

	expected := []int{1, 2, 3, 4, 5}
	var result []int

	for {
		value, ok := iter.Next()
		if !ok {
			break
		}
		result = append(result, value)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyCollect_Convert(t *testing.T) {
	lazy := LazyRange(1, 6)
	collection := lazy.Collect()

	expected := []int{1, 2, 3, 4, 5}
	result := collection.All()

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyIsEmpty(t *testing.T) {
	empty := LazyCollect([]int{})
	if !empty.IsEmpty() {
		t.Error("Expected empty collection to be empty")
	}

	notEmpty := LazyRange(1, 3)
	if notEmpty.IsEmpty() {
		t.Error("Expected non-empty collection to not be empty")
	}
}

func TestLazyIsNotEmpty(t *testing.T) {
	notEmpty := LazyRange(1, 3)
	if !notEmpty.IsNotEmpty() {
		t.Error("Expected non-empty collection to not be empty")
	}
}

func TestLazyWhen(t *testing.T) {
	lazy := LazyRange(1, 6)
	result := lazy.When(true, func(lc *LazyCollection[int]) *LazyCollection[int] {
		return lc.Filter(func(n int, _ int) bool { return n > 3 })
	})

	expected := []int{4, 5}
	if !reflect.DeepEqual(result.All(), expected) {
		t.Errorf("Expected %v, got %v", expected, result.All())
	}
}

func TestLazyFromChannel(t *testing.T) {
	ch := make(chan int, 5)
	ch <- 1
	ch <- 2
	ch <- 3
	ch <- 4
	ch <- 5
	close(ch)

	lazy := LazyFromChannel(ch)
	expected := []int{1, 2, 3, 4, 5}
	result := lazy.All()

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
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

	expected := []int{1, 2, 3, 4, 5}
	result := lazy.All()

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyChunk(t *testing.T) {
	lazy := LazyRange(1, 11)
	chunks := lazy.Chunk(3)

	expected := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}}
	if !reflect.DeepEqual(chunks, expected) {
		t.Errorf("Expected %v, got %v", expected, chunks)
	}
}

func TestLazyJoin(t *testing.T) {
	lazy := LazyRange(1, 6)
	joined := lazy.Join(",")

	expected := "1,2,3,4,5"
	if joined != expected {
		t.Errorf("Expected %s, got %s", expected, joined)
	}
}

func TestLazyToJSON(t *testing.T) {
	lazy := LazyRange(1, 4)
	json, err := lazy.ToJSON()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := "[1,2,3]"
	if json != expected {
		t.Errorf("Expected %s, got %s", expected, json)
	}
}

func TestLazyEach(t *testing.T) {
	lazy := LazyRange(1, 6)
	sum := 0

	lazy.Each(func(n int, _ int) {
		sum += n
	}).All() // Need to consume the lazy collection

	expected := 15
	if sum != expected {
		t.Errorf("Expected sum %d, got %d", expected, sum)
	}
}

func TestLazyForEach(t *testing.T) {
	lazy := LazyRange(1, 6)
	sum := 0

	lazy.ForEach(func(n int) {
		sum += n
	})

	expected := 15
	if sum != expected {
		t.Errorf("Expected sum %d, got %d", expected, sum)
	}
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
	if executed {
		t.Error("Expected generator to not execute until consumed")
	}

	// Now consume the result
	result := mapped.All()

	// Generator should have executed
	if !executed {
		t.Error("Expected generator to execute when consumed")
	}

	expected := []int{8, 10}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
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

	expected := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

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
	lc := LazyCollect(users)

	// Test 1: Two parameters (field, value) - implies '=' operator
	frenchUsers := lc.Where("Country", "FR")
	if frenchUsers.Count() != 2 {
		t.Errorf("Expected 2 French users, got %d", frenchUsers.Count())
	}

	youngUsers := lc.Where("Age", 25)
	if youngUsers.Count() != 2 {
		t.Errorf("Expected 2 users aged 25, got %d", youngUsers.Count())
	}

	// Test 2: Three parameters (field, operator, value)
	richUsers := lc.Where("Balance", ">", 100.0)
	if richUsers.Count() != 3 {
		t.Errorf("Expected 3 rich users, got %d", richUsers.Count())
	}

	nonFrenchUsers := lc.Where("Country", "!=", "FR")
	if nonFrenchUsers.Count() != 3 {
		t.Errorf("Expected 3 non-French users, got %d", nonFrenchUsers.Count())
	}

	seniorUsers := lc.Where("Age", ">=", 35)
	if seniorUsers.Count() != 2 {
		t.Errorf("Expected 2 senior users, got %d", seniorUsers.Count())
	}

	youngAdults := lc.Where("Age", "<", 35)
	if youngAdults.Count() != 3 {
		t.Errorf("Expected 3 young adults, got %d", youngAdults.Count())
	}

	// Test 3: Single parameter (callback function)
	customFilter := lc.Where(func(u User) bool {
		return u.Age > 25 && u.Country == "US"
	})
	if customFilter.Count() != 2 {
		t.Errorf("Expected 2 users matching custom filter, got %d", customFilter.Count())
	}

	// Test 4: Null comparisons
	activeUsers := lc.Where("DeletedAt", "=", nil)
	if activeUsers.Count() != 4 {
		t.Errorf("Expected 4 active users, got %d", activeUsers.Count())
	}

	deletedUsers := lc.Where("DeletedAt", "!=", nil)
	if deletedUsers.Count() != 1 {
		t.Errorf("Expected 1 deleted user, got %d", deletedUsers.Count())
	}

	// Test 5: String operations (like/not like)
	nameWithA := lc.Where("Name", "like", "a")
	if nameWithA.Count() != 3 { // Alice, Charlie, and David
		t.Errorf("Expected 3 users with 'a' in name, got %d", nameWithA.Count())
	}

	nameNotLikeTest := lc.Where("Name", "not like", "test")
	if nameNotLikeTest.Count() != 5 { // All users since none have 'test' in name
		t.Errorf("Expected 5 users without 'test' in name, got %d", nameNotLikeTest.Count())
	}
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
	lc := LazyCollect(users)

	// Test invalid parameter counts
	result := lc.Where()
	if result.Count() != 2 {
		t.Errorf("Expected original collection for no parameters, got %d items", result.Count())
	}

	result = lc.Where("Age", "=", 25, "extra")
	if result.Count() != 2 {
		t.Errorf("Expected original collection for too many parameters, got %d items", result.Count())
	}

	// Test invalid callback type
	result = lc.Where("not a callback")
	if result.Count() != 2 {
		t.Errorf("Expected original collection for invalid callback, got %d items", result.Count())
	}

	// Test invalid field name (non-string)
	result = lc.Where(123, 25)
	if result.Count() != 2 {
		t.Errorf("Expected original collection for invalid field name, got %d items", result.Count())
	}

	// Test invalid operator (non-string)
	result = lc.Where("Age", 123, 25)
	if result.Count() != 2 {
		t.Errorf("Expected original collection for invalid operator, got %d items", result.Count())
	}
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
	if executed {
		t.Error("Expected generator to not execute until consumed")
	}

	// Now consume the result
	result := filtered.Take(3).All()

	// Generator should have executed
	if !executed {
		t.Error("Expected generator to execute when consumed")
	}

	expected := []int{51, 52, 53}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLazyMapMethod(t *testing.T) {
	// Test with integers
	numbers := LazyRange(1, 6)

	// Test basic transformation
	doubled := numbers.Map(func(n int, _ int) interface{} {
		return n * 2
	})

	expected := []interface{}{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(doubled.All(), expected) {
		t.Errorf("Expected %v, got %v", expected, doubled.All())
	}

	// Test with index
	withIndex := numbers.Map(func(n int, i int) interface{} {
		return fmt.Sprintf("item_%d_%d", i, n)
	})

	expectedWithIndex := []interface{}{"item_0_1", "item_1_2", "item_2_3", "item_3_4", "item_4_5"}
	if !reflect.DeepEqual(withIndex.All(), expectedWithIndex) {
		t.Errorf("Expected %v, got %v", expectedWithIndex, withIndex.All())
	}

	// Test with structs
	type User struct {
		ID   int
		Name string
		Age  int
	}

	users := LazyCollect([]User{
		{ID: 1, Name: "Alice", Age: 25},
		{ID: 2, Name: "Bob", Age: 30},
		{ID: 3, Name: "Charlie", Age: 35},
	})

	// Extract names
	names := users.Map(func(u User, _ int) interface{} {
		return u.Name
	})

	expectedNames := []interface{}{"Alice", "Bob", "Charlie"}
	if !reflect.DeepEqual(names.All(), expectedNames) {
		t.Errorf("Expected %v, got %v", expectedNames, names.All())
	}

	// Complex transformation
	summaries := users.Map(func(u User, i int) interface{} {
		return map[string]interface{}{
			"id":      u.ID,
			"summary": fmt.Sprintf("%s (%d years)", u.Name, u.Age),
			"index":   i,
		}
	})

	if summaries.Count() != 3 {
		t.Errorf("Expected 3 summaries, got %d", summaries.Count())
	}

	// Test chaining with other methods
	result := numbers.
		Map(func(n int, _ int) interface{} {
			return n * 2
		}).
		Filter(func(item interface{}, _ int) bool {
			return item.(int) > 5
		})

	expectedChained := []interface{}{6, 8, 10}
	if !reflect.DeepEqual(result.All(), expectedChained) {
		t.Errorf("Expected %v, got %v", expectedChained, result.All())
	}
}

func TestLazyMapMethodTypes(t *testing.T) {
	// Test different return types
	numbers := LazyRange(1, 4)

	// Map to strings
	strings := numbers.Map(func(n int, _ int) interface{} {
		return fmt.Sprintf("number_%d", n)
	})

	expectedStrings := []interface{}{"number_1", "number_2", "number_3"}
	if !reflect.DeepEqual(strings.All(), expectedStrings) {
		t.Errorf("Expected %v, got %v", expectedStrings, strings.All())
	}

	// Map to booleans
	booleans := numbers.Map(func(n int, _ int) interface{} {
		return n%2 == 0
	})

	expectedBooleans := []interface{}{false, true, false}
	if !reflect.DeepEqual(booleans.All(), expectedBooleans) {
		t.Errorf("Expected %v, got %v", expectedBooleans, booleans.All())
	}

	// Test empty collection
	empty := LazyCollect([]int{})
	emptyMapped := empty.Map(func(n int, _ int) interface{} {
		return n * 2
	})

	if emptyMapped.Count() != 0 {
		t.Errorf("Expected empty mapped collection, got %d items", emptyMapped.Count())
	}
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
	if executed {
		t.Error("Expected generator to not execute until consumed")
	}

	// Now consume the result
	result := mapped.Take(3).All()

	// Generator should have executed
	if !executed {
		t.Error("Expected generator to execute when consumed")
	}

	expected := []interface{}{2, 4, 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
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

	expected := []interface{}{12, 14, 16, 18, 20}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Should be very fast since it stops after finding 5 items
	if duration > 100*time.Millisecond {
		t.Logf("Lazy Map evaluation took %v (should be fast)", duration)
	}
}
