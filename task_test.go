package main

import (
	"fmt"
	"testing"
	"time"
)

func TestData(t *testing.T) {
	const (
		integerVal     = 123456789
		zeroIntegerVal = 0

		floatVal     = 1234.56789
		zeroFloatVal = 0.0

		strVal      = "1234567890qwertyuiop!@#$%^&*()_+йцукенгшщзхъ"
		emptyStrVal = ""
	)
	var (
		result   interface{}
		status   bool
		myStruct = struct {
			a int
		}{
			integerVal,
		}
		testValues = []struct {
			val1    interface{}
			val2    interface{}
			zeroVal interface{}
		}{
			{integerVal, -integerVal, zeroIntegerVal},
			{floatVal, -floatVal, zeroFloatVal},
			{strVal, emptyStrVal, nil},
			{true, false, nil},
			{[10]interface{}{integerVal, -integerVal, zeroIntegerVal, floatVal, -floatVal, zeroFloatVal, strVal, emptyStrVal, true, false}, [0]interface{}{}, nil},
			{myStruct, struct{}{}, nil},
		}
	)
	for _, testValue := range testValues {
		cache := NewCache(6)

		cache.Add(testValue.val1, testValue.val1)
		result, status = cache.Get(testValue.val1)
		if status != true || result != testValue.val1 {
			xType := fmt.Sprintf("%T", testValue.val1)
			t.Errorf("The expected %T was not received", xType)
		}

		cache.Add(testValue.val2, testValue.val2)
		result, status = cache.Get(testValue.val2)
		if status != true || result != testValue.val2 {
			xType := fmt.Sprintf("%T", testValue.val2)
			t.Errorf("The expected %T was not received", xType)
		}

		cache.Add(testValue.val1, testValue.val2)
		result, status = cache.Get(testValue.val1)
		if status != true || result != testValue.val2 {
			xType := fmt.Sprintf("%T", testValue.val2)
			t.Errorf("The expected %s was not received", xType)
		}

		if testValue.zeroVal != nil {
			cache.Add(testValue.zeroVal, testValue.zeroVal)
			result, status = cache.Get(testValue.zeroVal)
			if status != true || result != testValue.zeroVal {
				xType := fmt.Sprintf("%T", testValue.zeroVal)
				t.Errorf("The expected %s was not received", xType)
			}

			cache.Add(testValue.zeroVal, testValue.val1)
			result, status = cache.Get(testValue.zeroVal)
			if status != true || result != testValue.val1 {
				xType := fmt.Sprintf("%T", testValue.val1)
				t.Errorf("The expected %s was not received", xType)
			}

			cache.Add(testValue.val1, testValue.zeroVal)
			result, status = cache.Get(testValue.val1)
			if status != true || result != testValue.zeroVal {
				xType := fmt.Sprintf("%T", testValue.zeroVal)
				t.Errorf("The expected %s was not received", xType)
			}
		}
	}
}

func TestAddWithTTL(t *testing.T) {
	const (
		testCap = 2
		time2   = time.Second * 2
		time5   = time.Second * 5
	)
	cache := NewCache(testCap)
	testValues := [testCap]int{0, 1}
	cache.AddWithTTL(testValues[0], testValues[0], time2)
	cache.AddWithTTL(testValues[1], testValues[1], time5)
	time.Sleep(1 * time.Second)
	val, ok := cache.Get(testValues[0])
	if !ok || val == nil {
		t.Errorf("the expected element has not expired, but it has not been detected")
	}
	val, ok = cache.Get(testValues[1])
	if !ok || val == nil {
		t.Errorf("the expected element has not expired, but it has not been detected")
	}

	time.Sleep(2 * time.Second)
	val, ok = cache.Get(testValues[0])
	if ok || val != nil {
		t.Errorf("An expired element has been detected")
	}
	val, ok = cache.Get(testValues[1])
	if !ok || val == nil {
		t.Errorf("the expected element has not expired, but it has not been detected")
	}

	time.Sleep(3 * time.Second)
	val, ok = cache.Get(testValues[1])
	if ok || val != nil {
		t.Errorf("An expired element has been detected")
	}
}

func TestDisplacement(t *testing.T) {
	const (
		testLen = 10
		newVal  = 100
	)
	cache := NewCache(testLen)
	testValues := [testLen]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, value := range testValues {
		cache.Add(value, value)
	}
	for _, value := range testValues {
		val, ok := cache.Get(value)
		if !ok || val != value {
			t.Errorf("expected non-displaced value is missing")
		}
	}

	cache.Add(newVal, newVal)
	val, ok := cache.Get(testValues[0])
	if val != nil || ok != false {
		t.Errorf("expected displaced value is not missing")
	}
	val, ok = cache.Get(newVal)
	if val != newVal || ok != true {
		t.Errorf("expected added value is missing")
	}
	for i := 1; i < testLen; i++ {
		val, ok = cache.Get(testValues[i])
		if !ok || val != testValues[i] {
			t.Errorf("expected non-displaced value is missing")
		}
	}
}

func TestCap(t *testing.T) {
	const testCap = 10
	cache := NewCache(testCap)
	resultCap := cache.Cap()
	if resultCap != testCap {
		t.Errorf("the element has an unexpected capacity")
	}
	for i := 0; i < testCap+1; i++ {
		cache.Add(i, i)
	}
	resultCap = cache.Cap()
	if resultCap != testCap {
		t.Errorf("the element has an unexpected capacity")
	}
}

func TestLen(t *testing.T) {
	const (
		Capacity  = 10
		testLen0  = 0
		testLen5  = 5
		testLen10 = 10
	)

	cache := NewCache(Capacity)
	resultLen := cache.Len()
	if resultLen != testLen0 {
		t.Errorf("the element has an unexpected capacity")
	}
	for i := 0; i < testLen5; i++ {
		cache.Add(i, i)
	}
	resultLen = cache.Len()
	if resultLen != testLen5 {
		t.Errorf("the element has an unexpected capacity")
	}

	cache = NewCache(Capacity)
	for i := 0; i < testLen10; i++ {
		cache.Add(i, i)
	}
	resultLen = cache.Len()
	if resultLen != testLen10 {
		t.Errorf("the element has an unexpected capacity")
	}
}

func TestClear(t *testing.T) {
	const (
		Capacity = 10
		testLen0 = 0
	)
	cache := NewCache(Capacity)
	for i := 0; i < Capacity; i++ {
		cache.Add(i, i)
	}
	cache.Clear()
	resultLen := cache.Len()
	if resultLen != testLen0 {
		t.Errorf("elements that were not deleted were detected or the element has an unexpected length")
	}
}

func TestRemove(t *testing.T) {
	const (
		Capacity     = 10
		DeletedValue = 5
	)
	testValues := [Capacity]int{0, 1, 2, 3, 4, DeletedValue, 6, 7, 8, 9}
	cache := NewCache(Capacity)
	for _, value := range testValues {
		cache.Add(value, value)
	}
	cache.Remove(DeletedValue)
	for _, value := range testValues {
		val, ok := cache.Get(value)
		if value == DeletedValue && (val != nil || ok != false) {
			t.Errorf("the expected deleted item has not been deleted")
		}
		if value != DeletedValue && (val != value || ok != true) {
			t.Errorf("expected value was not received")
		}

	}
}
