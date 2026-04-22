package utils

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func ShortTitle(title string, max ...int) string {
	title = html.UnescapeString(title)
	limit := 25
	if len(max) > 0 {
		limit = max[0]
	}
	runes := []rune(title)
	if len(runes) <= limit {
		return title
	}
	return string(runes[:limit]) + "..."
}

// EscapeHTML escapes characters for Telegram HTML mode (only &, <, >).
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func CleanURL(raw string) string {
	before, _, _ := strings.Cut(raw, "?")
	return before
}

func MentionHTML(u *tg.UserObj) string {
	if u == nil {
		return "Unknown"
	}

	fullName := strings.TrimSpace(u.FirstName + " " + u.LastName)
	if fullName == "" {
		fullName = "User"
	}
	fullName = EscapeHTML(ShortTitle(fullName, 15))

	return fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", u.ID, fullName)
}

// IfElse returns `a` if condition is true, else returns `b`.
func IfElse[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

// ParseBool converts strings like "on", "off", "enable", "disable", "true", "false"
// into a boolean value. Returns an error if input is invalid.
func ParseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "on", "enable", "enabled", "true", "1", "yes", "y":
		return true, nil
	case "off", "disable", "disabled", "false", "0", "no", "n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean string: %q", s)
	}
}

// IntToStr converts any signed integer type to string.
func IntToStr[T ~int | ~int8 | ~int16 | ~int32 | ~int64](v T) string {
	return strconv.FormatInt(int64(v), 10)
}
