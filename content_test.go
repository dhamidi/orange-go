package main

import (
	"errors"
	"fmt"
	"testing"
	"time"
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

func Test_OnFrontpage_SubmissionsAreSortedByVotes(t *testing.T) {
	scenario := setup(t)
	for i := range 10 {
		scenario.must(scenario.postLink("https://news.ycombinator.com", fmt.Sprintf("%d", i)))
		scenario.upvoteN(scenario.PostIDs[i], 10-i)
	}

	submissions := scenario.frontpage()
	for i, s := range submissions {
		t.Logf("%d: %s", i, s.Title)
	}
	first := mustFind(submissions, "Title", "0")
	if first.ItemID != submissions[0].ItemID {
		t.Fatalf("expected first submission to be %q, got %q", "0", submissions[0].Title)
	}
}

func Test_OnFrontpage_SubmissionsAreSortedByScores_ThenSubmissionTime(t *testing.T) {
	scenario := setup(t)
	for i := range 10 {
		postLink := scenario.postLink("https://news.ycombinator.com", fmt.Sprintf("%d", i))
		postLink.SubmittedAt = postLink.SubmittedAt.Add(-time.Duration(10-i) * 24 * time.Hour)
		scenario.must(postLink)
	}

	// The oldest submission got 5 upvotes, but votes are multiplied by 0.9 every day.
	// Since ten days have passed, it will have a score of 5 * 0.9**10 = 1.74.
	//
	// The submission three days ago gets 3 upvotes, and will have a higher score.
	scenario.upvoteN(scenario.PostIDs[0], 5)
	scenario.upvoteN(scenario.PostIDs[10-3], 3)

	submissions := scenario.frontpage()
	for _, s := range submissions {
		t.Logf("votes: %d, score: %.2f: %s %s",
			s.VoteCount,
			s.Score,
			s.Title,
			s.SubmittedAt.Format(time.DateOnly),
		)
	}
	first := submissions[0]
	second := submissions[1]

	if f, s := first.Score, second.Score; f < s {
		t.Fatalf("Expected first score to be higher than second, got %.2f < %.2f", f, s)
	}

	for i := 2; i < 10; i += 2 {
		if a, b := submissions[i].SubmittedAt, submissions[i+1].SubmittedAt; b.After(a) {
			t.Fatalf("submission %d: Expected %s to be before %s",
				i,
				b.Format(time.DateOnly),
				a.Format(time.DateOnly))
		}
	}
}

func Test_OnFrontpage_SubmissionsCanBePaged_WithAfter(t *testing.T) {
	scenario := setup(t)
	for i := range 10 {
		postLink := scenario.postLink("https://news.ycombinator.com", fmt.Sprintf("%d", i))
		postLink.SubmittedAt = postLink.SubmittedAt.Add(-time.Duration(10-i) * 24 * time.Hour)
		scenario.must(postLink)
	}
	submissions := scenario.frontpageAfter(0)
	if act, exp := len(submissions), 10; act != exp {
		t.Fatalf("expected %d submissions, got %d", exp, act)
	}

	submissions = scenario.frontpageAfter(5)
	if act, exp := len(submissions), 5; act != exp {
		t.Fatalf("expected %d submissions, got %d", exp, act)
	}

	submissions = scenario.frontpageAfter(11)
	if act, exp := len(submissions), 0; act != exp {
		t.Fatalf("expected %d submissions, got %d", exp, act)
	}
}
