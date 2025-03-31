package app

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// VideoProcessingWorkflow processes a video file using FFmpeg
func VideoProcessingWorkflow(ctx workflow.Context, params FFmpegProcessingParams) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting video processing workflow",
		"inputPath", params.InputPath,
		"outputPath", params.OutputPath)

	// Activity options with heartbeat timeout and retry policy
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Hour,    // Max time for the activity to complete
		HeartbeatTimeout:    30 * time.Second, // Max time between heartbeats
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,      // Start with 1 second between retries
			BackoffCoefficient: 2.0,              // Double the interval on each retry
			MaximumInterval:    time.Minute * 10, // Up to 10 minutes
			MaximumAttempts:    3,                // Retry up to 3 times
		},
	}

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Execute the FFmpeg processing activity
	var result FFmpegProcessingResult
	err := workflow.ExecuteActivity(ctx, ProcessVideoWithFFmpeg, params).Get(ctx, &result)
	if err != nil {
		logger.Error("FFmpeg processing failed", "error", err)
		return err
	}

	logger.Info("Video processing completed successfully",
		"processingTime", result.ProcessingTime.String(),
		"videoLength", result.DurationMs)

	return nil
}
