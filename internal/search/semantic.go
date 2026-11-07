package search

import (
	"math"
	"strings"
	"sync"
)

// Document represents a single document for semantic search
type Document struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Source   string `json:"source"`
	Category string `json:"category"`
}

// SearchResult represents a search result with relevance score
type SearchResult struct {
	Document Document `json:"document"`
	Score    float64  `json:"score"`
}

// SemanticSearcher provides simple semantic search capabilities
type SemanticSearcher struct {
	mu         sync.RWMutex
	docs       []Document
	stopWords  map[string]bool
	termFreqs  map[string]map[string]int // docID -> term -> count
	docFreqs   map[string]int            // term -> doc count
	vocab      map[string]int            // term -> index
	termCounts []int
	nDocs      int
}

// NewSemanticSearcher creates a new SemanticSearcher instance
func NewSemanticSearcher() *SemanticSearcher {
	return &SemanticSearcher{
		stopWords:  make(map[string]bool),
		termFreqs:  make(map[string]map[string]int),
		docFreqs:   make(map[string]int),
		vocab:      make(map[string]int),
		termCounts: make([]int, 0),
	}
}

// AddDocument adds a document to the search index
func (s *SemanticSearcher) AddDocument(doc Document) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if document already exists
	for i, d := range s.docs {
		if d.ID == doc.ID {
			s.docs[i] = doc
			return
		}
	}
	s.docs = append(s.docs, doc)
}

// RemoveDocument removes a document by ID
func (s *SemanticSearcher) RemoveDocument(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, doc := range s.docs {
		if doc.ID == id {
			s.docs = append(s.docs[:i], s.docs[i+1:]...)
			return
		}
	}
}

// IndexDocuments builds the search index from documents
func (s *SemanticSearcher) IndexDocuments(docs []Document) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.docs = docs
	s.nDocs = len(docs)

	// Build vocabulary and compute statistics
	termFreqs := make(map[string]map[string]int)
	docFreqs := make(map[string]int)

	for _, doc := range docs {
		terms := s.tokenize(doc.Content + " " + doc.Title)
		if termFreqs[doc.ID] == nil {
			termFreqs[doc.ID] = make(map[string]int)
		}
		for _, term := range terms {
			termFreqs[doc.ID][term]++
			docFreqs[term]++
		}
	}

	s.termFreqs = termFreqs
	s.docFreqs = docFreqs

	// Build vocabulary
	vocab := make(map[string]int)
	idx := 0
	for term := range docFreqs {
		vocab[term] = idx
		idx++
	}
	s.vocab = vocab

	// Build term counts
	termCounts := make([]int, len(docFreqs))
	for _, count := range docFreqs {
		termCounts = append(termCounts, count)
	}
	s.termCounts = termCounts
}

// Search performs a semantic search and returns top-k results
func (s *SemanticSearcher) Search(query string, topK int) []SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	terms := s.tokenize(query)
	if len(terms) == 0 {
		return []SearchResult{}
	}

	// Compute IDF for each term
	idf := make(map[string]float64)
	for term := range s.docFreqs {
		df := s.docFreqs[term]
		if df > 0 {
			idf[term] = math.Log(float64(s.nDocs+1) / float64(df))
		}
	}

	// Compute TF-IDF scores for each document
	scores := make(map[string]float64)
	for docID, tf := range s.termFreqs {
		for term, count := range tf {
			// Inverse document frequency
			idfVal := math.Log(float64(s.nDocs+1) / float64(s.docFreqs[term]+1))

			// Term frequency with sublinear scaling
			tfIdf := math.Log1p(float64(count))
			if tfIdf == 0 {
				tfIdf = 0.5
			} else {
				tfIdf = 1 + math.Log(tfIdf)
			}

			// Document length normalization
			docLen := 1.0
			for _, c := range tf {
				docLen += float64(c)
			}
			docLenNorm := 0.5 * (1 + math.Log(math.Max(1, docLen)))

			scores[docID] += tfIdf * idfVal / docLenNorm
		}
	}

	// Sort documents by score
	type scoreEntry struct {
		docID string
		score float64
	}
	entries := make([]scoreEntry, 0, len(scores))
	for docID, score := range scores {
		entries = append(entries, scoreEntry{docID, score})
	}

	// Simple selection sort for top-k
	for i := 0; i < len(entries); i++ {
		maxIdx := i
		for j := i + 1; j < len(entries); j++ {
			if entries[j].score > entries[maxIdx].score {
				maxIdx = j
			}
		}
		entries[i], entries[maxIdx] = entries[maxIdx], entries[i]
	}

	// Build results
	results := make([]SearchResult, 0, len(entries))
	for i, entry := range entries {
		if i >= topK {
			break
		}
		// Find the document
		for _, doc := range s.docs {
			if doc.ID == entry.docID {
				results = append(results, SearchResult{
					Document: doc,
					Score:    entry.score,
				})
				break
			}
		}
	}

	return results
}

