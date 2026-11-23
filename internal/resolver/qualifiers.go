package resolver

import "github.com/kryptamine/herdr-auto-title/internal/state"

// sessionIDWidth is how much of a session id a qualifier keeps: the first
// group of its UUID.
const sessionIDWidth = 8

// Qualifiers finds the tabs whose titles would read the same and says, by tab
// ID, what tells each one apart. The ladder it climbs is in
// docs/architecture/title-resolution.md.
func (d *Deterministic) Qualifiers(tabs []state.TabState) map[string]string {
	byName := make(map[string][]state.TabState, len(tabs))

	for _, tab := range tabs {
		name := d.Resolve(tab).Name
		byName[name] = append(byName[name], tab)
	}

	qualifiers := make(map[string]string)

	for _, group := range byName {
		if len(group) > 1 {
			d.qualify(group, qualifiers)
		}
	}

	return qualifiers
}

// qualify gives each tab of one group the first value that tells every tab in
// it apart, leaving the group alone when none does.
func (d *Deterministic) qualify(group []state.TabState, qualifiers map[string]string) {
	for _, valueOf := range []func(state.TabState) string{d.contextOf, sessionOf} {
		values := make([]string, len(group))
		for i, tab := range group {
			values[i] = valueOf(tab)
		}

		if !tellApart(values) {
			continue
		}

		for i, tab := range group {
			if values[i] != "" {
				qualifiers[tab.ID] = values[i]
			}
		}

		return
	}
}

// contextOf is where the chain places a tab before any row above, a hidden
// directory or a bound has taken it away.
func (d *Deterministic) contextOf(tab state.TabState) string {
	return d.collect(tab.Context).parts.Context
}

// sessionOf is the start of the session id the tab's agent holds, or "" when it
// reported none by id.
func sessionOf(tab state.TabState) string {
	if tab.Context == nil {
		return ""
	}

	id, ok := tab.Context.AgentSession.ID()
	if !ok {
		return ""
	}

	return id[:sessionIDWidth]
}

// tellApart reports that no two values are the same and one says something. A
// tab with nothing to add keeps its title, which its qualified twins now differ
// from.
func tellApart(values []string) bool {
	seen := make(map[string]bool, len(values))
	said := false

	for _, value := range values {
		if seen[value] {
			return false
		}

		seen[value] = true
		said = said || value != ""
	}

	return said
}
