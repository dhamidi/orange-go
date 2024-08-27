package main

import (
	"testing"
	"time"
)

func Test_PutComment_InsertsCommentAndBumpsCount(t *testing.T) {
	state := NewInMemoryContentState()
	submission := &Submission{
		ItemID:      "item",
		Submitter:   "alice",
		SubmittedAt: time.Now(),
	}
	comment := &Comment{
		ParentID: NewTreeID("item"),
		Author:   "alice",
		Content:  "content",
		PostedAt: time.Now(),
	}
	if err := state.PutSubmission(submission); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := state.PutComment(comment); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	nestedComment := &Comment{
		ParentID: comment.ID(),
	}

	if err := state.PutComment(nestedComment); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(submission.Comments) != 1 {
		t.Errorf("expected 1 child, got %d", len(submission.Comments))
	}

	if len(submission.Comments[0].Children) != 1 {
		t.Errorf("expected 1 nested child, got %d", len(submission.Comments[0].Children))
	}

	if submission.CommentCount != 2 {
		t.Errorf("expected 2 comments, got %d", submission.CommentCount)
	}
}

func Test_GetSubmissionForComment_ReturnsSubmissionForComment(t *testing.T) {
	submission := &Submission{
		ItemID:      "item",
		Submitter:   "alice",
		SubmittedAt: time.Now(),
	}
	comment := &Comment{
		ParentID: NewTreeID("item"),
		Index:    0,
		Author:   "alice",
		Content:  "content",
		PostedAt: time.Now(),
	}
	submission.Comments = append(submission.Comments, comment)

	c := submission.Comment(NewTreeID("item", "0"))
	if c == nil {
		t.Fatalf("expected comment, got nil")
	}
}
