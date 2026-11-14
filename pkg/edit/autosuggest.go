package edit

import (
	"os"
	"strings"
	"unicode"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vars"
)

func defaultHistoryProvider(hs *histStore, code string) string {
	if hs == nil {
		return ""
	}
	if code == "" {
		return ""
	}

	// Get commands from history that start with the current input
	// TODO: Optimize this to avoid fetching all commands
	cmds, err := hs.AllCmds()
	if err != nil {
		return ""
	}

	// Search backwards through history for the most recent match. The full
	// command is returned: the code area computes the part to display, so the
	// provider never has to reason about where the buffer ends.
	for i := len(cmds) - 1; i >= 0; i-- {
		text := trimSpaceRight(cmds[i].Text)
		// Also handle multiline commands - only match against the first line
		if len(text) > len(code) && strings.HasPrefix(text, code) {
			return text
		}
	}

	return ""
}

func trimSpaceRight(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

func initAutoSuggestionSpec(appSpec *cli.AppSpec, hs *histStore, enabled vars.PtrVar, provider vars.PtrVar, ev *eval.Evaler) {
	appSpec.AutoSuggestionProvider = func(code string) string {
		if !enabled.GetRaw().(bool) {
			return ""
		}

		// Check if user has set a custom provider function
		providerFn := provider.GetRaw()
		if fn, ok := providerFn.(eval.Callable); ok {
			// Call user-defined provider function
			var result string
			valuesCb := func(ch <-chan any) {
				for v := range ch {
					if s, ok := v.(string); ok && result == "" {
						result = s
						return
					}
				}
			}
			bytesCb := func(_ *os.File) {}

			port, done, err := eval.PipePort(valuesCb, bytesCb)
			if err != nil {
				return ""
			}
			err = ev.Call(fn,
				eval.CallCfg{Args: []any{code}, From: "[auto-suggestion provider]"},
				eval.EvalCfg{Ports: []*eval.Port{nil, port, nil}})
			done()
			if err == nil && result != "" {
				return result
			}
			return ""
		}

		// Fall back to default history-based provider
		return defaultHistoryProvider(hs, code)
	}
}

func initAutoSuggestionAPI(app cli.App, enabled vars.PtrVar, provider vars.PtrVar, nb eval.NsBuilder) {
	acceptFn := func() {
		codeArea, ok := focusedCodeArea(app)
		if !ok {
			return
		}
		codeArea.MutateState(func(s *tk.CodeAreaState) {
			// Only accept if there's an autosuggestion pending
			if s.Pending.Content != "" && s.Pending.AutoSuggestion {
				s.ApplyPending()
			}
		})
	}

	nb.AddNs("auto-suggestion",
		eval.BuildNsNamed("edit:auto-suggestion").
			AddVar("enabled", enabled).
			AddVar("provider", provider).
			AddGoFn("accept", acceptFn))
}
