package app

import (
	"slices"
	"testing"

	"github.com/kryptamine/herdr-auto-title/internal/herdr"
	"github.com/kryptamine/herdr-auto-title/internal/herdr/herdrtest"
)

// twinSession is two workspaces each named after its directory, whose agents
// gave their tabs the same title, so each tab drops its directory.
func twinSession() *herdrtest.Client {
	client := herdrtest.New(
		[]herdr.TabInfo{
			{TabID: "w1:t1", WorkspaceID: "w1", Label: "1"},
			{TabID: "w2:t1", WorkspaceID: "w2", Label: "1"},
		},
		[]herdr.PaneInfo{
			{
				PaneID: "w1:p1", TabID: "w1:t1", Focused: true,
				CWD:                   dashboard,
				TerminalTitleStripped: "Fix OAuth redirect",
			},
			{
				PaneID: "w2:p1", TabID: "w2:t1", Focused: true,
				CWD:                   api,
				TerminalTitleStripped: "Fix OAuth redirect",
			},
		},
	)
	client.SetWorkspaces(
		herdr.WorkspaceInfo{WorkspaceID: "w1", Label: "dashboard"},
		herdr.WorkspaceInfo{WorkspaceID: "w2", Label: "api"},
	)

	return client
}

func TestTwinTitlesAreLeftAlikeByDefault(t *testing.T) {
	t.Parallel()

	// Off by default, a tab's title stays a matter of that tab alone.
	h := startConfigured(t, twinSession(), testConfig())
	h.poll()

	want := []string{"Fix OAuth redirect", "Fix OAuth redirect"}
	if got := tabLabels(h); !slices.Equal(got, want) {
		t.Errorf("renames = %q, want %q", got, want)
	}
}

func TestTwinTitlesSayWhereEachTabIs(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.DistinctTitles = true

	h := startConfigured(t, twinSession(), cfg)
	h.poll()

	want := []string{"dashboard › Fix OAuth redirect", "api › Fix OAuth redirect"}
	if got := tabLabels(h); !slices.Equal(got, want) {
		t.Errorf("renames = %q, want %q", got, want)
	}
}

func TestATwinThatGoesTakesTheQualifierWithIt(t *testing.T) {
	t.Parallel()

	// The qualifier is decided from each poll's tabs, so a tab left on its own
	// reads as it would have without one.
	cfg := testConfig()
	cfg.DistinctTitles = true

	h := startConfigured(t, twinSession(), cfg)
	h.poll()

	h.client.CloseTab("w2:t1")
	h.client.ClosePane("w2:p1")
	h.poll()

	renames := h.client.Renames()
	if got := renames[len(renames)-1].Label; got != "Fix OAuth redirect" {
		t.Errorf("rename = %q, want %q", got, "Fix OAuth redirect")
	}
}
