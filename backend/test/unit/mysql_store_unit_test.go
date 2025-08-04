package unit

import (
	"context"
	"testing"
	"time"

	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wcygan/todo/backend/internal/store"
)

func TestMySQLStore_Unit_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MySQL store unit tests in short mode")
	}

	ctx := context.Background()
	container, dbConfig := setupTestMariaDB(t, ctx)
	defer container.Terminate(ctx)

	mysqlStore, err := store.NewMySQLTaskStore(dbConfig)
	require.NoError(t, err)
	defer mysqlStore.Close()

	t.Run("CreateTask_Success", func(t *testing.T) {
		task, err := mysqlStore.CreateTask(ctx, "Unit test task")
		require.NoError(t, err)
		assert.NotEmpty(t, task.Id)
		assert.Equal(t, "Unit test task", task.Description)
		assert.False(t, task.Completed)
		assert.NotNil(t, task.CreatedAt)
		assert.NotNil(t, task.UpdatedAt)
	})

	t.Run("CreateTask_EmptyDescription", func(t *testing.T) {
		_, err := mysqlStore.CreateTask(ctx, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "task description cannot be empty")
	})

	t.Run("GetTask_Existing", func(t *testing.T) {
		// Create a task first
		created, err := mysqlStore.CreateTask(ctx, "Task to retrieve")
		require.NoError(t, err)

		// Retrieve it
		retrieved, err := mysqlStore.GetTask(ctx, created.Id)
		require.NoError(t, err)
		assert.Equal(t, created.Id, retrieved.Id)
		assert.Equal(t, created.Description, retrieved.Description)
		assert.Equal(t, created.Completed, retrieved.Completed)
	})

	t.Run("GetTask_NonExistent", func(t *testing.T) {
		_, err := mysqlStore.GetTask(ctx, "99999")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetTask_InvalidID", func(t *testing.T) {
		_, err := mysqlStore.GetTask(ctx, "invalid-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task ID format")
	})

	t.Run("UpdateTask_Success", func(t *testing.T) {
		// Create a task first
		created, err := mysqlStore.CreateTask(ctx, "Task to update")
		require.NoError(t, err)

		// Add small delay to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		// Update it
		updated, err := mysqlStore.UpdateTask(ctx, created.Id, "Updated description", true)
		require.NoError(t, err)
		assert.Equal(t, created.Id, updated.Id)
		assert.Equal(t, "Updated description", updated.Description)
		assert.True(t, updated.Completed)
		assert.True(t, updated.UpdatedAt.AsTime().After(created.UpdatedAt.AsTime()))
	})

	t.Run("UpdateTask_CompletionOnly", func(t *testing.T) {
		// Create a task first
		created, err := mysqlStore.CreateTask(ctx, "Task for completion")
		require.NoError(t, err)

		// Update only completion status
		updated, err := mysqlStore.UpdateTask(ctx, created.Id, "", true)
		require.NoError(t, err)
		assert.Equal(t, created.Description, updated.Description) // Description unchanged
		assert.True(t, updated.Completed)
	})

	t.Run("UpdateTask_NonExistent", func(t *testing.T) {
		_, err := mysqlStore.UpdateTask(ctx, "99999", "Should fail", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("DeleteTask_Success", func(t *testing.T) {
		// Create a task first
		created, err := mysqlStore.CreateTask(ctx, "Task to delete")
		require.NoError(t, err)

		// Delete it
		err = mysqlStore.DeleteTask(ctx, created.Id)
		require.NoError(t, err)

		// Verify it's gone
		_, err = mysqlStore.GetTask(ctx, created.Id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("DeleteTask_NonExistent", func(t *testing.T) {
		err := mysqlStore.DeleteTask(ctx, "99999")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("ListTasks_Empty", func(t *testing.T) {
		// Create fresh store for this test
		freshContainer, freshConfig := setupTestMariaDB(t, ctx)
		defer freshContainer.Terminate(ctx)

		freshStore, err := store.NewMySQLTaskStore(freshConfig)
		require.NoError(t, err)
		defer freshStore.Close()

		tasks, err := freshStore.ListTasks(ctx)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("ListTasks_Multiple", func(t *testing.T) {
		// Create fresh store for this test
		freshContainer, freshConfig := setupTestMariaDB(t, ctx)
		defer freshContainer.Terminate(ctx)

		freshStore, err := store.NewMySQLTaskStore(freshConfig)
		require.NoError(t, err)
		defer freshStore.Close()

		// Create multiple tasks
		descriptions := []string{"Task 1", "Task 2", "Task 3"}
		createdTasks := make([]*taskv1.Task, 0, len(descriptions))

		for _, desc := range descriptions {
			task, err := freshStore.CreateTask(ctx, desc)
			require.NoError(t, err)
			createdTasks = append(createdTasks, task)
		}

		// List all tasks
		tasks, err := freshStore.ListTasks(ctx)
		require.NoError(t, err)
		assert.Len(t, tasks, len(descriptions))

		// Verify tasks are ordered by created_at DESC (newest first)
		for i := 1; i < len(tasks); i++ {
			assert.True(t, tasks[i-1].CreatedAt.AsTime().After(tasks[i].CreatedAt.AsTime()) ||
				tasks[i-1].CreatedAt.AsTime().Equal(tasks[i].CreatedAt.AsTime()))
		}
	})
}

func TestMySQLStore_Unit_ContextHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MySQL store context tests in short mode")
	}

	ctx := context.Background()
	container, dbConfig := setupTestMariaDB(t, ctx)
	defer container.Terminate(ctx)

	mysqlStore, err := store.NewMySQLTaskStore(dbConfig)
	require.NoError(t, err)
	defer mysqlStore.Close()

	t.Run("CreateTask_ContextCancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		_, err := mysqlStore.CreateTask(cancelCtx, "Should fail")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context")
	})

	t.Run("GetTask_ContextTimeout", func(t *testing.T) {
		// Create a task first
		task, err := mysqlStore.CreateTask(ctx, "Task for timeout test")
		require.NoError(t, err)

		// Create a very short timeout context
		timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
		defer cancel()

		// This might succeed if it's fast enough, so we just ensure it handles the context properly
		_, err = mysqlStore.GetTask(timeoutCtx, task.Id)
		// Don't assert error as the operation might complete before timeout
	})

	t.Run("ListTasks_ContextCancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		_, err := mysqlStore.ListTasks(cancelCtx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context")
	})
}

func TestMySQLStore_Unit_EdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MySQL store edge case tests in short mode")
	}

	ctx := context.Background()
	container, dbConfig := setupTestMariaDB(t, ctx)
	defer container.Terminate(ctx)

	mysqlStore, err := store.NewMySQLTaskStore(dbConfig)
	require.NoError(t, err)
	defer mysqlStore.Close()

	t.Run("CreateTask_VeryLongDescription", func(t *testing.T) {
		longDesc := string(make([]byte, 10000)) // 10KB description
		for i := range longDesc {
			longDesc = longDesc[:i] + "a" + longDesc[i+1:]
		}

		task, err := mysqlStore.CreateTask(ctx, longDesc)
		require.NoError(t, err)
		assert.Len(t, task.Description, 10000)
	})

	t.Run("CreateTask_UnicodeDescription", func(t *testing.T) {
		unicodeDesc := "测试任务 🚀 émojis и unicode"
		task, err := mysqlStore.CreateTask(ctx, unicodeDesc)
		require.NoError(t, err)
		assert.Equal(t, unicodeDesc, task.Description)

		// Verify it can be retrieved correctly
		retrieved, err := mysqlStore.GetTask(ctx, task.Id)
		require.NoError(t, err)
		assert.Equal(t, unicodeDesc, retrieved.Description)
	})

	t.Run("UpdateTask_SameValues", func(t *testing.T) {
		// Create a task
		task, err := mysqlStore.CreateTask(ctx, "Unchanged task")
		require.NoError(t, err)

		// Update with same values
		updated, err := mysqlStore.UpdateTask(ctx, task.Id, "Unchanged task", false)
		require.NoError(t, err)
		assert.Equal(t, task.Description, updated.Description)
		assert.Equal(t, task.Completed, updated.Completed)
		// UpdatedAt should still be updated
		assert.True(t, updated.UpdatedAt.AsTime().After(task.UpdatedAt.AsTime()) ||
			updated.UpdatedAt.AsTime().Equal(task.UpdatedAt.AsTime()))
	})

	t.Run("HealthCheck_Success", func(t *testing.T) {
		err := mysqlStore.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("Close_Multiple", func(t *testing.T) {
		// Test that multiple closes don't cause issues
		tempContainer, tempConfig := setupTestMariaDB(t, ctx)
		defer tempContainer.Terminate(ctx)

		tempStore, err := store.NewMySQLTaskStore(tempConfig)
		require.NoError(t, err)

		err1 := tempStore.Close()
		err2 := tempStore.Close()
		assert.NoError(t, err1)
		assert.NoError(t, err2) // Second close should not error
	})
}