package result_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/LigeronAhill/luxcarpets-go/pkg/result"
)

// TestOk тестирует конструктор Ok
func TestOk(t *testing.T) {
	t.Run("Ok с int", func(t *testing.T) {
		r := result.Ok(42)

		if !r.IsOk() {
			t.Errorf("Ok(42) должен быть успешным")
		}
		if r.IsErr() {
			t.Errorf("Ok(42) не должен содержать ошибку")
		}
		if r.Value != 42 {
			t.Errorf("Ok(42).Value = %v, ожидается 42", r.Value)
		}
		if r.Error != nil {
			t.Errorf("Ok(42).Error = %v, ожидается nil", r.Error)
		}
	})

	t.Run("Ok с string", func(t *testing.T) {
		r := result.Ok("test")

		if !r.IsOk() {
			t.Errorf("Ok(\"test\") должен быть успешным")
		}
		if r.Value != "test" {
			t.Errorf("Ok(\"test\").Value = %v, ожидается 'test'", r.Value)
		}
	})

	t.Run("Ok с nil значением", func(t *testing.T) {
		r := result.Ok[error](nil)

		if !r.IsOk() {
			t.Errorf("Ok(nil) должен быть успешным")
		}
		if r.Value != nil {
			t.Errorf("Ok(nil).Value = %v, ожидается nil", r.Value)
		}
	})
}

// TestErr тестирует конструктор Err
func TestErr(t *testing.T) {
	t.Run("Err с int", func(t *testing.T) {
		err := errors.New("test error")
		r := result.Err[int](err)

		if r.IsOk() {
			t.Errorf("Err(error) не должен быть успешным")
		}
		if !r.IsErr() {
			t.Errorf("Err(error) должен содержать ошибку")
		}
		if r.Value != 0 {
			t.Errorf("Err[int](error).Value = %v, ожидается 0", r.Value)
		}
		if r.Error != err {
			t.Errorf("Err(error).Error = %v, ожидается %v", r.Error, err)
		}
	})
}

// TestTry тестирует конструктор Try
func TestTry(t *testing.T) {
	t.Run("успешный Try", func(t *testing.T) {
		r := result.Try(42, nil)

		if !r.IsOk() {
			t.Error("Try(42, nil) должен быть успешным")
		}
		if r.Value != 42 {
			t.Errorf("Try(42, nil).Value = %v, ожидается 42", r.Value)
		}
	})

	t.Run("Try с ошибкой", func(t *testing.T) {
		err := errors.New("parse error")
		r := result.Try(0, err)

		if !r.IsErr() {
			t.Error("Try(0, error) должен содержать ошибку")
		}
		if r.Error != err {
			t.Errorf("Try(0, error).Error = %v, ожидается %v", r.Error, err)
		}
	})

	t.Run("Try имитирует вызов функции", func(t *testing.T) {
		value, err := strconv.Atoi("123")
		r := result.Try(value, err)

		if !r.IsOk() {
			t.Error("Try(strconv.Atoi(\"123\")) должен быть успешным")
		}
		if r.Value != 123 {
			t.Errorf("Try(strconv.Atoi(\"123\")).Value = %v, ожидается 123", r.Value)
		}
	})
}

// TestUnwrap тестирует метод Unwrap
func TestUnwrap(t *testing.T) {
	t.Run("успешный Unwrap", func(t *testing.T) {
		r := result.Ok("success")
		value, err := r.Unwrap()

		if err != nil {
			t.Errorf("Ok.Unwrap() вернул ошибку: %v", err)
		}
		if value != "success" {
			t.Errorf("Ok.Unwrap() вернул значение: %v, ожидается 'success'", value)
		}
	})

	t.Run("Unwrap с ошибкой", func(t *testing.T) {
		err := errors.New("failed")
		r := result.Err[string](err)
		value, actualErr := r.Unwrap()

		if actualErr != err {
			t.Errorf("Err.Unwrap() вернул ошибку: %v, ожидается %v", actualErr, err)
		}
		if value != "" {
			t.Errorf("Err.Unwrap() вернул значение: %v, ожидается пустая строка", value)
		}
	})
}

// TestMust тестирует метод Must
func TestMust(t *testing.T) {
	t.Run("успешный Must", func(t *testing.T) {
		r := result.Ok(100)
		value := r.Must()

		if value != 100 {
			t.Errorf("Must() вернул %v, ожидается 100", value)
		}
	})

	t.Run("Must с ошибкой - паника", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Must() должен вызвать панику при ошибке")
			}
		}()

		r := result.Err[int](errors.New("error"))
		_ = r.Must()
	})
}

// TestIsOkIsErr тестирует методы IsOk и IsErr
func TestIsOkIsErr(t *testing.T) {
	t.Run("IsOk на успешном Result", func(t *testing.T) {
		r := result.Ok(true)
		if !r.IsOk() {
			t.Error("IsOk() должен вернуть true для успешного Result")
		}
		if r.IsErr() {
			t.Error("IsErr() должен вернуть false для успешного Result")
		}
	})

	t.Run("IsErr на Result с ошибкой", func(t *testing.T) {
		r := result.Err[bool](errors.New("error"))
		if r.IsOk() {
			t.Error("IsOk() должен вернуть false для Result с ошибкой")
		}
		if !r.IsErr() {
			t.Error("IsErr() должен вернуть true для Result с ошибкой")
		}
	})
}