// GetDocuments returns all indexed documents
func (s *SemanticSearcher) GetDocuments() []Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.docs
}

// GetDocument returns a document by ID
func (s *SemanticSearcher) GetDocument(id string) (Document, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, doc := range s.docs {
		if doc.ID == id {
			return doc, true
		}
	}
	return Document{}, false
}

// Tokenize splits text into lowercase tokens, filtering stop words
func (s *SemanticSearcher) tokenize(text string) []string {
	// Simple whitespace tokenization
	words := strings.Fields(strings.ToLower(text))
	tokens := make([]string, 0, len(words))
	for _, w := range words {
		// Clean punctuation
		cleaned := strings.TrimLeft(w, ".,!?;:\"'()-—")
		cleaned = strings.TrimRight(cleaned, ".,!?;:\"'()-—")
		if cleaned != "" && !s.stopWords[cleaned] {
			tokens = append(tokens, cleaned)
		}
	}
	return tokens
}

// AddStopWord adds a stop word to filter out
func (s *SemanticSearcher) AddStopWord(word string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopWords[word] = true
}

// GetDocCount returns the number of indexed documents
func (s *SemanticSearcher) GetDocCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nDocs
}

// ===== SEMANTIC NAVIGATION =====

// ConceptRelationship represents a relationship between two documents via shared concepts
type ConceptRelationship struct {
	SourceID       string   `json:"source_id"`
	TargetID       string   `json:"target_id"`
	TargetTitle    string   `json:"target_title"`
	TargetSource   string   `json:"target_source"`
	TargetCategory string   `json:"target_category"`
	SharedConcepts []string `json:"shared_concepts"`
	Strength       float64  `json:"strength"`
}

// conceptIndex maps concepts to the documents that contain them
type conceptIndex struct {
	concepts    map[string]map[string]bool // concept -> docID -> true
	conceptList []string                   // ordered list of all concepts
}

// newConceptIndex creates a new concept index
func newConceptIndex() *conceptIndex {
	return &conceptIndex{
		concepts:    make(map[string]map[string]bool),
		conceptList: make([]string, 0),
	}
}

// addDocumentConcepts adds all concepts from a document to the index
func (ci *conceptIndex) addDocumentConcepts(docID string, concepts []string) {
	for _, concept := range concepts {
		if ci.concepts[concept] == nil {
			ci.concepts[concept] = make(map[string]bool)
			ci.conceptList = append(ci.conceptList, concept)
		}
		ci.concepts[concept][docID] = true
	}
}

// getRelatedConcepts returns concepts shared between docID and other documents
func (ci *conceptIndex) getRelatedConcepts(docID string) map[string][]string {
	// Collect all concepts for this document
	docConcepts := make(map[string]bool)
	for _, concept := range ci.conceptList {
		if ci.concepts[concept][docID] {
			docConcepts[concept] = true
		}
	}

	// Map other docs to shared concept count
	relatedDocs := make(map[string]int)
	relatedConcepts := make(map[string][]string)

	for _, concept := range ci.conceptList {
		if docConcepts[concept] {
			for otherDoc := range ci.concepts[concept] {
				if otherDoc != docID {
					relatedDocs[otherDoc]++
					relatedConcepts[otherDoc] = append(relatedConcepts[otherDoc], concept)
				}
			}
		}
	}

	// Sort by strength (number of shared concepts)
	type docCount struct {
		docID string
		count int
	}
	sorted := make([]docCount, 0, len(relatedDocs))
	for docID, count := range relatedDocs {
		sorted = append(sorted, docCount{docID, count})
	}
	for i := 0; i < len(sorted); i++ {
		maxIdx := i
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[maxIdx].count {
				maxIdx = j
			}
		}
		sorted[i], sorted[maxIdx] = sorted[maxIdx], sorted[i]
	}

	// Keep top 10
	if len(sorted) > 10 {
		sorted = sorted[:10]
	}

	// Normalize counts to strength (0-1)
	maxCount := sorted[0].count
	result := make(map[string][]string)
	for _, dc := range sorted {
		result[dc.docID] = relatedConcepts[dc.docID]
	}
	_ = maxCount // used for normalization if needed

	return result
}

