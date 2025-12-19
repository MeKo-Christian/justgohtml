package dom

import (
	"testing"
)

const (
	testContainer = "container"
	testModified  = "modified"
)

// =============================================================================
// Node Interface Tests (baseNode)
// =============================================================================

func TestBaseNodeChildren(t *testing.T) {
	parent := NewElement("div")
	child1 := NewElement("span")
	child2 := NewElement("p")

	// Initially no children
	if len(parent.Children()) != 0 {
		t.Errorf("expected 0 children initially, got %d", len(parent.Children()))
	}

	parent.AppendChild(child1)
	parent.AppendChild(child2)

	children := parent.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if children[0] != child1 {
		t.Error("first child should be child1")
	}
	if children[1] != child2 {
		t.Error("second child should be child2")
	}
}

func TestBaseNodeHasChildNodes(t *testing.T) {
	parent := NewElement("div")

	if parent.HasChildNodes() {
		t.Error("expected HasChildNodes() to return false for empty element")
	}

	child := NewElement("span")
	parent.AppendChild(child)

	if !parent.HasChildNodes() {
		t.Error("expected HasChildNodes() to return true after adding child")
	}
}

func TestBaseNodeRemoveChild(t *testing.T) {
	parent := NewElement("div")
	child1 := NewElement("span")
	child2 := NewElement("p")
	child3 := NewElement("a")

	parent.AppendChild(child1)
	parent.AppendChild(child2)
	parent.AppendChild(child3)

	// Remove middle child
	parent.RemoveChild(child2)

	if len(parent.Children()) != 2 {
		t.Fatalf("expected 2 children after removal, got %d", len(parent.Children()))
	}
	if child2.Parent() != nil {
		t.Error("removed child's parent should be nil")
	}

	// Verify remaining children
	if parent.Children()[0] != child1 || parent.Children()[1] != child3 {
		t.Error("wrong children remaining after removal")
	}

	// Remove non-existent child should be no-op
	nonChild := NewElement("div")
	parent.RemoveChild(nonChild) // Should not panic
	if len(parent.Children()) != 2 {
		t.Error("removing non-child should not affect children count")
	}
}

func TestBaseNodeReplaceChild(t *testing.T) {
	parent := NewElement("div")
	child1 := NewElement("span")
	child2 := NewElement("p")
	newChild := NewElement("a")

	parent.AppendChild(child1)
	parent.AppendChild(child2)

	// Replace child2 with newChild
	replaced := parent.ReplaceChild(newChild, child2)

	if replaced != child2 {
		t.Error("ReplaceChild should return the replaced child")
	}
	if child2.Parent() != nil {
		t.Error("replaced child's parent should be nil")
	}
	if newChild.Parent() != parent {
		t.Error("new child's parent should be parent")
	}
	if parent.Children()[1] != newChild {
		t.Error("new child should be at old child's position")
	}

	// Replace non-existent child should return nil
	nonChild := NewElement("div")
	result := parent.ReplaceChild(NewElement("x"), nonChild)
	if result != nil {
		t.Error("replacing non-existent child should return nil")
	}
}

func TestBaseNodeInsertBeforeNilRef(t *testing.T) {
	parent := NewElement("div")
	child := NewElement("span")

	// InsertBefore with nil reference should append
	parent.InsertBefore(child, nil)

	if len(parent.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Children()))
	}
	if parent.Children()[0] != child {
		t.Error("child should be appended when refChild is nil")
	}
}

func TestBaseNodeInsertBeforeNotFound(t *testing.T) {
	parent := NewElement("div")
	child := NewElement("span")
	nonChild := NewElement("p")

	// InsertBefore with non-existent reference should append
	parent.InsertBefore(child, nonChild)

	if len(parent.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Children()))
	}
	if parent.Children()[0] != child {
		t.Error("child should be appended when refChild not found")
	}
}

// =============================================================================
// Element Tests
// =============================================================================

func TestNewElement(t *testing.T) {
	elem := NewElement("DIV") // Should be lowercased

	if elem.TagName != "div" {
		t.Errorf("expected TagName 'div', got '%s'", elem.TagName)
	}
	if elem.Namespace != NamespaceHTML {
		t.Errorf("expected namespace %s, got %s", NamespaceHTML, elem.Namespace)
	}
	if elem.Attributes == nil {
		t.Error("Attributes should not be nil")
	}
	if elem.Type() != ElementNodeType {
		t.Errorf("expected ElementNodeType, got %d", elem.Type())
	}
}

func TestNewElementNS(t *testing.T) {
	elem := NewElementNS("svg", NamespaceSVG)

	if elem.TagName != "svg" {
		t.Errorf("expected TagName 'svg', got '%s'", elem.TagName)
	}
	if elem.Namespace != NamespaceSVG {
		t.Errorf("expected namespace %s, got %s", NamespaceSVG, elem.Namespace)
	}
}

