// Пакет result предоставляет тип Result для обработки операций,
// которые могут завершиться успешно или с ошибкой.
// Это альтернатива стандартному подходу Go с возвратом (value, error),
// позволяющая строить цепочки вызовов и комбинировать результаты.
//
// Пример использования:
//
//	func parseAndProcess(input string) result.Result[int] {
//	    return result.Try(strconv.Atoi(input)).
//	        Map(func(x int) int { return x * 2 }).
//	        MapErr(func(err error) error {
//	            return fmt.Errorf("обработка failed: %w", err)
//	        })
//	}
package result

import (
	"errors"
	"fmt"
)

// Result представляет результат операции, которая может завершиться
// успешно (содержит значение типа T) или с ошибкой.
// Это монадический контейнер, вдохновленный аналогичными типами
// из Rust и других функциональных языков.
//
// T может быть любым типом (any).
//
// Использование вместо стандартного (T, error):
// - Позволяет строить цепочки операций через Map и MapErr
// - Упрощает комбинирование нескольких результатов
// - Делает обработку ошибок более декларативной
type Result[T any] struct {
	Value T
	Error error
}

// Unwrap возвращает значение и ошибку, содержащиеся в Result.
// Это обратное преобразование к конструкторам Ok, Err и Try.
//
// Пример:
//
//	val, err := r.Unwrap()
//	if err != nil {
//	    // обработка ошибки
//	}
func (r Result[T]) Unwrap() (T, error) {
	return r.Value, r.Error
}

// Must возвращает значение, содержащееся в Result.
// Если Result содержит ошибку, вызывает panic.
//
// Используйте только в случаях, когда ошибка невозможна
// или является фатальной для программы.
//
// Пример:
//
//	value := r.Must() // panic при ошибке
func (r Result[T]) Must() T {
	if r.Error != nil {
		panic(r.Error)
	}
	return r.Value
}

// IsOk возвращает true, если Result не содержит ошибки
// (операция завершилась успешно).
//
// Пример:
//
//	if result.IsOk() {
//	    // работаем со значением
//	}
func (r Result[T]) IsOk() bool {
	return r.Error == nil
}

// IsErr возвращает true, если Result содержит ошибку.
// Противоположность IsOk().
//
// Пример:
//
//	if result.IsErr() {
//	    // обрабатываем ошибку
//	}
func (r Result[T]) IsErr() bool {
	return r.Error != nil
}

// OrElse возвращает значение из Result, если нет ошибки.
// Если Result содержит ошибку, возвращает fallback.
//
// Полезно для задания значений по умолчанию.
//
// Пример:
//
//	value := result.OrElse(0) // 0 если была ошибка
func (r Result[T]) OrElse(fallback T) T {
	if r.Error != nil {
		return fallback
	}
	return r.Value
}

// Map применяет функцию f к значению в Result, если нет ошибки.
// Если Result содержит ошибку, возвращает её без изменений.
//
// Функция f должна принимать и возвращать один и тот же тип T.
// Для преобразования между разными типами используйте функции пакета.
//
// Пример:
//
//	result.Map(func(x int) int { return x * 2 })
func (r Result[T]) Map(f func(T) T) Result[T] {
	if r.Error != nil {
		return Result[T]{Error: r.Error}
	}
	return Result[T]{Value: f(r.Value)}
}

// MapErr применяет функцию f к ошибке в Result, если она есть.
// Если Result не содержит ошибки, возвращает его без изменений.
//
// Полезно для оборачивания ошибок в контекст.
//
// Пример:
//
//	result.MapErr(func(err error) error {
//	    return fmt.Errorf("в контексте: %w", err)
//	})
func (r Result[T]) MapErr(f func(error) error) Result[T] {
	if r.Error == nil {
		return r
	}
	return Result[T]{Value: r.Value, Error: f(r.Error)}
}

// AndThen применяет функцию, возвращающую Result,
// к значению в текущем Result.
// Позволяет строить цепочки операций, каждая из которых может вернуть ошибку.
//
// Пример:
//
//	result := result.Ok(42).
//	    AndThen(func(x int) result.Result[string] {
//	        return result.Try(strconv.Itoa(x))
//	    }).
//	    AndThen(func(s string) result.Result[int] {
//	        return result.Try(strconv.Atoi(s + "0"))
//	    })
func (r Result[T]) AndThen(f func(T) Result[T]) Result[T] {
	if r.Error != nil {
		return Result[T]{Error: r.Error}
	}
	return f(r.Value)
}

// Match выполняет одну из двух функций в зависимости от того,
// содержит Result значение или ошибку.
// Аналог match из Rust или case из функциональных языков.
//
// Пример:
//
//	message := result.Ok(42).Match(
//	    func(value int) string { return fmt.Sprintf("Значение: %d", value) },
//	    func(err error) string { return fmt.Sprintf("Ошибка: %v", err) },
//	)
func (r Result[T]) Match(onSuccess func(T) T, onError func(error) T) T {
	if r.Error != nil {
		return onError(r.Error)
	}
	return onSuccess(r.Value)
}

// WrapErr оборачивает ошибку в Result в новый контекст с сообщением.
// Если Result не содержит ошибки, возвращает его без изменений.
//
// Удобная альтернатива MapErr для простого добавления контекста.
//
// Пример:
//
//	result := result.Try(strconv.Atoi("invalid")).
//	    WrapErr("ошибка парсинга числа")
//
// Эквивалентно:
//
//	result.MapErr(func(err error) error {
//	    return fmt.Errorf("ошибка парсинга числа: %w", err)
//	})
func (r Result[T]) WrapErr(message string) Result[T] {
	if r.Error == nil || message == "" {
		return r
	}
	return Result[T]{Value: r.Value, Error: fmt.Errorf("%s: %w", message, r.Error)}
}

