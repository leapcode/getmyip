package algo

import (
	"math/rand"
	"sort"

	"github.com/dropbox/godropbox/errors"
)

// [left, right]
func Partition(array sort.Interface, left int, right int,
	pivotIndex int) (int, error) {
	if pivotIndex < left || pivotIndex > right {
		return -1, errors.New("Pivot index out of range")
	}
	if left < 0 || right >= array.Len() {
		return -1, errors.New("Array index out of range")
	}
	lastIndex := right
	array.Swap(pivotIndex, lastIndex)
	leftIndex := left
	for i := left; i < lastIndex; i++ {
		if array.Less(i, lastIndex) {
			array.Swap(leftIndex, i)
			leftIndex++
		}
	}
	array.Swap(leftIndex, lastIndex)
	return leftIndex, nil
}

// k is started from 1
func QuickSelect(array sort.Interface, k int) error {
	if k <= 0 || k > array.Len() {
		return errors.New("Parameter k is invalid")
	}
	k--
	left := 0
	right := array.Len() - 1
	for {
		if left == right {
			return nil
		}
		pivotIndex := left + rand.Intn(right-left+1)
		pivotIndex, err := Partition(array, left, right, pivotIndex)
		if err != nil {
			return errors.Wrap(err, "Partition returns error")
		}
		if k == pivotIndex {
			return nil
		} else if k < pivotIndex {
			right = pivotIndex - 1
		} else {
			left = pivotIndex + 1
		}
	}
}
