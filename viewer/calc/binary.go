package calc

// GeneratePowersOfTwo generates a slice of powers of two in the range [from, to].
// It returns a slice containing 2^0, 2^1, 2^2, ... , 2^n, where n is the largest
// integer such that 2^n is less than or equal to 'to'.
//
// If 'from' is negative or 'to' is less than 'from', it returns nil.
func GeneratePowersOfTwo(from, to int) []int {
	if from < 0 || to < from {
		return nil
	}

	powers := make([]int, 0)
	for i := 0; ; i++ {
		power := 1 << uint(i) // Calculate 2 to the power of i
		if power < from {
			continue
		}
		if power > to {
			break
		}
		powers = append(powers, power)
	}

	return powers
}
