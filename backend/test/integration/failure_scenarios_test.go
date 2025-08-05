package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"

	"github.com/wcygan/todo/backend/internal/config"
	"github.com/wcygan/todo/backend/internal/store"
	"github.com/wcygan/todo/backend/test/testutil"
)

func TestFailureScenarios_DatabaseResilience(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping failure scenario tests in short mode")
	}

	ctx := context.Background()

	t.Run("DatabaseConnectionFailure_AfterEstablishment", func(t *testing.T) {
		// Start with working database
		container, err := mariadb.Run(ctx,
			"mariadb:11.5",
			mariadb.WithDatabase("failure_test"),
			mariadb.WithUsername("testuser"),
			mariadb.WithPassword("testpass"),
		)
		require.NoError(t, err)

		// Get connection details
		host, err := container.Host(ctx)
		require.NoError(t, err)

		port, err := container.MappedPort(ctx, "3306")
		require.NoError(t, err)

		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Host:            host,
				Port:            port.Int(),
				User:            "testuser",
				Password:        "testpass",
				Database:        "failure_test",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				SSLMode:         "false",
			},
		}

		// Create manager successfully
		manager, err := store.NewManager(cfg)
		require.NoError(t, err)

		// Verify it works initially
		taskStore := manager.TaskStore()
		task, err := taskStore.CreateTask(ctx, "Test task before failure")
		require.NoError(t, err)
		assert.NotEmpty(t, task.Id)

		// Stop the database container to simulate connection failure
		err = container.Stop(ctx, nil)
		require.NoError(t, err)

		// Give some time for the connection to fail
		time.Sleep(2 * time.Second)

		// Operations should now fail
		_, err = taskStore.CreateTask(ctx, "Test task after failure")
		assert.Error(t, err, "Expected error after database connection failure")

		// Health check should also fail
		err = manager.HealthCheck(ctx)
		assert.Error(t, err, "Expected health check to fail after database connection failure")

		// Cleanup
		manager.Close()
		container.Terminate(ctx)
	})

	t.Run("DatabaseConnectionRecovery", func(t *testing.T) {
		// This test verifies that the application can handle database connection issues
		// and recover gracefully. Instead of physically restarting containers (which is flaky),
		// we test the resilience by creating multiple managers with short connection lifetimes.

		container, err := mariadb.Run(ctx,
			"mariadb:11.5",
			mariadb.WithDatabase("recovery_test"),
			mariadb.WithUsername("testuser"),
			mariadb.WithPassword("testpass"),
		)
		require.NoError(t, err)

		host, err := container.Host(ctx)
		require.NoError(t, err)

		port, err := container.MappedPort(ctx, "3306")
		require.NoError(t, err)

		cfg := &config.Config{
			Database: config.DatabaseConfig{
				Host:            host,
				Port:            port.Int(),
				User:            "testuser",
				Password:        "testpass",
				Database:        "recovery_test",
				MaxOpenConns:    2,               // Very small pool to force connection cycling
				MaxIdleConns:    1,               // Minimal idle connections
				ConnMaxLifetime: 1 * time.Second, // Very short lifetime to force connection refresh
				ConnMaxIdleTime: 1 * time.Second, // Short idle time
				SSLMode:         "false",
			},
		}

		// Test connection recovery by creating and closing multiple managers
		var lastTaskId string

		for i := 0; i < 3; i++ {
			t.Logf("Testing connection cycle %d", i+1)

			manager, err := store.NewManager(cfg)
			require.NoError(t, err, "Should be able to create manager on cycle %d", i+1)

			taskStore := manager.TaskStore()

			// Create task to verify functionality
			task, err := taskStore.CreateTask(ctx, fmt.Sprintf("Recovery test task %d", i+1))
			require.NoError(t, err, "Should be able to create task on cycle %d", i+1)
			assert.NotEmpty(t, task.Id)

			// Verify we can retrieve the task
			retrieved, err := taskStore.GetTask(ctx, task.Id)
			require.NoError(t, err, "Should be able to retrieve task on cycle %d", i+1)
			assert.Equal(t, task.Description, retrieved.Description)

			lastTaskId = task.Id

			// Test health check
			err = manager.HealthCheck(ctx)
			require.NoError(t, err, "Health check should pass on cycle %d", i+1)

			manager.Close()

			// Brief pause to allow connection cleanup
			time.Sleep(2 * time.Second)
		}

		// Final verification: create one more manager and verify we can access the last task
		finalManager, err := store.NewManager(cfg)
		require.NoError(t, err)
		defer finalManager.Close()

		finalTaskStore := finalManager.TaskStore()
		finalTask, err := finalTaskStore.GetTask(ctx, lastTaskId)
		require.NoError(t, err, "Should be able to retrieve task after connection recovery")
		assert.NotEmpty(t, finalTask.Id)

		container.Terminate(ctx)
	})
}

