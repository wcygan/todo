package integration

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	taskconnect "buf.build/gen/go/wcygan/todo/connectrpc/go/task/v1/taskv1connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



func TestPerformance_DatabaseOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	suite := setupIntegrationTest(t)

	ctx := context.Background()

	t.Run("BenchmarkTaskCreation", func(t *testing.T) {
		const numTasks = 1000
		tasksToCleanup := make([]string, 0, numTasks)

		defer func() {
			// Cleanup
			for _, taskID := range tasksToCleanup {
				suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
			}
		}()

		start := time.Now()

		for i := 0; i < numTasks; i++ {
			resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
				Description: fmt.Sprintf("Benchmark task %d", i),
			}))
			require.NoError(t, err)
			tasksToCleanup = append(tasksToCleanup, resp.Msg.Task.Id)
		}

		duration := time.Since(start)
		throughput := float64(numTasks) / duration.Seconds()

		t.Logf("Created %d tasks in %v (%.2f tasks/sec)", numTasks, duration, throughput)
		
		// Performance assertion - should be able to create at least 100 tasks/sec
		assert.Greater(t, throughput, 100.0, "Task creation throughput is too low")
	})

	t.Run("BenchmarkConcurrentTaskCreation", func(t *testing.T) {
		const numGoroutines = 50
		const tasksPerGoroutine = 20
		const totalTasks = numGoroutines * tasksPerGoroutine

		var wg sync.WaitGroup
		var successCount int64
		var errorCount int64
		taskIDs := make(chan string, totalTasks)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				
				for j := 0; j < tasksPerGoroutine; j++ {
					resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
						Description: fmt.Sprintf("Concurrent task G%d-T%d", goroutineID, j),
					}))
					
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
						taskIDs <- resp.Msg.Task.Id
					}
				}
			}(i)
		}

		wg.Wait()
		close(taskIDs)

		duration := time.Since(start)
		throughput := float64(successCount) / duration.Seconds()

		t.Logf("Created %d tasks concurrently in %v (%.2f tasks/sec, %d errors)", 
			successCount, duration, throughput, errorCount)

		// Cleanup
		go func() {
			for taskID := range taskIDs {
				suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
			}
		}()

		// Performance assertions
		assert.Equal(t, int64(totalTasks), successCount, "Not all tasks were created successfully")
		assert.Equal(t, int64(0), errorCount, "There should be no errors during concurrent creation")
		assert.Greater(t, throughput, 200.0, "Concurrent task creation throughput is too low")
	})

	t.Run("BenchmarkTaskRetrieval", func(t *testing.T) {
		// Create test tasks first
		const numTasks = 100
		taskIDs := make([]string, 0, numTasks)

		for i := 0; i < numTasks; i++ {
			resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
				Description: fmt.Sprintf("Retrieval benchmark task %d", i),
			}))
			require.NoError(t, err)
			taskIDs = append(taskIDs, resp.Msg.Task.Id)
		}

		defer func() {
			// Cleanup
			for _, taskID := range taskIDs {
				suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
			}
		}()

		// Benchmark individual task retrieval
		const numRetrieves = 1000
		start := time.Now()

		for i := 0; i < numRetrieves; i++ {
			taskID := taskIDs[i%len(taskIDs)]
			_, err := suite.Client.GetTask(ctx, connect.NewRequest(&taskv1.GetTaskRequest{
				Id: taskID,
			}))
			require.NoError(t, err)
		}

		duration := time.Since(start)
		throughput := float64(numRetrieves) / duration.Seconds()

		t.Logf("Retrieved tasks %d times in %v (%.2f retrievals/sec)", numRetrieves, duration, throughput)
		
		// Should be able to retrieve tasks quickly
		assert.Greater(t, throughput, 500.0, "Task retrieval throughput is too low")
	})

	t.Run("BenchmarkTaskListing", func(t *testing.T) {
		// Create a moderate number of tasks
		const numTasks = 500
		taskIDs := make([]string, 0, numTasks)

		for i := 0; i < numTasks; i++ {
			resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
				Description: fmt.Sprintf("Listing benchmark task %d", i),
			}))
			require.NoError(t, err)
			taskIDs = append(taskIDs, resp.Msg.Task.Id)
		}

		defer func() {
			// Cleanup
			for _, taskID := range taskIDs {
				suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
			}
		}()

		// Benchmark listing operations
		const numListings = 100
		start := time.Now()

		for i := 0; i < numListings; i++ {
			resp, err := suite.Client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(resp.Msg.Tasks), numTasks)
		}

		duration := time.Since(start)
		throughput := float64(numListings) / duration.Seconds()

		t.Logf("Listed %d tasks %d times in %v (%.2f listings/sec)", numTasks, numListings, duration, throughput)
		
		// Should be able to list tasks reasonably quickly even with many tasks
		assert.Greater(t, throughput, 50.0, "Task listing throughput is too low")
	})

	t.Run("BenchmarkMixedOperations", func(t *testing.T) {
		const duration = 30 * time.Second
		const numWorkers = 20

		var wg sync.WaitGroup
		var operations int64
		var errors int64
		taskIDs := make(chan string, 1000)

		// Start workers
		start := time.Now()
		deadline := start.Add(duration)

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				localTaskIDs := make([]string, 0, 10)
				
				for time.Now().Before(deadline) {
					operation := time.Now().UnixNano() % 4
					
					switch operation {
					case 0: // Create task (40% of operations)
						resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
							Description: fmt.Sprintf("Mixed ops task W%d", workerID),
						}))
						if err != nil {
							atomic.AddInt64(&errors, 1)
						} else {
							localTaskIDs = append(localTaskIDs, resp.Msg.Task.Id)
							select {
							case taskIDs <- resp.Msg.Task.Id:
							default:
							}
						}
						
					case 1: // Get task (30% of operations)
						if len(localTaskIDs) > 0 {
							taskID := localTaskIDs[len(localTaskIDs)-1]
							_, err := suite.Client.GetTask(ctx, connect.NewRequest(&taskv1.GetTaskRequest{
								Id: taskID,
							}))
							if err != nil {
								atomic.AddInt64(&errors, 1)
							}
						}
						
					case 2: // List tasks (20% of operations)
						_, err := suite.Client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
						if err != nil {
							atomic.AddInt64(&errors, 1)
						}
						
					case 3: // Update task (10% of operations)
						if len(localTaskIDs) > 0 {
							taskID := localTaskIDs[len(localTaskIDs)-1]
							_, err := suite.Client.UpdateTask(ctx, connect.NewRequest(&taskv1.UpdateTaskRequest{
								Id:          taskID,
								Description: fmt.Sprintf("Updated mixed ops task W%d", workerID),
								Completed:   true,
							}))
							if err != nil {
								atomic.AddInt64(&errors, 1)
							}
						}
					}
					
					atomic.AddInt64(&operations, 1)
				}
			}(i)
		}

		wg.Wait()
		close(taskIDs)
		
		actualDuration := time.Since(start)
		throughput := float64(operations) / actualDuration.Seconds()
		errorRate := float64(errors) / float64(operations) * 100

		t.Logf("Performed %d mixed operations in %v (%.2f ops/sec, %.2f%% error rate)", 
			operations, actualDuration, throughput, errorRate)

		// Cleanup created tasks
		go func() {
			for taskID := range taskIDs {
				suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
			}
		}()

		// Performance assertions
		assert.Greater(t, throughput, 100.0, "Mixed operations throughput is too low")
		assert.Less(t, errorRate, 1.0, "Error rate is too high")
	})
}

