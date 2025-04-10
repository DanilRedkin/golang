package go_digest

import (
	"math/cmplx"
	"strings"
	"unsafe"
)

func GetCharByIndex(str string, idx int) rune {
	runeCounter := 0
	for _, r := range str {
		if runeCounter == idx {
			return r
		}
		runeCounter++
	}
	panic("Index out of range")
}

func GetStringBySliceOfIndexes(str string, indexes []int) string {
	var stringBuilder strings.Builder
	stringBuilder.Grow(len(indexes))
	for _, idx := range indexes {
		stringBuilder.WriteRune(GetCharByIndex(str, idx))
	}
	return stringBuilder.String()
}

func ShiftPointer(pointer **int, shift int) {
	if *pointer == nil {
		panic("Pointer is nil")
	}
	*pointer = (*int)(unsafe.Add(unsafe.Pointer(*pointer), uintptr(shift)))
}

const epsilon = 1e-6

func IsComplexEqual(a, b complex128) bool {
	if cmplx.IsInf(a) && cmplx.IsInf(b) {
		return real(a) == real(b) && imag(a) == imag(b)
	}
	return cmplx.Abs(a-b) < epsilon
}

func GetRootsOfQuadraticEquation(a, b, c float64) (complex128, complex128) {
	d := b*b - 4*a*c
	dSquared := cmplx.Sqrt(complex(d, 0))
	x1 := (-complex(b, 0) + dSquared) / (2 * complex(a, 0))
	x2 := (-complex(b, 0) - dSquared) / (2 * complex(a, 0))
	return x1, x2
}

var tmp []int

func Sort(source []int) {
	if len(source) < 2 {
		return
	}
	if len(tmp) < len(source) {
		tmp = make([]int, len(source))
	}
	mid := len(source) / 2
	Sort(source[:mid])
	Sort(source[mid:])
	merge(source, mid)
}

func merge(source []int, mid int) {
	copy(tmp, source)

	leftIdx, rightIdx, mergedIdx := 0, mid, 0
	for leftIdx < mid && rightIdx < len(source) {
		if tmp[leftIdx] < tmp[rightIdx] {
			source[mergedIdx] = tmp[leftIdx]
			leftIdx++
		} else {
			source[mergedIdx] = tmp[rightIdx]
			rightIdx++
		}
		mergedIdx++
	}

	for leftIdx < mid {
		source[mergedIdx] = tmp[leftIdx]
		leftIdx++
		mergedIdx++
	}

	for rightIdx < len(source) {
		source[mergedIdx] = tmp[rightIdx]
		rightIdx++
		mergedIdx++
	}
}

func ReverseSliceOne(s []int) {
	leftIdx, rightIdx := 0, len(s)-1

	for leftIdx < rightIdx {
		s[leftIdx], s[rightIdx] = s[rightIdx], s[leftIdx]
		leftIdx++
		rightIdx--
	}
}

func ReverseSliceTwo(s []int) []int {
	res := make([]int, len(s))
	copy(res, s)
	ReverseSliceOne(res)
	return res
}

func SwapPointers(a, b *int) {
	if a == nil && b == nil {
		panic("Pointers are nil")
	}
	if a == nil {
		panic("First pointer is nil")
	}
	if b == nil {
		panic("Second pointer is nil")
	}
	*a, *b = *b, *a
}

func IsSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, elem := range a {
		if elem != b[i] {
			return false
		}
	}
	return true
}

func DeleteByIndex(s []int, idx int) []int {
	if idx < 0 || idx >= len(s) {
		panic("Index out of range")
	}
	res := make([]int, len(s)-1)
	copy(res, s[:idx])
	copy(res[idx:], s[idx+1:])
	return res
}