func TestElementClone(t *testing.T) {
	elem := NewElement("div")
	elem.SetAttr("class", "test")
	child := NewElement("span")
	child.SetAttr("id", "inner")
	elem.AppendChild(child)

	// Shallow clone
	shallowClone := elem.Clone(false).(*Element)
	if shallowClone.TagName != "div" {
		t.Error("shallow clone should have same tag name")
	}
	if shallowClone.Attr("class") != "test" {
		t.Error("shallow clone should have same attributes")
	}
	if len(shallowClone.Children()) != 0 {
		t.Error("shallow clone should have no children")
	}
	if shallowClone == elem {
		t.Error("clone should be a different object")
	}

	// Deep clone
	deepClone := elem.Clone(true).(*Element)
	if len(deepClone.Children()) != 1 {
		t.Fatal("deep clone should have children")
	}
	clonedChild := deepClone.Children()[0].(*Element)
	if clonedChild.Attr("id") != "inner" {
		t.Error("deep cloned child should have same attributes")
	}
	if clonedChild == child {
		t.Error("deep cloned child should be a different object")
	}
}

func TestElementCloneWithTemplateContent(t *testing.T) {
	template := NewElement("template")
	content := NewDocumentFragment()
	inner := NewElement("div")
	content.AppendChild(inner)
	template.TemplateContent = content

	// Deep clone should clone template content
	clone := template.Clone(true).(*Element)
	if clone.TemplateContent == nil {
		t.Fatal("cloned template should have TemplateContent")
	}
	if clone.TemplateContent == content {
		t.Error("cloned TemplateContent should be a different object")
	}
	if len(clone.TemplateContent.Children()) != 1 {
		t.Error("cloned TemplateContent should have children")
	}
}

func TestElementInsertBefore(t *testing.T) {
	parent := NewElement("div")
	child1 := NewElement("span")
	child2 := NewElement("p")
	child3 := NewElement("a")

	parent.AppendChild(child1)
	parent.AppendChild(child3)
	parent.InsertBefore(child2, child3)

	children := parent.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}
	if children[0] != child1 || children[1] != child2 || children[2] != child3 {
		t.Error("children not in expected order")
	}
	if child2.Parent() != parent {
		t.Error("inserted child's parent should be set")
	}
}

func TestElementRemoveChild(t *testing.T) {
	parent := NewElement("div")
	child := NewElement("span")
	parent.AppendChild(child)

	parent.RemoveChild(child)

	if len(parent.Children()) != 0 {
		t.Error("expected no children after removal")
	}
	if child.Parent() != nil {
		t.Error("removed child's parent should be nil")
	}
}

func TestElementReplaceChild(t *testing.T) {
	parent := NewElement("div")
	oldChild := NewElement("span")
	newChild := NewElement("p")
	parent.AppendChild(oldChild)

	result := parent.ReplaceChild(newChild, oldChild)

	if result != oldChild {
		t.Error("should return old child")
	}
	if parent.Children()[0] != newChild {
		t.Error("new child should replace old")
	}
}

func TestElementHasChildNodes(t *testing.T) {
	elem := NewElement("div")

	if elem.HasChildNodes() {
		t.Error("empty element should not have child nodes")
	}

	elem.AppendChild(NewElement("span"))

	if !elem.HasChildNodes() {
		t.Error("element with child should have child nodes")
	}
}

func TestElementText(t *testing.T) {
	div := NewElement("div")
	div.AppendChild(NewText("Hello "))
	span := NewElement("span")
	span.AppendChild(NewText("World"))
	div.AppendChild(span)

	text := div.Text()
	if text != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", text)
	}
}

func TestElementTextWithComments(t *testing.T) {
	div := NewElement("div")
	div.AppendChild(NewText("Hello"))
	div.AppendChild(NewComment("ignored"))
	div.AppendChild(NewText("World"))

	text := div.Text()
	if text != "HelloWorld" {
		t.Errorf("expected 'HelloWorld', got '%s'", text)
	}
}

func TestElementAttributes(t *testing.T) {
	elem := NewElement("div")

	// SetAttr and Attr
	elem.SetAttr("class", testContainer)
	if elem.Attr("class") != testContainer {
		t.Error("SetAttr/Attr failed")
	}

	// HasAttr
	if !elem.HasAttr("class") {
		t.Error("HasAttr should return true")
	}
	if elem.HasAttr("nonexistent") {
		t.Error("HasAttr should return false for nonexistent")
	}

	// RemoveAttr
	elem.RemoveAttr("class")
	if elem.HasAttr("class") {
		t.Error("RemoveAttr failed")
	}
}

func TestElementID(t *testing.T) {
	elem := NewElement("div")

	if elem.ID() != "" {
		t.Error("ID should be empty initially")
	}

	elem.SetAttr("id", "main")
	if elem.ID() != "main" {
		t.Errorf("expected ID 'main', got '%s'", elem.ID())
	}
}

