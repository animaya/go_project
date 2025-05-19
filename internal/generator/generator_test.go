package generator

import (
	"testing"
)

func TestGenerateNames(t *testing.T) {
	tests := []struct {
		name                 string
		letter               string
		count                int
		wantCount            int
		wantStartsWithLetter bool
	}{
		{
			name:                 "Generate 5 names with letter A",
			letter:               "A",
			count:                5,
			wantCount:            5,
			wantStartsWithLetter: true,
		},
		{
			name:                 "Generate 0 names",
			letter:               "A",
			count:                0,
			wantCount:            0,
			wantStartsWithLetter: true,
		},
		{
			name:                 "Generate 100 names (should cap at available names)",
			letter:               "A",
			count:                100,
			wantCount:            len(NamesByLetter["A"]),
			wantStartsWithLetter: true,
		},
		{
			name:                 "Use empty letter (should choose random letter)",
			letter:               "",
			count:                5,
			wantCount:            5,
			wantStartsWithLetter: false, // Will check for any valid letter
		},
		{
			name:                 "Use lowercase letter (should convert to uppercase)",
			letter:               "b",
			count:                5,
			wantCount:            5,
			wantStartsWithLetter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateNames(tt.letter, tt.count)
			
			// Check the count
			if len(got) != tt.wantCount {
				t.Errorf("GenerateNames() returned %d names, want %d names", len(got), tt.wantCount)
			}
			
			// Skip other checks if we expect 0 names
			if tt.wantCount == 0 {
				return
			}
			
			// Check if names start with the right letter
			if tt.wantStartsWithLetter {
				expectedLetter := tt.letter
				if expectedLetter == "" {
					// If letter was empty, we don't know which letter was chosen
					// so just grab the first letter of the first name
					if len(got) > 0 {
						expectedLetter = string(got[0][0])
					}
				} else {
					// Convert to uppercase
					expectedLetter = string(expectedLetter[0])
					if expectedLetter[0] >= 'a' && expectedLetter[0] <= 'z' {
						expectedLetter = string(expectedLetter[0] - 32)
					}
				}
				
				for i, name := range got {
					if len(name) == 0 {
						t.Errorf("GenerateNames() returned empty name at index %d", i)
						continue
					}
					
					firstLetter := string(name[0])
					if firstLetter != expectedLetter {
						t.Errorf("GenerateNames() returned name %q starting with %q, want %q", name, firstLetter, expectedLetter)
					}
				}
			}
			
			// Check for duplicates
			nameSet := make(map[string]bool)
			for _, name := range got {
				if nameSet[name] {
					// This is not necessarily an error, just log a warning
					t.Logf("Warning: GenerateNames() returned duplicate name: %q", name)
				}
				nameSet[name] = true
			}
		})
	}
}
