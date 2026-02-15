package utils

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateTrxCode generates a unique transaction code
func GenerateTrxCode() string {
	now := time.Now()
	dateStr := now.Format("20060102")

	// Generate random 4-digit number
	n, _ := rand.Int(rand.Reader, big.NewInt(9999))
	randomNum := fmt.Sprintf("%04d", n.Int64())

	return fmt.Sprintf("TRX-%s-%s", dateStr, randomNum)
}

// FormatCurrency formats amount to Indonesian Rupiah format
func FormatCurrency(amount float64) string {
	return fmt.Sprintf("Rp %.2f", amount)
}

// ParsePhoneNumber parses and normalizes phone number
func ParsePhoneNumber(phone string) string {
	// Remove all non-digit characters
	re := regexp.MustCompile(`[^\d]`)
	phone = re.ReplaceAllString(phone, "")

	// Remove leading 0 if present and add 62 for Indonesia
	if strings.HasPrefix(phone, "0") {
		phone = "62" + phone[1:]
	}

	return phone
}

// ValidatePhoneNumber validates Indonesian phone number format
func ValidatePhoneNumber(phone string) bool {
	// Parse and normalize first
	normalized := ParsePhoneNumber(phone)

	// Check if it starts with 62 and has correct length (10-13 digits after 62)
	if !strings.HasPrefix(normalized, "62") {
		return false
	}

	// Remove 62 prefix and check remaining digits
	digits := normalized[2:]
	return len(digits) >= 9 && len(digits) <= 13
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// ValidatePassword validates password strength
func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	// Check for at least one uppercase, one lowercase, and one digit
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)

	return hasUpper && hasLower && hasDigit
}

// HashPassword creates a hash for the password (placeholder - use bcrypt in production)
func HashPassword(password string) string {
	// In production, use bcrypt: bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	// For now, return a simple hash (NOT SECURE FOR PRODUCTION)
	return fmt.Sprintf("hashed_%s", password)
}

// VerifyPassword verifies password against hash (placeholder - use bcrypt in production)
func VerifyPassword(password, hash string) bool {
	// In production, use bcrypt: bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	// For now, simple verification (NOT SECURE FOR PRODUCTION)
	return hash == fmt.Sprintf("hashed_%s", password)
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// GenerateAPIKey generates a random API key
func GenerateAPIKey() string {
	return GenerateRandomString(32)
}

// RoundToDecimal rounds float64 to specified decimal places
func RoundToDecimal(value float64, places int) float64 {
	multiplier := math.Pow(10, float64(places))
	return float64(int(value*multiplier+0.5)) / multiplier
}

// CalculatePercentage calculates percentage
func CalculatePercentage(part, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (part / total) * 100
}

// IsValidAmount validates monetary amount
func IsValidAmount(amount float64) bool {
	return amount >= 0 && amount < 999999999.99
}

// FormatAmount formats amount for display
func FormatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

// ParseAmount parses amount from string
func ParseAmount(amountStr string) (float64, error) {
	// Remove commas and other formatting
	re := regexp.MustCompile(`[^\d.]`)
	amountStr = re.ReplaceAllString(amountStr, "")

	return strconv.ParseFloat(amountStr, 64)
}

// SanitizeString sanitizes string for SQL injection prevention
func SanitizeString(s string) string {
	// Basic sanitization - in production, use parameterized queries
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "")
	s = strings.ReplaceAll(s, "--", "")
	return s
}

// TruncateString truncates string to specified length
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// Contains checks if slice contains string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicate removes duplicate strings from slice
func RemoveDuplicate(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Time helpers
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", timeStr)
}

func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// IsToday checks if given time is today
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

// StartOfDay returns start of the day for given time
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns end of the day for given time
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// AddDays adds days to time
func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// DifferenceInDays calculates difference in days between two times
func DifferenceInDays(t1, t2 time.Time) int {
	return int(t2.Sub(t1).Hours() / 24)
}
