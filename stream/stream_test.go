package stream

import (
	"testing"
)

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{StartTagEvent, "StartTag"},
		{EndTagEvent, "EndTag"},
		{TextEvent, "Text"},
		{CommentEvent, "Comment"},
		{DoctypeEvent, "Doctype"},
		{EventType(100), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.eventType.String()
		if got != tt.expected {
			t.Errorf("EventType(%d).String() = %q, want %q", tt.eventType, got, tt.expected)
		}
	}
}

func TestStreamBasicHTML(t *testing.T) {
	html := "<html><head><title>Test</title></head><body><p>Hello</p></body></html>"

	var events []Event
	for event := range Stream(html) {
		events = append(events, event)
	}

	// Verify we got some events
	if len(events) == 0 {
		t.Fatal("expected events, got none")
	}

	// Check first event is html start tag
	if events[0].Type != StartTagEvent || events[0].Name != "html" {
		t.Errorf("first event = {Type: %v, Name: %q}, want StartTag 'html'",
			events[0].Type, events[0].Name)
	}
}

func TestStreamStartTag(t *testing.T) {
	html := `<div id="main" class="container">`

	var events []Event
	for event := range Stream(html) {
		events = append(events, event)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != StartTagEvent {
		t.Errorf("Type = %v, want StartTagEvent", event.Type)
	}
	if event.Name != "div" {
		t.Errorf("Name = %q, want %q", event.Name, "div")
	}
	if event.Attrs["id"] != "main" {
		t.Errorf("Attrs[id] = %q, want %q", event.Attrs["id"], "main")
	}
	if event.Attrs["class"] != "container" {
		t.Errorf("Attrs[class] = %q, want %q", event.Attrs["class"], "container")
	}
}

func TestStreamEndTag(t *testing.T) {
	html := "</div>"

	var events []Event
	for event := range Stream(html) {
		events = append(events, event)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != EndTagEvent {
		t.Errorf("Type = %v, want EndTagEvent", event.Type)
	}
	if event.Name != "div" {
		t.Errorf("Name = %q, want %q", event.Name, "div")
	}
}

func TestStreamText(t *testing.T) {
	html := "Hello, World!"

	var events []Event
	for event := range Stream(html) {
		events = append(events, event)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != TextEvent {
		t.Errorf("Type = %v, want TextEvent", event.Type)
	}
	if event.Data != "Hello, World!" {
		t.Errorf("Data = %q, want %q", event.Data, "Hello, World!")
	}
}

func TestStreamComment(t *testing.T) {
	html := "<!-- This is a comment -->"

	var events []Event
	for event := range Stream(html) {
		events = append(events, event)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != CommentEvent {
		t.Errorf("Type = %v, want CommentEvent", event.Type)
	}
	if event.Data != " This is a comment " {
		t.Errorf("Data = %q, want %q", event.Data, " This is a comment ")
	}
}

func TestStreamDoctype(t *testing.T) {
	html := "<!DOCTYPE html>"

	var events []Event
	for event := range Stream(html) {
		events = append(events, event)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != DoctypeEvent {
		t.Errorf("Type = %v, want DoctypeEvent", event.Type)
	}
	if event.Name != "html" {
		t.Errorf("Name = %q, want %q", event.Name, "html")
	}
}

func TestStreamDoctypeWithPublicSystemID(t *testing.T) {
	html := `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">`

	var events []Event
	for event := range Stream(html) {
		events = append(events, event)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != DoctypeEvent {
		t.Errorf("Type = %v, want DoctypeEvent", event.Type)
	}
	if event.Name != "html" {
		t.Errorf("Name = %q, want %q", event.Name, "html")
	}
	if event.PublicID != "-//W3C//DTD XHTML 1.0 Strict//EN" {
		t.Errorf("PublicID = %q, want %q", event.PublicID, "-//W3C//DTD XHTML 1.0 Strict//EN")
	}
	if event.SystemID != "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd" {
		t.Errorf("SystemID = %q, want %q", event.SystemID, "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd")
	}
}

func TestStreamCompleteDocument(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<!-- comment -->
<p class="intro">Hello</p>
</body>
</html>`

	eventTypes := make(map[EventType]int)
	for event := range Stream(html) {
		eventTypes[event.Type]++
	}

	// Should have at least one of each type
	if eventTypes[DoctypeEvent] != 1 {
		t.Errorf("DoctypeEvent count = %d, want 1", eventTypes[DoctypeEvent])
	}
	if eventTypes[StartTagEvent] < 5 {
		t.Errorf("StartTagEvent count = %d, want >= 5", eventTypes[StartTagEvent])
	}
	if eventTypes[EndTagEvent] < 5 {
		t.Errorf("EndTagEvent count = %d, want >= 5", eventTypes[EndTagEvent])
	}
	if eventTypes[TextEvent] < 1 {
		t.Errorf("TextEvent count = %d, want >= 1", eventTypes[TextEvent])
	}
	if eventTypes[CommentEvent] != 1 {
		t.Errorf("CommentEvent count = %d, want 1", eventTypes[CommentEvent])
	}
}

func TestStreamEmpty(t *testing.T) {
	var events []Event
	for event := range Stream("") {
		events = append(events, event)
	}

	if len(events) != 0 {
		t.Errorf("expected 0 events for empty input, got %d", len(events))
	}
}

func TestStreamSelfClosingTag(t *testing.T) {
	html := "<br/><hr /><img src='test.png'/>"

	var startTags []string
	for event := range Stream(html) {
		if event.Type == StartTagEvent {
			startTags = append(startTags, event.Name)
		}
	}

	expected := []string{"br", "hr", "img"}
	if len(startTags) != len(expected) {
		t.Fatalf("got %d start tags, want %d", len(startTags), len(expected))
	}
	for i, name := range expected {
		if startTags[i] != name {
			t.Errorf("startTags[%d] = %q, want %q", i, startTags[i], name)
		}
	}
}

func TestStreamBytes(t *testing.T) {
	html := []byte("<div>Hello</div>")

	var events []Event
	for event := range StreamBytes(html) {
		events = append(events, event)
	}

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].Type != StartTagEvent || events[0].Name != "div" {
		t.Errorf("events[0] = {%v, %q}, want {StartTagEvent, 'div'}", events[0].Type, events[0].Name)
	}
	if events[1].Type != TextEvent || events[1].Data != "Hello" {
		t.Errorf("events[1] = {%v, %q}, want {TextEvent, 'Hello'}", events[1].Type, events[1].Data)
	}
	if events[2].Type != EndTagEvent || events[2].Name != "div" {
		t.Errorf("events[2] = {%v, %q}, want {EndTagEvent, 'div'}", events[2].Type, events[2].Name)
	}
}

func TestStreamBytesWithBOM(t *testing.T) {
	// UTF-8 BOM followed by HTML
	html := []byte{0xEF, 0xBB, 0xBF}
	html = append(html, []byte("<p>Test</p>")...)

	var events []Event
	for event := range StreamBytes(html) {
		events = append(events, event)
	}

	// Should still parse correctly
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].Type != StartTagEvent || events[0].Name != "p" {
		t.Errorf("events[0] = {%v, %q}, want {StartTagEvent, 'p'}", events[0].Type, events[0].Name)
	}
}

func TestStreamWithEncodingOption(t *testing.T) {
	html := []byte("<p>Test</p>")

	var events []Event
	for event := range StreamBytes(html, WithEncoding("utf-8")) {
		events = append(events, event)
	}

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
}

func TestStreamWithOptions(t *testing.T) {
	html := "<div>Test</div>"

	var events []Event
	for event := range Stream(html, WithEncoding("utf-8")) {
		events = append(events, event)
	}

	// Options don't affect string input, but should not cause errors
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
}

func TestStreamNestedElements(t *testing.T) {
	html := "<div><span><a>link</a></span></div>"

	var events []Event
	for event := range Stream(html) {
		events = append(events, event)
	}

	// Should be: div start, span start, a start, text, a end, span end, div end
	expected := []struct {
		typ  EventType
		name string
		data string
	}{
		{StartTagEvent, "div", ""},
		{StartTagEvent, "span", ""},
		{StartTagEvent, "a", ""},
		{TextEvent, "", "link"},
		{EndTagEvent, "a", ""},
		{EndTagEvent, "span", ""},
		{EndTagEvent, "div", ""},
	}

	if len(events) != len(expected) {
		t.Fatalf("expected %d events, got %d", len(expected), len(events))
	}

	for i, exp := range expected {
		if events[i].Type != exp.typ {
			t.Errorf("events[%d].Type = %v, want %v", i, events[i].Type, exp.typ)
		}
		if exp.name != "" && events[i].Name != exp.name {
			t.Errorf("events[%d].Name = %q, want %q", i, events[i].Name, exp.name)
		}
		if exp.data != "" && events[i].Data != exp.data {
			t.Errorf("events[%d].Data = %q, want %q", i, events[i].Data, exp.data)
		}
	}
}

func TestStreamMultipleAttributes(t *testing.T) {
	html := `<input type="text" name="username" value="test" disabled>`

	var events []Event
	for event := range Stream(html) {
		events = append(events, event)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != StartTagEvent {
		t.Fatalf("expected StartTagEvent, got %v", event.Type)
	}

	expectedAttrs := map[string]string{
		"type":     "text",
		"name":     "username",
		"value":    "test",
		"disabled": "",
	}

	for key, expected := range expectedAttrs {
		if got := event.Attrs[key]; got != expected {
			t.Errorf("Attrs[%q] = %q, want %q", key, got, expected)
		}
	}
}

func TestStreamScript(t *testing.T) {
	html := "<script>var x = '<div>';</script>"

	var events []Event
	for event := range Stream(html) {
		events = append(events, event)
	}

	// Should have: script start, text (script content), script end
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].Type != StartTagEvent || events[0].Name != "script" {
		t.Errorf("events[0] = {%v, %q}, want {StartTagEvent, 'script'}", events[0].Type, events[0].Name)
	}
	if events[1].Type != TextEvent || events[1].Data != "var x = '<div>';" {
		t.Errorf("events[1] = {%v, %q}, want {TextEvent, \"var x = '<div>';\"}", events[1].Type, events[1].Data)
	}
	if events[2].Type != EndTagEvent || events[2].Name != "script" {
		t.Errorf("events[2] = {%v, %q}, want {EndTagEvent, 'script'}", events[2].Type, events[2].Name)
	}
}

func TestStreamStyle(t *testing.T) {
	html := "<style>.class { color: red; }</style>"

	var events []Event
	for event := range Stream(html) {
		events = append(events, event)
	}

	// Should have: style start, text (style content), style end
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].Type != StartTagEvent || events[0].Name != "style" {
		t.Errorf("events[0] = {%v, %q}, want {StartTagEvent, 'style'}", events[0].Type, events[0].Name)
	}
	if events[1].Type != TextEvent || events[1].Data != ".class { color: red; }" {
		t.Errorf("events[1] = {%v, %q}, want {TextEvent, '.class { color: red; }'}", events[1].Type, events[1].Data)
	}
	if events[2].Type != EndTagEvent || events[2].Name != "style" {
		t.Errorf("events[2] = {%v, %q}, want {EndTagEvent, 'style'}", events[2].Type, events[2].Name)
	}
}

func TestPtrToString(t *testing.T) {
	// Test nil case
	if got := ptrToString(nil); got != "" {
		t.Errorf("ptrToString(nil) = %q, want empty string", got)
	}

	// Test non-nil case
	s := "test"
	if got := ptrToString(&s); got != "test" {
		t.Errorf("ptrToString(&\"test\") = %q, want \"test\"", got)
	}
}

func BenchmarkStream(b *testing.B) {
	html := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<div id="main">
<p class="intro">Hello, World!</p>
<ul>
<li>Item 1</li>
<li>Item 2</li>
<li>Item 3</li>
</ul>
</div>
</body>
</html>`

	b.ResetTimer()
	for range b.N {
		for range Stream(html) {
			// Consume all events
		}
	}
}

func BenchmarkStreamBytes(b *testing.B) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<div id="main">
<p class="intro">Hello, World!</p>
<ul>
<li>Item 1</li>
<li>Item 2</li>
<li>Item 3</li>
</ul>
</div>
</body>
</html>`)

	b.ResetTimer()
	for range b.N {
		for range StreamBytes(html) {
			// Consume all events
		}
	}
}
