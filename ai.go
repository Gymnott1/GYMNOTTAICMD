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

// new
const (
	groqModel     = "meta-llama/llama-4-scout-17b-16e-instruct"
	groqTextModel = "llama-3.3-70b-versatile"
	geminiModel   = "gemini-flash-latest"
)

func takeScreenshotFile(crop bool) (string, error) {
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

	return path, nil
}

func takeScreenshot(crop bool) (string, error) {
	path, err := takeScreenshotFile(crop)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	os.Remove(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

const systemPrompt = `You are a senior technical assistant. Your output is pasted line by line into a terminal — each line is typed then Enter is pressed.
STRICT rules:
- Output ONLY raw commands and # comments — no markdown, no code fences, no backticks wrapping the output
- NEVER use interactive wizards or commands that open a sub-prompt (e.g. never use '/ip hotspot setup' — use '/ip hotspot add' with explicit parameters instead)
- # comments must be short labels only: e.g. # create pool, # add user. Never write sentences in comments
- Placeholders the user must change: write inline as <placeholder>
- Every command must be complete and runnable on its own line
- No numbering, no bullets, no blank prose lines
- Never truncate or skip steps — write every command in full
- Assume Linux terminal unless context says otherwise (MikroTik = RouterOS CLI)`

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

func getGeminiAPIKey() string {
	if k := os.Getenv("GEMINI_API_KEY"); k != "" {
		return k
	}
	envFile := os.Getenv("HOME") + "/.config/gymnott_ai.env"
	data, err := os.ReadFile(envFile)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "GEMINI_API_KEY=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "GEMINI_API_KEY="))
		}
	}
	return ""
}

func stripFences(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			continue
		}
		out = append(out, l)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func askAI(query string, withScreenshot, crop, textExtract bool) string {
	if withScreenshot && textExtract {
		return askGeminiWithExtractedText(query, crop)
	}

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
		"max_completion_tokens": 4096,
		"temperature":           0.7,
		"stream":                false,
	}

	result := stripFences(callGroq(payload, apiKey))

	// Append assistant reply to history (text only — vision content not kept)
	chatHistory = append(chatHistory, map[string]any{
		"role":    "assistant",
		"content": result,
	})

	return result
}

func askGeminiWithExtractedText(query string, crop bool) string {
	apiKey := getGeminiAPIKey()
	if apiKey == "" {
		return "Error: GEMINI_API_KEY environment variable not set."
	}

	imgPath, err := takeScreenshotFile(crop)
	if err != nil {
		return "Screenshot error: " + err.Error()
	}
	defer os.Remove(imgPath)

	extractedText, err := extractTextFromImage(imgPath)
	if err != nil {
		return "Text extraction error: " + err.Error() + "\nInstall OCR support with: sudo apt install tesseract-ocr"
	}
	if strings.TrimSpace(extractedText) == "" {
		extractedText = "(No text was detected in the captured image.)"
	}

	userPrompt := strings.TrimSpace(query)
	if userPrompt == "" {
		userPrompt = "Use the extracted screen text to help me."
	}
	prompt := systemPrompt +
		"\n\nUser request:\n" + userPrompt +
		"\n\nExtracted text from screenshot:\n" + extractedText

	payload := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]any{
					{"text": prompt},
				},
			},
		},
	}

	result := stripFences(callGemini(payload, apiKey))
	chatHistory = append(chatHistory,
		map[string]any{"role": "user", "content": prompt},
		map[string]any{"role": "assistant", "content": result},
	)
	return result
}

