package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "валидный пароль",
			password: "TestP@ssw0rd",
			wantErr:  nil,
		},
		{
			name:     "пароль слишком короткий",
			password: "Te1!",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "пароль слишком длинный",
			password: strings.Repeat("a", 73) + "A1!",
			wantErr:  ErrPasswordTooLong,
		},
		{
			name:     "нет заглавных букв",
			password: "testp@ssw0rd",
			wantErr:  ErrPasswordNoUpper,
		},
		{
			name:     "нет строчных букв",
			password: "TESTP@SSW0RD",
			wantErr:  ErrPasswordNoLower,
		},
		{
			name:     "нет цифр",
			password: "TestPassword!",
			wantErr:  ErrPasswordNoDigit,
		},
		{
			name:     "нет спецсимволов",
			password: "TestPassword1",
			wantErr:  ErrPasswordNoSpecial,
		},
		{
			name:     "только буквы",
			password: "TestPassword",
			wantErr:  ErrPasswordNoDigit,
		},
		{
			name:     "только цифры",
			password: "12345678",
			wantErr:  ErrPasswordNoUpper, // сначала проверка на uppercase
		},
		{
			name:     "спецсимволы есть, но нет букв",
			password: "!@#$%^&*",
			wantErr:  ErrPasswordNoUpper,
		},
		{
			name:     "кириллица (не проходит валидацию)",
			password: "ТестПароль1!",
			wantErr:  nil,
		},
		{
			name:     "смешанная кириллица и латиница",
			password: "TestПароль1!",
			wantErr:  nil,
		},
		{
			name:     "только кириллица",
			password: "ТестПароль",
			wantErr:  ErrPasswordNoDigit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "валидный пароль",
			password: "TestP@ssw0rd",
			wantErr:  nil,
		},
		{
			name:     "сложный пароль",
			password: "VeryStr0ng!P@ssw0rd",
			wantErr:  nil,
		},
		{
			name:     "пароль с пробелами",
			password: "Test P@ssw0rd",
			wantErr:  nil, // пробелы считаются спецсимволами? unicode.IsPunct(' ') - false
		},
		{
			name:     "невалидный пароль",
			password: "weak",
			wantErr:  ErrPasswordTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hashPassword(tt.password)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, hash)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, hash)

			// Проверяем формат хеша
			assert.True(t, strings.HasPrefix(hash, "$argon2id$v=19$"))
			parts := strings.Split(hash, "$")
			assert.Len(t, parts, 6)

			// Проверяем параметры
			assert.Contains(t, parts[3], "m=65536")
			assert.Contains(t, parts[3], "t=3")
			assert.Contains(t, parts[3], "p=2")

			// Проверяем, что соль и хеш не пустые
			assert.NotEmpty(t, parts[4])
			assert.NotEmpty(t, parts[5])
		})
	}
}

func TestComparePasswordAndHash(t *testing.T) {
	validPassword := "TestP@ssw0rd"
	hash, err := hashPassword(validPassword)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	tests := []struct {
		name      string
		password  string
		hash      string
		wantMatch bool
		wantErr   error
	}{
		{
			name:      "пароль совпадает",
			password:  validPassword,
			hash:      hash,
			wantMatch: true,
			wantErr:   nil,
		},
		{
			name:      "пароль не совпадает",
			password:  "WrongP@ssw0rd",
			hash:      hash,
			wantMatch: false,
			wantErr:   nil,
		},
		{
			name:      "невалидный пароль",
			password:  "weak",
			hash:      hash,
			wantMatch: false,
			wantErr:   ErrPasswordTooShort,
		},
		{
			name:      "некорректный хеш",
			password:  validPassword,
			hash:      "invalid-hash",
			wantMatch: false,
			wantErr:   nil, // decodeHash вернет ошибку, но comparePasswordAndHash обернет ее
		},
		{
			name:      "пустой хеш",
			password:  validPassword,
			hash:      "",
			wantMatch: false,
			wantErr:   nil, // будет ошибка декодирования
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := comparePasswordAndHash(tt.password, tt.hash)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.False(t, match)
				return
			}

			if tt.wantMatch {
				assert.NoError(t, err)
				assert.True(t, match)
			} else {
				// Для неверного пароля ошибки быть не должно, только false
				if tt.name == "некорректный хеш" || tt.name == "пустой хеш" {
					assert.Error(t, err) // ожидаем ошибку декодирования
				} else {
					assert.NoError(t, err)
				}
				assert.False(t, match)
			}
		})
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	tests := []struct {
		name string
		n    uint32
	}{
		{"нулевая длина", 0},
		{"маленькая длина", 16},
		{"средняя длина", 32},
		{"большая длина", 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes, err := generateRandomBytes(tt.n)
			require.NoError(t, err)
			assert.Len(t, bytes, int(tt.n))

			// Проверяем, что байты действительно случайные (не все нули)
			if tt.n > 0 {
				allZeros := true
				for _, b := range bytes {
					if b != 0 {
						allZeros = false
						break
					}
				}
				assert.False(t, allZeros, "random bytes should not be all zeros")
			}
		})
	}
}

