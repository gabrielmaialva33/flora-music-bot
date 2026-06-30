// Package rtmp provides RTMP group-call streaming for Telegram.
//
// It is ported from gogram v1.7.3 (telegram/tgcalls.go), whose high-level RTMP
// helper was removed upstream in v1.7.6. The only coupling to gogram is the
// stable PhoneGetGroupCallStreamRtmpURL / ResolvePeer client methods; everything
// else is ffmpeg orchestration. Keeping it in-tree decouples Flora from gogram's
// call API churn.
package rtmp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

type State int

const (
	StateIdle State = iota
	StatePlaying
	StatePaused
	StateStopped
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StatePlaying:
		return "playing"
	case StatePaused:
		return "paused"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

var (
	ErrFFmpegNotFound  = errors.New("ffmpeg not found in PATH")
	ErrStreamPlaying   = errors.New("stream already playing")
	ErrStreamNotPaused = errors.New("stream not paused")
	ErrNoRTMPURL       = errors.New("RTMP URL not set, call FetchRTMPURL() first")
	ErrNoInputSource   = errors.New("no input source available")
	ErrFileNotFound    = errors.New("input file not found")
)

type Stream struct {
	chatID     int64
	rtmpURL    string
	rtmpKey    string
	client     *tg.Client
	state      State
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stderr     bytes.Buffer
	cancelFunc context.CancelFunc
	mu         sync.Mutex
	inputFile  string
	inputData  []byte
	loopCount  int
	bitrate    string
	audioBit   string
	frameRate  int
	startTime  time.Time
	pausedAt   time.Duration
	seekPos    time.Duration
	lastError  error
	onError    func(int64, error)
	onEnd      func(int64)
	audioOnly  bool
	imageFile  string
	muted      bool
}

type Config struct {
	Bitrate   string // Video bitrate (e.g., "2000k")
	AudioBit  string // Audio bitrate (e.g., "96k")
	FrameRate int    // Video frame rate (default: 30)
	LoopCount int    // Number of times to loop (-1 for infinite)
}

func DefaultConfig() *Config {
	return &Config{
		Bitrate:   "2000k",
		AudioBit:  "96k",
		FrameRate: 30,
		LoopCount: -1,
	}
}

func New(client *tg.Client, chatID int64, config ...*Config) (*Stream, error) {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, ErrFFmpegNotFound
	}

	if len(config) == 0 {
		config = append(config, DefaultConfig())
	}

	rtmpConfig := config[0]

	return &Stream{
		chatID:    chatID,
		client:    client,
		state:     StateIdle,
		loopCount: rtmpConfig.LoopCount,
		bitrate:   rtmpConfig.Bitrate,
		audioBit:  rtmpConfig.AudioBit,
		frameRate: rtmpConfig.FrameRate,
	}, nil
}

// FetchRTMPURL fetches the RTMP URL and stream key from Telegram.
// NOTE: This can only be called by user accounts, not bot accounts.
// Bot accounts will receive an error. For bots, use SetURL() and SetKey() manually.
func (s *Stream) FetchRTMPURL() error {
	peer, err := s.client.ResolvePeer(s.chatID)
	if err != nil {
		return fmt.Errorf("failed to resolve peer: %w", err)
	}
	rtmpInfo, err := s.client.PhoneGetGroupCallStreamRtmpURL(false, peer, false)
	if err != nil {
		return fmt.Errorf("failed to fetch RTMP URL: %w", err)
	}
	s.mu.Lock()
	s.rtmpURL = rtmpInfo.URL
	s.rtmpKey = rtmpInfo.Key
	s.mu.Unlock()
	return nil
}

func (s *Stream) SetURL(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rtmpURL = url
}

func (s *Stream) SetKey(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rtmpKey = key
}

func (s *Stream) SetFullURL(fullURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if fullURL == "" {
		return fmt.Errorf("RTMP URL cannot be empty")
	}

	if !bytes.Contains([]byte(fullURL), []byte("rtmp://")) &&
		!bytes.Contains([]byte(fullURL), []byte("rtmps://")) {
		return fmt.Errorf("invalid RTMP URL: must start with rtmp:// or rtmps://")
	}

	var url, key string

	// Try /s/ separator first (Telegram format)
	if parts := bytes.SplitN([]byte(fullURL), []byte("/s/"), 2); len(parts) == 2 {
		url = string(parts[0]) + "/s/"
		key = string(parts[1])
	} else if parts := bytes.SplitN([]byte(fullURL), []byte("/"), 4); len(parts) >= 4 {
		url = string(parts[0]) + "//" + string(parts[1]) + "/" + string(parts[2]) + "/"
		key = string(parts[3])
	} else {
		return fmt.Errorf("invalid RTMP URL format: expected rtmp://host/app/streamkey or similar")
	}

	if url == "" || key == "" {
		return fmt.Errorf("failed to parse URL: both URL and key must be non-empty")
	}

	s.rtmpURL = url
	s.rtmpKey = key
	return nil
}

