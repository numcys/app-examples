package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

type Annotation struct {
	Key   string      `yaml:"key"`
	Value interface{} `yaml:"value"`
}

type Annotations []Annotation

// UnmarshalYAML handles custom unmarshaling of annotations.
func (a *Annotations) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var rawAnnotations []interface{}
	if err := unmarshal(&rawAnnotations); err != nil {
		return fmt.Errorf("failed to unmarshal annotations: %w", err)
	}

	var annotations []Annotation

	for _, raw := range rawAnnotations {
		switch item := raw.(type) {
		case string:
			// Handle "key=value" style annotations
			parts := strings.SplitN(item, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid annotation format: %s", item)
			}
			annotations = append(annotations, Annotation{Key: parts[0], Value: parts[1]})

		case map[interface{}]interface{}:
			// Handle annotations as map[interface{}]interface{}
			for k, v := range item {
				keyStr, ok := k.(string)
				if !ok {
					return fmt.Errorf("annotation key is not a string: %v", k)
				}

				// Try to parse the value as JSON
				var jsonValue interface{}
				if jsonStr, ok := v.(string); ok {
					if err := json.Unmarshal([]byte(jsonStr), &jsonValue); err == nil {
						// Successfully unmarshaled JSON string into a map
						annotations = append(annotations, Annotation{Key: keyStr, Value: jsonValue})
					} else {
						// If not a valid JSON, store the string as-is
						annotations = append(annotations, Annotation{Key: keyStr, Value: jsonStr})
					}
				} else {
					// Store as is if not a JSON string
					annotations = append(annotations, Annotation{Key: keyStr, Value: v})
				}
			}

		case map[string]interface{}:
			// Handle annotations as map[string]interface{}
			for k, v := range item {
				// Try to parse the value as JSON
				var jsonValue interface{}
				if jsonStr, ok := v.(string); ok {
					if err := json.Unmarshal([]byte(jsonStr), &jsonValue); err == nil {
						// Successfully unmarshaled JSON string into a map
						annotations = append(annotations, Annotation{Key: k, Value: jsonValue})
					} else {
						// If not a valid JSON, store the string as-is
						annotations = append(annotations, Annotation{Key: k, Value: jsonStr})
					}
				} else {
					// Store as is if not a JSON string
					annotations = append(annotations, Annotation{Key: k, Value: v})
				}
			}

		default:
			return fmt.Errorf("unsupported annotation type: %T", item)
		}
	}

	*a = annotations
	return nil
}

// AnnotationsMap will return a map of string to []string// AnnotationsMap will return a map of string to []string
func (a Annotations) AnnotationsMap() map[string]string {
	annotations := make(map[string]string)

	for _, annotation := range a {
		// Marshal the value to a JSON string
		jsonValue, err := json.Marshal(annotation.Value)
		if err != nil {
			// Handle JSON marshaling errors if needed
			fmt.Printf("Error marshaling annotation value for key %s: %v\n", annotation.Key, err)
			continue
		}

		jsonStr := string(jsonValue)

		if containsJSONStructure(jsonStr) {
			key := fmt.Sprintf("%s : |", annotation.Key)
			annotations[key] = jsonStr
		} else {
			key := fmt.Sprintf("%s :", annotation.Key)
			annotations[key] = jsonStr
		}
	}

	return annotations
}

func containsJSONStructure(str string) bool {
	return strings.Contains(str, "{") || strings.Contains(str, "}")
}

// Helper function to check if a string is a valid JSON object or array
func isJSON(str string) bool {
	str = strings.TrimSpace(str)
	// Check if the string starts with '{' or '[' and ends with '}' or ']'
	return (len(str) > 1 && ((str[0] == '{' && str[len(str)-1] == '}') || (str[0] == '[' && str[len(str)-1] == ']')))
}

// Helper function to split a string into lines (for regular strings)
func splitLines(str string) []string {
	// Split string by newlines, handling multiline content
	return strings.Split(str, "\n")
}

// Convert multiline string or YAML into a slice of strings, one per line
func splitIntoLines(value interface{}) []string {
	var lines []string
	var valueStr string

	// Convert the value to string if it's not already
	switch v := value.(type) {
	case string:
		valueStr = v
	case []byte:
		valueStr = string(v)
	default:
		// Marshal the interface value to a string (YAML or JSON)
		bytes, err := yaml.Marshal(v)
		if err != nil {
			log.Printf("Error marshaling interface to YAML: %v", err)
			valueStr = fmt.Sprintf("%v", v)
		} else {
			valueStr = string(bytes)
		}
	}

	// Split based on newlines
	for _, line := range strings.Split(valueStr, "\n") {
		lines = append(lines, line)
	}
	return lines
}

