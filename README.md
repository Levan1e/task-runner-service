# Task runner service

It's a service you can use for for queuing tasks and tracking their status, that is written in Go.

## Project description:

Task Runner Service is designed for efficient task management. The service supports the queuing of tasks, checking the status of each task, and retrieving results for completed tasks.

### Features:
+ Task queuing with the ability to track the status.
+ Ability to process tasks asynchronously.
+ For faster task processing, Redis is used as a message broker, allowing for quick task queuing and retrieval. It ensures low-latency operations, making the system highly responsive.
+ Healthcheck endpoint for monitoring the service.
+ Docker and Docker Compose support

## Launch options: 
+ ## launch:
        go run .\cmd\runner\main.go   

## API endpoints:
+ ### POST /api/v1/tasks
  This endpoint allows you to create a new task and add it to the queue.
  
  ### Request:
      {
      "name": "task_name",
      "args": [
        {"type": "string", "value": "example_value"}
      ],
      "queue": "optional_queue_name"
      }

  ### Retrieval:
       {
      "id": "task_8b06143a-9012-4cdf-a0cd-2d44c110febd",
      "status": "PENDING"
      }
  
+ ### GET /api/v1/tasks/{id}
  No body required. The id of the task is passed as part of the URL.
  
  ### Retrieval:
      {
      "id": "task_8b06143a-9012-4cdf-a0cd-2d44c110febd",
      "status": "PENDING",
      "created_at": "2025-04-23T13:55:18+03:00"
      }
  
+ ### GET /api/v1/tasks
  This endpoint returns a list of tasks, with the option to filter by status, and paginate the results.
  
  ### Retrieval:
      {
        "tasks": [
          {
            "id": "task_1",
            "status": "SUCCESS"
          }
        ],
        "meta": {
          "limit": 10,
          "offset": 0,
          "total": 1
        }
      }

+ ### GET /api/v1/health
  Checking the service status

  ### Retrieval:
        {"status": "ok"}     
