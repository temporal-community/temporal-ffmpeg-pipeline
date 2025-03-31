package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
)

// FFmpegProcessingParams contains the parameters for the FFmpeg video processing activity
type FFmpegProcessingParams struct {
	InputPath  string   `json:"inputPath"`
	OutputPath string   `json:"outputPath"`
	FFmpegArgs []string `json:"ffmpegArgs"`
}

// FFmpegProcessingResult contains the results of the processing
type FFmpegProcessingResult struct {
	OutputPath     string        `json:"outputPath"`
	DurationMs     int64         `json:"durationMs"`
	ProcessingTime time.Duration `json:"processingTime"`
}

// ProcessVideoWithFFmpeg runs ffmpeg command with the given parameters and sends heartbeats periodically
func ProcessVideoWithFFmpeg(ctx context.Context, params FFmpegProcessingParams) (*FFmpegProcessingResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting FFmpeg video processing", "inputPath", params.InputPath, "outputPath", params.OutputPath)

	startTime := time.Now()

	// Prepare the ffmpeg command
	args := []string{"-i", params.InputPath}
	args = append(args, params.FFmpegArgs...)
	args = append(args, params.OutputPath)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	// Create pipes for stdout and stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Regular expression to extract progress information
	progressRegex := regexp.MustCompile(`time=(\d+:\d+:\d+\.\d+)`)
	durationRegex := regexp.MustCompile(`Duration: (\d+:\d+:\d+\.\d+)`)

	var totalDurationMs int64
	var currentProgressMs int64

	// Read from stderr pipe
	scanner := bufio.NewScanner(io.LimitReader(stderr, 1024*1024*10)) // Limit to 10MB to prevent memory issues

	// Start heartbeating in a separate goroutine
	heartbeatTicker := time.NewTicker(5 * time.Second)
	defer heartbeatTicker.Stop()

	// Create a channel to signal when to stop heartbeating
	done := make(chan struct{})
	defer close(done)

	// Start heartbeat goroutine
	go func() {
		for {
			select {
			case <-heartbeatTicker.C:
				details := map[string]interface{}{
					"progress": currentProgressMs,
					"total":    totalDurationMs,
					"percent":  calculatePercentage(currentProgressMs, totalDurationMs),
				}
				activity.RecordHeartbeat(ctx, details)
			case <-done:
				return
			}
		}
	}()

	// Process ffmpeg output
	for scanner.Scan() {
		line := scanner.Text()

		// Extract total duration (runs once)
		if totalDurationMs == 0 {
			durationMatch := durationRegex.FindStringSubmatch(line)
			if len(durationMatch) > 1 {
				totalDurationMs = timeToMilliseconds(durationMatch[1])
				logger.Info("Video duration", "durationMs", totalDurationMs)
			}
		}

		// Extract current progress
		progressMatch := progressRegex.FindStringSubmatch(line)
		if len(progressMatch) > 1 {
			currentProgressMs = timeToMilliseconds(progressMatch[1])
			logger.Debug("Processing progress",
				"current", currentProgressMs,
				"total", totalDurationMs,
				"percent", calculatePercentage(currentProgressMs, totalDurationMs))
		}

		// Check if we need to manually heartbeat (in case our ticker hasn't triggered yet)
		activity.RecordHeartbeat(ctx, map[string]interface{}{
			"progress": currentProgressMs,
			"total":    totalDurationMs,
			"percent":  calculatePercentage(currentProgressMs, totalDurationMs),
		})
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("ffmpeg processing failed: %w", err)
	}

	processingTime := time.Since(startTime)
	logger.Info("FFmpeg processing completed successfully",
		"processingTime", processingTime.String(),
		"outputPath", params.OutputPath)

	return &FFmpegProcessingResult{
		OutputPath:     params.OutputPath,
		DurationMs:     totalDurationMs,
		ProcessingTime: processingTime,
	}, nil
}

// timeToMilliseconds converts ffmpeg time format (HH:MM:SS.MS) to milliseconds
func timeToMilliseconds(timeStr string) int64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, _ := strconv.ParseInt(parts[0], 10, 64)
	minutes, _ := strconv.ParseInt(parts[1], 10, 64)

	secondsParts := strings.Split(parts[2], ".")
	seconds, _ := strconv.ParseInt(secondsParts[0], 10, 64)

	var milliseconds int64
	if len(secondsParts) > 1 {
		// Handle milliseconds part, padding with zeros if needed
		msStr := secondsParts[1]
		if len(msStr) > 3 {
			msStr = msStr[:3]
		} else if len(msStr) < 3 {
			msStr = msStr + strings.Repeat("0", 3-len(msStr))
		}
		milliseconds, _ = strconv.ParseInt(msStr, 10, 64)
	}

	return (hours*3600+minutes*60+seconds)*1000 + milliseconds
}

// calculatePercentage calculates the progress percentage
func calculatePercentage(current, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(current) / float64(total) * 100
}
