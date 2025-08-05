package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	taskconnect "buf.build/gen/go/wcygan/todo/connectrpc/go/task/v1/taskv1connect"
	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// E2ETestConfig holds configuration for end-to-end tests
type E2ETestConfig struct {
	BackendURL  string
	HealthURL   string
	Timeout     time.Duration
	SkipCleanup bool
}

func getE2EConfig() *E2ETestConfig {
	backendURL := os.Getenv("E2E_BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080" // Default for port-forwarded backend
	}

	healthURL := backendURL + "/health"

	timeout := 30 * time.Second
	if timeoutStr := os.Getenv("E2E_TIMEOUT"); timeoutStr != "" {
		if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = parsedTimeout
		}
	}

	return &E2ETestConfig{
		BackendURL:  backendURL,
		HealthURL:   healthURL,
		Timeout:     timeout,
		SkipCleanup: os.Getenv("E2E_SKIP_CLEANUP") == "true",
	}
}

func TestE2E_KubernetesDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	config := getE2EConfig()

	// Check if backend is available
	if !isBackendAvailable(t, config) {
		t.Skip("Backend not available for E2E testing. Make sure to run 'skaffold dev' first.")
	}

	client := taskconnect.NewTaskServiceClient(
		&http.Client{Timeout: config.Timeout},
		config.BackendURL,
	)

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	t.Run("SystemHealth", func(t *testing.T) {
		testSystemHealth(t, config)
	})

	t.Run("DatabasePersistence", func(t *testing.T) {
		testDatabasePersistence(t, ctx, client, config)
	})

	t.Run("FullWorkflow", func(t *testing.T) {
		testFullWorkflow(t, ctx, client, config)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		testErrorHandling(t, ctx, client, config)
	})

	t.Run("PerformanceBaseline", func(t *testing.T) {
		testPerformanceBaseline(t, ctx, client, config)
	})
}

func isBackendAvailable(t *testing.T, config *E2ETestConfig) bool {
	t.Helper()

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(config.HealthURL)
	if err != nil {
		t.Logf("Backend not available: %v", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func testSystemHealth(t *testing.T, config *E2ETestConfig) {
	client := &http.Client{Timeout: config.Timeout}

	t.Run("HealthEndpoint", func(t *testing.T) {
		resp, err := client.Get(config.HealthURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var healthResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&healthResp)
		require.NoError(t, err)

		assert.Equal(t, "healthy", healthResp["status"])
		assert.Equal(t, "mysql", healthResp["database"])
		assert.Equal(t, "mysql", healthResp["store"])
	})

	t.Run("HealthEndpoint_RepeatedCalls", func(t *testing.T) {
		// Test health endpoint stability
		for i := 0; i < 10; i++ {
			resp, err := client.Get(config.HealthURL)
			require.NoError(t, err, "Health check failed on iteration %d", i)
			resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
	})
}

func testDatabasePersistence(t *testing.T, ctx context.Context, client taskconnect.TaskServiceClient, config *E2ETestConfig) {
	// Create a task
	createResp, err := client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
		Description: "E2E persistence test task",
	}))
	require.NoError(t, err)
	require.NotNil(t, createResp.Msg.Task)

	taskID := createResp.Msg.Task.Id

	if !config.SkipCleanup {
		defer func() {
			// Cleanup
			client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
		}()
	}

	// Verify task persists across multiple reads
	for i := 0; i < 5; i++ {
		getResp, err := client.GetTask(ctx, connect.NewRequest(&taskv1.GetTaskRequest{Id: taskID}))
		require.NoError(t, err, "Failed to get task on iteration %d", i)
		assert.Equal(t, taskID, getResp.Msg.Task.Id)
		assert.Equal(t, "E2E persistence test task", getResp.Msg.Task.Description)
	}

	// Update the task
	updateResp, err := client.UpdateTask(ctx, connect.NewRequest(&taskv1.UpdateTaskRequest{
		Id:          taskID,
		Description: "E2E updated persistence test task",
		Completed:   true,
	}))
	require.NoError(t, err)
	assert.Equal(t, "E2E updated persistence test task", updateResp.Msg.Task.Description)
	assert.True(t, updateResp.Msg.Task.Completed)

	// Verify update persisted
	getResp, err := client.GetTask(ctx, connect.NewRequest(&taskv1.GetTaskRequest{Id: taskID}))
	require.NoError(t, err)
	assert.Equal(t, "E2E updated persistence test task", getResp.Msg.Task.Description)
	assert.True(t, getResp.Msg.Task.Completed)
}

func testFullWorkflow(t *testing.T, ctx context.Context, client taskconnect.TaskServiceClient, config *E2ETestConfig) {
	tasksToCleanup := make([]string, 0)

	if !config.SkipCleanup {
		defer func() {
			// Cleanup all created tasks
			for _, taskID := range tasksToCleanup {
				client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
			}
		}()
	}

	// Get initial task count
	initialResp, err := client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
	require.NoError(t, err)
	initialCount := len(initialResp.Msg.Tasks)

	// Create multiple tasks
	taskDescriptions := []string{
		"E2E workflow task 1",
		"E2E workflow task 2",
		"E2E workflow task 3",
		"E2E workflow task with unicode: 测试 🚀",
	}

	createdTasks := make([]*taskv1.Task, 0, len(taskDescriptions))

	for _, desc := range taskDescriptions {
		createResp, err := client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
			Description: desc,
		}))
		require.NoError(t, err)
		require.NotNil(t, createResp.Msg.Task)

		task := createResp.Msg.Task
		createdTasks = append(createdTasks, task)
		tasksToCleanup = append(tasksToCleanup, task.Id)

		assert.NotEmpty(t, task.Id)
		assert.Equal(t, desc, task.Description)
		assert.False(t, task.Completed)
		assert.NotNil(t, task.CreatedAt)
		assert.NotNil(t, task.UpdatedAt)
	}

	// Verify all tasks exist
	listResp, err := client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(listResp.Msg.Tasks), initialCount+len(taskDescriptions))

	// Verify our tasks are in the list
	taskMap := make(map[string]*taskv1.Task)
	for _, task := range listResp.Msg.Tasks {
		taskMap[task.Id] = task
	}

	for _, createdTask := range createdTasks {
		foundTask, exists := taskMap[createdTask.Id]
		assert.True(t, exists, "Created task %s not found in list", createdTask.Id)
		if exists {
			assert.Equal(t, createdTask.Description, foundTask.Description)
		}
	}

	// Update first task
	firstTask := createdTasks[0]
	updateResp, err := client.UpdateTask(ctx, connect.NewRequest(&taskv1.UpdateTaskRequest{
		Id:          firstTask.Id,
		Description: "E2E updated workflow task",
		Completed:   true,
	}))
	require.NoError(t, err)
	assert.Equal(t, "E2E updated workflow task", updateResp.Msg.Task.Description)
	assert.True(t, updateResp.Msg.Task.Completed)

	// Delete second task
	secondTask := createdTasks[1]
	deleteResp, err := client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{
		Id: secondTask.Id,
	}))
	require.NoError(t, err)
	assert.True(t, deleteResp.Msg.Success)

	// Remove from cleanup list since it's already deleted
	for i, id := range tasksToCleanup {
		if id == secondTask.Id {
			tasksToCleanup = append(tasksToCleanup[:i], tasksToCleanup[i+1:]...)
			break
		}
	}

	// Verify final state
	finalResp, err := client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
	require.NoError(t, err)

	finalTaskMap := make(map[string]*taskv1.Task)
	for _, task := range finalResp.Msg.Tasks {
		finalTaskMap[task.Id] = task
	}

	// First task should be updated
	if updatedTask, exists := finalTaskMap[firstTask.Id]; exists {
		assert.Equal(t, "E2E updated workflow task", updatedTask.Description)
		assert.True(t, updatedTask.Completed)
	}

	// Second task should be deleted
	_, exists := finalTaskMap[secondTask.Id]
	assert.False(t, exists, "Deleted task still exists in list")
}

