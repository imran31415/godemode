package browsertools

import (
	"fmt"
)

// Registry holds all browser tools
type Registry struct{}

// NewRegistry creates a new browser tools registry
func NewRegistry() *Registry {
	return &Registry{}
}

// Call executes a tool by name with the given arguments
func (r *Registry) Call(toolName string, args map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "navigate":
		url, _ := args["url"].(string)
		waitSelector, _ := args["wait_selector"].(string)
		return Navigate(url, waitSelector)

	case "get_title":
		return GetTitle()

	case "get_url":
		return GetURL()

	case "screenshot":
		filename, _ := args["filename"].(string)
		fullPage, _ := args["full_page"].(bool)
		return Screenshot(filename, fullPage)

	case "get_text":
		selector, _ := args["selector"].(string)
		return GetText(selector)

	case "get_attribute":
		selector, _ := args["selector"].(string)
		attribute, _ := args["attribute"].(string)
		return GetAttribute(selector, attribute)

	case "click":
		selector, _ := args["selector"].(string)
		return Click(selector)

	case "type":
		selector, _ := args["selector"].(string)
		text, _ := args["text"].(string)
		return Type(selector, text)

	case "wait_for_selector":
		selector, _ := args["selector"].(string)
		timeout := 10
		if t, ok := args["timeout"].(float64); ok {
			timeout = int(t)
		}
		if t, ok := args["timeout"].(int); ok {
			timeout = t
		}
		return WaitForSelector(selector, timeout)

	case "count_elements":
		selector, _ := args["selector"].(string)
		return CountElements(selector)

	case "get_all_text":
		selector, _ := args["selector"].(string)
		limit := 10
		if l, ok := args["limit"].(float64); ok {
			limit = int(l)
		}
		if l, ok := args["limit"].(int); ok {
			limit = l
		}
		return GetAllText(selector, limit)

	case "get_all_attributes":
		selector, _ := args["selector"].(string)
		attribute, _ := args["attribute"].(string)
		limit := 10
		if l, ok := args["limit"].(float64); ok {
			limit = int(l)
		}
		if l, ok := args["limit"].(int); ok {
			limit = l
		}
		return GetAllAttributes(selector, attribute, limit)

	case "evaluate":
		script, _ := args["script"].(string)
		return Evaluate(script)

	case "scroll_to":
		selector, _ := args["selector"].(string)
		return ScrollTo(selector)

	case "element_exists":
		selector, _ := args["selector"].(string)
		return ElementExists(selector)

	case "get_inner_html":
		selector, _ := args["selector"].(string)
		return GetInnerHTML(selector)

	case "sleep":
		ms := 1000
		if m, ok := args["milliseconds"].(float64); ok {
			ms = int(m)
		}
		if m, ok := args["milliseconds"].(int); ok {
			ms = m
		}
		return Sleep(ms)

	case "get_page_source":
		return GetPageSource()

	case "init_browser":
		err := InitBrowser()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil

	case "close_browser":
		CloseBrowser()
		return map[string]interface{}{"success": true}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// GetToolDocumentation returns documentation for all tools
func GetToolDocumentation() string {
	return `Available Browser/Puppeteer Tools:

1. navigate(url, wait_selector)
   Navigate to a URL and optionally wait for a selector
   Args: url (string, required), wait_selector (string, optional)

2. get_title()
   Get the page title
   Returns: title (string)

3. get_url()
   Get the current page URL
   Returns: url (string)

4. screenshot(filename, full_page)
   Take a screenshot
   Args: filename (string), full_page (bool)
   Returns: success, filename, size

5. get_text(selector)
   Get text content from an element
   Args: selector (string, CSS selector)
   Returns: text (string)

6. get_attribute(selector, attribute)
   Get an attribute value from an element
   Args: selector (string), attribute (string)
   Returns: value (string), exists (bool)

7. click(selector)
   Click on an element
   Args: selector (string)
   Returns: success (bool)

8. type(selector, text)
   Type text into an input field
   Args: selector (string), text (string)
   Returns: success (bool)

9. wait_for_selector(selector, timeout)
   Wait for an element to appear
   Args: selector (string), timeout (int, seconds, default 10)
   Returns: success (bool)

10. count_elements(selector)
    Count elements matching a selector
    Args: selector (string)
    Returns: count (int)

11. get_all_text(selector, limit)
    Get text from all matching elements
    Args: selector (string), limit (int, default 10)
    Returns: texts ([]string), count (int)

12. get_all_attributes(selector, attribute, limit)
    Get attribute values from all matching elements
    Args: selector (string), attribute (string), limit (int)
    Returns: values ([]string), count (int)

13. evaluate(script)
    Execute JavaScript in page context
    Args: script (string)
    Returns: result (interface{})

14. scroll_to(selector)
    Scroll to an element
    Args: selector (string)
    Returns: success (bool)

15. element_exists(selector)
    Check if an element exists
    Args: selector (string)
    Returns: exists (bool)

16. get_inner_html(selector)
    Get innerHTML of an element
    Args: selector (string)
    Returns: html (string)

17. sleep(milliseconds)
    Wait for specified time
    Args: milliseconds (int)
    Returns: success (bool)

18. get_page_source()
    Get full page HTML source
    Returns: source (string, truncated), length (int)
`
}
