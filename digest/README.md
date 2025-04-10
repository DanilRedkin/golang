# Go digest

## Необходимо поддержать следующие функции

```go
package go_digest

// GetCharByIndex returns the i-th character from the given string.
func GetCharByIndex(str string, idx int) rune

// GetStringBySliceOfIndexes returns a string formed by concatenating specific characters from the input string based
// on the provided indexes.
func GetStringBySliceOfIndexes(str string, indexes []int) string

// ShiftPointer shifts the given pointer by the specified number of bytes using unsafe.Add.
func ShiftPointer(pointer **int, shift int)

// IsComplexEqual compares two complex numbers and determines if they are equal.
func IsComplexEqual(a, b complex128) bool

// GetRootsOfQuadraticEquation returns two roots of a quadratic equation ax^2 + bx + c = 0.
func GetRootsOfQuadraticEquation(a, b, c float64) (complex128, complex128)

// Sort sorts in-place the given slice of integers in ascending order.
func Sort(source []int)

// ReverseSliceOne in-place reverses the order of elements in the given slice.
func ReverseSliceOne(s []int)

// ReverseSliceTwo returns a new slice of integers with elements in reverse order compared to the input slice.
// The original slice remains unmodified.
func ReverseSliceTwo(s []int) []int

// SwapPointers swaps the values of two pointers.
func SwapPointers(a, b *int)

// IsSliceEqual compares two slices of integers and returns true if they contain the same elements in the same order.
func IsSliceEqual(a, b []int) bool

// DeleteByIndex deletes the element at the specified index from the slice and returns a new slice.
// The original slice remains unmodified.
func DeleteByIndex(s []int, idx int) []int
```

## Сдача
* Все функции реализовать в файле [`main.go`](main.go)
* Открыть pull request из ветки `hw` в ветку `main` **вашего репозитория**
* Отправить заявку на ревью в соответствующей форме.
* Время для дедлайна фиксируется отправкой формы

## Makefile
Для удобств локальной разработки сделан [`Makefile`](Makefile). Имеются следующие команды:

Запустить полный цикл (линтер, тесты):
```bash 
make all
```

Запустить только тесты:
```bash
make test
``` 

Запустить линтер:
```bash
make lint
```

Подтянуть новые тесты:
```bash
make update
```

При разработке на Windows рекомендуется использовать [WSL](https://learn.microsoft.com/en-us/windows/wsl/install), чтобы была возможность пользоваться вспомогательными скриптами.
