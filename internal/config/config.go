package config

import (
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	_ "github.com/joho/godotenv/autoload"
)

var (

	// Required Variables

	APIID          int32
	APIHash        string
	Token          string
	LoggerID       int64
	MongoURI       string
	StringSessions []string
	SessionType    string

	// Optional Variables

	OwnerID             int64
	SpotifyClientID     string
	SpotifyClientSecret string
	FallenAPIURL        string
	FallenAPIKey        string
	BetomatoToken       string
	BetomatoBaseURL     string
	DefaultLang         string
	DurationLimit       int
	LeaveOnDemoted      bool
	QueueLimit          int
	SupportChat         string
	SupportChannel      string
	CookiesLink         string
	SetCmds             bool
	MaxAuthUsers        int
	StartImage          string
	PingImage           string
	Port                string

	// System & Logging

	StartTime   time.Time
	LogFileName = "logs.txt"
	LogWriter   io.Writer

	// Internal
	logger  = gologging.GetLogger("config")
	logFile *os.File
)

func init() {
	initLogging()
	loadConfig()
	validateConfig()
}

func initLogging() {
	_ = os.Remove(LogFileName)

	file, err := os.OpenFile(
		LogFileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		logger.FatalF("Failed to open log file: %v", err)
	}

	logFile = file
	LogWriter = io.MultiWriter(file, os.Stderr)
}

func loadConfig() {
	StartTime = time.Now()

	// Load Required
	APIID = getInt[int32]("API_ID", 0)
	APIHash = getString("API_HASH", "")
	// Checks TOKEN, fallbacks to BOT_TOKEN
	Token = getString(
		"TOKEN",
		getString("BOT_TOKEN", ""),
	)
	LoggerID = getInt[int64]("LOGGER_ID", 0)
	MongoURI = getString("MONGO_DB_URI", "")
	SessionType = getString("SESSION_TYPE", "pyrogram")
	StringSessions = getStringSlice(
		"STRING_SESSIONS",
		getStringSlice("STRING_SESSION", nil),
	)
	// Load Optional
	OwnerID = getInt[int64]("OWNER_ID", 0)
	SpotifyClientID = getString("SPOTIFY_CLIENT_ID", "")
	SpotifyClientSecret = getString("SPOTIFY_CLIENT_SECRET", "")
	FallenAPIURL = getString("FALLEN_API_URL", "https://beta.fallenapi.fun")
	FallenAPIKey = getString("FALLEN_API_KEY", "")
	BetomatoToken = getString("BETOMATO_TOKEN", "")
	BetomatoBaseURL = getString("BETOMATO_BASE_URL", "https://edge.betomato.com")

	DefaultLang = getString("DEFAULT_LANG", "ptbr")
	DurationLimit = getInt("DURATION_LIMIT", 4200) // In seconds
	LeaveOnDemoted = getBool("LEAVE_ON_DEMOTED", false)
	QueueLimit = getInt("QUEUE_LIMIT", 7)
	SupportChat = getString("SUPPORT_CHAT", "https://t.me/WinxCallGroup")
	SupportChannel = getString("SUPPORT_CHANNEL", "https://t.me/WinxCallChannel")
	CookiesLink = getString("COOKIES_LINK", "")
	SetCmds = getBool("SET_CMDS", false)
	MaxAuthUsers = getInt("MAX_AUTH_USERS", 25)

	StartImage = getString(
		"START_IMG_URL",
		"https://raw.githubusercontent.com/gabrielmaialva33/flora-music-bot/refs/heads/main/.github/assets/start.png",
	)
	PingImage = getString(
		"PING_IMG_URL",
		"https://raw.githubusercontent.com/gabrielmaialva33/flora-music-bot/refs/heads/main/.github/assets/ping.png",
	)
	Port = getString("PORT", "8000")
}

func validateConfig() {
	required := []struct {
		ok   bool
		name string
	}{
		{APIID != 0, "API_ID"},
		{APIHash != "", "API_HASH"},
		{LoggerID != 0, "LOGGER_ID"},
		{MongoURI != "", "MONGO_DB_URI"},
		{Token != "", "TOKEN (or BOT_TOKEN)"},
		{len(StringSessions) != 0, "STRING_SESSIONS (or STRING_SESSION)"},
	}

	var missing []string
	for _, field := range required {
		if !field.ok {
			missing = append(missing, field.name)
		}
	}

	if len(missing) > 0 {
		logger.FatalF(
			"missing required config: %s",
			strings.Join(missing, ", "),
		)
	}

	if SpotifyClientID == "" || SpotifyClientSecret == "" {
		logger.Warn(
			"Spotify credentials not configured - Spotify links won't work",
		)
	}
}

// --- Helper Functions ---

// lookupEnv checks multiple variations of a key (e.g., lowercase, uppercase, no underscore)
func lookupEnv(baseKey string) (string, bool) {
	variants := []string{
		baseKey,
		strings.ToUpper(baseKey),
		strings.ToLower(baseKey),
		strings.ReplaceAll(baseKey, "_", ""),
	}

	for _, key := range variants {
		if val, ok := os.LookupEnv(key); ok {
			val = strings.TrimSpace(val)
			if val != "" {
				return val, true
			}
		}
	}
	return "", false
}

func getString(key, fallback string) string {
	if val, ok := lookupEnv(key); ok {
		return val
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	val, ok := lookupEnv(key)
	if !ok {
		return fallback
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		logger.FatalF("Invalid boolean for %s: %v", key, err)
	}
	return boolVal
}

func getInt[T ~int | ~int32 | ~int64](key string, fallback T) T {
	val, ok := lookupEnv(key)
	if !ok {
		return fallback
	}

	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		logger.FatalF("Invalid integer for %s: %v", key, err)
	}
	return T(num)
}

func getStringSlice(key string, fallback []string) []string {
	val, ok := lookupEnv(key)
	if !ok {
		return fallback
	}

	normalized := strings.NewReplacer(",", " ", ";", " ").Replace(val)
	parts := strings.Fields(normalized)

	if len(parts) > 0 {
		return parts
	}
	return fallback
}

func CloseLogging() {
	if logFile != nil {
		_ = logFile.Close()
	}
}