func (s *Stream) GetURL() string {
	return s.rtmpURL
}

func (s *Stream) GetKey() string {
	return s.rtmpKey
}

func (s *Stream) GetFullURL() string {
	return s.rtmpURL + s.rtmpKey
}

// RefreshRTMPURL fetches a new RTMP URL and stream key (revokes the old one).
// NOTE: This can only be called by user accounts, not bot accounts.
func (s *Stream) RefreshRTMPURL() error {
	peer, err := s.client.ResolvePeer(s.chatID)
	if err != nil {
		return fmt.Errorf("failed to resolve peer: %w", err)
	}

	rtmpInfo, err := s.client.PhoneGetGroupCallStreamRtmpURL(false, peer, true)
	if err != nil {
		return fmt.Errorf("failed to refresh RTMP URL: %w", err)
	}
	s.mu.Lock()
	s.rtmpURL = rtmpInfo.URL
	s.rtmpKey = rtmpInfo.Key
	s.mu.Unlock()
	return nil
}

func (s *Stream) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// Play starts streaming from a file path (string) or raw bytes ([]byte)
func (s *Stream) Play(source any) error {
	// Validate RTMP URL is set
	if s.rtmpURL == "" || s.rtmpKey == "" {
		return ErrNoRTMPURL
	}

	s.mu.Lock()
	if s.state == StatePlaying {
		s.mu.Unlock()
		return ErrStreamPlaying
	}

	switch src := source.(type) {
	case string:
		// Check if file exists
		if _, err := os.Stat(src); err != nil {
			s.mu.Unlock()
			if os.IsNotExist(err) {
				return fmt.Errorf("%w: %s", ErrFileNotFound, src)
			}
			return fmt.Errorf("failed to access file: %w", err)
		}
		s.inputFile = src
		s.inputData = nil
		s.mu.Unlock()
		return s.startFFmpeg(src, false)
	case []byte:
		if len(src) == 0 {
			s.mu.Unlock()
			return errors.New("empty byte input")
		}
		s.inputData = src
		s.inputFile = ""
		s.mu.Unlock()
		return s.startFFmpeg("pipe:0", true)
	default:
		s.mu.Unlock()
		return fmt.Errorf("unsupported source type: expected string or []byte, got %T", source)
	}
}

func (s *Stream) startFFmpeg(input string, pipeInput bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	args := s.buildFFmpegArgs(input)
	s.cmd = exec.CommandContext(ctx, "ffmpeg", args...)

	// Reset and capture stderr for error messages
	s.stderr.Reset()
	s.cmd.Stderr = &s.stderr
	s.cmd.Stdout = nil

	if pipeInput {
		stdin, err := s.cmd.StdinPipe()
		if err != nil {
			cancel()
			return fmt.Errorf("failed to get stdin pipe: %w", err)
		}
		s.stdin = stdin
	}

	if err := s.cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	s.mu.Lock()
	s.state = StatePlaying
	s.startTime = time.Now()
	s.lastError = nil
	s.mu.Unlock()

	if pipeInput && s.inputData != nil {
		go func() {
			defer s.stdin.Close()
			s.stdin.Write(s.inputData)
		}()
	}

	go func() {
		err := s.cmd.Wait()
		s.mu.Lock()
		// Don't reset state if paused or stopped
		if s.state == StatePlaying {
			s.state = StateIdle
			// Check for ffmpeg errors (non-zero exit and not cancelled)
			if err != nil && ctx.Err() == nil {
				errMsg := s.stderr.String()
				if errMsg != "" {
					s.lastError = fmt.Errorf("ffmpeg error: %s", errMsg)
				} else {
					s.lastError = fmt.Errorf("ffmpeg exited with error: %w", err)
				}
				if s.onError != nil {
					go s.onError(s.chatID, s.lastError)
				}
			} else {
				// Stream ended normally (no error)
				if s.onEnd != nil {
					go s.onEnd(s.chatID)
				}
			}
		}
		s.mu.Unlock()
	}()

	return nil
}

