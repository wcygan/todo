package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"connectrpc.com/connect"
	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupIntegrationTest sets up the shared integration test suite
func setupIntegrationTest(t *testing.T) *SharedIntegrationSuite {
	return GetSharedIntegrationSuite(t)
}


func TestIntegration_DatabasePersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite := setupIntegrationTest(t)

	ctx := context.Background()

	t.Run("TaskPersistence_BasicWorkflow", func(t *testing.T) {
		// Create multiple tasks
		tasks := []string{
			"Integration test task 1",
			"Integration test task 2",
			"Integration test task 3",
		}

		createdTasks := make([]*taskv1.Task, 0, len(tasks))

		// Create tasks
		for _, desc := range tasks {
			resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
				Description: desc,
			}))
			require.NoError(t, err)
			require.NotNil(t, resp.Msg.Task)

			task := resp.Msg.Task
			assert.NotEmpty(t, task.Id)
			assert.Equal(t, desc, task.Description)
			assert.False(t, task.Completed)
			assert.NotNil(t, task.CreatedAt)
			assert.NotNil(t, task.UpdatedAt)

			createdTasks = append(createdTasks, task)
		}

		// Verify all tasks exist
		listResp, err := suite.Client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
		require.NoError(t, err)
		assert.Len(t, listResp.Msg.Tasks, len(tasks))

		// Update a task
		taskToUpdate := createdTasks[0]
		_, err = suite.Client.UpdateTask(ctx, connect.NewRequest(&taskv1.UpdateTaskRequest{
			Id:          taskToUpdate.Id,
			Description: "Updated integration task",
			Completed:   true,
		}))
		require.NoError(t, err)

		// Verify update persisted
		listResp, err = suite.Client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
		require.NoError(t, err)

		var updatedTask *taskv1.Task
		for _, task := range listResp.Msg.Tasks {
			if task.Id == taskToUpdate.Id {
				updatedTask = task
				break
			}
		}
		require.NotNil(t, updatedTask)
		assert.Equal(t, "Updated integration task", updatedTask.Description)
		assert.True(t, updatedTask.Completed)

		// Delete a task
		taskToDelete := createdTasks[1]
		deleteResp, err := suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{
			Id: taskToDelete.Id,
		}))
		require.NoError(t, err)
		assert.True(t, deleteResp.Msg.Success)

		// Verify deletion persisted
		listResp, err = suite.Client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
		require.NoError(t, err)
		assert.Len(t, listResp.Msg.Tasks, len(tasks)-1)

		// Verify deleted task is not in list
		for _, task := range listResp.Msg.Tasks {
			assert.NotEqual(t, taskToDelete.Id, task.Id)
		}
	})

	t.Run("TaskPersistence_LargeDataset", func(t *testing.T) {
		const numTasks = 100

		// Create many tasks
		createdIDs := make([]string, 0, numTasks)
		for i := 0; i < numTasks; i++ {
			resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
				Description: fmt.Sprintf("Bulk task %d", i),
			}))
			require.NoError(t, err)
			createdIDs = append(createdIDs, resp.Msg.Task.Id)
		}

		// Verify all tasks exist
		listResp, err := suite.Client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listResp.Msg.Tasks), numTasks)

		// Verify tasks are ordered correctly (newest first)
		tasks := listResp.Msg.Tasks
		for i := 1; i < len(tasks); i++ {
			assert.True(t, tasks[i-1].CreatedAt.AsTime().After(tasks[i].CreatedAt.AsTime()) ||
				tasks[i-1].CreatedAt.AsTime().Equal(tasks[i].CreatedAt.AsTime()))
		}

		// Clean up - delete all created tasks
		for _, id := range createdIDs {
			_, err := suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{
				Id: id,
			}))
			assert.NoError(t, err)
		}
	})
}

func TestIntegration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite := setupIntegrationTest(t)

	ctx := context.Background()

	t.Run("DatabaseError_InvalidOperations", func(t *testing.T) {
		// Try to get non-existent task
		_, err := suite.Client.GetTask(ctx, connect.NewRequest(&taskv1.GetTaskRequest{
			Id: "99999",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())

		// Try to update non-existent task
		_, err = suite.Client.UpdateTask(ctx, connect.NewRequest(&taskv1.UpdateTaskRequest{
			Id:          "99999",
			Description: "Should fail",
			Completed:   false,
		}))
		require.Error(t, err)
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())

		// Try to delete non-existent task
		deleteResp, err := suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{
			Id: "99999",
		}))
		require.NoError(t, err) // Delete endpoint returns success=false instead of error
		assert.False(t, deleteResp.Msg.Success)
		assert.Contains(t, deleteResp.Msg.Message, "not found")
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		// Empty description
		_, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
			Description: "",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
		assert.Contains(t, connectErr.Message(), "description cannot be empty")
	})
}

