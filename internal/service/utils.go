package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/argon2"
)

// Ошибки, возвращаемые функциями хеширования паролей
var (
	// ErrPasswordTooLong возвращается, когда пароль превышает максимальную длину
	ErrPasswordTooLong = errors.New("password too long")
	// ErrPasswordTooShort возвращается, когда пароль короче минимальной длины
	ErrPasswordTooShort = errors.New("password too short")
	// ErrPasswordNoUpper возвращается, когда в пароле нет заглавных букв
	ErrPasswordNoUpper = errors.New("password must contain at least one uppercase letter")
	// ErrPasswordNoLower возвращается, когда в пароле нет строчных букв
	ErrPasswordNoLower = errors.New("password must contain at least one lowercase letter")
	// ErrPasswordNoDigit возвращается, когда в пароле нет цифр
	ErrPasswordNoDigit = errors.New("password must contain at least one digit")
	// ErrPasswordNoSpecial возвращается, когда в пароле нет спецсимволов
	ErrPasswordNoSpecial = errors.New("password must contain at least one special character")
)

// Максимальная длина пароля для предотвращения DoS-атак
// 72 байта - ограничение для большинства алгоритмов хеширования
const maxPasswordLength = 72

// Минимальная длина пароля для обеспечения базовой безопасности
const minPasswordLength = 8

// params содержит параметры для алгоритма Argon2id
type params struct {
	memory      uint32 // объем памяти в килобайтах
	iterations  uint32 // количество итераций
	parallelism uint8  // степень параллелизма (количество потоков)
	saltLength  uint32 // длина соли в байтах
	keyLength   uint32 // длина ключа в байтах
}

// Рекомендуемые параметры для продакшена
// - memory: 64 MB - достаточно для большинства случаев
// - iterations: 3 - оптимальное количество итераций
// - parallelism: 2 - использует 2 потока
// - saltLength: 16 - стандартная длина соли
// - keyLength: 32 - длина ключа 256 бит
var defaultParams = &params{
	memory:      64 * 1024, // 64 MB
	iterations:  3,
	parallelism: 2,
	saltLength:  16,
	keyLength:   32,
}

// generateRandomBytes генерирует криптостойкие случайные байты
//
// Параметры:
//   - n: количество байт для генерации
//
// Возвращает:
//   - []byte: срез со случайными байтами
//   - error: ошибка, если генерация не удалась
func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return b, nil
}

// validatePassword проверяет пароль на соответствие требованиям безопасности
//
// Требования:
//   - длина от 8 до 72 символов
//   - минимум одна заглавная буква
//   - минимум одна строчная буква
//   - минимум одна цифра
//   - минимум один спецсимвол
//
// Параметры:
//   - password: пароль для проверки
//
// Возвращает:
//   - error: nil если пароль валидный, иначе одна из ошибок валидации
func validatePassword(password string) error {
	if len(password) > maxPasswordLength {
		return ErrPasswordTooLong
	}
	if len(password) < minPasswordLength {
		return ErrPasswordTooShort
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return ErrPasswordNoUpper
	}
	if !hasLower {
		return ErrPasswordNoLower
	}
	if !hasDigit {
		return ErrPasswordNoDigit
	}
	if !hasSpecial {
		return ErrPasswordNoSpecial
	}

	return nil
}

// generateFromPassword создает хеш пароля с использованием Argon2id
//
// Параметры:
//   - password: пароль для хеширования
//   - p: параметры Argon2id
//
// Возвращает:
//   - string: закодированный хеш в формате "$argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>"
//   - error: ошибка, если хеширование не удалось
//
// Формат хеша:
//   - алгоритм: argon2id
//   - версия: 19
//   - параметры: m=память, t=итерации, p=параллелизм
//   - соль: base64
//   - хеш: base64
func generateFromPassword(password string, p *params) (encodedHash string, err error) {
	if err := validatePassword(password); err != nil {
		return "", err
	}

	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return a string using the standard encoded hash representation.
	encodedHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// hashPassword создает хеш пароля с параметрами по умолчанию
//
// Параметры:
//   - password: пароль для хеширования
//
// Возвращает:
//   - string: закодированный хеш
//   - error: ошибка валидации или хеширования
//
// Пример использования:
//
//	hash, err := hashPassword("MyP@ssw0rd")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(hash) // $argon2id$v=19$m=65536,t=3,p=2$...
func hashPassword(password string) (string, error) {
	return generateFromPassword(password, defaultParams)
}

// comparePasswordAndHash сравнивает пароль с хешем
//
// Параметры:
//   - password: пароль для проверки
//   - encodedHash: хеш в формате, созданном hashPassword
//
// Возвращает:
//   - bool: true если пароль соответствует хешу
//   - error: ошибка валидации или декодирования хеша
//
// Безопасность:
//   - использует ConstantTimeCompare для предотвращения timing-атак
//   - валидирует пароль перед сравнением
//
// Пример использования:
//
//	match, err := comparePasswordAndHash("MyP@ssw0rd", hash)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if match {
//	    fmt.Println("Password is correct")
//	}
func comparePasswordAndHash(password, encodedHash string) (match bool, err error) {
	if err := validatePassword(password); err != nil {
		return false, err
	}

	// Extract the parameters, salt and derived key from the encoded password hash.
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

// decodeHash декодирует хеш в параметры, соль и хеш
//
// Параметры:
//   - encodedHash: хеш в формате "$argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>"
//
// Возвращает:
//   - p: параметры Argon2id
//   - salt: соль в байтах
//   - hash: хеш в байтах
//   - error: ошибка декодирования
//
// Возможные ошибки:
//   - неверный формат хеша
//   - неподдерживаемый алгоритм
//   - несовместимая версия Argon2
//   - неверные параметры
//   - ошибки декодирования base64
func decodeHash(encodedHash string) (p *params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errors.New("the encoded hash is not in the correct format")
	}

	if vals[1] != "argon2id" {
		return nil, nil, nil, errors.New("unsupported algorithm")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse version: %w", err)
	}
	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("incompatible version of argon2: expected %d, got %d", argon2.Version, version)
	}

	p = &params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Validate parameters
	if p.memory == 0 || p.iterations == 0 || p.parallelism == 0 {
		return nil, nil, nil, errors.New("invalid parameters in hash")
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}
	if len(salt) == 0 {
		return nil, nil, nil, errors.New("salt cannot be empty")
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}
	if len(hash) == 0 {
		return nil, nil, nil, errors.New("hash cannot be empty")
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}