func (s *Stream) buildFFmpegArgs(input string) []string {
	args := []string{}

	if s.seekPos > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", s.seekPos.Seconds()))
	}

	args = append(args, "-re")

	// If audio-only with static image
	if s.audioOnly && s.imageFile != "" {
		// Add static image as loop input
		args = append(
			args,
			"-loop", "1",
			"-i", s.imageFile,
		)

		// Add audio input
		if s.loopCount != 0 {
			args = append(args, "-stream_loop", fmt.Sprintf("%d", s.loopCount))
		}
		args = append(args, "-i", input)

		// Map video from image and audio from input
		args = append(
			args,
			"-map", "0:v", // Video from image (first input)
			"-map", "1:a", // Audio from audio file (second input)
			"-c:v", "libx264",
			"-preset", "superfast",
			"-b:v", s.bitrate,
			"-maxrate", s.bitrate,
			"-bufsize", s.doubleBitrate(),
			"-pix_fmt", "yuv420p",
			"-r", fmt.Sprintf("%d", s.frameRate),
			"-g", fmt.Sprintf("%d", s.frameRate),
			"-threads", "0",
		)

		if s.muted {
			// anullsrc
			args = append(
				args,
				"-c:a", "aac",
				"-b:a", s.audioBit,
				"-ac", "2",
				"-ar", "44100",
				"-af", "volume=0",
			)
		} else {
			args = append(
				args,
				"-c:a", "aac",
				"-b:a", s.audioBit,
				"-ac", "2",
				"-ar", "44100",
			)
		}

		args = append(
			args,
			"-shortest",
			"-f", "flv",
			"-rtmp_buffer", "100",
			"-rtmp_live", "live",
			s.GetFullURL(),
		)
	} else {
		if s.loopCount != 0 {
			args = append(args, "-stream_loop", fmt.Sprintf("%d", s.loopCount))
		}

		args = append(
			args,
			"-i", input,
			"-c:v", "libx264",
			"-preset", "superfast",
			"-b:v", s.bitrate,
			"-maxrate", s.bitrate,
			"-bufsize", s.doubleBitrate(),
			"-pix_fmt", "yuv420p",
			"-g", fmt.Sprintf("%d", s.frameRate),
			"-threads", "0",
		)

		if s.muted {
			args = append(
				args,
				"-c:a", "aac",
				"-b:a", s.audioBit,
				"-ac", "2",
				"-ar", "44100",
				"-af", "volume=0",
			)
		} else {
			args = append(
				args,
				"-c:a", "aac",
				"-b:a", s.audioBit,
				"-ac", "2",
				"-ar", "44100",
			)
		}

		args = append(
			args,
			"-f", "flv",
			"-rtmp_buffer", "100",
			"-rtmp_live", "live",
			s.GetFullURL(),
		)
	}

	return args
}

func (s *Stream) doubleBitrate() string {
	var val int
	fmt.Sscanf(s.bitrate, "%dk", &val)
	return fmt.Sprintf("%dk", val*2)
}

func (s *Stream) Pause() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != StatePlaying {
		return fmt.Errorf("cannot pause: stream is %s", s.state)
	}

	s.pausedAt = time.Since(s.startTime) + s.seekPos

	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	s.state = StatePaused
	return nil
}

func (s *Stream) Resume() error {
	s.mu.Lock()
	if s.state != StatePaused {
		s.mu.Unlock()
		return ErrStreamNotPaused
	}

	s.seekPos = s.pausedAt
	s.mu.Unlock()

	if s.inputFile != "" {
		return s.startFFmpeg(s.inputFile, false)
	} else if s.inputData != nil {
		return s.startFFmpeg("pipe:0", true)
	}
	return ErrNoInputSource
}

// Seek to a specific position (only works with file input)
func (s *Stream) Seek(position time.Duration) error {
	s.mu.Lock()
	if s.inputFile == "" {
		s.mu.Unlock()
		return errors.New("seek only supported for file input")
	}

	if position < 0 {
		s.mu.Unlock()
		return errors.New("seek position cannot be negative")
	}

	wasPlaying := s.state == StatePlaying
	s.seekPos = position
	s.mu.Unlock()

	if wasPlaying {
		s.Stop()
		return s.startFFmpeg(s.inputFile, false)
	}
	return nil
}

