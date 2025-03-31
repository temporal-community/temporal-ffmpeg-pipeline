# Temporal ffmpeg Pipeline

This is a sample Temporal project that can be used to convert video with ffmpeg, using heartbeating. 

## Instructions

Ensure you have Go 1.23 or later installed locally, and a local [Temporal Cluster](https://docs.temporal.io/cli) running.

You should be able to view your local cluster's Temporal Web UI at <http://localhost:8233>.

Clone this repository:

```bash
git clone https://github.com/temporal-community/temporal-ffmpeg-pipeline
```

Run the worker included in the project:

```bash
go run worker/main.go
```

Then, run the starter, and provide both an `--input` and an `--output` path:

```bash
go run start/main.go --input path/to/video.mp4 --output path/to/output.mkv
```
