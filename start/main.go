package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"go.temporal.io/sdk/client"

	"temporal-ffmpeg-pipeline/app"
)

// Command-line flags
var (
	startWorkflowMode = flag.Bool("start", false, "Start a new workflow")
	workflowID        = flag.String("workflow-id", "", "Workflow ID (defaults to auto-generated)")
	inputPath         = flag.String("input", "", "Input video file path")
	outputPath        = flag.String("output", "", "Output video file path")
	ffmpegArgs        = flag.String("ffmpeg-args", "-c:v libx264 -preset medium -crf 23 -c:a copy", "FFmpeg arguments (space-separated)")
)

// In a separate file or program, you can start the workflow:
func main() {
	flag.Parse()

	// Create a Temporal client
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	// Validate required flags
	if *inputPath == "" {
		log.Fatal("Input path is required. Use -input flag.")
	}
	if *outputPath == "" {
		log.Fatal("Output path is required. Use -output flag.")
	}

	// Parse FFmpeg arguments
	ffmpegArgsList := strings.Fields(*ffmpegArgs)

	// Prepare the workflow parameters
	params := app.FFmpegProcessingParams{
		InputPath:  *inputPath,
		OutputPath: *outputPath,
		FFmpegArgs: *ffmpegArgs,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        "video-processing-" + time.Now().Format("20060102-150405"),
		TaskQueue: "video-processing-task-queue",
	}

	log.Printf("Starting workflow with ID: %s", workflowOptions.ID)
	log.Printf("Input: %s", params.InputPath)
	log.Printf("Output: %s", params.OutputPath)
	log.Printf("FFmpeg args: %v", params.FFmpegArgs)

	// Start the workflow
	execution, err := c.ExecuteWorkflow(
		context.Background(),
		workflowOptions,
		app.VideoProcessingWorkflow,
		params,
	)
	if err != nil {
		log.Fatalf("Failed to start workflow: %v", err)
	}

	log.Printf("Started workflow, ID: %s, RunID: %s", execution.GetID(), execution.GetRunID())

	// Wait for workflow completion
	var result interface{}
	err = execution.Get(context.Background(), &result)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	log.Printf("Workflow completed successfully")
}