// CurrentPosition returns the current playback position
func (s *Stream) CurrentPosition() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch s.state {
	case StatePlaying:
		return time.Since(s.startTime) + s.seekPos
	case StatePaused:
		return s.pausedAt
	default:
		return 0
	}
}

// LastError returns the last error that occurred during streaming
func (s *Stream) LastError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastError
}

// OnError sets a callback function that will be called when an error occurs
func (s *Stream) OnError(fn func(int64, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onError = fn
}

// OnEnd sets a callback function that will be called when the stream ends normally
func (s *Stream) OnEnd(fn func(int64)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onEnd = fn
}

// Mute mutes the audio stream by restarting FFmpeg with volume=0 filter
func (s *Stream) Mute() error {
	s.mu.Lock()
	if s.muted {
		s.mu.Unlock()
		return nil
	}

	if s.state != StatePlaying {
		s.mu.Unlock()
		return fmt.Errorf("cannot mute: stream is %s", s.state)
	}

	s.muted = true
	wasPlaying := s.state == StatePlaying
	currentPos := time.Since(s.startTime) + s.seekPos
	s.mu.Unlock()

	if wasPlaying {
		s.Stop()
		s.mu.Lock()
		s.seekPos = currentPos
		s.mu.Unlock()

		if s.inputFile != "" {
			return s.startFFmpeg(s.inputFile, false)
		} else if s.inputData != nil {
			return s.startFFmpeg("pipe:0", true)
		}
	}

	return nil
}

// Unmute unmutes the audio stream by restarting FFmpeg without volume filter
func (s *Stream) Unmute() error {
	s.mu.Lock()
	if !s.muted {
		s.mu.Unlock()
		return nil
	}

	if s.state != StatePlaying {
		s.mu.Unlock()
		return fmt.Errorf("cannot unmute: stream is %s", s.state)
	}

	s.muted = false
	wasPlaying := s.state == StatePlaying
	currentPos := time.Since(s.startTime) + s.seekPos
	s.mu.Unlock()

	if wasPlaying {
		s.Stop()
		s.mu.Lock()
		s.seekPos = currentPos
		s.mu.Unlock()

		if s.inputFile != "" {
			return s.startFFmpeg(s.inputFile, false)
		} else if s.inputData != nil {
			return s.startFFmpeg("pipe:0", true)
		}
	}

	return nil
}

// IsMuted returns whether the audio is currently muted
func (s *Stream) IsMuted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.muted
}

func (s *Stream) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == StateIdle || s.state == StateStopped {
		return nil
	}

	s.state = StateStopped

	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	if s.stdin != nil {
		s.stdin.Close()
	}

	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}

	return nil
}

func (s *Stream) SetBitrate(bitrate string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bitrate = bitrate
}

func (s *Stream) SetAudioBitrate(bitrate string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audioBit = bitrate
}

func (s *Stream) SetFrameRate(fps int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.frameRate = fps
}

func (s *Stream) SetLoopCount(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loopCount = count
}

// StartPipe starts the RTMP stream expecting data to be fed via FeedChunk().
// Use this when you want to stream data progressively in chunks.
func (s *Stream) StartPipe() error {
	if s.rtmpURL == "" || s.rtmpKey == "" {
		return ErrNoRTMPURL
	}

	s.mu.Lock()
	if s.state == StatePlaying {
		s.mu.Unlock()
		return ErrStreamPlaying
	}
	s.inputFile = ""
	s.inputData = nil
	s.mu.Unlock()

	return s.startFFmpegPipe()
}

func (s *Stream) startFFmpegPipe() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	args := s.buildFFmpegPipeArgs()
	s.cmd = exec.CommandContext(ctx, "ffmpeg", args...)

	s.stderr.Reset()
	s.cmd.Stderr = &s.stderr
	s.cmd.Stdout = nil

	stdin, err := s.cmd.StdinPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	s.stdin = stdin

	if err := s.cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	s.mu.Lock()
	s.state = StatePlaying
	s.startTime = time.Now()
	s.lastError = nil
	s.mu.Unlock()

	go func() {
		err := s.cmd.Wait()
		s.mu.Lock()
		if s.state == StatePlaying {
			s.state = StateIdle
			if err != nil && ctx.Err() == nil {
				errMsg := s.stderr.String()
				if errMsg != "" {
					s.lastError = fmt.Errorf("ffmpeg error: %s", errMsg)
				} else {
					s.lastError = fmt.Errorf("ffmpeg exited with error: %w", err)
				}
				if s.onError != nil {
					go s.onError(s.chatID, s.lastError)
				}
			} else {
				if s.onEnd != nil {
					go s.onEnd(s.chatID)
				}
			}
		}
		s.mu.Unlock()
	}()

	return nil
}