// TestOrElse тестирует метод OrElse
func TestOrElse(t *testing.T) {
	t.Run("успешный OrElse", func(t *testing.T) {
		r := result.Ok("value")
		result := r.OrElse("fallback")

		if result != "value" {
			t.Errorf("OrElse() вернул %v, ожидается 'value'", result)
		}
	})

	t.Run("OrElse с ошибкой", func(t *testing.T) {
		r := result.Err[string](errors.New("error"))
		result := r.OrElse("fallback")

		if result != "fallback" {
			t.Errorf("OrElse() вернул %v, ожидается 'fallback'", result)
		}
	})

	t.Run("OrElse с нулевым значением", func(t *testing.T) {
		r := result.Ok(0)
		result := r.OrElse(42)

		if result != 0 {
			t.Errorf("OrElse() вернул %v, ожидается 0", result)
		}
	})
}

// TestMap тестирует метод Map
func TestMap(t *testing.T) {
	t.Run("успешный Map", func(t *testing.T) {
		r := result.Ok(10)
		mapped := r.Map(func(x int) int { return x * 2 })

		if !mapped.IsOk() {
			t.Error("Map() на успешном Result должен быть успешным")
		}
		if mapped.Value != 20 {
			t.Errorf("Map() вернул %v, ожидается 20", mapped.Value)
		}
	})

	t.Run("Map с ошибкой", func(t *testing.T) {
		err := errors.New("original error")
		r := result.Err[int](err)
		mapped := r.Map(func(x int) int { return x * 2 })

		if !mapped.IsErr() {
			t.Error("Map() на Result с ошибкой должен содержать ошибку")
		}
		if mapped.Error != err {
			t.Errorf("Map() сохранил ошибку %v, ожидается %v", mapped.Error, err)
		}
	})

	t.Run("Map с тем же типом", func(t *testing.T) {
		r := result.Ok(42)
		mapped := r.Map(func(x int) int { return x * 10 })

		if !mapped.IsOk() {
			t.Error("Map() должен быть успешным")
		}
		if mapped.Value != 420 {
			t.Errorf("Map() вернул %v, ожидается 420", mapped.Value)
		}
	})
}

// TestMapErr тестирует метод MapErr
func TestMapErr(t *testing.T) {
	t.Run("MapErr на успешном Result", func(t *testing.T) {
		r := result.Ok(42)
		mapped := r.MapErr(func(err error) error {
			return fmt.Errorf("wrapped: %w", err)
		})

		if !mapped.IsOk() {
			t.Error("MapErr() на успешном Result должен остаться успешным")
		}
		if mapped.Value != 42 {
			t.Errorf("MapErr() изменил значение: %v, ожидается 42", mapped.Value)
		}
	})

	t.Run("MapErr с ошибкой", func(t *testing.T) {
		originalErr := errors.New("original")
		r := result.Err[int](originalErr)
		mapped := r.MapErr(func(err error) error {
			return fmt.Errorf("wrapped: %w", err)
		})

		if !mapped.IsErr() {
			t.Error("MapErr() на Result с ошибкой должен содержать ошибку")
		}
		if mapped.Error.Error() != "wrapped: original" {
			t.Errorf("MapErr() вернул ошибку: %v, ожидается 'wrapped: original'", mapped.Error)
		}
		if !errors.Is(mapped.Error, originalErr) {
			t.Error("MapErr() должен сохранять оригинальную ошибку через %w")
		}
	})
}

// TestAndThenMethod тестирует метод AndThen (теперь возвращает Result[T])
func TestAndThenMethod(t *testing.T) {
	t.Run("успешный AndThen (метод с тем же типом)", func(t *testing.T) {
		r := result.Ok(42)
		result := r.AndThen(func(x int) result.Result[int] {
			return result.Ok(x * 2)
		})

		if !result.IsOk() {
			t.Error("AndThen() должен быть успешным")
		}
		if result.Value != 84 {
			t.Errorf("AndThen() вернул %v, ожидается 84", result.Value)
		}
	})

	t.Run("AndThen с ошибкой в исходном Result (метод)", func(t *testing.T) {
		r := result.Err[int](errors.New("source error"))
		result := r.AndThen(func(x int) result.Result[int] {
			return result.Ok(x * 2)
		})

		if !result.IsErr() {
			t.Error("AndThen() с ошибкой в исходном Result должен содержать ошибку")
		}
		if result.Error.Error() != "source error" {
			t.Errorf("AndThen() вернул ошибку: %v, ожидается 'source error'", result.Error)
		}
	})

	t.Run("AndThen с ошибкой в функции (метод)", func(t *testing.T) {
		r := result.Ok(42)
		result := r.AndThen(func(x int) result.Result[int] {
			return result.Err[int](errors.New("function error"))
		})

		if !result.IsErr() {
			t.Error("AndThen() с ошибкой в функции должен содержать ошибку")
		}
		if result.Error.Error() != "function error" {
			t.Errorf("AndThen() вернул ошибку: %v, ожидается 'function error'", result.Error)
		}
	})

	t.Run("AndThen цепочка с тем же типом", func(t *testing.T) {
		r := result.Ok(10)
		result := r.
			AndThen(func(x int) result.Result[int] { return result.Ok(x * 2) }).
			AndThen(func(x int) result.Result[int] { return result.Ok(x + 5) })

		if !result.IsOk() {
			t.Error("Цепочка AndThen() должна быть успешной")
		}
		if result.Value != 25 { // (10 * 2) + 5
			t.Errorf("AndThen() цепочка вернула %v, ожидается 25", result.Value)
		}
	})
}