func extractTextFromImage(path string) (string, error) {
	out, err := exec.Command("tesseract", path, "stdout").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
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

func callGemini(payload map[string]any, apiKey string) string {
	body, _ := json.Marshal(payload)
	tmpFile := fmt.Sprintf("/tmp/ai_gemini_req_%d.json", time.Now().UnixMilli())
	os.WriteFile(tmpFile, body, 0600)
	defer os.Remove(tmpFile)
	out, err := exec.Command("curl", "-s",
		"https://generativelanguage.googleapis.com/v1beta/models/"+geminiModel+":generateContent",
		"-H", "Content-Type: application/json",
		"-H", "X-goog-api-key: "+apiKey,
		"-X", "POST",
		"-d", "@"+tmpFile,
	).Output()
	if err != nil {
		return "curl error: " + err.Error()
	}
	return parseGeminiResponse(out)
}

func parseGeminiResponse(data []byte) string {
	var resp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
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
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "No response from Gemini."
	}
	var parts []string
	for _, part := range resp.Candidates[0].Content.Parts {
		if strings.TrimSpace(part.Text) != "" {
			parts = append(parts, part.Text)
		}
	}
	if len(parts) == 0 {
		return "No text response from Gemini."
	}
	return strings.Join(parts, "\n")
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
		h1:     mk("h1", map[string]any{"weight": pango.WEIGHT_BOLD, "scale": pango.SCALE_X_LARGE}),
		h2:     mk("h2", map[string]any{"weight": pango.WEIGHT_BOLD, "scale": pango.SCALE_LARGE}),
		h3:     mk("h3", map[string]any{"weight": pango.WEIGHT_BOLD, "scale": pango.SCALE_MEDIUM}),
		bold:   mk("bold", map[string]any{"weight": pango.WEIGHT_BOLD}),
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

	insertCopyBtn := func(code string) {
		anchor, _ := buf.CreateChildAnchor(iter)
		btn, _ := gtk.ButtonNewWithLabel(" 📋 ")
		btn.SetRelief(gtk.RELIEF_NONE)
		codeCopy := code
		btn.Connect("clicked", func() {
			clip, _ := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
			clip.SetText(codeCopy)
			btn.SetLabel(" ✓ ")
			go func() {
				time.Sleep(1500 * time.Millisecond)
				scheduleOnMain(func() { btn.SetLabel(" 📋 ") })
			}()
		})
		btn.ShowAll()
		responseView.AddChildAtAnchor(btn, anchor)
	}

	flushCode := func() {
		code := strings.Join(codeLines, "\n")
		codeLines = codeLines[:0]
		inCode = false
		insTag(code+"\n", t.codeBlock)
		insertCopyBtn(code)
		ins("\n")
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// ── fenced code block: only ``` triggers a fence, not single backticks ──
		if strings.HasPrefix(trimmed, "```") {
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
			renderInlineSpans(buf, iter, trimmed[2:], t, responseView, insertCopyBtn)
			ins("\n")
			continue
		}

		// ── numbered list ──
		if len(trimmed) > 2 && trimmed[0] >= '1' && trimmed[0] <= '9' && trimmed[1] == '.' {
			dot := strings.Index(trimmed, ".")
			ins("  " + trimmed[:dot+1] + " ")
			renderInlineSpans(buf, iter, strings.TrimSpace(trimmed[dot+1:]), t, responseView, insertCopyBtn)
			ins("\n")
			continue
		}

		// ── blank line ──
		if trimmed == "" {
			ins("\n")
			continue
		}

		// ── normal paragraph line ──
		renderInlineSpans(buf, iter, trimmed, t, responseView, insertCopyBtn)
		ins("\n")
	}

	// flush unclosed code block
	if inCode && len(codeLines) > 0 {
		flushCode()
	}
}

func renderInlineSpans(buf *gtk.TextBuffer, iter *gtk.TextIter, line string, t mdTags, responseView *gtk.TextView, insertCopyBtn func(string)) {
	type span struct {
		marker string
		tag    *gtk.TextTag
	}
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
		// insert copy button after inline code spans (only if looks like a command)
		if best == "`" && (strings.Contains(rest[:end], " ") || len(rest[:end]) > 8) {
			insertCopyBtn(rest[:end])
		}
		rest = rest[end+len(best):]
	}
}