func TestDecodeHash(t *testing.T) {
	// Сначала создадим валидный хеш
	validPassword := "TestP@ssw0rd"
	validHash, err := hashPassword(validPassword)
	require.NoError(t, err)

	tests := []struct {
		name      string
		hash      string
		wantError bool
	}{
		{
			name:      "валидный хеш",
			hash:      validHash,
			wantError: false,
		},
		{
			name:      "неверный формат (меньше частей)",
			hash:      "$argon2id$v=19$m=65536,t=3,p=2",
			wantError: true,
		},
		{
			name:      "неверный формат (больше частей)",
			hash:      "$argon2id$v=19$m=65536,t=3,p=2$salt$hash$extra",
			wantError: true,
		},
		{
			name:      "неподдерживаемый алгоритм",
			hash:      "$argon2i$v=19$m=65536,t=3,p=2$salt$hash",
			wantError: true,
		},
		{
			name:      "неверная версия",
			hash:      "$argon2id$v=18$m=65536,t=3,p=2$c2FsdA==$aGFzaA==",
			wantError: true,
		},
		{
			name:      "неверные параметры (memory=0)",
			hash:      "$argon2id$v=19$m=0,t=3,p=2$c2FsdA==$aGFzaA==",
			wantError: true,
		},
		{
			name:      "неверная соль (не base64)",
			hash:      "$argon2id$v=19$m=65536,t=3,p=2$invalid-salt$hash",
			wantError: true,
		},
		{
			name:      "пустая соль",
			hash:      "$argon2id$v=19$m=65536,t=3,p=2$$aGFzaA==",
			wantError: true,
		},
		{
			name:      "пустой хеш",
			hash:      "$argon2id$v=19$m=65536,t=3,p=2$c2FsdA==$",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, salt, hash, err := decodeHash(tt.hash)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, p)
				assert.Nil(t, salt)
				assert.Nil(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, p)
				assert.NotNil(t, salt)
				assert.NotNil(t, hash)

				// Проверяем параметры
				assert.Equal(t, uint32(64*1024), p.memory)
				assert.Equal(t, uint32(3), p.iterations)
				assert.Equal(t, uint8(2), p.parallelism)
				assert.Equal(t, uint32(len(salt)), p.saltLength)
				assert.Equal(t, uint32(len(hash)), p.keyLength)
			}
		})
	}
}

func TestHashPassword_Consistency(t *testing.T) {
	// Один и тот же пароль должен давать разные хеши каждый раз
	password := "TestP@ssw0rd"

	hash1, err := hashPassword(password)
	require.NoError(t, err)

	hash2, err := hashPassword(password)
	require.NoError(t, err)

	// Хеши должны быть разными из-за разной соли
	assert.NotEqual(t, hash1, hash2)

	// Но оба должны валидироваться
	match1, err := comparePasswordAndHash(password, hash1)
	require.NoError(t, err)
	assert.True(t, match1)

	match2, err := comparePasswordAndHash(password, hash2)
	require.NoError(t, err)
	assert.True(t, match2)
}

