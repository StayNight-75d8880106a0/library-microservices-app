package helper

import (
	"strings"
)

var subjectBlacklist = map[string]bool{
	"open library staff picks": true,
	"accessible book":          true,
	"protected daisy":          true,
	"in library":               true,
	"overdrive":                true,
	"large type books":         true,
	"reading level":            true,
}

type categoryRule struct {
	Category string
	Keywords []string
}

var categoryRules = []categoryRule{
	{"Fantasy", []string{"fantasy", "fairy tale", "magic"}},
	{"Science Fiction", []string{"science fiction", "sci-fi"}},
	{"Mystery", []string{"mystery", "detective", "crime"}},
	{"Thriller", []string{"thriller", "suspense", "espionage"}},
	{"Horror", []string{"horror", "ghost"}},
	{"Romance", []string{"romance", "love stories"}},
	{"Biography", []string{"biography", "autobiograph", "memoir"}},
	{"History", []string{"history", "historical"}},
	{"Poetry", []string{"poetry", "poems"}},
	{"Drama", []string{"drama", "plays", "theater"}},
	{"Comics", []string{"comic", "graphic novel", "manga"}},
	{"Children", []string{"juvenile", "children", "picture book"}},
	{"Education", []string{"textbook", "study guide", "education", "teaching"}},
	{"Technology", []string{"computer", "programming", "software", "technology", "engineering"}},
	{"Business", []string{"business", "economics", "management", "finance", "marketing"}},
	{"Self-Help", []string{"self-help", "self actualization", "motivation", "personal development"}},
	{"Religion", []string{"religion", "islam", "christian", "spiritual", "theology"}},
	{"Philosophy", []string{"philosophy", "ethics", "logic"}},
	{"Cooking", []string{"cooking", "cookbook", "recipes"}},
	{"Travel", []string{"travel", "guidebook"}},
	{"Health", []string{"health", "medicine", "fitness", "nutrition"}},
	{"Science", []string{"science", "mathematics", "physics", "chemistry", "biology"}},
	{"Fiction", []string{"fiction"}},
}

func ResolveCategory(subjects []string) string {
	if len(subjects) == 0 {
		return ""
	}

	normalized := make([]string, 0, len(subjects))

	for _, subject := range subjects {
		subject = strings.ToLower(strings.TrimSpace(subject))
		if subject == "" || subjectBlacklist[subject] {
			continue
		}
		normalized = append(normalized, subject)
	}

	for _, rule := range categoryRules {
		for _, subject := range normalized {
			for _, keyword := range rule.Keywords {
				if strings.Contains(subject, keyword) {
					return rule.Category
				}
			}
		}
	}

	return ""
}