func (s *Stream) buildFFmpegPipeArgs() []string {
	return []string{
		"-re",
		"-i", "pipe:0",
		"-c:v", "libx264",
		"-preset", "superfast",
		"-b:v", s.bitrate,
		"-maxrate", s.bitrate,
		"-bufsize", s.doubleBitrate(),
		"-pix_fmt", "yuv420p",
		"-g", fmt.Sprintf("%d", s.frameRate),
		"-threads", "0",
		"-c:a", "aac",
		"-b:a", s.audioBit,
		"-ac", "2",
		"-ar", "44100",
		"-f", "flv",
		"-rtmp_buffer", "100",
		"-rtmp_live", "live",
		s.GetFullURL(),
	}
}

// FeedChunk writes a chunk of data to the RTMP stream.
// Must call StartPipe() first to initialize the stream.
// Returns error if stream is not playing or write fails.
func (s *Stream) FeedChunk(data []byte) error {
	s.mu.Lock()
	if s.state != StatePlaying {
		s.mu.Unlock()
		return fmt.Errorf("cannot feed chunk: stream is %s", s.state)
	}
	if s.stdin == nil {
		s.mu.Unlock()
		return errors.New("stdin pipe not initialized, call StartPipe() first")
	}
	s.mu.Unlock()

	_, err := s.stdin.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write chunk: %w", err)
	}
	return nil
}

// FeedReader reads from an io.Reader and feeds data to the RTMP stream.
// This is useful for streaming from HTTP responses, files, etc.
// Must call StartPipe() first.
func (s *Stream) FeedReader(r io.Reader) error {
	s.mu.Lock()
	if s.state != StatePlaying {
		s.mu.Unlock()
		return fmt.Errorf("cannot feed reader: stream is %s", s.state)
	}
	if s.stdin == nil {
		s.mu.Unlock()
		return errors.New("stdin pipe not initialized, call StartPipe() first")
	}
	s.mu.Unlock()

	_, err := io.Copy(s.stdin, r)
	if err != nil {
		return fmt.Errorf("failed to copy from reader: %w", err)
	}
	return nil
}

// ClosePipe closes the stdin pipe, signaling EOF to ffmpeg.
// Call this when you're done feeding data.
func (s *Stream) ClosePipe() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stdin != nil {
		return s.stdin.Close()
	}
	return nil
}

// PlayAudioWithImage streams audio with a static image as background.
func (s *Stream) PlayAudioWithImage(audioSource any, imageSource string) error {
	if s.rtmpURL == "" || s.rtmpKey == "" {
		return ErrNoRTMPURL
	}

	isURL := strings.HasPrefix(imageSource, "http://") ||
		strings.HasPrefix(imageSource, "https://")

	if !isURL {
		if _, err := os.Stat(imageSource); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%w: %s", ErrFileNotFound, imageSource)
			}
			return fmt.Errorf("failed to access image file: %w", err)
		}
	}

	s.mu.Lock()
	if s.state == StatePlaying {
		s.mu.Unlock()
		return ErrStreamPlaying
	}

	s.audioOnly = true
	s.imageFile = imageSource

	switch src := audioSource.(type) {
	case string:
		if _, err := os.Stat(src); err != nil {
			s.mu.Unlock()
			if os.IsNotExist(err) {
				return fmt.Errorf("%w: %s", ErrFileNotFound, src)
			}
			return fmt.Errorf("failed to access audio file: %w", err)
		}
		s.inputFile = src
		s.inputData = nil
		s.mu.Unlock()
		return s.startFFmpeg(src, false)
	case []byte:
		if len(src) == 0 {
			s.mu.Unlock()
			return errors.New("empty byte input")
		}
		s.inputData = src
		s.inputFile = ""
		s.mu.Unlock()
		return s.startFFmpeg("pipe:0", true)
	default:
		s.mu.Unlock()
		return fmt.Errorf("unsupported source type: expected string or []byte, got %T", audioSource)
	}
}