func TestElementClasses(t *testing.T) {
	elem := NewElement("div")

	// No class attribute
	classes := elem.Classes()
	if classes != nil {
		t.Error("Classes should be nil when no class attribute")
	}

	// With class attribute
	elem.SetAttr("class", "foo bar baz")
	classes = elem.Classes()
	if len(classes) != 3 {
		t.Fatalf("expected 3 classes, got %d", len(classes))
	}
	if classes[0] != "foo" || classes[1] != "bar" || classes[2] != "baz" {
		t.Error("classes not parsed correctly")
	}
}

func TestElementHasClass(t *testing.T) {
	elem := NewElement("div")
	elem.SetAttr("class", "foo bar")

	if !elem.HasClass("foo") {
		t.Error("should have class 'foo'")
	}
	if !elem.HasClass("bar") {
		t.Error("should have class 'bar'")
	}
	if elem.HasClass("baz") {
		t.Error("should not have class 'baz'")
	}
}

func TestElementQuery(t *testing.T) {
	elem := NewElement("div")

	// Query is a stub that returns nil, nil
	results, err := elem.Query("span")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if results != nil {
		t.Error("stub should return nil results")
	}
}

func TestElementQueryFirst(t *testing.T) {
	elem := NewElement("div")

	// QueryFirst delegates to Query
	result, err := elem.QueryFirst("span")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("stub should return nil result")
	}
}

// =============================================================================
// Text Node Tests
// =============================================================================

func TestNewText(t *testing.T) {
	text := NewText("Hello World")

	if text.Data != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", text.Data)
	}
	if text.Type() != TextNodeType {
		t.Errorf("expected TextNodeType, got %d", text.Type())
	}
}

func TestTextParent(t *testing.T) {
	text := NewText("test")

	if text.Parent() != nil {
		t.Error("parent should be nil initially")
	}

	parent := NewElement("div")
	text.SetParent(parent)

	if text.Parent() != parent {
		t.Error("parent should be set")
	}
}

func TestTextLeafNode(t *testing.T) {
	text := NewText("test")

	// Text nodes have no children
	if text.Children() != nil {
		t.Error("Children should return nil")
	}
	if text.HasChildNodes() {
		t.Error("HasChildNodes should return false")
	}

	// Child operations are no-ops
	text.AppendChild(NewText("ignored"))
	text.InsertBefore(NewText("a"), NewText("b"))
	text.RemoveChild(NewText("c"))
	result := text.ReplaceChild(NewText("a"), NewText("b"))
	if result != nil {
		t.Error("ReplaceChild should return nil for leaf node")
	}
}

func TestTextClone(t *testing.T) {
	text := NewText("Hello")
	clone := text.Clone(true).(*Text)

	if clone.Data != "Hello" {
		t.Error("clone should have same data")
	}
	if clone == text {
		t.Error("clone should be a different object")
	}
}

// =============================================================================
// Comment Node Tests
// =============================================================================

func TestNewComment(t *testing.T) {
	comment := NewComment("This is a comment")

	if comment.Data != "This is a comment" {
		t.Errorf("expected 'This is a comment', got '%s'", comment.Data)
	}
	if comment.Type() != CommentNodeType {
		t.Errorf("expected CommentNodeType, got %d", comment.Type())
	}
}

func TestCommentParent(t *testing.T) {
	comment := NewComment("test")

	if comment.Parent() != nil {
		t.Error("parent should be nil initially")
	}

	parent := NewElement("div")
	comment.SetParent(parent)

	if comment.Parent() != parent {
		t.Error("parent should be set")
	}
}

func TestCommentLeafNode(t *testing.T) {
	comment := NewComment("test")

	// Comment nodes have no children
	if comment.Children() != nil {
		t.Error("Children should return nil")
	}
	if comment.HasChildNodes() {
		t.Error("HasChildNodes should return false")
	}

	// Child operations are no-ops
	comment.AppendChild(NewText("ignored"))
	comment.InsertBefore(NewText("a"), NewText("b"))
	comment.RemoveChild(NewText("c"))
	result := comment.ReplaceChild(NewText("a"), NewText("b"))
	if result != nil {
		t.Error("ReplaceChild should return nil for leaf node")
	}
}

func TestCommentClone(t *testing.T) {
	comment := NewComment("Test")
	clone := comment.Clone(true).(*Comment)

	if clone.Data != "Test" {
		t.Error("clone should have same data")
	}
	if clone == comment {
		t.Error("clone should be a different object")
	}
}

// =============================================================================
// Document Tests
// =============================================================================

func TestNewDocument(t *testing.T) {
	doc := NewDocument()

	if doc.Type() != DocumentNodeType {
		t.Errorf("expected DocumentNodeType, got %d", doc.Type())
	}
	if doc.QuirksMode != NoQuirks {
		t.Error("default quirks mode should be NoQuirks")
	}
}