func TestLoad_DatabaseConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	suite := setupIntegrationTest(t)

	ctx := context.Background()

	t.Run("LoadTest_MultipleConnections", func(t *testing.T) {
		const numClients = 50
		const operationsPerClient = 20

		var wg sync.WaitGroup
		var successCount int64
		var errorCount int64
		results := make(chan time.Duration, numClients*operationsPerClient)

		// Create multiple clients to simulate multiple connections
		clients := make([]taskconnect.TaskServiceClient, numClients)
		for i := range clients {
			clients[i] = suite.Client // Reuse the same client for simplicity
		}

		start := time.Now()

		for i := 0; i < numClients; i++ {
			wg.Add(1)
			go func(clientID int, client taskconnect.TaskServiceClient) {
				defer wg.Done()
				
				for j := 0; j < operationsPerClient; j++ {
					opStart := time.Now()
					
					resp, err := client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
						Description: fmt.Sprintf("Load test task C%d-O%d", clientID, j),
					}))
					
					opDuration := time.Since(opStart)
					results <- opDuration
					
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
						// Clean up immediately
						go func(taskID string) {
							client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
						}(resp.Msg.Task.Id)
					}
				}
			}(i, clients[i])
		}

		wg.Wait()
		close(results)

		totalDuration := time.Since(start)
		totalOps := numClients * operationsPerClient
		throughput := float64(successCount) / totalDuration.Seconds()

		// Calculate response time statistics
		var durations []time.Duration
		for duration := range results {
			durations = append(durations, duration)
		}

		var totalLatency time.Duration
		var maxLatency time.Duration
		var minLatency time.Duration = time.Hour // Initialize to a large value

		for _, d := range durations {
			totalLatency += d
			if d > maxLatency {
				maxLatency = d
			}
			if d < minLatency {
				minLatency = d
			}
		}

		avgLatency := totalLatency / time.Duration(len(durations))

		t.Logf("Load test results:")
		t.Logf("  - Clients: %d", numClients)
		t.Logf("  - Operations per client: %d", operationsPerClient)
		t.Logf("  - Total operations: %d", totalOps)
		t.Logf("  - Successful operations: %d", successCount)
		t.Logf("  - Failed operations: %d", errorCount)
		t.Logf("  - Total duration: %v", totalDuration)
		t.Logf("  - Throughput: %.2f ops/sec", throughput)
		t.Logf("  - Average latency: %v", avgLatency)
		t.Logf("  - Min latency: %v", minLatency)
		t.Logf("  - Max latency: %v", maxLatency)

		// Assertions
		assert.Equal(t, int64(totalOps), successCount, "Not all operations succeeded")
		assert.Equal(t, int64(0), errorCount, "There should be no errors under normal load")
		assert.Greater(t, throughput, 50.0, "Throughput under load is too low")
		assert.Less(t, avgLatency, 500*time.Millisecond, "Average latency is too high")
		assert.Less(t, maxLatency, 2*time.Second, "Max latency is too high")
	})
}