// TestAndThenFunction тестирует функцию AndThen (которая возвращает Result[U])
func TestAndThenFunction(t *testing.T) {
	t.Run("успешный AndThen (функция с разными типами)", func(t *testing.T) {
		r := result.Ok("123")
		result := result.AndThen(r, func(s string) result.Result[int] {
			return result.Try(strconv.Atoi(s))
		})

		if !result.IsOk() {
			t.Error("AndThen() должен быть успешным")
		}
		if result.Value != 123 {
			t.Errorf("AndThen() вернул %v, ожидается 123", result.Value)
		}
	})

	t.Run("AndThen с изменением типа (функция)", func(t *testing.T) {
		r := result.Ok(42)
		result := result.AndThen(r, func(x int) result.Result[string] {
			return result.Try(strconv.Itoa(x), nil)
		})

		if !result.IsOk() {
			t.Error("AndThen() должен быть успешным")
		}
		if result.Value != "42" {
			t.Errorf("AndThen() вернул %v, ожидается '42'", result.Value)
		}
	})

	t.Run("AndThen цепочка с разными типами", func(t *testing.T) {
		// Метод AndThen не работает с разными типами, поэтому используем функцию
		result := result.AndThen(
			result.AndThen(
				result.Ok("123"),
				func(s string) result.Result[int] { return result.Try(strconv.Atoi(s)) },
			),
			func(x int) result.Result[string] { return result.Try(strconv.Itoa(x*2), nil) },
		)

		if !result.IsOk() {
			t.Error("Цепочка AndThen() должна быть успешной")
		}
		if result.Value != "246" { // 123 * 2
			t.Errorf("AndThen() цепочка вернула %v, ожидается '246'", result.Value)
		}
	})
}

// TestMatchMethod тестирует метод Match (теперь возвращает T)
func TestMatchMethod(t *testing.T) {
	t.Run("Match на успешном Result (метод)", func(t *testing.T) {
		r := result.Ok(42)
		result := r.Match(
			func(value int) int { return value * 2 },
			func(err error) int { return -1 },
		)

		if result != 84 {
			t.Errorf("Match() вернул %v, ожидается 84", result)
		}
	})

	t.Run("Match на Result с ошибкой (метод)", func(t *testing.T) {
		r := result.Err[int](errors.New("failed"))
		result := r.Match(
			func(value int) int { return value * 2 },
			func(err error) int { return -1 },
		)

		if result != -1 {
			t.Errorf("Match() вернул %v, ожидается -1", result)
		}
	})

	t.Run("Match с обработкой ошибки", func(t *testing.T) {
		r := result.Err[int](errors.New("not found"))
		result := r.Match(
			func(value int) int { return value },
			func(err error) int {
				if err.Error() == "not found" {
					return 404
				}
				return 500
			},
		)

		if result != 404 {
			t.Errorf("Match() вернул %v, ожидается 404", result)
		}
	})
}

// TestMatchFunction тестирует функцию Match
func TestMatchFunction(t *testing.T) {
	t.Run("Match с сохранением типа (функция)", func(t *testing.T) {
		r := result.Ok(42)
		result := result.Match(r,
			func(value int) string { return fmt.Sprintf("value: %d", value) },
			func(err error) string { return fmt.Sprintf("error: %v", err) },
		)

		expected := "value: 42"
		if result != expected {
			t.Errorf("Match() вернул %v, ожидается %v", result, expected)
		}
	})

	t.Run("Match с преобразованием типа", func(t *testing.T) {
		r := result.Err[int](errors.New("division by zero"))
		result := result.Match(r,
			func(value int) float64 { return float64(value) },
			func(err error) float64 { return 0.0 },
		)

		if result != 0.0 {
			t.Errorf("Match() вернул %v, ожидается 0.0", result)
		}
	})
}

// TestCombine тестирует функцию Combine
func TestCombine(t *testing.T) {
	t.Run("Combine двух успешных Results", func(t *testing.T) {
		r1 := result.Ok("hello")
		r2 := result.Ok(42)

		result := result.Combine(r1, r2)

		if !result.IsOk() {
			t.Error("Combine() двух успешных Results должен быть успешным")
		}
		if result.Value.First != "hello" {
			t.Errorf("Combine().First = %v, ожидается 'hello'", result.Value.First)
		}
		if result.Value.Second != 42 {
			t.Errorf("Combine().Second = %v, ожидается 42", result.Value.Second)
		}
	})

	t.Run("Combine с ошибкой в первом Result", func(t *testing.T) {
		err1 := errors.New("error1")
		r1 := result.Err[string](err1)
		r2 := result.Ok(42)

		result := result.Combine(r1, r2)

		if !result.IsErr() {
			t.Error("Combine() с ошибкой в первом Result должен содержать ошибку")
		}
		if result.Error != err1 {
			t.Errorf("Combine() вернул ошибку: %v, ожидается %v", result.Error, err1)
		}
	})

	t.Run("Combine с ошибкой во втором Result", func(t *testing.T) {
		r1 := result.Ok("hello")
		err2 := errors.New("error2")
		r2 := result.Err[int](err2)

		result := result.Combine(r1, r2)

		if !result.IsErr() {
			t.Error("Combine() с ошибкой во втором Result должен содержать ошибку")
		}
		if result.Error != err2 {
			t.Errorf("Combine() вернул ошибку: %v, ожидается %v", result.Error, err2)
		}
	})

	t.Run("Combine с двумя ошибками", func(t *testing.T) {
		err1 := errors.New("error1")
		err2 := errors.New("error2")
		r1 := result.Err[string](err1)
		r2 := result.Err[int](err2)

		result := result.Combine(r1, r2)

		if !result.IsErr() {
			t.Error("Combine() с двумя ошибками должен содержать ошибку")
		}
		joinedErr := errors.Join(err1, err2)
		if result.Error.Error() != joinedErr.Error() {
			t.Errorf("Combine() вернул ошибку: %v, ожидается %v", result.Error, joinedErr)
		}
	})
}