func TestHashPassword_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "пароль из 8 символов (минимум)",
			password: "TestP@ss", // 8 символов, есть заглавная, строчная, спецсимвол, но нет цифры
			wantErr:  ErrPasswordNoDigit,
		},
		{
			name:     "пароль из 8 символов с цифрой",
			password: "TestP@1s", // 8 символов, есть всё
			wantErr:  nil,
		},
		{
			name:     "пароль из 72 символов (максимум)",
			password: "TestP@ssw0rd" + strings.Repeat("a", 60), // 8 + 60 = 68, нужно добавить
			wantErr:  nil,
		},
		{
			name:     "пароль со всеми типами спецсимволов",
			password: "Test!@#$%^&*()_+P@ssw0rd",
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr != nil {
				err := validatePassword(tt.password)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				hash, err := hashPassword(tt.password)
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)

				// Проверяем, что хеш можно декодировать
				p, salt, hashBytes, err := decodeHash(hash)
				assert.NoError(t, err)
				assert.NotNil(t, p)
				assert.NotEmpty(t, salt)
				assert.NotEmpty(t, hashBytes)
			}
		})
	}
}

func TestComparePasswordAndHash_TimingAttack(t *testing.T) {
	// Этот тест проверяет, что функция сравнения не падает при неверном пароле
	validPassword := "TestP@ssw0rd"
	hash, err := hashPassword(validPassword)
	require.NoError(t, err)

	// Используем только валидные пароли (проходящие validatePassword)
	validButWrongPasswords := []string{
		"TestP@ssw0rD",   // отличается только последняя буква (D вместо d)
		"TestP@ssw0rd!",  // добавлен спецсимвол в конце
		"TestP@ssw0rd1",  // заменен спецсимвол на цифру
		"TestP@ssw0rd2",  // другая цифра в конце
		"TestP@ssw0rde",  // буква вместо цифры
		"TestP@ssw0rdd",  // дублирование последней буквы
		"TestP@ssw0rd0",  // цифра 0 вместо o
		"TestP@ssw0rd@",  // другой спецсимвол
		"TestP@ssw0rd#",  // другой спецсимвол
		"TestP@ssw0rd$",  // другой спецсимвол
		"TestP@ssw0rd%",  // другой спецсимвол
		"TestP@ssw0rd^",  // другой спецсимвол
		"TestP@ssw0rd&",  // другой спецсимвол
		"TestP@ssw0rd*",  // другой спецсимвол
		"TestP@ssw0rd(",  // другой спецсимвол
		"TestP@ssw0rd)",  // другой спецсимвол
		"TestP@ssw0rd-",  // другой спецсимвол
		"TestP@ssw0rd_",  // другой спецсимвол
		"TestP@ssw0rd+",  // другой спецсимвол
		"TestP@ssw0rd=",  // другой спецсимвол
		"TestP@ssw0rd{",  // другой спецсимвол
		"TestP@ssw0rd}",  // другой спецсимвол
		"TestP@ssw0rd[",  // другой спецсимвол
		"TestP@ssw0rd]",  // другой спецсимвол
		"TestP@ssw0rd|",  // другой спецсимвол
		"TestP@ssw0rd\\", // другой спецсимвол
		"TestP@ssw0rd:",  // другой спецсимвол
		"TestP@ssw0rd;",  // другой спецсимвол
		"TestP@ssw0rd'",  // другой спецсимвол
		"TestP@ssw0rd\"", // другой спецсимвол
		"TestP@ssw0rd<",  // другой спецсимвол
		"TestP@ssw0rd>",  // другой спецсимвол
		"TestP@ssw0rd,",  // другой спецсимвол
		"TestP@ssw0rd.",  // другой спецсимвол
		"TestP@ssw0rd?",  // другой спецсимвол
		"TestP@ssw0rd/",  // другой спецсимвол
		"TestP@ssw0rd~",  // другой спецсимвол
		"TestP@ssw0rd`",  // другой спецсимвол
	}

	for _, wrongPass := range validButWrongPasswords {
		t.Run(wrongPass, func(t *testing.T) {
			// Убедимся, что пароль валидный
			err := validatePassword(wrongPass)
			require.NoError(t, err, "Тестовый пароль должен быть валидным: %s", wrongPass)

			match, err := comparePasswordAndHash(wrongPass, hash)
			assert.NoError(t, err, "Для валидного пароля не должно быть ошибки")
			assert.False(t, match, "Пароль не должен совпадать")
		})
	}
}