// ExtractConcepts extracts key concepts from document content
func ExtractConcepts(content string, title string) []string {
	// Combine title and content for concept extraction
	text := strings.ToLower(title + " " + content)

	// Extract meaningful words (skip very short/long words)
	words := strings.Fields(text)
	conceptSet := make(map[string]bool)

	// Common English stop words to filter
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "shall": true, "can": true, "need": true,
		"must": true, "and": true, "but": true, "nor": true,
		"not": true, "so": true, "yet": true, "both": true, "either": true,
		"neither": true, "each": true, "every": true, "all": true, "any": true,
		"few": true, "more": true, "most": true, "other": true, "some": true,
		"such": true, "only": true, "own": true, "same": true, "that": true,
		"this": true, "these": true, "those": true, "which": true, "who": true,
		"whom": true, "what": true, "where": true, "when": true, "how": true,
		"why": true, "if": true, "then": true, "else": true, "than": true,
		"too": true, "very": true, "just": true, "about": true, "above": true,
		"after": true, "again": true, "against": true, "am": true, "as": true,
		"at": true, "for": true, "from": true, "he": true, "in": true,
		"into": true, "it": true, "its": true, "i": true, "me": true,
		"my": true, "myself": true, "of": true, "on": true, "off": true,
		"once": true, "our": true, "out": true, "over": true, "she": true,
		"through": true, "to": true, "up": true, "we": true, "with": true,
		"down": true, "during": true, "no": true, "us": true,
		"by": true, "further": true,
	}

	// Extract concepts: words longer than 3 chars that aren't stop words
	for _, word := range words {
		cleaned := strings.TrimLeft(word, ".,!?;:\"'()-—_:[]{}()")
		cleaned = strings.TrimRight(cleaned, ".,!?;:\"'()-—_:[]{}()")
		if len(cleaned) > 3 && !stopWords[cleaned] && !conceptSet[cleaned] {
			conceptSet[cleaned] = true
		}
	}

	concepts := make([]string, 0, len(conceptSet))
	for concept := range conceptSet {
		concepts = append(concepts, concept)
	}

	return concepts
}

// GetRelatedDocuments returns documents related to the given document ID via shared concepts
func (s *SemanticSearcher) GetRelatedDocuments(docID string, maxResults int) []ConceptRelationship {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find the source document
	var sourceDoc Document
	found := false
	for _, doc := range s.docs {
		if doc.ID == docID {
			sourceDoc = doc
			found = true
			break
		}
	}
	if !found {
		return []ConceptRelationship{}
	}

	// Extract concepts from source document
	sourceConcepts := ExtractConcepts(sourceDoc.Content, sourceDoc.Title)

	// Build temporary concept index for this query
	ci := newConceptIndex()
	for _, doc := range s.docs {
		if doc.ID != docID {
			concepts := ExtractConcepts(doc.Content, doc.Title)
			ci.addDocumentConcepts(doc.ID, concepts)
		}
	}

	// Get related docs and their shared concepts
	relatedMap := ci.getRelatedConcepts(docID)

	// Build results
	relationships := make([]ConceptRelationship, 0, len(relatedMap))
	for targetID, sharedConcepts := range relatedMap {
		// Find target document
		var targetDoc Document
		for _, doc := range s.docs {
			if doc.ID == targetID {
				targetDoc = doc
				break
			}
		}

		// Calculate strength based on shared concept count and TF-IDF overlap
		strength := float64(len(sharedConcepts)) / float64(len(sourceConcepts))
		if strength > 1.0 {
			strength = 1.0
		}

		relationships = append(relationships, ConceptRelationship{
			SourceID:       docID,
			TargetID:       targetID,
			TargetTitle:    targetDoc.Title,
			TargetSource:   targetDoc.Source,
			TargetCategory: targetDoc.Category,
			SharedConcepts: sharedConcepts,
			Strength:       strength,
		})
	}

	// Sort by strength descending
	for i := 0; i < len(relationships); i++ {
		maxIdx := i
		for j := i + 1; j < len(relationships); j++ {
			if relationships[j].Strength > relationships[maxIdx].Strength {
				maxIdx = j
			}
		}
		relationships[i], relationships[maxIdx] = relationships[maxIdx], relationships[i]
	}

	// Limit results
	if maxResults > 0 && len(relationships) > maxResults {
		relationships = relationships[:maxResults]
	}

	return relationships
}

// GetConceptsForDocument returns the extracted concepts for a document
func (s *SemanticSearcher) GetConceptsForDocument(docID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, doc := range s.docs {
		if doc.ID == docID {
			return ExtractConcepts(doc.Content, doc.Title)
		}
	}
	return []string{}
}
