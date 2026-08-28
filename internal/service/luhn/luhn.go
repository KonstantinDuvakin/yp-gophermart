// Package luhn проверяет корректность номера по алгоритму Луна.
package luhn

// CheckNumber возвращает true, если строка состоит только из цифр и проходит
// проверку по алгоритму Луна. Пустая строка считается некорректной.
func CheckNumber(number string) bool {
	if number == "" {
		return false
	}

	sum := 0
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		num := number[i]

		if num < '0' || num > '9' {
			return false
		}

		d := int(num - '0')

		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d

		double = !double
	}

	return sum%10 == 0
}