// TestCombine3 тестирует функцию Combine3
func TestCombine3(t *testing.T) {
	t.Run("Combine3 трех успешных Results", func(t *testing.T) {
		r1 := result.Ok("a")
		r2 := result.Ok(1)
		r3 := result.Ok(true)

		result := result.Combine3(r1, r2, r3)

		if !result.IsOk() {
			t.Error("Combine3() трех успешных Results должен быть успешным")
		}
		if result.Value.First != "a" {
			t.Errorf("Combine3().First = %v, ожидается 'a'", result.Value.First)
		}
		if result.Value.Second != 1 {
			t.Errorf("Combine3().Second = %v, ожидается 1", result.Value.Second)
		}
		if result.Value.Third != true {
			t.Errorf("Combine3().Third = %v, ожидается true", result.Value.Third)
		}
	})
	t.Run("Combine3 с одной ошибкой", func(t *testing.T) {
		err := errors.New("error")
		r1 := result.Ok("a")
		r2 := result.Err[int](err)
		r3 := result.Ok(true)

		result := result.Combine3(r1, r2, r3)

		if !result.IsErr() {
			t.Error("Combine3() с ошибкой должен содержать ошибку")
		}
		// errors.Join создает новую ошибку, поэтому нужно проверять содержимое
		if result.Error == nil {
			t.Error("Combine3() должен вернуть ошибку")
		}
		// Проверяем что ошибка содержит оригинальную
		if !errors.Is(result.Error, err) {
			t.Errorf("Combine3() вернул ошибку которая не содержит оригинальную: %v", result.Error)
		}
	})

	t.Run("Combine3 с несколькими ошибками", func(t *testing.T) {
		err1 := errors.New("error1")
		err2 := errors.New("error2")
		r1 := result.Err[string](err1)
		r2 := result.Ok(1)
		r3 := result.Err[bool](err2)

		result := result.Combine3(r1, r2, r3)

		if !result.IsErr() {
			t.Error("Combine3() с ошибками должен содержать ошибку")
		}
		joinedErr := errors.Join(err1, err2)
		if result.Error.Error() != joinedErr.Error() {
			t.Errorf("Combine3() вернул ошибку: %v, ожидается %v", result.Error, joinedErr)
		}
	})
}

// TestCombineSlice тестирует функцию CombineSlice
func TestCombineSlice(t *testing.T) {
	t.Run("CombineSlice успешных Results", func(t *testing.T) {
		results := []result.Result[int]{
			result.Ok(1),
			result.Ok(2),
			result.Ok(3),
		}

		result := result.CombineSlice(results)

		if !result.IsOk() {
			t.Error("CombineSlice() успешных Results должен быть успешным")
		}
		if len(result.Value) != 3 {
			t.Errorf("CombineSlice() вернул срез длиной %v, ожидается 3", len(result.Value))
		}
		expected := []int{1, 2, 3}
		for i, v := range result.Value {
			if v != expected[i] {
				t.Errorf("CombineSlice()[%d] = %v, ожидается %v", i, v, expected[i])
			}
		}
	})

	t.Run("CombineSlice с ошибками", func(t *testing.T) {
		err1 := errors.New("error1")
		err2 := errors.New("error2")
		results := []result.Result[int]{
			result.Ok(1),
			result.Err[int](err1),
			result.Ok(3),
			result.Err[int](err2),
		}

		result := result.CombineSlice(results)

		if !result.IsErr() {
			t.Error("CombineSlice() с ошибками должен содержать ошибку")
		}
		joinedErr := errors.Join(err1, err2)
		if result.Error.Error() != joinedErr.Error() {
			t.Errorf("CombineSlice() вернул ошибку: %v, ожидается %v", result.Error, joinedErr)
		}
	})

	t.Run("CombineSlice пустого среза", func(t *testing.T) {
		var results []result.Result[int]
		result := result.CombineSlice(results)

		if !result.IsOk() {
			t.Error("CombineSlice() пустого среза должен быть успешным")
		}
		if len(result.Value) != 0 {
			t.Errorf("CombineSlice() пустого среза вернул срез длиной %v, ожидается 0", len(result.Value))
		}
	})
}

