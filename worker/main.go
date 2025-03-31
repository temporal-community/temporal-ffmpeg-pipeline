package main

import (
	"log"
	"temporal-ffmpeg-pipeline/app"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// Create the client object just once per process
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()

	// Create a worker
	w := worker.New(c, "video-processing-task-queue", worker.Options{})

	// Register workflow and activity
	w.RegisterWorkflow(app.VideoProcessingWorkflow)
	w.RegisterActivity(app.ProcessVideoWithFFmpeg)

	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}