func TestDocumentAppendChild(t *testing.T) {
	doc := NewDocument()
	html := NewElement("html")

	doc.AppendChild(html)

	if len(doc.Children()) != 1 {
		t.Error("document should have 1 child")
	}
	if html.Parent() != doc {
		t.Error("child's parent should be document")
	}
}

func TestDocumentInsertBeforeSetsParent(t *testing.T) {
	doc := NewDocument()
	html := NewElement("html")
	head := NewElement("head")
	body := NewElement("body")

	doc.AppendChild(html)
	html.AppendChild(body)
	html.InsertBefore(head, body)

	if head.Parent() != html {
		t.Fatalf("head.Parent() = %T, want html element", head.Parent())
	}
	if body.Parent() != html {
		t.Fatalf("body.Parent() = %T, want html element", body.Parent())
	}
	if doc.Parent() != nil {
		t.Fatalf("doc.Parent() = %T, want nil", doc.Parent())
	}
}

func TestDocumentElement(t *testing.T) {
	doc := NewDocument()

	// No document element initially
	if doc.DocumentElement() != nil {
		t.Error("should return nil when no elements")
	}

	// Add non-element child first (e.g., doctype placeholder)
	comment := NewComment("test")
	doc.AppendChild(comment)

	if doc.DocumentElement() != nil {
		t.Error("should return nil when only non-element children")
	}

	// Add HTML element
	html := NewElement("html")
	doc.AppendChild(html)

	if doc.DocumentElement() != html {
		t.Error("should return html element")
	}
}

func TestDocumentHead(t *testing.T) {
	doc := NewDocument()

	// No head when no document element
	if doc.Head() != nil {
		t.Error("should return nil when no document element")
	}

	html := NewElement("html")
	doc.AppendChild(html)

	// No head when html has no head child
	if doc.Head() != nil {
		t.Error("should return nil when no head element")
	}

	// Add non-head element first
	body := NewElement("body")
	html.AppendChild(body)

	if doc.Head() != nil {
		t.Error("should return nil when no head element")
	}

	// Add head element
	head := NewElement("head")
	html.InsertBefore(head, body)

	if doc.Head() != head {
		t.Error("should return head element")
	}
}

func TestDocumentBody(t *testing.T) {
	doc := NewDocument()

	// No body when no document element
	if doc.Body() != nil {
		t.Error("should return nil when no document element")
	}

	html := NewElement("html")
	doc.AppendChild(html)

	// No body when html has no body child
	if doc.Body() != nil {
		t.Error("should return nil when no body element")
	}

	head := NewElement("head")
	html.AppendChild(head)

	if doc.Body() != nil {
		t.Error("should return nil when only head element")
	}

	body := NewElement("body")
	html.AppendChild(body)

	if doc.Body() != body {
		t.Error("should return body element")
	}
}

func TestDocumentTitle(t *testing.T) {
	doc := NewDocument()

	// No title when no head
	if doc.Title() != "" {
		t.Error("should return empty string when no head")
	}

	html := NewElement("html")
	head := NewElement("head")
	doc.AppendChild(html)
	html.AppendChild(head)

	// No title when head has no title element
	if doc.Title() != "" {
		t.Error("should return empty string when no title element")
	}

	// Add non-title element first
	meta := NewElement("meta")
	head.AppendChild(meta)

	if doc.Title() != "" {
		t.Error("should return empty string when no title element")
	}

	// Add title element
	title := NewElement("title")
	title.AppendChild(NewText("Test Page"))
	head.AppendChild(title)

	if doc.Title() != "Test Page" {
		t.Errorf("expected 'Test Page', got '%s'", doc.Title())
	}
}

func TestDocumentQuery(t *testing.T) {
	doc := NewDocument()

	// No root element
	results, err := doc.Query("div")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if results != nil {
		t.Error("should return nil when no root element")
	}

	// With root element (delegates to element's Query)
	html := NewElement("html")
	doc.AppendChild(html)

	results, err = doc.Query("div")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Currently Query is a stub returning nil
	if results != nil {
		t.Error("stub should return nil")
	}
}

func TestDocumentQueryFirst(t *testing.T) {
	doc := NewDocument()

	// No root element
	result, err := doc.QueryFirst("div")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("should return nil when no root element")
	}
}

func TestDocumentClone(t *testing.T) {
	doc := NewDocument()
	doc.QuirksMode = Quirks
	doc.Doctype = NewDocumentType("html", "", "")
	html := NewElement("html")
	body := NewElement("body")
	doc.AppendChild(html)
	html.AppendChild(body)

	// Shallow clone
	shallowClone := doc.Clone(false).(*Document)
	if shallowClone.QuirksMode != Quirks {
		t.Error("shallow clone should preserve quirks mode")
	}
	if shallowClone.Doctype == nil {
		t.Error("shallow clone should have doctype")
	}
	if len(shallowClone.Children()) != 0 {
		t.Error("shallow clone should have no children")
	}

	// Deep clone
	deepClone := doc.Clone(true).(*Document)
	if len(deepClone.Children()) != 1 {
		t.Fatal("deep clone should have children")
	}
	clonedHTML := deepClone.Children()[0].(*Element)
	if clonedHTML == html {
		t.Error("deep cloned child should be different object")
	}
	if len(clonedHTML.Children()) != 1 {
		t.Error("deep clone should include grandchildren")
	}
}