func TestStress_DatabaseLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress tests in short mode")
	}

	suite := setupIntegrationTest(t)

	ctx := context.Background()

	t.Run("StressTest_LargeTaskData", func(t *testing.T) {
		// Test with very large task descriptions
		largeSizes := []int{1000, 10000, 50000} // 1KB, 10KB, 50KB
		
		for _, size := range largeSizes {
			t.Run(fmt.Sprintf("TaskSize_%dB", size), func(t *testing.T) {
				largeDesc := string(make([]byte, size))
				for i := range largeDesc {
					largeDesc = largeDesc[:i] + "A" + largeDesc[i+1:]
				}

				start := time.Now()
				resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
					Description: largeDesc,
				}))
				createDuration := time.Since(start)
				
				require.NoError(t, err, "Failed to create task with %d byte description", size)
				assert.Len(t, resp.Msg.Task.Description, size)

				taskID := resp.Msg.Task.Id

				// Test retrieval
				start = time.Now()
				getResp, err := suite.Client.GetTask(ctx, connect.NewRequest(&taskv1.GetTaskRequest{
					Id: taskID,
				}))
				getDuration := time.Since(start)

				require.NoError(t, err)
				assert.Len(t, getResp.Msg.Task.Description, size)

				t.Logf("Task with %d bytes: create=%v, get=%v", size, createDuration, getDuration)

				// Cleanup
				suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))

				// Performance assertions - large tasks should still be reasonably fast
				assert.Less(t, createDuration, 5*time.Second, "Large task creation is too slow")
				assert.Less(t, getDuration, 2*time.Second, "Large task retrieval is too slow")
			})
		}
	})

	t.Run("StressTest_TaskLimit", func(t *testing.T) {
		// Test system behavior with many tasks
		const maxTasks = 10000
		const batchSize = 100
		
		taskIDs := make([]string, 0, maxTasks)
		defer func() {
			// Cleanup all tasks
			t.Logf("Cleaning up %d tasks...", len(taskIDs))
			for _, taskID := range taskIDs {
				suite.Client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{Id: taskID}))
			}
		}()

		// Create tasks in batches
		start := time.Now()
		for i := 0; i < maxTasks; i += batchSize {
			batchStart := time.Now()
			
			for j := 0; j < batchSize && i+j < maxTasks; j++ {
				resp, err := suite.Client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
					Description: fmt.Sprintf("Stress test task %d", i+j),
				}))
				require.NoError(t, err)
				taskIDs = append(taskIDs, resp.Msg.Task.Id)
			}
			
			batchDuration := time.Since(batchStart)
			if i%1000 == 0 {
				t.Logf("Created %d tasks (batch %d took %v)", i+batchSize, i/batchSize+1, batchDuration)
			}
		}
		
		totalCreateTime := time.Since(start)
		createThroughput := float64(len(taskIDs)) / totalCreateTime.Seconds()

		t.Logf("Created %d tasks in %v (%.2f tasks/sec)", len(taskIDs), totalCreateTime, createThroughput)

		// Test listing performance with many tasks
		start = time.Now()
		listResp, err := suite.Client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
		listDuration := time.Since(start)
		
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listResp.Msg.Tasks), len(taskIDs))

		t.Logf("Listed %d tasks in %v", len(listResp.Msg.Tasks), listDuration)

		// Performance assertions
		assert.Greater(t, createThroughput, 50.0, "Task creation throughput degraded with many tasks")
		assert.Less(t, listDuration, 10*time.Second, "Task listing is too slow with many tasks")
	})
}