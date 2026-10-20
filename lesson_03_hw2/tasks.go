package homework

// PairInput соответствует объекту nums/target из условия контеста.
type PairInput struct {
	Nums   []int
	Target int
}

// longestPrefix возвращает самый длинный общий префикс строк.
func longestPrefix(words []string) string {
	if 0 == len(words) {
		return ""
	}

	prefix := []rune(words[0])

	for _, word := range words[1:] {
		runes := []rune(word)
		n := 0

		for n < len(prefix) && n < len(runes) && prefix[n] == runes[n] {
			n++
		}

		prefix = prefix[:n]

		if 0 == len(prefix) {
			return ""
		}
	}

	return string(prefix)
}

// pairIndices возвращает индексы двух разных элементов по возрастанию.
func pairIndices(input PairInput) [2]int {
	visited := make(map[int]int, len(input.Nums))

	for i, num := range input.Nums {
		if j, ok := visited[input.Target-num]; ok {
			return [2]int{j, i}
		}

		if _, ok := visited[num]; !ok {
			visited[num] = i
		}
	}

	return [2]int{}
}

// hasDuplicate сообщает, встречается ли какое-либо значение повторно.
func hasDuplicate(nums []int) bool {
	visited := make(map[int]struct{}, len(nums))

	for _, num := range nums {
		if _, ok := visited[num]; ok {
			return true
		}

		visited[num] = struct{}{}
	}

	return false
}

// compactSorted создаёт независимый слайс уникальных значений.
func compactSorted(nums []int) []int {
	if 0 == len(nums) {
		return nil
	}

	ans := make([]int, 0, len(nums))
	ans = append(ans, nums[0])

	for i := 1; i < len(nums); i++ {
		if nums[i] != nums[i-1] {
			ans = append(ans, nums[i])
		}
	}

	return ans
}

// balancedBrackets проверяет пары и порядок закрытия скобок.
func balancedBrackets(text string) bool {
	if 1 == len(text)%2 {
		return false
	}

	pairs := map[byte]byte{
		')': '(',
		']': '[',
		'}': '{',
	}

	stack := make([]byte, 0, len(text))

	for i := 0; i < len(text); i++ {
		ch := text[i]

		switch ch {
		case '(', '[', '{':
			stack = append(stack, ch)
		case ')', ']', '}':
			if 0 == len(stack) || stack[len(stack)-1] != pairs[ch] {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}

	return 0 == len(stack)
}