// TestIntegration тестирует интеграционные сценарии
func TestIntegration(t *testing.T) {
	t.Run("цепочка Map", func(t *testing.T) {
		successResult := result.Ok(10).
			Map(func(x int) int { return x * 2 }).
			Map(func(x int) int { return x + 5 })

		if !successResult.IsOk() {
			t.Error("Цепочка Map должна быть успешной")
		}
		if successResult.Value != 25 { // (10 * 2) + 5
			t.Errorf("Цепочка Map вернула %v, ожидается 25", successResult.Value)
		}
	})

	t.Run("цепочка AndThen (метод)", func(t *testing.T) {
		successResult := result.Ok(10).
			AndThen(func(x int) result.Result[int] { return result.Ok(x * 2) }).
			AndThen(func(x int) result.Result[int] { return result.Ok(x + 3) })

		if !successResult.IsOk() {
			t.Error("Цепочка AndThen должна быть успешной")
		}
		if successResult.Value != 23 { // (10 * 2) + 3
			t.Errorf("Цепочка AndThen вернула %v, ожидается 23", successResult.Value)
		}
	})

	t.Run("комбинация Map и AndThen", func(t *testing.T) {
		successResult := result.Ok(5).
			Map(func(x int) int { return x * 2 }).
			AndThen(func(x int) result.Result[int] {
				if x > 0 {
					return result.Ok(x)
				}
				return result.Err[int](errors.New("negative"))
			})

		if !successResult.IsOk() {
			t.Error("Комбинация должна быть успешной")
		}
		if successResult.Value != 10 {
			t.Errorf("Комбинация вернула %v, ожидается 10", successResult.Value)
		}
	})

	t.Run("реальный сценарий: парсинг и валидация", func(t *testing.T) {
		parseInput := func(input string) result.Result[int] {
			return result.Try(strconv.Atoi(input))
		}

		validatePositive := func(x int) result.Result[int] {
			if x > 0 {
				return result.Ok(x)
			}
			return result.Err[int](errors.New("число должно быть положительным"))
		}

		// Используем функцию AndThen для типобезопасности
		successResult := result.AndThen(
			parseInput("42"),
			validatePositive,
		)

		if !successResult.IsOk() {
			t.Error("Парсинг и валидация должны быть успешными")
		}
		if successResult.Value != 42 {
			t.Errorf("Результат %v, ожидается 42", successResult.Value)
		}

		// Тест с ошибкой
		errorResult := result.AndThen(
			parseInput("-10"),
			validatePositive,
		)

		if !errorResult.IsErr() {
			t.Error("Отрицательное число должно вызывать ошибку")
		}
	})

	t.Run("использование Match для форматирования", func(t *testing.T) {
		r := result.Ok(42)
		message := result.Match(
			r,
			func(value int) string { return fmt.Sprintf("Значение: %d", value) },
			func(err error) string { return fmt.Sprintf("Ошибка: %v", err) },
		)

		if message != "Значение: 42" {
			t.Errorf("Match() вернул %v, ожидается 'Значение: 42'", message)
		}
	})

	t.Run("использование Match с ошибкой", func(t *testing.T) {
		r := result.Err[int](errors.New("not found"))
		defaultValue := r.Match(
			func(value int) int { return value },
			func(err error) int { return 0 },
		)

		if defaultValue != 0 {
			t.Errorf("Match() с ошибкой вернул %v, ожидается 0", defaultValue)
		}
	})
}

// TestEdgeCases тестирует граничные случаи
func TestEdgeCases(t *testing.T) {
	t.Run("нулевое значение в успешном Result", func(t *testing.T) {
		r := result.Ok(0)
		if !r.IsOk() {
			t.Error("Ok(0) должен быть успешным")
		}
		if r.Value != 0 {
			t.Errorf("Ok(0).Value = %v, ожидается 0", r.Value)
		}
	})

	t.Run("Map с nil функцией (паника)", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Map с nil функцией должен вызывать панику")
			}
		}()

		r := result.Ok(42)
		var f func(int) int
		_ = r.Map(f)
	})

	t.Run("AndThen с nil функцией (паника)", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("AndThen с nil функцией должен вызывать панику")
			}
		}()

		r := result.Ok(42)
		var f func(int) result.Result[int]
		_ = r.AndThen(f)
	})
}

// Benchmark тесты для измерения производительности
func BenchmarkResultOperations(b *testing.B) {
	b.Run("Ok создание", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = result.Ok(i)
		}
	})

	b.Run("Map выполнение", func(b *testing.B) {
		r := result.Ok(42)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.Map(func(x int) int { return x * 2 })
		}
	})

	b.Run("AndThen метод", func(b *testing.B) {
		r := result.Ok(42)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.AndThen(func(x int) result.Result[int] { return result.Ok(x * 2) })
		}
	})

	b.Run("AndThen функция", func(b *testing.B) {
		r := result.Ok("123")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = result.AndThen(r, func(s string) result.Result[int] {
				return result.Try(strconv.Atoi(s))
			})
		}
	})

	b.Run("Combine двух Results", func(b *testing.B) {
		r1 := result.Ok("test")
		r2 := result.Ok(42)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = result.Combine(r1, r2)
		}
	})
}

// Example-тесты для документации
func ExampleOk() {
	r := result.Ok(42)
	fmt.Println(r.IsOk())
	// Output: true
}

func ExampleErr() {
	r := result.Err[int](errors.New("error"))
	fmt.Println(r.IsErr())
	// Output: true
}

