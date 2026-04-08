package main

// ai.go - screenshot, Groq API call via curl, markdown rendering

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
)

const (
	groqModel     = "meta-llama/llama-4-scout-17b-16e-instruct"
	groqTextModel = "llama-3.3-70b-versatile"
)

func takeScreenshot(crop bool) (string, error) {
	path := fmt.Sprintf("/tmp/ai_screenshot_%d.png", time.Now().UnixMilli())

	if crop {
		// Use slop to let user select a region, then scrot crops to it
		out, err := exec.Command("slop", "-f", "%x %y %w %h").Output()
		if err != nil {
			return "", fmt.Errorf("slop: %w", err)
		}
		var x, y, w, h int
		fmt.Sscanf(strings.TrimSpace(string(out)), "%d %d %d %d", &x, &y, &w, &h)
		if w == 0 || h == 0 {
			return "", fmt.Errorf("selection cancelled")
		}
		cmd := exec.Command("scrot", "-a", fmt.Sprintf("%d,%d,%d,%d", x, y, w, h), path)
		if err := cmd.Run(); err != nil {
			return "", err
		}
	} else {
		cmd := exec.Command("scrot", "-u", path)
		if err := cmd.Run(); err != nil {
			cmd = exec.Command("scrot", path)
			if err2 := cmd.Run(); err2 != nil {
				return "", err2
			}
		}
	}

	data, err := os.ReadFile(path)
	os.Remove(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

const systemPrompt = `You are a helpful technical assistant. Be concise and avoid repetition.
- If the answer involves commands or config, include them in full — no truncation
- Keep explanations brief and relevant
- Skip filler phrases like "here are some steps" or restating the question
- Do not include example commands or placeholder commands — only real, complete, runnable commands
- Do not pad responses with obvious information or generic advice`

// chatHistory holds the conversation turns for multi-turn context.
// Each entry is a map ready to be serialised into the messages array.
var chatHistory []map[string]any

func clearHistory() {
	chatHistory = nil
}

func getAPIKey() string {
	if k := os.Getenv("GROQ_API_KEY"); k != "" {
		return k
	}
	envFile := os.Getenv("HOME") + "/.config/gymnott_ai.env"
	data, err := os.ReadFile(envFile)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "GROQ_API_KEY=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "GROQ_API_KEY="))
		}
	}
	return ""
}

func askAI(query string, withScreenshot, crop bool) string {
	apiKey := getAPIKey()
	if apiKey == "" {
		return "Error: GROQ_API_KEY environment variable not set."
	}

	var userMsg map[string]any
	var model string

	if withScreenshot {
		imgB64, err := takeScreenshot(crop)
		if err != nil {
			withScreenshot = false
		} else {
			model = groqModel
			userMsg = map[string]any{
				"role": "user",
				"content": []map[string]any{
					{"type": "text", "text": query},
					{"type": "image_url", "image_url": map[string]string{
						"url": "data:image/png;base64," + imgB64,
					}},
				},
			}
		}
	}
	if !withScreenshot {
		model = groqTextModel
		userMsg = map[string]any{"role": "user", "content": query}
	}

	chatHistory = append(chatHistory, userMsg)

	messages := append([]map[string]any{
		{"role": "system", "content": systemPrompt},
	}, chatHistory...)

	payload := map[string]any{
		"model":                 model,
		"messages":              messages,
		"max_completion_tokens": 1024,
		"temperature":           0.7,
		"stream":                false,
	}

	result := callGroq(payload, apiKey)

	// Append assistant reply to history (text only — vision content not kept)
	chatHistory = append(chatHistory, map[string]any{
		"role":    "assistant",
		"content": result,
	})

	return result
}

func callGroq(payload map[string]any, apiKey string) string {
	body, _ := json.Marshal(payload)
	tmpFile := fmt.Sprintf("/tmp/ai_req_%d.json", time.Now().UnixMilli())
	os.WriteFile(tmpFile, body, 0600)
	defer os.Remove(tmpFile)
	out, err := exec.Command("curl", "-s", "-X", "POST",
		"https://api.groq.com/openai/v1/chat/completions",
		"-H", "Content-Type: application/json",
		"-H", "Authorization: Bearer "+apiKey,
		"-d", "@"+tmpFile,
	).Output()
	if err != nil {
		return "curl error: " + err.Error()
	}
	return parseGroqResponse(out)
}

func parseGroqResponse(data []byte) string {
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "Parse error: " + err.Error() + "\nRaw: " + string(data)
	}
	if resp.Error.Message != "" {
		return "API Error: " + resp.Error.Message
	}
	if len(resp.Choices) == 0 {
		return "No response from API."
	}
	return resp.Choices[0].Message.Content
}

// ── Markdown renderer ────────────────────────────────────────────────────────

type mdTags struct {
	h1, h2, h3      *gtk.TextTag
	bold, italic    *gtk.TextTag
	code, codeBlock *gtk.TextTag
	copyHint        *gtk.TextTag
}