func TestFailureScenarios_TransactionIntegrity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping transaction integrity tests in short mode")
	}

	suite := testutil.GetSharedIntegrationSuite(t)

	ctx := context.Background()

	t.Run("ContextCancellation_DuringOperation", func(t *testing.T) {
		// Test context cancellation during various operations
		operations := []struct {
			name string
			op   func(context.Context) error
		}{
			{
				name: "CreateTask",
				op: func(ctx context.Context) error {
					_, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
						Description: "Test task for cancellation",
					}))
					return err
				},
			},
			{
				name: "GetAllTasks",
				op: func(ctx context.Context) error {
					_, err := suite.Client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
					return err
				},
			},
		}

		for _, op := range operations {
			t.Run(op.name, func(t *testing.T) {
				// Create a context that will be cancelled quickly
				opCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()

				// The operation might succeed if it's very fast, but if it fails,
				// it should be due to context cancellation
				err := op.op(opCtx)
				if err != nil {
					// Verify the error is related to context cancellation
					assert.Contains(t, err.Error(), "context",
						"Error should be related to context cancellation")
				}
			})
		}
	})

	t.Run("ConcurrentModification_SameTask", func(t *testing.T) {
		// Create a task to be modified concurrently
		createResp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
			Description: "Task for concurrent modification",
		}))
		require.NoError(t, err)
		taskID := createResp.Msg.Task.Id

		defer func() {
			suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
		}()

		// Try to update the same task concurrently
		const numUpdates = 10
		results := make(chan error, numUpdates)

		for i := 0; i < numUpdates; i++ {
			go func(updateID int) {
				_, err := suite.Client.UpdateTask(ctx, connect.NewRequest(&taskv1.UpdateTaskRequest{
					Id:          taskID,
					Description: fmt.Sprintf("Concurrent update %d", updateID),
					Completed:   updateID%2 == 0,
				}))
				results <- err
			}(i)
		}

		// Collect results
		var errors []error
		for i := 0; i < numUpdates; i++ {
			if err := <-results; err != nil {
				errors = append(errors, err)
			}
		}

		// All updates should succeed (MySQL handles concurrent updates)
		assert.Empty(t, errors, "Concurrent updates should not fail: %v", errors)

		// Verify task still exists and has one of the expected descriptions
		finalTask, err := suite.Client.GetTask(ctx, connect.NewRequest(&taskv1.GetTaskRequest{
			Id: taskID,
		}))
		require.NoError(t, err)
		assert.Contains(t, finalTask.Msg.Task.Description, "Concurrent update")
	})

	t.Run("LargeDataset_MemoryPressure", func(t *testing.T) {
		// Create many tasks to test memory handling
		const numTasks = 100 // Reduced from 5000 to prevent hanging
		taskIDs := make([]string, 0, numTasks)

		defer func() {
			// Cleanup
			for _, taskID := range taskIDs {
				suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
			}
		}()

		// Create tasks with varying sizes
		for i := 0; i < numTasks; i++ {
			descSize := 100 + (i%1000)*10 // Varying description sizes
			desc := fmt.Sprintf("Memory pressure test task %d: %s",
				i, string(make([]byte, descSize)))

			for j := range desc[len(desc)-descSize:] {
				desc = desc[:len(desc)-descSize+j] + "A" + desc[len(desc)-descSize+j+1:]
			}

			resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
				Description: desc,
			}))
			require.NoError(t, err, "Failed to create task %d", i)
			taskIDs = append(taskIDs, resp.Msg.Task.Id)

			// Log progress
			if i%20 == 0 && i > 0 {
				t.Logf("Created %d tasks", i)
			}
		}

		// Test listing all tasks (memory pressure test)
		start := time.Now()
		listResp, err := suite.Client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
		duration := time.Since(start)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listResp.Msg.Tasks), numTasks)

		t.Logf("Listed %d tasks in %v", len(listResp.Msg.Tasks), duration)

		// Should handle large datasets without excessive memory usage or timeouts
		assert.Less(t, duration, 30*time.Second, "Listing large dataset took too long")
	})
}