func splitIntoLinesPretty(s string) ([]string, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, []byte(s), "", "  ")
	if err != nil {
		return nil, err // Return the error if JSON is invalid
	}
	return strings.Split(prettyJSON.String(), "\n"), nil
}

func splitIntoLinesUnified(value interface{}) []string {
	var lines []string
	var valueStr string

	// Convert the value to string if it's not already
	switch v := value.(type) {
	case string:
		valueStr = v
	case []byte:
		valueStr = string(v)
	default:
		// Marshal the interface value to YAML
		bytes, err := yaml.Marshal(v)
		if err != nil {
			log.Printf("Error marshaling interface to YAML: %v", err)
			valueStr = fmt.Sprintf("%v", v)
		} else {
			valueStr = string(bytes)
		}
	}

	// Check if the value is a valid JSON string
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(valueStr), "", "  "); err == nil {
		// If valid JSON, use the pretty-printed JSON
		valueStr = prettyJSON.String()
	}

	// Split the final string into lines
	for _, line := range strings.Split(valueStr, "\n") {
		lines = append(lines, line)
	}
	return lines
}

func main() {

	yamlData := `
annotations:
  - "key1=value1"
  - key2: "value2"
  - key3: |
      this is it
      line2
  - ad.datadoghq.com/test.checks: |
      {   
          "my-test": {
          "init_config": {
              "is_jmx": true,
              "collect_default_metrics": true
          },
          "instances": [{
              "host": "localhost",
              "port": "1234"
          }]
      }
      }
`

	// Parse the YAML into Annotations
	var data struct {
		Annotations Annotations `yaml:"annotations"`
	}

	if err := yaml.Unmarshal([]byte(yamlData), &data); err != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Get the annotations as a map
	annotationsMap := data.Annotations.AnnotationsMap()
	fmt.Println(annotationsMap)

	// Prepare the template
	tmplStr := `Annotations:
{{- range $key, $value := . }}
    {{ $key }}
    {{- $lines := splitIntoLinesUnified $value }}
    {{- range $lines }}
      {{ . }}
    {{- end }}
{{- end }}
`

	// Register the custom function to split string into lines
	funcMap := template.FuncMap{
		"splitIntoLines":        splitIntoLines,
		"splitIntoLinesPretty":  splitIntoLinesPretty,
		"splitIntoLinesUnified": splitIntoLinesUnified,
	}

	tmpl, err := template.New("annotations").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	// Use os.Stdout as the writer for the template output
	err = tmpl.ExecuteTemplate(os.Stdout, "annotations", annotationsMap)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, annotationsMap); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	fmt.Println(buf.Bytes())

	dat2, err := FormatYAML(buf.Bytes())

	fmt.Println(string(dat2))

}

var (
	yamlSplitter = regexp.MustCompile(`(?m)^\s*---\s*$`)
)

// FormatYAML processes the YAML data, adds the '|' pipe symbol before unmarshalling, and ensures correct JSON formatting.
func FormatYAML(data []byte) ([]byte, error) {
	// First, unescape the HTML entities to make sure JSON is rendered properly
	unescapedData := []byte(html.UnescapeString(string(data)))

	fmt.Println(string(unescapedData))

	// Split the YAML by the document separator
	ps := yamlSplitter.Split(string(unescapedData), -1)
	bs := make([][]byte, len(ps))

	for i, p := range ps {
		var v interface{}
		// Unmarshal the split YAML content
		if err := yaml.Unmarshal([]byte(p), &v); err != nil {
			return nil, fmt.Errorf("Error while Doing final Formatting - Unmarshalling: %w", err)
		}

		// Marshal back to YAML after modifications
		data, err := yaml.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("Error while Doing final Formatting - Marshalling: %w", err)
		}

		bs[i] = data
	}

	// Join all YAML segments back with the '---' separator
	return bytes.Join(bs, []byte("---\n")), nil
}

// addPipeBeforeUnmarshalling modifies the YAML string to add '|' for multiline annotations with JSON-like content
func addPipeBeforeUnmarshalling(data []byte) []byte {
	// Split the data into lines
	lines := strings.Split(string(data), "\n")
	var modifiedLines []string

	for _, line := range lines {
		// Check if the line looks like it contains a JSON object
		if isJSON(line) {
			// Prepend the pipe symbol ('|') to this line
			// Example: "key: {...}" -> "key: | {...}"
			modifiedLines = append(modifiedLines, strings.Replace(line, ":", ": |", 1))
		} else {
			// Keep the line as is if it's not JSON-like
			modifiedLines = append(modifiedLines, line)
		}
	}

	// Join the modified lines back into a single string
	return []byte(strings.Join(modifiedLines, "\n"))
}

// isJSON checks if a string looks like a JSON object (starts with { and ends with })
