package bibleit

import (
	"context"
	"errors"
	"strconv"
)

// SSHLiveOverlayDesign identifies the immutable design selected for a Live.
// YouTubeURL is absent until the authorized workspace event is actually live.
type SSHLiveOverlayDesign struct {
	DesignID       string
	DesignRevision int64
	Name           string
	YouTubeURL     *string
}

// SSHLiveOverlay retains the sequence needed for a conditional overlay write.
// A nil Overlay restores normal Live translation output; it does not stop/start playback.
type SSHLiveOverlay struct {
	LiveID   string
	Sequence int64
	Overlay  *SSHLiveOverlayDesign
	Fields   map[string]string
	Raw      CommandResponse
}

// GetLiveOverlayCommand reads the selected overlay without changing playback.
func GetLiveOverlayCommand(id string) (Command, error) { return command("LIVE", id, "OVERLAY", "GET") }

// SetLiveOverlayCommand selects a saved or built-in design at the given sequence.
func SetLiveOverlayCommand(id, designID string, sequence int64) (Command, error) {
	if len(designID) == 0 || len(designID) > 64 || sequence < 0 {
		return Command{}, errors.New("invalid Live overlay selection")
	}
	return command("LIVE", id, "OVERLAY", "SET", designID, strconv.FormatInt(sequence, 10))
}

// ClearLiveOverlayCommand restores normal translation output at the given sequence.
func ClearLiveOverlayCommand(id string, sequence int64) (Command, error) {
	if sequence < 0 {
		return Command{}, errors.New("invalid Live sequence")
	}
	return command("LIVE", id, "OVERLAY", "CLEAR", strconv.FormatInt(sequence, 10))
}
func decodeLiveOverlay(raw CommandResponse) (SSHLiveOverlay, error) {
	f, err := single(raw)
	if err != nil {
		return SSHLiveOverlay{}, err
	}
	result := SSHLiveOverlay{Fields: f, Raw: raw}
	if result.LiveID, err = required(f, "live_id"); err != nil {
		return SSHLiveOverlay{}, err
	}
	if result.Sequence, err = number(f, "sequence"); err != nil {
		return SSHLiveOverlay{}, err
	}
	active, err := boolean(f, "active")
	if err != nil {
		return SSHLiveOverlay{}, err
	}
	if active {
		selection := SSHLiveOverlayDesign{YouTubeURL: optionalString(f, "youtube_url")}
		if selection.DesignID, err = required(f, "design_id"); err != nil {
			return SSHLiveOverlay{}, err
		}
		if selection.DesignRevision, err = number(f, "design_revision"); err != nil {
			return SSHLiveOverlay{}, err
		}
		if selection.Name, err = present(f, "name"); err != nil {
			return SSHLiveOverlay{}, err
		}
		result.Overlay = &selection
	}
	return result, nil
}
func (c *SSHClient) liveOverlay(ctx context.Context, cmd Command) (SSHLiveOverlay, error) {
	raw, err := c.Execute(ctx, cmd)
	if err != nil {
		return SSHLiveOverlay{}, err
	}
	return decodeLiveOverlay(raw)
}

// GetLiveOverlay reads the overlay and current sequence over the caller-owned SSH connection.
func (c *SSHClient) GetLiveOverlay(ctx context.Context, id string) (SSHLiveOverlay, error) {
	cmd, err := GetLiveOverlayCommand(id)
	if err != nil {
		return SSHLiveOverlay{}, err
	}
	return c.liveOverlay(ctx, cmd)
}

// SetLiveOverlay selects a design conditionally. It never retries or starts a stopped Live.
func (c *SSHClient) SetLiveOverlay(ctx context.Context, id, designID string, sequence int64) (SSHLiveOverlay, error) {
	cmd, err := SetLiveOverlayCommand(id, designID, sequence)
	if err != nil {
		return SSHLiveOverlay{}, err
	}
	return c.liveOverlay(ctx, cmd)
}

// ClearLiveOverlay restores normal output conditionally, without changing the retained queue.
func (c *SSHClient) ClearLiveOverlay(ctx context.Context, id string, sequence int64) (SSHLiveOverlay, error) {
	cmd, err := ClearLiveOverlayCommand(id, sequence)
	if err != nil {
		return SSHLiveOverlay{}, err
	}
	return c.liveOverlay(ctx, cmd)
}