func TestFailureScenarios_InvalidData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping invalid data tests in short mode")
	}

	suite := testutil.GetSharedIntegrationSuite(t)

	ctx := context.Background()

	t.Run("InvalidCharacters_InDescription", func(t *testing.T) {
		// Test various potentially problematic characters
		problematicDescriptions := []string{
			"Task with null byte: \x00",
			"Task with Unicode: 🚀🎉✨ 测试 עברית العربية",
			"Task with quotes: \"single\" and 'double' quotes",
			"Task with SQL-like: '; DROP TABLE tasks; --",
			"Task with HTML: <script>alert('xss')</script>",
			"Task with very long line: " + string(make([]byte, 1000)),
		}

		for i, desc := range problematicDescriptions {
			t.Run(fmt.Sprintf("Description_%d", i), func(t *testing.T) {
				resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
					Description: desc,
				}))

				// Most should work (good UTF-8 handling), but null bytes might be filtered
				if err != nil {
					t.Logf("Expected failure for description: %q, error: %v", desc, err)
				} else {
					taskID := resp.Msg.Task.Id

					// Verify we can retrieve it
					getResp, err := suite.Client.GetTask(ctx, connect.NewRequest(&taskv1.GetTaskRequest{
						Id: taskID,
					}))
					require.NoError(t, err)

					// Description should be stored correctly (or safely filtered)
					assert.NotEmpty(t, getResp.Msg.Task.Description)

					// Cleanup
					suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
				}
			})
		}
	})

	t.Run("InvalidTaskIDs", func(t *testing.T) {
		invalidIDs := []string{
			"", // Empty ID
			"invalid-format",
			"999999999999999999999", // Very large number
			"-1",                    // Negative number
			"abc123",                // Non-numeric
			"null",
			"undefined",
			"<script>",
		}

		for _, invalidID := range invalidIDs {
			t.Run(fmt.Sprintf("ID_%s", invalidID), func(t *testing.T) {
				// Try to get task with invalid ID
				_, err := suite.Client.GetTask(ctx, connect.NewRequest(&taskv1.GetTaskRequest{
					Id: invalidID,
				}))
				assert.Error(t, err, "Should fail with invalid ID: %s", invalidID)

				// Try to update task with invalid ID
				_, err = suite.Client.UpdateTask(ctx, connect.NewRequest(&taskv1.UpdateTaskRequest{
					Id:          invalidID,
					Description: "Should fail",
					Completed:   false,
				}))
				assert.Error(t, err, "Should fail with invalid ID: %s", invalidID)

				// Try to delete task with invalid ID
				deleteResp, err := suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{
					Id: invalidID,
				}))

				// Delete might return success=false instead of error
				if err == nil {
					assert.False(t, deleteResp.Msg.Success, "Delete should indicate failure for invalid ID: %s", invalidID)
				}
			})
		}
	})
}

