package util

// ToBase32 将数字转换为字符串
func ToBase32(num int64) string {
	num = num + 1888888

	chars := "5dysjk0replh3tn2og4w7ca9bf6um8vx1iqz"
	result := ""
	for num > 0 {
		remainder := num % 36
		result = string(chars[remainder]) + result
		num /= 36
	}

	return result
}