func testErrorHandling(t *testing.T, ctx context.Context, client taskconnect.TaskServiceClient, config *E2ETestConfig) {
	t.Run("InvalidRequests", func(t *testing.T) {
		// Empty description
		_, err := client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
			Description: "",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
	})

	t.Run("NotFoundErrors", func(t *testing.T) {
		// Get non-existent task
		_, err := client.GetTask(ctx, connect.NewRequest(&taskv1.GetTaskRequest{
			Id: "999999",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())

		// Update non-existent task
		_, err = client.UpdateTask(ctx, connect.NewRequest(&taskv1.UpdateTaskRequest{
			Id:          "999999",
			Description: "Should fail",
			Completed:   false,
		}))
		require.Error(t, err)
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())

		// Delete non-existent task (returns success=false instead of error)
		deleteResp, err := client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{
			Id: "999999",
		}))
		require.NoError(t, err)
		assert.False(t, deleteResp.Msg.Success)
		assert.Contains(t, deleteResp.Msg.Message, "not found")
	})
}

func testPerformanceBaseline(t *testing.T, ctx context.Context, client taskconnect.TaskServiceClient, config *E2ETestConfig) {
	tasksToCleanup := make([]string, 0)

	if !config.SkipCleanup {
		defer func() {
			// Cleanup
			for _, taskID := range tasksToCleanup {
				client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
			}
		}()
	}

	t.Run("CreateTaskPerformance", func(t *testing.T) {
		const numTasks = 50
		start := time.Now()

		for i := 0; i < numTasks; i++ {
			createResp, err := client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
				Description: fmt.Sprintf("Performance test task %d", i),
			}))
			require.NoError(t, err)
			tasksToCleanup = append(tasksToCleanup, createResp.Msg.Task.Id)
		}

		duration := time.Since(start)
		avgDuration := duration / numTasks

		t.Logf("Created %d tasks in %v (avg: %v per task)", numTasks, duration, avgDuration)

		// Baseline: should be able to create tasks reasonably fast
		// This is a loose check - adjust based on your performance requirements
		assert.Less(t, avgDuration, 200*time.Millisecond, "Task creation is too slow")
	})

	t.Run("ListTasksPerformance", func(t *testing.T) {
		start := time.Now()

		listResp, err := client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
		require.NoError(t, err)

		duration := time.Since(start)
		taskCount := len(listResp.Msg.Tasks)

		t.Logf("Listed %d tasks in %v", taskCount, duration)

		// Should be able to list tasks quickly
		assert.Less(t, duration, 1*time.Second, "Task listing is too slow")
	})

	t.Run("GetTaskPerformance", func(t *testing.T) {
		if len(tasksToCleanup) == 0 {
			t.Skip("No tasks available for get performance test")
		}

		taskID := tasksToCleanup[0]
		const numGets = 20

		start := time.Now()
		for i := 0; i < numGets; i++ {
			_, err := client.GetTask(ctx, connect.NewRequest(&taskv1.GetTaskRequest{
				Id: taskID,
			}))
			require.NoError(t, err)
		}
		duration := time.Since(start)
		avgDuration := duration / numGets

		t.Logf("Got task %d times in %v (avg: %v per get)", numGets, duration, avgDuration)

		// Should be able to get individual tasks quickly
		assert.Less(t, avgDuration, 100*time.Millisecond, "Task retrieval is too slow")
	})
}