func ExampleTry() {
	r := result.Try(strconv.Atoi("42"))
	value, _ := r.Unwrap()
	fmt.Println(value)
	// Output: 42
}

func ExampleResult_Map() {
	r := result.Ok(10).
		Map(func(x int) int { return x * 2 }).
		Map(func(x int) int { return x + 1 })
	value, _ := r.Unwrap()
	fmt.Println(value)
	// Output: 21
}

func ExampleResult_AndThen() {
	r := result.Ok(10).
		AndThen(func(x int) result.Result[int] { return result.Ok(x * 2) }).
		AndThen(func(x int) result.Result[int] { return result.Ok(x + 3) })
	value, _ := r.Unwrap()
	fmt.Println(value)
	// Output: 23
}

func ExampleResult_Match() {
	r := result.Ok(42)
	testResult := result.Match(
		r,
		func(value int) string { return fmt.Sprintf("Значение: %d", value) },
		func(err error) string { return fmt.Sprintf("Ошибка: %v", err) },
	)
	fmt.Println(testResult)
	// Output: Значение: 42
}

func ExampleAndThen() {
	testResult := result.AndThen(
		result.Ok("123"),
		func(s string) result.Result[int] {
			return result.Try(strconv.Atoi(s))
		},
	)
	value, _ := testResult.Unwrap()
	fmt.Println(value)
	// Output: 123
}

func ExampleMatch() {
	r := result.Ok(42)
	testResult := result.Match(
		r,
		func(value int) string { return fmt.Sprintf("value: %d", value) },
		func(err error) string { return fmt.Sprintf("error: %v", err) },
	)
	fmt.Println(testResult)
	// Output: value: 42
}

func ExampleCombine() {
	r1 := result.Ok("hello")
	r2 := result.Ok(42)
	testResult := result.Combine(r1, r2)
	value, _ := testResult.Unwrap()
	fmt.Printf("%s %d", value.First, value.Second)
	// Output: hello 42
}

// TestWrapErr тестирует метод WrapErr
func TestWrapErr(t *testing.T) {
	t.Run("WrapErr на успешном Result", func(t *testing.T) {
		r := result.Ok(42)
		wrapped := r.WrapErr("дополнительный контекст")

		if !wrapped.IsOk() {
			t.Error("WrapErr() на успешном Result должен остаться успешным")
		}
		if wrapped.Value != 42 {
			t.Errorf("WrapErr() изменил значение: %v, ожидается 42", wrapped.Value)
		}
	})

	t.Run("WrapErr с ошибкой", func(t *testing.T) {
		originalErr := errors.New("оригинальная ошибка")
		r := result.Err[int](originalErr)
		wrapped := r.WrapErr("не удалось выполнить операцию")

		if !wrapped.IsErr() {
			t.Error("WrapErr() на Result с ошибкой должен содержать ошибку")
		}
		expectedError := "не удалось выполнить операцию: оригинальная ошибка"
		if wrapped.Error.Error() != expectedError {
			t.Errorf("WrapErr() вернул ошибку: '%v', ожидается '%v'", wrapped.Error.Error(), expectedError)
		}
		if !errors.Is(wrapped.Error, originalErr) {
			t.Error("WrapErr() должен сохранять оригинальную ошибку через errors.Is")
		}
	})

	t.Run("WrapErr цепочка вызовов", func(t *testing.T) {
		originalErr := errors.New("базовая ошибка")
		r := result.Err[int](originalErr).
			WrapErr("первый уровень").
			WrapErr("второй уровень")

		if !r.IsErr() {
			t.Error("Цепочка WrapErr() должна сохранять ошибку")
		}
		expectedError := "второй уровень: первый уровень: базовая ошибка"
		if r.Error.Error() != expectedError {
			t.Errorf("Цепочка WrapErr() вернула: '%v', ожидается '%v'", r.Error.Error(), expectedError)
		}
		if !errors.Is(r.Error, originalErr) {
			t.Error("Цепочка WrapErr() должна сохранять оригинальную ошибку")
		}
	})

	t.Run("WrapErr после MapErr", func(t *testing.T) {
		originalErr := errors.New("исходная ошибка")
		r := result.Err[int](originalErr).
			MapErr(func(err error) error {
				return fmt.Errorf("перехвачено: %w", err)
			}).
			WrapErr("добавлен контекст")

		if !r.IsErr() {
			t.Error("Комбинация MapErr и WrapErr должна сохранять ошибку")
		}
		expectedError := "добавлен контекст: перехвачено: исходная ошибка"
		if r.Error.Error() != expectedError {
			t.Errorf("Комбинация вернула: '%v', ожидается '%v'", r.Error.Error(), expectedError)
		}
	})

	t.Run("WrapErr с пустым сообщением", func(t *testing.T) {
		originalErr := errors.New("ошибка")
		r := result.Err[int](originalErr).WrapErr("")

		if !r.IsErr() {
			t.Error("WrapErr с пустым сообщением должен сохранять ошибку")
		}
		// Пустое сообщение не должно добавлять префикс с двоеточием
		if r.Error.Error() != "ошибка" {
			t.Errorf("WrapErr с пустым сообщением вернул: '%v', ожидается 'ошибка'", r.Error.Error())
		}
	})
}