// =============================================================================
// DocumentType Tests
// =============================================================================

func TestNewDocumentType(t *testing.T) {
	dt := NewDocumentType("html", "-//W3C//DTD HTML 4.01//EN", "http://www.w3.org/TR/html4/strict.dtd")

	if dt.Name != "html" {
		t.Errorf("expected Name 'html', got '%s'", dt.Name)
	}
	if dt.PublicID != "-//W3C//DTD HTML 4.01//EN" {
		t.Error("PublicID not set correctly")
	}
	if dt.SystemID != "http://www.w3.org/TR/html4/strict.dtd" {
		t.Error("SystemID not set correctly")
	}
	if dt.Type() != DoctypeNodeType {
		t.Errorf("expected DoctypeNodeType, got %d", dt.Type())
	}
}

func TestDocumentTypeParent(t *testing.T) {
	dt := NewDocumentType("html", "", "")

	if dt.Parent() != nil {
		t.Error("parent should be nil initially")
	}

	doc := NewDocument()
	dt.SetParent(doc)

	if dt.Parent() != doc {
		t.Error("parent should be set")
	}
}

func TestDocumentTypeLeafNode(t *testing.T) {
	dt := NewDocumentType("html", "", "")

	// DocumentType nodes have no children
	if dt.Children() != nil {
		t.Error("Children should return nil")
	}
	if dt.HasChildNodes() {
		t.Error("HasChildNodes should return false")
	}

	// Child operations are no-ops
	dt.AppendChild(NewText("ignored"))
	dt.InsertBefore(NewText("a"), NewText("b"))
	dt.RemoveChild(NewText("c"))
	result := dt.ReplaceChild(NewText("a"), NewText("b"))
	if result != nil {
		t.Error("ReplaceChild should return nil for leaf node")
	}
}

func TestDocumentTypeClone(t *testing.T) {
	dt := NewDocumentType("html", "public", "system")
	clone := dt.Clone(true).(*DocumentType)

	if clone.Name != "html" || clone.PublicID != "public" || clone.SystemID != "system" {
		t.Error("clone should have same data")
	}
	if clone == dt {
		t.Error("clone should be a different object")
	}
}

// =============================================================================
// DocumentFragment Tests
// =============================================================================

func TestNewDocumentFragment(t *testing.T) {
	df := NewDocumentFragment()

	if df.Type() != DocumentNodeType {
		t.Errorf("expected DocumentNodeType, got %d", df.Type())
	}
}

func TestDocumentFragmentAppendChildSetsParent(t *testing.T) {
	df := NewDocumentFragment()
	div := NewElement("div")
	df.AppendChild(div)
	if div.Parent() != df {
		t.Fatalf("div.Parent() = %T, want DocumentFragment", div.Parent())
	}
}

func TestDocumentFragmentClone(t *testing.T) {
	df := NewDocumentFragment()
	div := NewElement("div")
	span := NewElement("span")
	df.AppendChild(div)
	div.AppendChild(span)

	// Shallow clone
	shallowClone := df.Clone(false).(*DocumentFragment)
	if len(shallowClone.Children()) != 0 {
		t.Error("shallow clone should have no children")
	}

	// Deep clone
	deepClone := df.Clone(true).(*DocumentFragment)
	if len(deepClone.Children()) != 1 {
		t.Fatal("deep clone should have children")
	}
	clonedDiv := deepClone.Children()[0].(*Element)
	if clonedDiv == div {
		t.Error("deep cloned child should be different object")
	}
	if len(clonedDiv.Children()) != 1 {
		t.Error("deep clone should include grandchildren")
	}
}

// =============================================================================
// Attributes Tests
// =============================================================================

func TestAttributesGet(t *testing.T) {
	attrs := NewAttributes()
	attrs.Set("class", "container")

	val, found := attrs.Get("class")
	if !found {
		t.Error("should find attribute")
	}
	if val != "container" {
		t.Errorf("expected 'container', got '%s'", val)
	}

	// Case-insensitive lookup
	val, found = attrs.Get("CLASS")
	if !found {
		t.Error("should find attribute case-insensitively")
	}
	if val != "container" {
		t.Error("case-insensitive lookup failed")
	}

	// Non-existent attribute
	val, found = attrs.Get("nonexistent")
	if found {
		t.Error("should not find nonexistent attribute")
	}
	if val != "" {
		t.Error("should return empty string for nonexistent")
	}
}

