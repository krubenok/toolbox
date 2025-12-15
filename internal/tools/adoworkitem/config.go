package adoworkitem

import (
	"os"

	"github.com/krubenok/toolbox/internal/config"
)

const configFile = "ado-work-item.json"

// FieldMode specifies when a field should be included in output.
type FieldMode string

const (
	FieldModeAlways   FieldMode = "always"   // Always include the field
	FieldModeNotEmpty FieldMode = "notEmpty" // Only include if not empty/null (default unless overridden)
	FieldModeNever    FieldMode = "never"    // Never include the field
)

// Config holds all configuration for the ado-work-item tool.
type Config struct {
	Output *OutputConfig `json:"output,omitempty"`
}

// OutputConfig controls which fields are included in output.
// Field names match the map keys emitted in TOON mode.
type OutputConfig struct {
	// Work item fields
	ID          FieldMode `json:"id,omitempty"`
	URL         FieldMode `json:"url,omitempty"`
	UIURL       FieldMode `json:"uiUrl,omitempty"`
	Rev         FieldMode `json:"rev,omitempty"`
	Title       FieldMode `json:"title,omitempty"`
	Type        FieldMode `json:"type,omitempty"`
	State       FieldMode `json:"state,omitempty"`
	AssignedTo  FieldMode `json:"assignedTo,omitempty"`
	Description FieldMode `json:"description,omitempty"`

	// Sections
	Discussion  FieldMode `json:"discussion,omitempty"`
	Children    FieldMode `json:"children,omitempty"`
	Attachments FieldMode `json:"attachments,omitempty"`

	// Discussion fields
	CommentID       FieldMode `json:"commentId,omitempty"`
	CommentAuthor   FieldMode `json:"commentAuthor,omitempty"`
	CommentCreated  FieldMode `json:"commentCreated,omitempty"`
	CommentModified FieldMode `json:"commentModified,omitempty"`
	CommentText     FieldMode `json:"commentText,omitempty"`

	// Child fields
	ChildID    FieldMode `json:"childId,omitempty"`
	ChildURL   FieldMode `json:"childUrl,omitempty"`
	ChildUIURL FieldMode `json:"childUiUrl,omitempty"`

	// Attachment fields
	AttachmentName        FieldMode `json:"attachmentName,omitempty"`
	AttachmentURL         FieldMode `json:"attachmentUrl,omitempty"`
	AttachmentDownloadURL FieldMode `json:"attachmentDownloadUrl,omitempty"`
}

// DefaultOutputConfig returns the default output config.
// Central sections are always included to provide a stable shape for LLM workflows.
func DefaultOutputConfig() *OutputConfig {
	return &OutputConfig{
		ID:          FieldModeNotEmpty,
		URL:         FieldModeNotEmpty,
		UIURL:       FieldModeNotEmpty,
		Rev:         FieldModeNotEmpty,
		Title:       FieldModeNotEmpty,
		Type:        FieldModeNotEmpty,
		State:       FieldModeNotEmpty,
		AssignedTo:  FieldModeNotEmpty,
		Description: FieldModeAlways,

		Discussion:  FieldModeAlways,
		Children:    FieldModeAlways,
		Attachments: FieldModeAlways,

		CommentID:       FieldModeNotEmpty,
		CommentAuthor:   FieldModeNotEmpty,
		CommentCreated:  FieldModeNotEmpty,
		CommentModified: FieldModeNotEmpty,
		CommentText:     FieldModeNotEmpty,

		ChildID:    FieldModeNotEmpty,
		ChildURL:   FieldModeNotEmpty,
		ChildUIURL: FieldModeNotEmpty,

		AttachmentName:        FieldModeNotEmpty,
		AttachmentURL:         FieldModeNotEmpty,
		AttachmentDownloadURL: FieldModeNotEmpty,
	}
}

// GetFieldMode returns the mode for a field, defaulting to notEmpty if not set.
func (oc *OutputConfig) GetFieldMode(field string) FieldMode {
	if oc == nil {
		return FieldModeNotEmpty
	}

	var mode FieldMode
	switch field {
	case "id":
		mode = oc.ID
	case "url":
		mode = oc.URL
	case "uiUrl":
		mode = oc.UIURL
	case "rev":
		mode = oc.Rev
	case "title":
		mode = oc.Title
	case "type":
		mode = oc.Type
	case "state":
		mode = oc.State
	case "assignedTo":
		mode = oc.AssignedTo
	case "description":
		mode = oc.Description
	case "discussion":
		mode = oc.Discussion
	case "children":
		mode = oc.Children
	case "attachments":
		mode = oc.Attachments
	case "commentId":
		mode = oc.CommentID
	case "commentAuthor":
		mode = oc.CommentAuthor
	case "commentCreated":
		mode = oc.CommentCreated
	case "commentModified":
		mode = oc.CommentModified
	case "commentText":
		mode = oc.CommentText
	case "childId":
		mode = oc.ChildID
	case "childUrl":
		mode = oc.ChildURL
	case "childUiUrl":
		mode = oc.ChildUIURL
	case "attachmentName":
		mode = oc.AttachmentName
	case "attachmentUrl":
		mode = oc.AttachmentURL
	case "attachmentDownloadUrl":
		mode = oc.AttachmentDownloadURL
	}

	if mode == "" {
		return FieldModeNotEmpty
	}
	return mode
}

// LoadConfig loads the ado-work-item config from ~/.toolbox/ado-work-item.json.
// Falls back to defaults if file doesn't exist.
func LoadConfig() (*Config, error) {
	var cfg Config
	err := config.Load(configFile, &cfg)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Output: DefaultOutputConfig()}, nil
		}
		return nil, err
	}

	if cfg.Output == nil {
		cfg.Output = DefaultOutputConfig()
	}

	return &cfg, nil
}