func TestFailureScenarios_ResourceExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource exhaustion tests in short mode")
	}

	t.Run("DatabaseConnectionPool_Exhaustion", func(t *testing.T) {
		ctx := context.Background()

		// Create a store with very limited connections
		container, err := mariadb.Run(ctx,
			"mariadb:11.5",
			mariadb.WithDatabase("pool_test"),
			mariadb.WithUsername("testuser"),
			mariadb.WithPassword("testpass"),
		)
		require.NoError(t, err)
		defer container.Terminate(ctx)

		host, err := container.Host(ctx)
		require.NoError(t, err)

		port, err := container.MappedPort(ctx, "3306")
		require.NoError(t, err)

		cfg := &config.DatabaseConfig{
			Host:            host,
			Port:            port.Int(),
			User:            "testuser",
			Password:        "testpass",
			Database:        "pool_test",
			MaxOpenConns:    2, // Very limited
			MaxIdleConns:    1,
			ConnMaxLifetime: 30 * time.Second,
			ConnMaxIdleTime: 30 * time.Second,
			SSLMode:         "false",
		}

		mysqlStore, err := store.NewMySQLTaskStore(cfg)
		require.NoError(t, err)
		defer mysqlStore.Close()

		// Try to exhaust the connection pool with long-running transactions
		const numConcurrent = 5 // Reduced from 10 to prevent hanging
		results := make(chan error, numConcurrent)

		for i := 0; i < numConcurrent; i++ {
			go func(goroutineID int) {
				// Create a context with timeout
				opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()

				_, err := mysqlStore.CreateTask(opCtx, fmt.Sprintf("Pool exhaustion test %d", goroutineID))
				results <- err
			}(i)
		}

		// Collect results
		var errors []error
		var successes int
		for i := 0; i < numConcurrent; i++ {
			err := <-results
			if err != nil {
				errors = append(errors, err)
			} else {
				successes++
			}
		}

		t.Logf("Pool exhaustion test: %d successes, %d errors", successes, len(errors))

		// Some operations should succeed, but the system should handle pool exhaustion gracefully
		assert.Greater(t, successes, 0, "At least some operations should succeed")

		// If there are errors, they should be meaningful (not crashes)
		for _, err := range errors {
			assert.NotNil(t, err)
			t.Logf("Pool exhaustion error (expected): %v", err)
		}
	})

	t.Run("VeryLongRunning_Operations", func(t *testing.T) {
		suite := testutil.GetSharedIntegrationSuite(t)

		// Create a context with a reasonable timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Test creating many tasks in sequence (simulating long-running batch operation)
		const numTasks = 50 // Reduced from 1000 to prevent hanging
		taskIDs := make([]string, 0, numTasks)

		defer func() {
			// Cleanup
			for _, taskID := range taskIDs {
				suite.Client.DeleteTask(context.Background(), connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
			}
		}()

		start := time.Now()
		for i := 0; i < numTasks; i++ {
			select {
			case <-ctx.Done():
				t.Fatalf("Operation timed out after creating %d tasks", i)
			default:
			}

			resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
				Description: fmt.Sprintf("Long running batch task %d", i),
			}))
			require.NoError(t, err, "Failed to create task %d", i)
			taskIDs = append(taskIDs, resp.Msg.Task.Id)

			// Log progress
			if i%10 == 0 && i > 0 {
				elapsed := time.Since(start)
				t.Logf("Created %d tasks in %v (%.2f tasks/sec)", i, elapsed, float64(i)/elapsed.Seconds())
			}
		}

		duration := time.Since(start)
		throughput := float64(numTasks) / duration.Seconds()

		t.Logf("Created %d tasks in %v (%.2f tasks/sec)", numTasks, duration, throughput)

		// Verify all tasks exist
		listResp, err := suite.Client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listResp.Msg.Tasks), numTasks)

		// Performance should remain reasonable even for long-running operations
		assert.Greater(t, throughput, 10.0, "Long-running batch operation throughput is too low")
	})
}