// WrapErrf оборачивает ошибку в Result с форматированием сообщения.
// Если Result не содержит ошибки, возвращает его без изменений.
//
// Пример:
//
//	result := result.Try(strconv.Atoi(input)).
//	    WrapErrf("не удалось распарсить '%s'", input)
func (r Result[T]) WrapErrf(format string, args ...any) Result[T] {
	if r.Error == nil || format == "" {
		return r
	}
	message := fmt.Sprintf(format, args...)
	return Result[T]{Value: r.Value, Error: fmt.Errorf("%s: %w", message, r.Error)}
}

// Ok создает успешный Result с указанным значением.
//
// Пример:
//
//	result := result.Ok(42)
//	result.IsOk() // true
func Ok[T any](value T) Result[T] {
	return Result[T]{Value: value}
}

// Err создает Result с указанной ошибкой.
// Значение будет нулевым для типа T.
//
// Пример:
//
//	result := result.Err[int](errors.New("ошибка"))
//	result.IsErr() // true
func Err[T any](err error) Result[T] {
	var zero T
	return Result[T]{Value: zero, Error: err}
}

// Try оборачивает стандартный возврат (value, error) в Result.
// Это основной способ создания Result из существующего кода.
//
// Пример:
//
//	// Вместо:
//	value, err := strconv.Atoi("42")
//	if err != nil {
//	    // обработка
//	}
//
//	// Используйте:
//	result := result.Try(strconv.Atoi("42"))
func Try[T any](value T, err error) Result[T] {
	return Result[T]{Value: value, Error: err}
}

// AndThen (FlatMap) применяет функцию, возвращающую Result,
// к значению в текущем Result.
// Позволяет строить цепочки операций, каждая из которых может вернуть ошибку.
//
// Пример:
//
//	result.AndThen(func(x int) result.Result[string] {
//	    return result.Try(strconv.Itoa(x))
//	})
func AndThen[T, U any](r Result[T], f func(T) Result[U]) Result[U] {
	if r.Error != nil {
		var zero U
		return Result[U]{Value: zero, Error: r.Error}
	}
	return f(r.Value)
}

// Match выполняет одну из двух функций в зависимости от того,
// содержит Result значение или ошибку.
// Аналог match из Rust или case из функциональных языков.
//
// Пример:
//
//	message := result.Match(
//	    func(value int) string { return fmt.Sprintf("Значение: %d", value) },
//	    func(err error) string { return fmt.Sprintf("Ошибка: %v", err) },
//	)
func Match[T, U any](r Result[T], onSuccess func(T) U, onError func(error) U) U {
	if r.Error != nil {
		return onError(r.Error)
	}
	return onSuccess(r.Value)
}

// Combine комбинирует два Result в один Result с парой значений.
// Возвращает ошибку, если любой из Results содержит ошибку.
// Если оба Results содержат ошибки, они объединяются через errors.Join.
//
// Пример:
//
//	func getUserAndPost(id int) result.Result[(User, Post)] {
//	    userRes := getUser(id)
//	    postRes := getPost(id)
//	    return result.Combine(userRes, postRes)
//	}
func Combine[T, U any](r1 Result[T], r2 Result[U]) Result[struct {
	First  T
	Second U
}] {
	// Обрабатываем случаи с ошибками
	if r1.Error != nil && r2.Error != nil {
		return Err[struct {
			First  T
			Second U
		}](errors.Join(r1.Error, r2.Error))
	}
	if r1.Error != nil {
		return Err[struct {
			First  T
			Second U
		}](r1.Error)
	}
	if r2.Error != nil {
		return Err[struct {
			First  T
			Second U
		}](r2.Error)
	}

	// Оба успешны
	return Ok(struct {
		First  T
		Second U
	}{
		First:  r1.Value,
		Second: r2.Value,
	})
}

// Combine3 комбинирует три Result (аналогично Combine)
func Combine3[T, U, V any](
	r1 Result[T],
	r2 Result[U],
	r3 Result[V],
) Result[struct {
	First  T
	Second U
	Third  V
}] {
	// Собираем все ошибки
	var errs []error
	if r1.Error != nil {
		errs = append(errs, r1.Error)
	}
	if r2.Error != nil {
		errs = append(errs, r2.Error)
	}
	if r3.Error != nil {
		errs = append(errs, r3.Error)
	}

	if len(errs) > 0 {
		return Err[struct {
			First  T
			Second U
			Third  V
		}](errors.Join(errs...))
	}

	return Ok(struct {
		First  T
		Second U
		Third  V
	}{
		First:  r1.Value,
		Second: r2.Value,
		Third:  r3.Value,
	})
}

// CombineSlice комбинирует срез Results
// Возвращает ошибку, если любой элемент содержит ошибку
func CombineSlice[T any](results []Result[T]) Result[[]T] {
	values := make([]T, 0, len(results))
	var errs []error

	for _, r := range results {
		if r.Error != nil {
			errs = append(errs, r.Error)
			// Можно продолжить собирать все ошибки
			// или прерваться на первой:
			// return Err[[]T](r.Error)
		} else {
			values = append(values, r.Value)
		}
	}

	if len(errs) > 0 {
		return Err[[]T](errors.Join(errs...))
	}

	return Ok(values)
}