func TestIntegration_HealthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite := setupIntegrationTest(t)

	t.Run("HealthCheck_DatabaseConnected", func(t *testing.T) {
		resp, err := http.Get(suite.Server.URL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		body := make([]byte, 1024)
		n, err := resp.Body.Read(body)
		require.NoError(t, err)

		bodyStr := string(body[:n])
		assert.Contains(t, bodyStr, `"status":"healthy"`)
		assert.Contains(t, bodyStr, `"database":"mysql"`)
	})
}

func TestIntegration_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite := setupIntegrationTest(t)

	ctx := context.Background()

	t.Run("ConcurrentTaskCreation", func(t *testing.T) {
		const numGoroutines = 20
		const tasksPerGoroutine = 10

		type result struct {
			taskID string
			err    error
		}

		resultChan := make(chan result, numGoroutines*tasksPerGoroutine)

		// Create tasks concurrently
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				for j := 0; j < tasksPerGoroutine; j++ {
					resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
						Description: fmt.Sprintf("Concurrent task G%d-T%d", goroutineID, j),
					}))
					if err != nil {
						resultChan <- result{"", err}
					} else {
						resultChan <- result{resp.Msg.Task.Id, nil}
					}
				}
			}(i)
		}

		// Collect results
		createdIDs := make(map[string]bool)
		var errors []error

		for i := 0; i < numGoroutines*tasksPerGoroutine; i++ {
			res := <-resultChan
			if res.err != nil {
				errors = append(errors, res.err)
			} else {
				assert.False(t, createdIDs[res.taskID], "Duplicate task ID: %s", res.taskID)
				createdIDs[res.taskID] = true
			}
		}

		// No errors should occur
		assert.Empty(t, errors, "Errors during concurrent creation: %v", errors)
		assert.Len(t, createdIDs, numGoroutines*tasksPerGoroutine)

		// Verify all tasks exist in database
		listResp, err := suite.Client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listResp.Msg.Tasks), numGoroutines*tasksPerGoroutine)

		// Clean up
		for id := range createdIDs {
			suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: id}))
		}
	})

	t.Run("ConcurrentMixedOperations", func(t *testing.T) {
		// Create some initial tasks
		initialTasks := make([]string, 10)
		for i := 0; i < 10; i++ {
			resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
				Description: fmt.Sprintf("Initial task %d", i),
			}))
			require.NoError(t, err)
			initialTasks[i] = resp.Msg.Task.Id
		}

		const numOperations = 50
		operationChan := make(chan error, numOperations)

		// Perform mixed operations concurrently
		for i := 0; i < numOperations; i++ {
			go func(opID int) {
				switch opID % 4 {
				case 0: // Create
					_, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
						Description: fmt.Sprintf("Concurrent create %d", opID),
					}))
					operationChan <- err

				case 1: // Update (if we have tasks)
					if len(initialTasks) > 0 {
						taskID := initialTasks[opID%len(initialTasks)]
						_, err := suite.Client.UpdateTask(ctx, connect.NewRequest(&taskv1.UpdateTaskRequest{
							Id:          taskID,
							Description: fmt.Sprintf("Updated %d", opID),
							Completed:   opID%2 == 0,
						}))
						operationChan <- err
					} else {
						operationChan <- nil
					}

				case 2: // List
					_, err := suite.Client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
					operationChan <- err

				case 3: // Get (if we have tasks)
					if len(initialTasks) > 0 {
						taskID := initialTasks[opID%len(initialTasks)]
						_, err := suite.Client.GetTask(ctx, connect.NewRequest(&taskv1.GetTaskRequest{
							Id: taskID,
						}))
						operationChan <- err
					} else {
						operationChan <- nil
					}
				}
			}(i)
		}

		// Collect results
		var errors []error
		for i := 0; i < numOperations; i++ {
			if err := <-operationChan; err != nil {
				errors = append(errors, err)
			}
		}

		assert.Empty(t, errors, "Errors during concurrent mixed operations: %v", errors)
	})
}