// TestWrapErrf тестирует метод WrapErrf
func TestWrapErrf(t *testing.T) {
	t.Run("WrapErrf на успешном Result", func(t *testing.T) {
		r := result.Ok("test")
		wrapped := r.WrapErrf("операция с %s", "аргументом")

		if !wrapped.IsOk() {
			t.Error("WrapErrf() на успешном Result должен остаться успешным")
		}
		if wrapped.Value != "test" {
			t.Errorf("WrapErrf() изменил значение: %v, ожидается 'test'", wrapped.Value)
		}
	})

	t.Run("WrapErrf с форматированием", func(t *testing.T) {
		originalErr := errors.New("ошибка чтения")
		input := "config.yaml"
		r := result.Err[int](originalErr)
		wrapped := r.WrapErrf("не удалось загрузить файл %s", input)

		if !wrapped.IsErr() {
			t.Error("WrapErrf() должен сохранять ошибку")
		}
		expectedError := "не удалось загрузить файл config.yaml: ошибка чтения"
		if wrapped.Error.Error() != expectedError {
			t.Errorf("WrapErrf() вернул: '%v', ожидается '%v'", wrapped.Error.Error(), expectedError)
		}
	})

	t.Run("WrapErrf с несколькими аргументами", func(t *testing.T) {
		originalErr := errors.New("connection failed")
		r := result.Err[string](originalErr)
		wrapped := r.WrapErrf("ошибка при подключении к %s:%d", "localhost", 5432)

		if !wrapped.IsErr() {
			t.Error("WrapErrf() с аргументами должен сохранять ошибку")
		}
		expectedError := "ошибка при подключении к localhost:5432: connection failed"
		if wrapped.Error.Error() != expectedError {
			t.Errorf("WrapErrf() вернул: '%v', ожидается '%v'", wrapped.Error.Error(), expectedError)
		}
	})

	t.Run("WrapErrf цепочка с разными форматами", func(t *testing.T) {
		originalErr := errors.New("permission denied")
		r := result.Err[int](originalErr).
			WrapErrf("файл %s", "data.txt").
			WrapErrf("попытка %d", 3)

		if !r.IsErr() {
			t.Error("Цепочка WrapErrf() должна сохранять ошибку")
		}
		expectedError := "попытка 3: файл data.txt: permission denied"
		if r.Error.Error() != expectedError {
			t.Errorf("Цепочка вернула: '%v', ожидается '%v'", r.Error.Error(), expectedError)
		}
	})

	t.Run("WrapErrf с специальными символами", func(t *testing.T) {
		originalErr := errors.New("invalid format")
		r := result.Err[int](originalErr)
		wrapped := r.WrapErrf("обработка JSON: ключ=%q, значение=%v", "name", "John")

		if !wrapped.IsErr() {
			t.Error("WrapErrf() со специальными символами должен сохранять ошибку")
		}
		expectedError := `обработка JSON: ключ="name", значение=John: invalid format`
		if wrapped.Error.Error() != expectedError {
			t.Errorf("WrapErrf() вернул: '%v', ожидается '%v'", wrapped.Error.Error(), expectedError)
		}
	})

	t.Run("WrapErrf с пустым форматом", func(t *testing.T) {
		originalErr := errors.New("ошибка")
		r := result.Err[int](originalErr).WrapErrf("")

		if !r.IsErr() {
			t.Error("WrapErrf с пустым форматом должен сохранять ошибку")
		}
		// Пустой формат не должен добавлять префикс
		if r.Error.Error() != "ошибка" {
			t.Errorf("WrapErrf с пустым форматом вернул: '%v', ожидается 'ошибка'", r.Error.Error())
		}
	})
}

// TestWrapErrIntegration тестирует интеграцию WrapErr с другими методами
func TestWrapErrIntegration(t *testing.T) {
	t.Run("реальный сценарий: загрузка конфигурации", func(t *testing.T) {
		// Имитация функции загрузки конфигурации
		loadConfig := func(filename string) result.Result[string] {
			if filename == "" {
				return result.Err[string](errors.New("пустое имя файла"))
			}
			return result.Ok("конфигурация загружена")
		}

		// Цепочка с WrapErr
		result := loadConfig("").
			WrapErr("инициализация приложения").
			MapErr(func(err error) error {
				return fmt.Errorf("критическая ошибка: %w", err)
			})

		if !result.IsErr() {
			t.Error("Цепочка должна завершиться ошибкой")
		}
		expectedError := "критическая ошибка: инициализация приложения: пустое имя файла"
		if result.Error.Error() != expectedError {
			t.Errorf("Интеграция вернула: '%v', ожидается '%v'", result.Error.Error(), expectedError)
		}
	})

	t.Run("комбинация AndThen и WrapErr", func(t *testing.T) {
		parseNumber := func(s string) result.Result[int] {
			return result.Try(strconv.Atoi(s))
		}

		validatePositive := func(n int) result.Result[int] {
			if n > 0 {
				return result.Ok(n)
			}
			return result.Err[int](errors.New("число должно быть положительным"))
		}

		// Используем функцию AndThen для типобезопасности
		testResult := result.AndThen(
			parseNumber("не число"),
			validatePositive,
		).WrapErr("обработка ввода пользователя")

		if !testResult.IsErr() {
			t.Error("Ожидается ошибка при парсинге нечисловой строки")
		}
		if !strings.Contains(testResult.Error.Error(), "обработка ввода пользователя") {
			t.Errorf("WrapErr должен добавлять контекст: %v", testResult.Error)
		}
	})

	t.Run("использование в цепочке Map", func(t *testing.T) {
		r := result.Ok("data").
			Map(func(s string) string { return s + " processed" }).
			WrapErr("обработка данных") // WrapErr на успешном результате не делает ничего

		if !r.IsOk() {
			t.Error("Цепочка с успешным Map должна остаться успешной")
		}
		if r.Value != "data processed" {
			t.Errorf("Цепочка вернула: %v, ожидается 'data processed'", r.Value)
		}
	})
}

