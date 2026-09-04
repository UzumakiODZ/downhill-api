package graph

import (
	"context"
	"downhill-api/database"
	"downhill-api/graph/model"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*Resolver, func()) {
	t.Helper()

	// Use in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open sqlite memory database: %v", err)
	}

	// Auto migrate all relevant schemas
	err = db.AutoMigrate(
		&database.User{},
		&database.Company{},
		&database.Role{},
		&database.QuestionBank{},
		&database.Post{},
		&database.Comment{},
	)
	if err != nil {
		t.Fatalf("Failed to auto migrate models: %v", err)
	}

	oldDB := database.DB
	database.DB = db

	// Initialize dummy redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   15,
	})

	resolver := &Resolver{
		RedisClient: rdb,
	}

	cleanup := func() {
		database.DB = oldDB
		_ = rdb.Close()
	}

	return resolver, cleanup
}

func TestCommentQueriesAndMutations(t *testing.T) {
	resolver, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	mutResolver := resolver.Mutation()
	queryResolver := resolver.Query()

	// 1. Seed user, company, post
	user := database.User{
		Username: "john_doe",
		Email:    "john.mitblr2024@learner.manipal.edu",
		RegID:    "24MITBLR001",
		Password: "hashedpassword",
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to seed user: %v", err)
	}

	company := database.Company{
		CompanyName: "Acme Corp",
	}
	if err := database.DB.Create(&company).Error; err != nil {
		t.Fatalf("Failed to seed company: %v", err)
	}

	post := database.Post{
		Title:     "Interview Experience at Acme Corp",
		Content:   "Great interview process with 3 rounds.",
		UserID:    user.ID,
		CompanyID: company.ID,
		CreatedAt: time.Now(),
	}
	if err := database.DB.Create(&post).Error; err != nil {
		t.Fatalf("Failed to seed post: %v", err)
	}

	postIDStr := fmt.Sprintf("%d", post.ID)
	userIDStr := fmt.Sprintf("%d", user.ID)

	// Test 1: CreateComment - Success
	t.Run("CreateComment_Success", func(t *testing.T) {
		commentContent := "This is a very helpful post! Thanks for sharing."
		createInput := model.CreateCommentInput{
			Content: commentContent,
			PostID:  postIDStr,
			UserID:  userIDStr,
		}

		comment, err := mutResolver.CreateComment(ctx, createInput)
		if err != nil {
			t.Fatalf("CreateComment failed: %v", err)
		}

		if comment == nil {
			t.Fatal("Expected created comment to be non-nil")
		}

		if comment.Content == nil || *comment.Content != commentContent {
			t.Errorf("Expected comment content '%s', got '%v'", commentContent, comment.Content)
		}

		if comment.PostID != postIDStr {
			t.Errorf("Expected postID '%s', got '%s'", postIDStr, comment.PostID)
		}

		if comment.UserID != userIDStr {
			t.Errorf("Expected userID '%s', got '%s'", userIDStr, comment.UserID)
		}

		if comment.ID == "" {
			t.Errorf("Expected non-empty comment ID")
		}

		// Verify persisted in DB
		var dbComment database.Comment
		if err := database.DB.First(&dbComment, comment.ID).Error; err != nil {
			t.Errorf("Comment not found in database: %v", err)
		}
	})

	// Test 2: CreateComment - Invalid PostID / UserID
	t.Run("CreateComment_InvalidIDs", func(t *testing.T) {
		invalidInputPost := model.CreateCommentInput{
			Content: "Invalid post test",
			PostID:  "abc",
			UserID:  userIDStr,
		}
		_, err := mutResolver.CreateComment(ctx, invalidInputPost)
		if err == nil {
			t.Errorf("Expected error for invalid postID 'abc', got nil")
		}

		invalidInputUser := model.CreateCommentInput{
			Content: "Invalid user test",
			PostID:  postIDStr,
			UserID:  "xyz",
		}
		_, err = mutResolver.CreateComment(ctx, invalidInputUser)
		if err == nil {
			t.Errorf("Expected error for invalid userID 'xyz', got nil")
		}
	})

	// Test 3: GetCommentsByPost - Query with Offset
	t.Run("GetCommentsByPost_Success", func(t *testing.T) {
		// Add more comments
		for i := 2; i <= 5; i++ {
			c := database.Comment{
				Content:   fmt.Sprintf("Comment number %d", i),
				PostID:    post.ID,
				UserID:    user.ID,
				CreatedAt: time.Now(),
			}
			if err := database.DB.Create(&c).Error; err != nil {
				t.Fatalf("Failed to seed comment: %v", err)
			}
		}

		offset0 := int32(0)
		comments, err := queryResolver.GetCommentsByPost(ctx, postIDStr, &offset0)
		if err != nil {
			t.Fatalf("GetCommentsByPost failed: %v", err)
		}

		if len(comments) != 5 {
			t.Errorf("Expected 5 comments, got %d", len(comments))
		}

		// Test pagination offset
		offset3 := int32(3)
		paginatedComments, err := queryResolver.GetCommentsByPost(ctx, postIDStr, &offset3)
		if err != nil {
			t.Fatalf("GetCommentsByPost with offset 3 failed: %v", err)
		}

		if len(paginatedComments) != 2 {
			t.Errorf("Expected 2 comments with offset 3, got %d", len(paginatedComments))
		}

		// Test empty result for large offset
		offset100 := int32(100)
		emptyComments, err := queryResolver.GetCommentsByPost(ctx, postIDStr, &offset100)
		if err != nil {
			t.Fatalf("GetCommentsByPost with offset 100 failed: %v", err)
		}
		if len(emptyComments) != 0 {
			t.Errorf("Expected 0 comments with offset 100, got %d", len(emptyComments))
		}
	})

	// Test 4: DeleteComment - Success and Non-existent
	t.Run("DeleteComment", func(t *testing.T) {
		// Create a comment to delete
		newComment := database.Comment{
			Content:   "To be deleted",
			PostID:    post.ID,
			UserID:    user.ID,
			CreatedAt: time.Now(),
		}
		if err := database.DB.Create(&newComment).Error; err != nil {
			t.Fatalf("Failed to create comment for deletion: %v", err)
		}

		commentIDStr := fmt.Sprintf("%d", newComment.ID)

		// Delete existing comment
		deleted, err := mutResolver.DeleteComment(ctx, commentIDStr)
		if err != nil {
			t.Fatalf("DeleteComment failed: %v", err)
		}
		if !deleted {
			t.Errorf("Expected DeleteComment to return true, got false")
		}

		// Verify comment is removed from DB
		var dbComment database.Comment
		err = database.DB.First(&dbComment, newComment.ID).Error
		if err == nil {
			t.Errorf("Expected comment to be deleted from database, but found it")
		}

		// Try deleting again (should fail because it's not found)
		deletedAgain, err := mutResolver.DeleteComment(ctx, commentIDStr)
		if err == nil {
			t.Errorf("Expected error when deleting non-existent comment, got nil")
		}
		if deletedAgain {
			t.Errorf("Expected DeleteComment on non-existent ID to return false")
		}
	})
}
