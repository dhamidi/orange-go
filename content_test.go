package main

import (
	"errors"
	"fmt"
	"testing"
)

func TestUpvoteFailsIfVoterIsMissing(t *testing.T) {
	scenario := setup(t)
	err := scenario.do(scenario.upvote("item-id", ""))
	if !errors.Is(err, ErrMissingVoter) {
		t.Fatalf("expected %#v, got %#v", ErrMissingVoter, err)
	}
}

func TestUpvoteFailsIfVoterAlreadyVoted(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.upvote("item-id", "voter"))
	err := scenario.do(scenario.upvote("item-id", "voter"))
	if !errors.Is(err, ErrAlreadyVoted) {
		t.Fatalf("expected %#v, got %#v", ErrAlreadyVoted, err)
	}

}

func TestFrontpageReturnsRecentSubmissionsFirst(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.postLink("https://news.ycombinator.com", "1"))
	scenario.must(scenario.postLink("https://err.ee", "2"))
	scenario.must(scenario.postLink("https://www.kathimerini.gr", "3"))

	submissions := scenario.frontpage()

	if submissions[0].Title != "3" {
		t.Fatalf("expected submission %q, got %#v", "3", submissions[0])
	}
}

func TestFrontPageReturns10Submissions(t *testing.T) {
	scenario := setup(t)
	for i := range 20 {
		scenario.must(scenario.postLink("https://news.ycombinator.com", fmt.Sprintf("%d", i)))
	}

	submissions := scenario.frontpage()

	if len(submissions) != 10 {
		t.Fatalf("expected %d entries for the frontpage, got %d", 10, len(submissions))
	}
}

func Test_OnFrontpage_CanVoteIsTrue_IfUserVotedAlready(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.postLink("https://err.ee", "Upvoted"))
	upvotedId := scenario.PostIDs[0]
	scenario.must(scenario.upvote(upvotedId, scenario.Viewer))

	scenario.must(scenario.postLink("https://err.ee", "Other"))

	submissions := scenario.frontpage()

	upvoted := mustFind(submissions, "Title", "Upvoted")
	other := mustFind(submissions, "Title", "Other")
	if upvoted.ViewerHasVoted == false {
		t.Fatalf("ViewerHasVoted is false for Upvoted")
	}

	if other.ViewerHasVoted == true {
		t.Fatalf("ViewerHasVoted is true for Other")
	}
}

func Test_OnFrontpage_VotingMultipleTimes(t *testing.T) {
	scenario := setup(t)
	scenario.must(scenario.postLink("https://err.ee", "Upvoted"))
	upvotedId := scenario.PostIDs[0]
	scenario.must(scenario.upvote(upvotedId, scenario.Viewer))

	submissions := scenario.frontpage()

	upvoted := mustFind(submissions, "Title", "Upvoted")
	if upvoted.VoteCount != 1 {
		t.Fatalf("VoteCount is %d, expected %d", upvoted.VoteCount, 1)
	}
}
