package main

import (
	"testing"
	"time"
)

func TestCommentAddChild(t *testing.T) {
	itemID := "item"
	parent := &Comment{
		ParentID: NewTreeID(itemID),
		Author:   "alice",
		PostedAt: time.Now(),
	}
	if parent.ID().String() != itemID+"/0" {
		t.Errorf("expected ID %s, got %s", itemID+"/0", parent.ID().String())
	}

	child := &Comment{ParentID: parent.ID(), Author: "bob", PostedAt: time.Now()}
	if !child.Of(parent) {
		t.Errorf("expected child to be of parent")
	}

	parent.AddChild(child)

	if len(parent.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(parent.Children))
	}

	if child.ID().String() != parent.ID().And("0").String() {
		t.Errorf("expected ID %s, got %s", parent.ID().And("0").String(), child.ID().String())
	}

	secondChild := &Comment{
		ParentID: parent.ParentID,
		Author:   "alice",
		PostedAt: time.Now(),
	}
	parent.AddChild(secondChild)

	if secondChild.Index != 1 {
		t.Errorf("expected index 1, got %d", secondChild.Index)
	}
}

func TestCommentAddNestedChildren(t *testing.T) {
	itemID := "item"
	parent := &Comment{
		ParentID: NewTreeID(itemID),
		Author:   "alice",
		PostedAt: time.Now(),
	}
	child := &Comment{ParentID: parent.ID(), Author: "bob", PostedAt: time.Now()}
	if !child.Of(parent) {
		t.Errorf("expected child to be of parent")
	}

	parent.AddChild(child)

	if len(parent.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(parent.Children))
	}

	secondChild := &Comment{
		ParentID: child.ID(),
		Author:   "alice",
		PostedAt: time.Now(),
	}
	child.AddChild(secondChild)

	expectedID := parent.ID().And("0").And("0").String()
	if nestedID := secondChild.ID().String(); nestedID != expectedID {
		t.Errorf("expected nested ID %s, got %s", expectedID, nestedID)
	}
}
