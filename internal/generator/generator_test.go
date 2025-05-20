package generator

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
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
		})
	}
}

func TestGenerateWithContext(t *testing.T) {
	// Create a new name generator
	generator := NewNameGenerator(4)
	defer generator.Shutdown()
	
	// Test context cancellation
	t.Run("ContextCancellation", func(t *testing.T) {
		// Create a context that is already canceled
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		// Try to generate names with the canceled context
		names := generator.GenerateWithContext(ctx, "A", 100)
		
		// Should return an empty slice or a partial result
		if len(names) >= 100 {
			t.Errorf("Expected context cancellation to limit results, got %d names", len(names))
		}
	})
	
	// Test context timeout
	t.Run("ContextTimeout", func(t *testing.T) {
		// Create a context with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		
		// Try to generate names with the timed-out context
		names := generator.GenerateWithContext(ctx, "A", 10000) // Use a large number to ensure it takes longer than the timeout
		
		// Should return a partial result
		if len(names) >= 10000 {
			t.Errorf("Expected context timeout to limit results, got %d names", len(names))
		}
	})
}

func TestCaching(t *testing.T) {
	// Create a new name generator
	generator := NewNameGenerator(4)
	defer generator.Shutdown()
	
	// Generate names first time
	letter := "C"
	count := 10
	firstNames := generator.Generate(letter, count)
	
	// Generate names second time with same parameters
	secondNames := generator.Generate(letter, count)
	
	// Check that the results are the same (from cache)
	if len(firstNames) != len(secondNames) {
		t.Errorf("Cache results have different lengths: first=%d, second=%d", len(firstNames), len(secondNames))
	} else {
		for i := 0; i < len(firstNames); i++ {
			if firstNames[i] != secondNames[i] {
				t.Errorf("Cache results differ at index %d: first=%q, second=%q", i, firstNames[i], secondNames[i])
			}
		}
	}
}

func TestConcurrentGeneration(t *testing.T) {
	// Create a new name generator
	generator := NewNameGenerator(4)
	defer generator.Shutdown()
	
	// Number of concurrent generations
	numConcurrent := 100
	
	// Create a wait group
	var wg sync.WaitGroup
	wg.Add(numConcurrent)
	
	// Channel to collect errors
	errCh := make(chan error, numConcurrent)
	
	// Generate names concurrently
	for i := 0; i < numConcurrent; i++ {
		go func(id int) {
			defer wg.Done()
			
			// Use different letters to avoid cache hits
			letter := string(rune('A' + id%26))
			count := 5
			
			names := generator.Generate(letter, count)
			
			// Check if the correct number of names was generated
			if len(names) != count {
				errCh <- fmt.Errorf("generator %d: expected %d names, got %d", id, count, len(names))
				return
			}
			
			// Check if the names start with the correct letter
			for j, name := range names {
				if len(name) == 0 {
					errCh <- fmt.Errorf("generator %d: empty name at index %d", id, j)
					return
				}
				
				if string(name[0]) != letter {
					errCh <- fmt.Errorf("generator %d: name %q does not start with %q", id, name, letter)
					return
				}
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
	close(errCh)
	
	// Check for errors
	for err := range errCh {
		t.Error(err)
	}
}

func BenchmarkGenerateNames(b *testing.B) {
	// Reset the generator to ensure we start fresh
	DefaultGenerator = nil
	
	for _, count := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("Count=%d", count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				GenerateNames("A", count)
			}
		})
	}
}

func BenchmarkGenerateNamesParallel(b *testing.B) {
	// Reset the generator to ensure we start fresh
	DefaultGenerator = nil
	
	for _, count := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("Count=%d", count), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					GenerateNames("A", count)
				}
			})
		})
	}
}
