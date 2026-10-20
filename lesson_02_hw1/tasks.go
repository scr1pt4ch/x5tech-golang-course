package lesson2

// FizzBuzz возвращает FizzBuzz, Fizz, Buzz или пустую строку для обычного числа.
// Для n <= 0 возвращает invalid.
func FizzBuzz(n int) string {
	switch {
	case n <= 0:
		return "invalid"
	case 0 == n%15:
		return "FizzBuzz"
	case 0 == n%3:
		return "Fizz"
	case 0 == n%5:
		return "Buzz"
	default:
		return ""
	}
}

// Sum возвращает сумму. Для пустого среза и nil возвращает 0.
func Sum(values []int64) int64 {
	var ans int64

	for _, val := range values {
		ans += val
	}

	return ans
}

// Min возвращает минимум. Для пустого среза и nil возвращает 0 по условию задачи.
func Min(values []int64) int64 {
	if 0 == len(values) {
		return 0
	}

	mini := values[0]

	for _, val := range values {
		if val < mini {
			mini = val
		}
	}

	return mini
}

// Max возвращает максимум. Для пустого среза и nil возвращает 0 по условию задачи.
func Max(values []int64) int64 {
	if 0 == len(values) {
		return 0
	}

	maxi := values[0]

	for _, val := range values {
		if maxi < val {
			maxi = val
		}
	}

	return maxi
}

// Average возвращает среднее как float64. Для пустого среза и nil возвращает 0.
func Average(values []int64) float64 {
	if 0 == len(values) {
		return 0
	}

	return float64(Sum(values)) / float64(len(values))
}

// Category возвращает invalid, small, regular или large по сумме в копейках.
func Category(amount int64) string {
	switch {
	case amount < 0:
		return "invalid"
	case amount <= 9999:
		return "small"
	case amount <= 49999:
		return "regular"
	default:
		return "large"
	}
}

// RuneCount считает кодовые точки Unicode в строке через range.
func RuneCount(text string) int {
	cnt := 0

	for range text {
		cnt++
	}

	return cnt
}
