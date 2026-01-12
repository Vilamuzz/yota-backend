package pkg

import "regexp"

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func IsValidLengthPassword(password string) bool {
	return len(password) >= 8
}

func IsStrongPassword(password string) bool {
	// has upper case
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// has lower case
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// has number
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)

	return hasUpper && hasLower && hasNumber
}

func IsValidPhoneNumber(phone string) bool {
	re := regexp.MustCompile(`^\[0-9]{9,12}$`)
	return re.MatchString(phone)
}
