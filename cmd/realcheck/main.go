package main

// Throwaway repro: real DB -> GetResumeState -> pdf.Generate, exactly as
// ExportPDF does. Prints structure only. Deleted after use.

import (
	"fmt"
	"os"

	"resumator/internal/pdf"
	"resumator/internal/store"
)

func main() {
	s, err := store.Open()
	if err != nil {
		fmt.Println("FAIL: open store:", err)
		os.Exit(1)
	}
	defer s.Close()

	presets, err := s.ListPresets()
	if err != nil {
		fmt.Println("FAIL: presets:", err)
		os.Exit(1)
	}
	fmt.Println("presets:", len(presets))
	for _, p := range presets {
		state, err := s.GetResumeState(p.ID)
		if err != nil {
			fmt.Println("FAIL: state:", err)
			os.Exit(1)
		}
		fmt.Printf("preset %q: %d sections\n", p.Name, len(state.Sections))
		for i, sec := range state.Sections {
			fmt.Printf("  [%d] id=%d type=%s visible=%v items=%d title=%q\n",
				i, sec.ID, sec.Type, sec.Visible, len(sec.Items), sec.Title)
		}
		bytes, err := pdf.Generate(state, "a4")
		if err != nil {
			fmt.Println("FAIL: generate:", err)
			os.Exit(1)
		}
		name := fmt.Sprintf("/tmp/resumator-real-%d.pdf", p.ID)
		if err := os.WriteFile(name, bytes, 0o644); err != nil {
			fmt.Println("FAIL: write:", err)
			os.Exit(1)
		}
		fmt.Println("  written", name)
	}
}