// TestWrapErrEdgeCases тестирует граничные случаи
func TestWrapErrEdgeCases(t *testing.T) {
	t.Run("WrapErr с nil ошибкой", func(t *testing.T) {
		// Создаем Result с nil ошибкой (необычный случай)
		r := result.Result[int]{Value: 42, Error: nil}
		wrapped := r.WrapErr("контекст")

		if !wrapped.IsOk() {
			t.Error("WrapErr с nil ошибкой должен остаться успешным")
		}
		if wrapped.Value != 42 {
			t.Errorf("Значение изменилось: %v, ожидается 42", wrapped.Value)
		}
	})

	t.Run("WrapErrf с пустым форматом", func(t *testing.T) {
		originalErr := errors.New("ошибка")
		r := result.Err[int](originalErr).WrapErrf("")

		if !r.IsErr() {
			t.Error("WrapErrf с пустым форматом должен сохранять ошибку")
		}
		if r.Error.Error() != "ошибка" {
			t.Errorf("WrapErrf с пустым форматом вернул: '%v', ожидается 'ошибка'", r.Error.Error())
		}
	})

	t.Run("многократное оборачивание", func(t *testing.T) {
		baseErr := errors.New("базовая ошибка")
		r := result.Err[int](baseErr).
			WrapErr("уровень 1").
			WrapErr("уровень 2").
			WrapErr("уровень 3")

		if !r.IsErr() {
			t.Error("Многократное оборачивание должно сохранять ошибку")
		}
		expectedError := "уровень 3: уровень 2: уровень 1: базовая ошибка"
		if r.Error.Error() != expectedError {
			t.Errorf("Многократное оборачивание вернуло: '%v', ожидается '%v'", r.Error.Error(), expectedError)
		}
		// Проверяем что оригинальная ошибка доступна
		if !errors.Is(r.Error, baseErr) {
			t.Error("Многократное оборачивание должно сохранять доступ к оригинальной ошибке")
		}
	})
}

// Benchmark тесты для WrapErr
func BenchmarkWrapErr(b *testing.B) {
	b.Run("WrapErr без ошибки", func(b *testing.B) {
		r := result.Ok(42)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.WrapErr("контекст")
		}
	})

	b.Run("WrapErr с ошибкой", func(b *testing.B) {
		r := result.Err[int](errors.New("ошибка"))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.WrapErr("дополнительный контекст")
		}
	})

	b.Run("WrapErrf с ошибкой", func(b *testing.B) {
		r := result.Err[int](errors.New("ошибка"))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.WrapErrf("ошибка в %s", "функции")
		}
	})

	b.Run("цепочка WrapErr", func(b *testing.B) {
		r := result.Err[int](errors.New("базовая ошибка"))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.
				WrapErr("уровень 1").
				WrapErr("уровень 2").
				WrapErr("уровень 3")
		}
	})
}

// Example-тесты для документации
func ExampleResult_WrapErr() {
	// Симуляция функции, которая может вернуть ошибку
	parseConfig := func(data string) result.Result[int] {
		if data == "" {
			return result.Err[int](errors.New("пустые данные"))
		}
		return result.Ok(42)
	}

	// Использование WrapErr для добавления контекста
	result := parseConfig("").
		WrapErr("загрузка конфигурации")

	if result.IsErr() {
		fmt.Println(result.Error.Error())
	}
	// Output: загрузка конфигурации: пустые данные
}

func ExampleResult_WrapErrf() {
	// Симуляция функции работы с файлом
	readFile := func(filename string) result.Result[[]byte] {
		return result.Err[[]byte](errors.New("файл не найден"))
	}

	// Использование WrapErrf с форматированием
	result := readFile("config.yaml").
		WrapErrf("не удалось прочитать %s", "config.yaml")

	if result.IsErr() {
		fmt.Println(result.Error.Error())
	}
	// Output: не удалось прочитать config.yaml: файл не найден
}

func ExampleResult_WrapErr_chain() {
	// Демонстрация цепочки вызовов с WrapErr
	processResult := result.Err[int](errors.New("ошибка базы данных")).
		WrapErr("запрос к БД").
		WrapErr("обработка пользователя")

	if processResult.IsErr() {
		fmt.Println(processResult.Error.Error())
	}
	// Output: обработка пользователя: запрос к БД: ошибка базы данных
}

func ExampleResult_WrapErrf_formatting() {
	// Демонстрация форматирования с различными типами
	connectDB := func(host string, port int) result.Result[bool] {
		return result.Err[bool](errors.New("timeout"))
	}

	result := connectDB("localhost", 5432).
		WrapErrf("подключение к %s:%d", "localhost", 5432)

	if result.IsErr() {
		fmt.Println(result.Error.Error())
	}
	// Output: подключение к localhost:5432: timeout
}
