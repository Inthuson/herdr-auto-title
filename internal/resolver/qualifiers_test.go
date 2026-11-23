package resolver

import (
	"maps"
	"slices"
	"testing"

	"github.com/kryptamine/herdr-auto-title/internal/herdr"
	"github.com/kryptamine/herdr-auto-title/internal/state"
)

// twin builds a tab in workspace whose one pane in dir titles its terminal like
// every other twin, under its own IDs so two of them can be told apart.
func twin(id, workspace, dir string) state.TabState {
	tab := state.TabFrom(herdr.TabInfo{TabID: id}, workspace, 1, []*state.PaneState{{
		ID:            id + ":p1",
		Dir:           dir,
		TerminalTitle: "Fix OAuth redirect",
		Focused:       true,
	}}, false)

	return tab
}

// session gives the tab's pane an agent session reported by id.
func session(tab state.TabState, id string) state.TabState {
	tab.Context.AgentSession = &herdr.AgentSessionInfo{
		Agent: "claude", Kind: herdr.SessionRefID, Value: id,
	}

	return tab
}

// qualified resolves each tab the way a poll does once Qualifiers has run.
func qualified(chain *Deterministic, tabs ...state.TabState) []string {
	qualifiers := chain.Qualifiers(tabs)

	names := make([]string, len(tabs))
	for i, tab := range tabs {
		tab.Qualifier = qualifiers[tab.ID]
		names[i] = chain.Resolve(tab).Name
	}

	return names
}

func TestQualifiersLeaveDistinctTitlesAlone(t *testing.T) {
	t.Parallel()

	got := defaultChain().Qualifiers([]state.TabState{
		twin("w1:t1", "", dashboard),
		twin("w2:t1", "", api),
	})
	if len(got) != 0 {
		t.Errorf("qualifiers = %v, want none", got)
	}
}

func TestTwinsInTheirOwnWorkspacesGetTheirDirectoryBack(t *testing.T) {
	t.Parallel()

	// Each tab dropped its directory for repeating its workspace, so two tabs
	// whose agents chose the same title read the same everywhere one is listed.
	got := qualified(defaultChain(),
		twin("w1:t1", "dashboard", dashboard),
		twin("w2:t1", "api", api),
	)

	want := []string{"dashboard › Fix OAuth redirect", "api › Fix OAuth redirect"}
	if !slices.Equal(got, want) {
		t.Errorf("names = %q, want %q", got, want)
	}
}

func TestOnlyTheTwinWithADirectoryIsQualified(t *testing.T) {
	t.Parallel()

	// A pane with no project directory, as in the home directory, has no
	// context to add, and once its twin says where it is the two differ.
	got := qualified(defaultChain(),
		twin("w1:t1", "dashboard", dashboard),
		twin("w2:t1", "finance", ""),
	)

	want := []string{"dashboard › Fix OAuth redirect", "Fix OAuth redirect"}
	if !slices.Equal(got, want) {
		t.Errorf("names = %q, want %q", got, want)
	}
}

func TestAHiddenDirectoryComesBackToTellTwinsApart(t *testing.T) {
	t.Parallel()

	got := qualified(directoryChain(0, true),
		twin("w1:t1", "", dashboard),
		twin("w1:t2", "", api),
	)

	want := []string{"dashboard › Fix OAuth redirect", "api › Fix OAuth redirect"}
	if !slices.Equal(got, want) {
		t.Errorf("names = %q, want %q", got, want)
	}
}

func TestTwinsInOneDirectoryAreToldApartByTheirSession(t *testing.T) {
	t.Parallel()

	got := qualified(defaultChain(),
		session(
			twin("w1:t1", "dashboard", dashboard),
			"3fa9c1d2-0000-4000-8000-000000000001",
		),
		session(
			twin("w2:t1", "dashboard", dashboard),
			"7b20e4aa-0000-4000-8000-000000000002",
		),
	)

	want := []string{"3fa9c1d2 › Fix OAuth redirect", "7b20e4aa › Fix OAuth redirect"}
	if !slices.Equal(got, want) {
		t.Errorf("names = %q, want %q", got, want)
	}
}

func TestTwinsWithNothingToTellThemApartKeepTheirTitle(t *testing.T) {
	t.Parallel()

	// Two shells in one directory: there is nothing true to add, and a made-up
	// difference would be worse than none.
	got := defaultChain().Qualifiers([]state.TabState{
		twin("w1:t1", "dashboard", dashboard),
		twin("w1:t2", "dashboard", dashboard),
	})
	if len(got) != 0 {
		t.Errorf("qualifiers = %v, want none", got)
	}
}

func TestAThirdTabSharingADirectoryFallsBackToSessions(t *testing.T) {
	t.Parallel()

	// A directory that two of three twins share tells only one of them apart,
	// so the session, which all three hold, is what qualifies.
	got := qualified(defaultChain(),
		session(
			twin("w1:t1", "dashboard", dashboard),
			"3fa9c1d2-0000-4000-8000-000000000001",
		),
		session(
			twin("w2:t1", "dashboard", dashboard),
			"7b20e4aa-0000-4000-8000-000000000002",
		),
		session(
			twin("w3:t1", "api", api),
			"c0ffee00-0000-4000-8000-000000000003",
		),
	)

	want := []string{
		"3fa9c1d2 › Fix OAuth redirect",
		"7b20e4aa › Fix OAuth redirect",
		"c0ffee00 › Fix OAuth redirect",
	}
	if !slices.Equal(got, want) {
		t.Errorf("names = %q, want %q", got, want)
	}
}

func TestAQualifiedTabIsNamedTheSameEveryTime(t *testing.T) {
	t.Parallel()

	tabs := []state.TabState{
		twin("w1:t1", "dashboard", dashboard),
		twin("w2:t1", "api", api),
	}

	first := defaultChain().Qualifiers(tabs)
	for range 20 {
		if again := defaultChain().Qualifiers(tabs); !maps.Equal(again, first) {
			t.Fatalf("qualifiers = %v, then %v", first, again)
		}
	}
}
