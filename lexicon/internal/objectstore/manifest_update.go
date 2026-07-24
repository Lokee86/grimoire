package objectstore

import "sort"

func (m Manifest) Language(language string) (LanguageEntry, bool) {
	for _, entry := range m.Languages {
		if entry.Language == language {
			return entry, true
		}
	}
	return LanguageEntry{}, false
}

func (m Manifest) WithLanguage(entry LanguageEntry) Manifest {
	languages := make([]LanguageEntry, 0, len(m.Languages)+1)
	replaced := false
	for _, current := range m.Languages {
		if current.Language == entry.Language {
			languages = append(languages, entry)
			replaced = true
			continue
		}
		languages = append(languages, current)
	}
	if !replaced {
		languages = append(languages, entry)
	}
	sort.Slice(languages, func(left, right int) bool { return languages[left].Language < languages[right].Language })
	m.Languages = languages
	return m
}

func (m Manifest) WithoutLanguage(language string) Manifest {
	languages := make([]LanguageEntry, 0, len(m.Languages))
	for _, entry := range m.Languages {
		if entry.Language != language {
			languages = append(languages, entry)
		}
	}
	m.Languages = languages
	return m
}
