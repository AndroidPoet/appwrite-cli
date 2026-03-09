package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Format represents the output format
type Format string

const (
	FormatJSON    Format = "json"
	FormatTable   Format = "table"
	FormatMinimal Format = "minimal"
	FormatTSV     Format = "tsv"
	FormatCSV     Format = "csv"
	FormatYAML    Format = "yaml"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

var (
	currentFormat Format
	prettyPrint   bool
	quietMode     bool
	writer        io.Writer = os.Stdout
)

// Setup initializes the output formatter
func Setup(format string, pretty, quiet bool) {
	currentFormat = Format(format)
	prettyPrint = pretty
	quietMode = quiet
}

// SetWriter sets the output writer (for testing)
func SetWriter(w io.Writer) {
	writer = w
}

// Print outputs data in the configured format
func Print(data interface{}) error {
	switch currentFormat {
	case FormatJSON:
		return printJSON(data)
	case FormatTable:
		return printTable(data)
	case FormatMinimal:
		return printMinimal(data)
	case FormatTSV:
		return printTSV(data)
	case FormatCSV:
		return printCSV(data)
	case FormatYAML:
		return printYAML(data)
	default:
		return printJSON(data)
	}
}

// PrintSuccess prints a success message (respects quiet mode)
func PrintSuccess(format string, args ...interface{}) {
	if !quietMode {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(writer, "%s%s✓ %s%s\n", colorBold, colorGreen, msg, colorReset)
	}
}

// PrintInfo prints an info message (respects quiet mode)
func PrintInfo(format string, args ...interface{}) {
	if !quietMode {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(writer, "%s%s%s\n", colorCyan, msg, colorReset)
	}
}

// PrintError prints an error message (always shown)
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s%s✗ Error: %s%s\n", colorBold, colorRed, fmt.Sprintf(format, args...), colorReset)
}

// PrintWarning prints a warning message
func PrintWarning(format string, args ...interface{}) {
	if !quietMode {
		fmt.Fprintf(os.Stderr, "%s⚠ Warning: %s%s\n", colorYellow, fmt.Sprintf(format, args...), colorReset)
	}
}

func printJSON(data interface{}) error {
	var out []byte
	var err error

	if prettyPrint {
		out, err = json.MarshalIndent(data, "", "  ")
	} else {
		out, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Fprintln(writer, string(out))
	return nil
}

func printTable(data interface{}) error {
	w := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
	defer w.Flush()

	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			fmt.Fprintln(writer, "(no results)")
			return nil
		}

		first := v.Index(0)
		if first.Kind() == reflect.Ptr {
			first = first.Elem()
		}

		headers := getStructHeaders(first)
		fmt.Fprintf(w, "%s%s%s\n", colorBold, strings.Join(headers, "\t"), colorReset)
		fmt.Fprintln(w, strings.Repeat("-\t", len(headers)))

		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			values := getStructValues(elem)
			fmt.Fprintln(w, strings.Join(values, "\t"))
		}
	} else if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct) {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		headers := getStructHeaders(v)
		values := getStructValues(v)
		for i, h := range headers {
			fmt.Fprintf(w, "%s:\t%s\n", h, values[i])
		}
	} else {
		return printJSON(data)
	}

	return nil
}

func printMinimal(data interface{}) error {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			if elem.Kind() == reflect.Struct && elem.NumField() > 0 {
				fmt.Fprintln(writer, elem.Field(0).Interface())
			}
		}
	} else if v.Kind() == reflect.Struct && v.NumField() > 0 {
		fmt.Fprintln(writer, v.Field(0).Interface())
	} else {
		fmt.Fprintln(writer, data)
	}

	return nil
}

func printTSV(data interface{}) error {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			return nil
		}

		first := v.Index(0)
		if first.Kind() == reflect.Ptr {
			first = first.Elem()
		}
		headers := getStructHeaders(first)
		fmt.Fprintln(writer, strings.Join(headers, "\t"))

		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			values := getStructValues(elem)
			fmt.Fprintln(writer, strings.Join(values, "\t"))
		}
	} else {
		return printJSON(data)
	}

	return nil
}

func printCSV(data interface{}) error {
	w := csv.NewWriter(writer)
	defer w.Flush()

	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			return nil
		}

		first := v.Index(0)
		if first.Kind() == reflect.Ptr {
			first = first.Elem()
		}
		headers := getStructHeaders(first)
		if err := w.Write(headers); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}

		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			values := getStructValues(elem)
			if err := w.Write(values); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	} else {
		return printJSON(data)
	}

	return nil
}

func printYAML(data interface{}) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	fmt.Fprint(writer, string(out))
	return nil
}

func getStructHeaders(v reflect.Value) []string {
	t := v.Type()
	headers := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			parts := strings.Split(tag, ",")
			headers = append(headers, strings.ToUpper(parts[0]))
		} else {
			headers = append(headers, strings.ToUpper(field.Name))
		}
	}
	return headers
}

func getStructValues(v reflect.Value) []string {
	values := make([]string, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		values = append(values, fmt.Sprintf("%v", field.Interface()))
	}
	return values
}

// Result wraps a successful operation result
type Result struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// PrintResult prints an operation result
func PrintResult(success bool, message string, data interface{}) error {
	return Print(&Result{
		Success: success,
		Message: message,
		Data:    data,
	})
}
