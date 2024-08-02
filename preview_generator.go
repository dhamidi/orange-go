package main

import (
	"log"
	"net/url"
	"time"

	"github.com/jimmysawczuk/recon"
)

type PreviewGenerator struct {
	Logger            *log.Logger
	App               *App
	Commands          CommandLog
	GeneratedPreviews map[string]time.Time
	Version           int
}

func NewPreviewGenerator(app *App, commands CommandLog, logger *log.Logger) *PreviewGenerator {
	return &PreviewGenerator{
		Logger:            logger,
		App:               app,
		Commands:          commands,
		GeneratedPreviews: make(map[string]time.Time),
		Version:           0,
	}
}

func (p *PreviewGenerator) HandleCommand(command Command) error {
	switch c := command.(type) {
	case *SetSubmissionPreview:
		return p.recordGeneratedPreview(c)
	default:
		return ErrCommandNotAccepted
	}
}

func (p *PreviewGenerator) recordGeneratedPreview(c *SetSubmissionPreview) error {
	p.GeneratedPreviews[c.ItemID] = time.Now()
	return nil
}

func (p *PreviewGenerator) shouldRegenerateFor(itemID string) bool {
	lastCheck, ok := p.GeneratedPreviews[itemID]
	if !ok {
		// we have never generated a preview for this item
		return true
	}

	// If the preview is older than a week, it should get regenerated
	return time.Now().Sub(lastCheck) > 24*time.Hour*7
}

func (p *PreviewGenerator) Start() func() {
	stop := make(chan struct{})

	commands, err := p.Commands.After(0)
	if err != nil {
		p.Logger.Printf("failed to fetch commands: %v", err)
		return func() {}
	}
	for command := range commands {
		p.HandleCommand(command.Message)
		p.Version = command.ID
	}
	p.Logger.Printf("PreviewGenerator started at version %d", p.Version)
	go p.loop(stop)
	return func() { close(stop) }
}

func (p *PreviewGenerator) loop(stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	tick := ticker.C
	for {
		select {
		case <-stop:
			return
		case <-tick:
			p.fetchPreviews()
		}
	}
}

func (p *PreviewGenerator) fetchPreviews() {
	commands, err := p.Commands.After(p.Version)
	if err != nil {
		p.Logger.Printf("failed to fetch commands: %v", err)
		return
	}
	for command := range commands {
		if postLink, ok := command.Message.(*PostLink); ok {
			if p.shouldRegenerateFor(postLink.ItemID) {
				p.fetchPreview(postLink.ItemID, postLink.Url)
			}
		}
		p.Version = command.ID
	}
}

func (p *PreviewGenerator) fetchPreview(itemID string, submissionURL string) {
	_, err := url.Parse(submissionURL)
	if err != nil {
		p.Logger.Printf("Invalid url %q: %s", submissionURL, err)
		return
	}

	result, err := recon.Parse(submissionURL)
	if err != nil {
		p.Logger.Printf("fetchPreview(%q): %s", submissionURL, err)
		return
	}

	var imageURL *string = nil
	if len(result.Images) > 0 {
		imageURL = &result.Images[0].URL
	}
	setPreview := &SetSubmissionPreview{
		ItemID:         itemID,
		ExtractedTitle: result.Title,
		ImageURL:       imageURL,
	}
	if err := p.App.HandleCommand(setPreview); err != nil {
		p.Logger.Printf("fetchPreview(%q): %s", submissionURL, err)
	}
}