func makeTags(buf *gtk.TextBuffer) mdTags {
	tt, _ := buf.GetTagTable()
	mk := func(name string, props map[string]any) *gtk.TextTag {
		t, _ := gtk.TextTagNew(name)
		for k, v := range props {
			t.SetProperty(k, v)
		}
		tt.Add(t)
		return t
	}
	return mdTags{
		h1:   mk("h1", map[string]any{"weight": pango.WEIGHT_BOLD, "scale": pango.SCALE_X_LARGE}),
		h2:   mk("h2", map[string]any{"weight": pango.WEIGHT_BOLD, "scale": pango.SCALE_LARGE}),
		h3:   mk("h3", map[string]any{"weight": pango.WEIGHT_BOLD, "scale": pango.SCALE_MEDIUM}),
		bold: mk("bold", map[string]any{"weight": pango.WEIGHT_BOLD}),
		italic: mk("italic", map[string]any{"style": pango.STYLE_ITALIC}),
		code: mk("code", map[string]any{
			"family": "monospace", "background": "#2d2d2d", "foreground": "#f8f8f2",
		}),
		codeBlock: mk("codeblock", map[string]any{
			"family": "monospace", "background": "#1e1e1e", "foreground": "#d4d4d4",
			"left-margin": 12, "right-margin": 12,
			"pixels-above-lines": 1, "pixels-below-lines": 1,
		}),
		copyHint: mk("copyhint", map[string]any{
			"foreground": "#666666", "style": pango.STYLE_ITALIC,
		}),
	}
}

// isFence returns true if the trimmed line starts with ``` or a single ` followed by a word (```bash / `bash)
func isFence(line string) bool {
	t := strings.TrimSpace(line)
	return strings.HasPrefix(t, "```") || (strings.HasPrefix(t, "`") && len(t) > 1 && t[1] != '`' && !strings.Contains(t[1:], " ") == false)
}

func renderMarkdown(buf *gtk.TextBuffer, t mdTags, responseView *gtk.TextView, text string) {
	buf.SetText("")
	iter := buf.GetEndIter()

	ins := func(s string) { buf.Insert(iter, s) }
	insTag := func(s string, tag *gtk.TextTag) { buf.InsertWithTag(iter, s, tag) }

	lines := strings.Split(text, "\n")
	inCode := false
	codeLines := []string{}

	flushCode := func() {
		code := strings.Join(codeLines, "\n")
		codeLines = codeLines[:0]
		inCode = false

		// Insert code block text
		insTag(code+"\n", t.codeBlock)

		// Insert a real GTK copy button as a child anchor
		anchor, _ := buf.CreateChildAnchor(iter)
		btn, _ := gtk.ButtonNewWithLabel("  📋 Copy  ")
		btn.SetRelief(gtk.RELIEF_NONE)

		// Style the button
		btnStyle, _ := btn.GetStyleContext()
		btnStyle.AddClass("copy-btn")

		codeCopy := code // capture for closure
		btn.Connect("clicked", func() {
			clip, _ := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
			clip.SetText(codeCopy)
			btn.SetLabel("  ✓ Copied  ")
			// Reset label after 1.5s
			go func() {
				time.Sleep(1500 * time.Millisecond)
				scheduleOnMain(func() { btn.SetLabel("  📋 Copy  ") })
			}()
		})
		btn.ShowAll()
		responseView.AddChildAtAnchor(btn, anchor)
		ins("\n")
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// ── fenced code block (handles indented fences + `bash style) ──
		if strings.HasPrefix(trimmed, "```") || (strings.HasPrefix(trimmed, "`") && !strings.HasPrefix(trimmed, "``") && len(trimmed) > 1) {
			if inCode {
				flushCode()
			} else {
				inCode = true
			}
			continue
		}

		if inCode {
			codeLines = append(codeLines, line)
			continue
		}

		// ── headings ──
		if strings.HasPrefix(trimmed, "### ") {
			insTag(strings.TrimPrefix(trimmed, "### ")+"\n", t.h3)
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			insTag(strings.TrimPrefix(trimmed, "## ")+"\n", t.h2)
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			insTag(strings.TrimPrefix(trimmed, "# ")+"\n", t.h1)
			continue
		}

		// ── bullet list ──
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			ins("  • ")
			content := trimmed[2:]
			renderInlineSpans(buf, iter, content, t)
			ins("\n")
			continue
		}

		// ── numbered list ──
		if len(trimmed) > 2 && trimmed[0] >= '1' && trimmed[0] <= '9' && trimmed[1] == '.' {
			dot := strings.Index(trimmed, ".")
			ins("  " + trimmed[:dot+1] + " ")
			renderInlineSpans(buf, iter, strings.TrimSpace(trimmed[dot+1:]), t)
			ins("\n")
			continue
		}

		// ── blank line ──
		if trimmed == "" {
			ins("\n")
			continue
		}

		// ── normal paragraph line ──
		renderInlineSpans(buf, iter, trimmed, t)
		ins("\n")
	}

	// flush unclosed code block
	if inCode && len(codeLines) > 0 {
		flushCode()
	}
}

func renderInlineSpans(buf *gtk.TextBuffer, iter *gtk.TextIter, line string, t mdTags) {
	type span struct {
		marker string
		tag    *gtk.TextTag
	}
	// order matters: check ** before *
	spans := []span{{"**", t.bold}, {"*", t.italic}, {"`", t.code}}

	ins := func(s string) { buf.Insert(iter, s) }
	insTag := func(s string, tag *gtk.TextTag) { buf.InsertWithTag(iter, s, tag) }

	rest := line
	for rest != "" {
		bestIdx := len(rest) + 1
		best := ""
		var bestTag *gtk.TextTag
		for _, sp := range spans {
			idx := strings.Index(rest, sp.marker)
			if idx != -1 && idx < bestIdx {
				best, bestIdx, bestTag = sp.marker, idx, sp.tag
			}
		}
		if best == "" {
			ins(rest)
			break
		}
		ins(rest[:bestIdx])
		rest = rest[bestIdx+len(best):]
		end := strings.Index(rest, best)
		if end == -1 {
			ins(best + rest)
			break
		}
		insTag(rest[:end], bestTag)
		rest = rest[end+len(best):]
	}
}
