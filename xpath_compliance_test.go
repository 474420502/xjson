package xjson

import (
	"testing"
)

const complianceTestJSON = `
{
  "store": {
    "book": [
      { "category": "reference",
        "author": "Nigel Rees",
        "title": "Sayings of the Century",
        "price": 8.95
      },
      { "category": "fiction",
        "author": "Evelyn Waugh",
        "title": "Sword of Honour",
        "price": 12.99,
        "available": true
      },
      { "category": "fiction",
        "author": "Herman Melville",
        "title": "Moby Dick",
        "isbn": "0-553-21311-3",
        "price": 8.99,
        "available": true
      },
      { "category": "fiction",
        "author": "J. R. R. Tolkien",
        "title": "The Lord of the Rings",
        "isbn": "0-395-19395-8",
        "price": 22.99,
        "available": false
      }
    ],
    "bicycle": {
      "color": "red",
      "price": 19.95
    }
  },
  "expensive": 10
}
`

func TestXPathCompliance(t *testing.T) {
	doc, err := ParseString(complianceTestJSON)
	if err != nil {
		t.Fatalf("Failed to parse test JSON: %v", err)
	}

	tests := []struct {
		name          string
		xpath         string
		expectedCount int
		expectedTitle string // Used for single match verification
	}{
		{
			name:          "Select by attribute equality",
			xpath:         "//book[@category='fiction']",
			expectedCount: 3,
		},
		{
			name:          "Select by attribute and numeric comparison",
			xpath:         "//book[@category='fiction' and @price < 10]",
			expectedCount: 1,
			expectedTitle: "Moby Dick",
		},
		{
			name:          "Select by numeric comparison",
			xpath:         "//book[@price > 20]",
			expectedCount: 1,
			expectedTitle: "The Lord of the Rings",
		},
		{
			name:          "Select by boolean attribute",
			xpath:         "//book[@available=true]",
			expectedCount: 2,
		},
		{
			name:          "Select by positional index (first)",
			xpath:         "//book[0]",
			expectedCount: 1,
			expectedTitle: "Sayings of the Century",
		},
		{
			name:          "Select by positional index (last)",
			xpath:         "//book[3]",
			expectedCount: 1,
			expectedTitle: "The Lord of the Rings",
		},
		// {
		// 	name:     "Select by positional index (last)",
		// 	xpath:    "//book[last()]",
		// 	expectedCount: 1,
		// 	expectedTitle: "The Lord of the Rings",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := doc.Query(tt.xpath)
			if result.Error() != nil {
				t.Errorf("Query failed with error: %v", result.Error())
				return
			}

			if result.MatchCount() != tt.expectedCount {
				t.Errorf("Expected %d matches, but got %d", tt.expectedCount, result.MatchCount())
			}

			if tt.expectedTitle != "" && result.MatchCount() == 1 {
				title, err := result.Get("title").String()
				if err != nil {
					t.Errorf("Failed to get title from result: %v", err)
				}
				if title != tt.expectedTitle {
					t.Errorf("Expected title '%s', but got '%s'", tt.expectedTitle, title)
				}
			}
		})
	}
}
