package main

type SetSubmissionPreview struct {
	ItemID         string
	ExtractedTitle string
	ImageURL       *string
	Metadata       *map[string]string
}

func (cmd *SetSubmissionPreview) CommandName() string { return "SetSubmissionPreview" }

func init() {
	DefaultCommandRegistry.Register("SetSubmissionPreview", func() Command { return &SetSubmissionPreview{} })
}

func (self *Content) handleSetSubmissionPreview(cmd *SetSubmissionPreview) error {
	preview := &SubmissionPreview{
		ItemID:   cmd.ItemID,
		Title:    &cmd.ExtractedTitle,
		ImageURL: cmd.ImageURL,
	}
	return self.state.PutSubmissionPreview(preview)
}