func TestAttributesGetNS(t *testing.T) {
	attrs := NewAttributes()
	attrs.SetNS("http://www.w3.org/1999/xlink", "href", "link.html")

	val, found := attrs.GetNS("http://www.w3.org/1999/xlink", "href")
	if !found {
		t.Error("should find namespaced attribute")
	}
	if val != "link.html" {
		t.Errorf("expected 'link.html', got '%s'", val)
	}

	// Wrong namespace
	_, found = attrs.GetNS("", "href")
	if found {
		t.Error("should not find with wrong namespace")
	}
}

func TestAttributesSetUpdatesExisting(t *testing.T) {
	attrs := NewAttributes()
	attrs.Set("class", "old")
	attrs.Set("class", "new")

	if attrs.Len() != 1 {
		t.Errorf("should have 1 attribute, got %d", attrs.Len())
	}
	val, _ := attrs.Get("class")
	if val != "new" {
		t.Error("should update existing value")
	}
}

func TestAttributesSetNSUpdatesExisting(t *testing.T) {
	attrs := NewAttributes()
	attrs.SetNS("ns", "attr", "old")
	attrs.SetNS("ns", "attr", "new")

	if attrs.Len() != 1 {
		t.Errorf("should have 1 attribute, got %d", attrs.Len())
	}
	val, _ := attrs.GetNS("ns", "attr")
	if val != "new" {
		t.Error("should update existing value")
	}
}

func TestAttributesHas(t *testing.T) {
	attrs := NewAttributes()
	attrs.Set("class", "test")

	if !attrs.Has("class") {
		t.Error("should have class attribute")
	}
	if attrs.Has("nonexistent") {
		t.Error("should not have nonexistent attribute")
	}
}

func TestAttributesHasNS(t *testing.T) {
	attrs := NewAttributes()
	attrs.SetNS("ns", "attr", "value")

	if !attrs.HasNS("ns", "attr") {
		t.Error("should have namespaced attribute")
	}
	if attrs.HasNS("other", "attr") {
		t.Error("should not have attribute with wrong namespace")
	}
}

func TestAttributesRemove(t *testing.T) {
	attrs := NewAttributes()
	attrs.Set("class", "test")
	attrs.Set("id", "main")

	attrs.Remove("class")

	if attrs.Has("class") {
		t.Error("should have removed class")
	}
	if !attrs.Has("id") {
		t.Error("should still have id")
	}

	// Remove nonexistent should be no-op
	attrs.Remove("nonexistent")
	if attrs.Len() != 1 {
		t.Error("removing nonexistent should not affect count")
	}
}

func TestAttributesRemoveNS(t *testing.T) {
	attrs := NewAttributes()
	attrs.SetNS("ns", "attr", "value")

	attrs.RemoveNS("ns", "attr")

	if attrs.HasNS("ns", "attr") {
		t.Error("should have removed namespaced attribute")
	}
}

func TestAttributesAll(t *testing.T) {
	attrs := NewAttributes()
	attrs.Set("class", "test")
	attrs.Set("id", "main")

	all := attrs.All()

	if len(all) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(all))
	}

	// Verify it's a copy
	all[0].Value = testModified
	original, _ := attrs.Get("class")
	if original == testModified {
		t.Error("All should return a copy")
	}
}

func TestAttributesLen(t *testing.T) {
	attrs := NewAttributes()

	if attrs.Len() != 0 {
		t.Error("empty attributes should have length 0")
	}

	attrs.Set("class", "test")
	if attrs.Len() != 1 {
		t.Error("should have length 1")
	}
}

func TestAttributesClone(t *testing.T) {
	attrs := NewAttributes()
	attrs.Set("class", "test")
	attrs.SetNS("ns", "attr", "value")

	clone := attrs.Clone()

	if clone.Len() != 2 {
		t.Error("clone should have same length")
	}
	if clone == attrs {
		t.Error("clone should be different object")
	}

	// Modifying clone should not affect original
	clone.Set("class", testModified)
	original, _ := attrs.Get("class")
	if original == testModified {
		t.Error("modifying clone should not affect original")
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestBuildHTMLDocument(t *testing.T) {
	doc := NewDocument()
	doc.Doctype = NewDocumentType("html", "", "")

	html := NewElement("html")
	doc.AppendChild(html)

	head := NewElement("head")
	html.AppendChild(head)

	title := NewElement("title")
	title.AppendChild(NewText("Test Page"))
	head.AppendChild(title)

	body := NewElement("body")
	html.AppendChild(body)

	div := NewElement("div")
	div.SetAttr("id", "main")
	div.SetAttr("class", "container")
	body.AppendChild(div)

	p := NewElement("p")
	p.AppendChild(NewText("Hello, "))
	strong := NewElement("strong")
	strong.AppendChild(NewText("World"))
	p.AppendChild(strong)
	p.AppendChild(NewText("!"))
	div.AppendChild(p)

	// Verify structure
	if doc.Title() != "Test Page" {
		t.Errorf("expected title 'Test Page', got '%s'", doc.Title())
	}
	if doc.Body() != body {
		t.Error("Body() should return body element")
	}
	if div.ID() != "main" {
		t.Error("div should have id 'main'")
	}
	if !div.HasClass("container") {
		t.Error("div should have class 'container'")
	}

	// Test text extraction
	text := p.Text()
	if text != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got '%s'", text)
	}
}

func TestNodeTypeConstants(t *testing.T) {
	// Verify node type constants match DOM spec values
	if ElementNodeType != 1 {
		t.Error("ElementNodeType should be 1")
	}
	if TextNodeType != 3 {
		t.Error("TextNodeType should be 3")
	}
	if CommentNodeType != 8 {
		t.Error("CommentNodeType should be 8")
	}
	if DocumentNodeType != 9 {
		t.Error("DocumentNodeType should be 9")
	}
	if DoctypeNodeType != 10 {
		t.Error("DoctypeNodeType should be 10")
	}
}

func TestQuirksModeConstants(t *testing.T) {
	if NoQuirks != 0 {
		t.Error("NoQuirks should be 0")
	}
	if Quirks != 1 {
		t.Error("Quirks should be 1")
	}
	if LimitedQuirks != 2 {
		t.Error("LimitedQuirks should be 2")
	}
}

func TestNamespaceConstants(t *testing.T) {
	if NamespaceHTML != "http://www.w3.org/1999/xhtml" {
		t.Error("NamespaceHTML constant incorrect")
	}
	if NamespaceSVG != "http://www.w3.org/2000/svg" {
		t.Error("NamespaceSVG constant incorrect")
	}
	if NamespaceMathML != "http://www.w3.org/1998/Math/MathML" {
		t.Error("NamespaceMathML constant incorrect")
	}
}

// =============================================================================
// Coverage Tests for no-op methods on leaf nodes
// =============================================================================

// TestTextNoOpMethods ensures all no-op methods on Text are covered
func TestTextNoOpMethods(t *testing.T) {
	text := NewText("test")
	child := NewText("child")

	// These are all no-ops but need to be called for coverage
	text.AppendChild(child)
	text.InsertBefore(child, nil)
	text.RemoveChild(child)
}

// TestCommentNoOpMethods ensures all no-op methods on Comment are covered
func TestCommentNoOpMethods(t *testing.T) {
	comment := NewComment("test")
	child := NewText("child")

	// These are all no-ops but need to be called for coverage
	comment.AppendChild(child)
	comment.InsertBefore(child, nil)
	comment.RemoveChild(child)
}

// TestDocumentTypeNoOpMethods ensures all no-op methods on DocumentType are covered
func TestDocumentTypeNoOpMethods(t *testing.T) {
	dt := NewDocumentType("html", "", "")
	child := NewText("child")

	// These are all no-ops but need to be called for coverage
	dt.AppendChild(child)
	dt.InsertBefore(child, nil)
	dt.RemoveChild(child)
}

// TestDocumentQueryFirstWithResults tests QueryFirst when Query returns results
func TestDocumentQueryFirstWithResults(t *testing.T) {
	// Currently Query is a stub that returns nil, so this just verifies the path
	// When selector is implemented, this will need real test data
	doc := NewDocument()
	html := NewElement("html")
	doc.AppendChild(html)

	result, err := doc.QueryFirst("html")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Currently Query is a stub returning nil
	if result != nil {
		t.Error("stub should return nil")
	}
}

// TestElementQueryFirstWithResults tests QueryFirst when Query returns results
func TestElementQueryFirstWithResults(t *testing.T) {
	// Currently Query is a stub that returns nil
	// When selector is implemented, this will need real test data
	elem := NewElement("div")
	elem.AppendChild(NewElement("span"))

	result, err := elem.QueryFirst("span")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Currently Query is a stub returning nil
	if result != nil {
		t.Error("stub should return nil")
	}
}

// =============================================================================
// Tests for baseNode methods via Document and DocumentFragment
// (These types inherit baseNode methods for InsertBefore, RemoveChild, etc.)
// =============================================================================

func TestDocumentBaseNodeInsertBefore(t *testing.T) {
	doc := NewDocument()
	comment := NewComment("before html")
	html := NewElement("html")
	afterComment := NewComment("after html")

	doc.AppendChild(html)
	doc.AppendChild(afterComment)

	// InsertBefore html - uses baseNode.InsertBefore
	doc.InsertBefore(comment, html)

	children := doc.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}
	if children[0] != comment {
		t.Error("comment should be first")
	}
	if children[1] != html {
		t.Error("html should be second")
	}
	if children[2] != afterComment {
		t.Error("afterComment should be third")
	}
}

func TestDocumentBaseNodeInsertBeforeNilRef(t *testing.T) {
	doc := NewDocument()
	html := NewElement("html")

	// InsertBefore with nil ref should append
	doc.InsertBefore(html, nil)

	if len(doc.Children()) != 1 {
		t.Fatal("expected 1 child")
	}
	if doc.Children()[0] != html {
		t.Error("html should be the child")
	}
}

func TestDocumentBaseNodeInsertBeforeNotFound(t *testing.T) {
	doc := NewDocument()
	html := NewElement("html")
	notChild := NewComment("not a child")

	// InsertBefore with non-existent ref should append
	doc.InsertBefore(html, notChild)

	if len(doc.Children()) != 1 {
		t.Fatal("expected 1 child")
	}
	if doc.Children()[0] != html {
		t.Error("html should be appended")
	}
}

func TestDocumentBaseNodeRemoveChild(t *testing.T) {
	doc := NewDocument()
	comment := NewComment("test")
	html := NewElement("html")

	doc.AppendChild(comment)
	doc.AppendChild(html)

	// RemoveChild uses baseNode.RemoveChild
	doc.RemoveChild(comment)

	if len(doc.Children()) != 1 {
		t.Fatal("expected 1 child after removal")
	}
	if doc.Children()[0] != html {
		t.Error("html should remain")
	}
	if comment.Parent() != nil {
		t.Error("removed child's parent should be nil")
	}
}

func TestDocumentBaseNodeRemoveChildNotFound(t *testing.T) {
	doc := NewDocument()
	html := NewElement("html")
	notChild := NewComment("not a child")

	doc.AppendChild(html)

	// RemoveChild with non-existent child should be no-op
	doc.RemoveChild(notChild)

	if len(doc.Children()) != 1 {
		t.Error("children count should not change")
	}
}

func TestDocumentBaseNodeReplaceChild(t *testing.T) {
	doc := NewDocument()
	oldChild := NewComment("old")
	newChild := NewComment("new")

	doc.AppendChild(oldChild)

	// ReplaceChild uses baseNode.ReplaceChild
	result := doc.ReplaceChild(newChild, oldChild)

	if result != oldChild {
		t.Error("should return old child")
	}
	if oldChild.Parent() != nil {
		t.Error("old child's parent should be nil")
	}
	if newChild.Parent() != doc {
		t.Error("new child's parent should be doc")
	}
	if doc.Children()[0] != newChild {
		t.Error("new child should replace old")
	}
}

func TestDocumentBaseNodeReplaceChildNotFound(t *testing.T) {
	doc := NewDocument()
	html := NewElement("html")
	notChild := NewComment("not a child")
	newChild := NewComment("new")

	doc.AppendChild(html)

	// ReplaceChild with non-existent oldChild should return nil
	result := doc.ReplaceChild(newChild, notChild)

	if result != nil {
		t.Error("should return nil when old child not found")
	}
}

func TestDocumentBaseNodeHasChildNodes(t *testing.T) {
	doc := NewDocument()

	// HasChildNodes uses baseNode.HasChildNodes
	if doc.HasChildNodes() {
		t.Error("empty doc should not have children")
	}

	doc.AppendChild(NewElement("html"))

	if !doc.HasChildNodes() {
		t.Error("doc with child should have children")
	}
}

func TestDocumentFragmentBaseNodeInsertBefore(t *testing.T) {
	df := NewDocumentFragment()
	div1 := NewElement("div")
	div2 := NewElement("div")
	div3 := NewElement("div")

	df.AppendChild(div1)
	df.AppendChild(div3)

	// InsertBefore uses baseNode.InsertBefore
	df.InsertBefore(div2, div3)

	children := df.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}
	if children[0] != div1 || children[1] != div2 || children[2] != div3 {
		t.Error("children not in expected order")
	}
}

func TestDocumentFragmentBaseNodeRemoveChild(t *testing.T) {
	df := NewDocumentFragment()
	div := NewElement("div")

	df.AppendChild(div)
	df.RemoveChild(div)

	if len(df.Children()) != 0 {
		t.Error("should have no children after removal")
	}
}

func TestDocumentFragmentBaseNodeReplaceChild(t *testing.T) {
	df := NewDocumentFragment()
	oldDiv := NewElement("div")
	newDiv := NewElement("span")

	df.AppendChild(oldDiv)
	result := df.ReplaceChild(newDiv, oldDiv)

	if result != oldDiv {
		t.Error("should return old child")
	}
	if df.Children()[0] != newDiv {
		t.Error("new child should be in place")
	}
}

func TestDocumentFragmentBaseNodeHasChildNodes(t *testing.T) {
	df := NewDocumentFragment()

	if df.HasChildNodes() {
		t.Error("empty fragment should not have children")
	}

	df.AppendChild(NewElement("div"))

	if !df.HasChildNodes() {
		t.Error("fragment with child should have children")
	}
}
