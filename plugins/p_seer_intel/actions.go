package p_seer_intel

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/UniquityVentures/lamu/plugins/p_google_genai"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

const intelSummarySystemPrompt = `You write concise factual summaries for an intelligence ingest pipeline.
Given raw source content, respond with a short plain-text summary only (no markdown headings, no preamble).
Aim for 2–6 sentences. If the content is empty or unusable, reply with a single sentence stating that.`

const intelTitleSystemPrompt = `You label rows in an intelligence ingest pipeline.
Given raw source content, respond with one short plain-text title only: no markdown, no preamble, no quotation marks.
At most 12 words. Describe the subject (what it is about), not the medium (avoid "post", "article", "document" unless necessary).`

const intelTitleMaxRunes = 200

// normalizeIntelTitle cleans model output to a single-line DB title.
func normalizeIntelTitle(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	s = strings.Trim(s, `"'`)
	if utf8.RuneCountInString(s) <= intelTitleMaxRunes {
		return s
	}
	runes := []rune(s)
	return strings.TrimSpace(string(runes[:intelTitleMaxRunes]))
}

func intelTitleFallback(content string) string {
	for line := range strings.SplitSeq(content, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "---") {
			continue
		}
		t = normalizeIntelTitle(t)
		if t != "" {
			return t
		}
	}
	return "Intel"
}

func runBatchGenerateJob(ctx context.Context, client *genai.Client, model string, prompts []string, systemInstruction string) ([]string, error) {
	if len(prompts) == 0 {
		return nil, nil
	}

	inlinedReqs := make([]*genai.InlinedRequest, len(prompts))
	for i, prompt := range prompts {
		inlinedReqs[i] = &genai.InlinedRequest{
			Model: model,
			Contents: []*genai.Content{
				genai.NewContentFromText(prompt, genai.RoleUser),
			},
			Config: &genai.GenerateContentConfig{
				SystemInstruction: genai.NewContentFromText(systemInstruction, genai.RoleUser),
			},
		}
	}

	src := &genai.BatchJobSource{
		InlinedRequests: inlinedReqs,
	}

	job, err := client.Batches.Create(ctx, model, src, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch job: %w", err)
	}

	completedStates := map[genai.JobState]bool{
		genai.JobStateSucceeded: true,
		genai.JobStateFailed:    true,
		genai.JobStateCancelled: true,
	}

	for {
		time.Sleep(10 * time.Second)
		j, err := client.Batches.Get(ctx, job.Name, nil)
		if err != nil {
			return nil, err
		}
		if completedStates[j.State] {
			if j.State != genai.JobStateSucceeded {
				return nil, fmt.Errorf("batch prediction job failed with state: %s", j.State)
			}
			job = j
			break
		}
	}

	orderedResults := make([]string, len(prompts))
	for i := range inlinedReqs {
		if job.Dest != nil && i < len(job.Dest.InlinedResponses) && job.Dest.InlinedResponses[i].Response != nil {
			orderedResults[i] = job.Dest.InlinedResponses[i].Response.Text()
		}
	}
	return orderedResults, nil
}

// NewFromIntelKind builds a slice of [Intel] from ks using text + embeddings from [p_google_genai] concurrently.
// [Intel.Kind] / [Intel.KindID] are set from each k.
func NewFromIntelKind(ctx context.Context, ks []IntelKind) ([]Intel, error) {
	if len(ks) == 0 {
		return nil, nil
	}

	client, err := p_google_genai.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	prompts := make([]string, len(ks))
	for i, k := range ks {
		prompts[i] = k.Content()
	}

	titles, err := runBatchGenerateJob(ctx, client, IntelConfigValue.TitleModel, prompts, intelTitleSystemPrompt)
	if err != nil {
		return nil, fmt.Errorf("title batch job failed: %w", err)
	}

	summaries, err := runBatchGenerateJob(ctx, client, IntelConfigValue.SummaryModel, prompts, intelSummarySystemPrompt)
	if err != nil {
		return nil, fmt.Errorf("summary batch job failed: %w", err)
	}

	out := make([]Intel, len(ks))
	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex

	for i, k := range ks {
		wg.Add(1)
		go func(idx int, item IntelKind) {
			defer wg.Done()
			content := item.Content()
			valuesResp, err := client.Models.EmbedContent(ctx,
				IntelConfigValue.EmbeddingModel,
				[]*genai.Content{genai.NewContentFromText(content, genai.RoleUser)},
				nil,
			)
			if err != nil {
				slog.Error("p_seer_intel: embed content", "error", err)
				errMu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("p_seer_intel: embed content: %w", err)
				}
				errMu.Unlock()
				return
			}
			values := valuesResp.Embeddings[0].Values
			if len(values) != SeerIntelEmbeddingDim {
				errMu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("p_seer_intel: embed dimension %d, want %d", len(values), SeerIntelEmbeddingDim)
				}
				errMu.Unlock()
				return
			}
			vec := pgvector.NewVector(values)
			title := normalizeIntelTitle(titles[idx])
			if title == "" {
				title = intelTitleFallback(content)
			}
			out[idx] = Intel{
				Title:     title,
				Kind:      strings.TrimSpace(item.Kind()),
				KindID:    item.IntelID(),
				Summary:   strings.TrimSpace(summaries[idx]),
				Datetime:  time.Now().UTC(),
				Embedding: &vec,
			}
		}(i, k)
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return out, nil
}

type IngestRequest struct {
	Kind IntelKind
}

var IntelChannel = make(chan IngestRequest, 1024)

var intelDB *gorm.DB

func StartIntelIngestWorker(db *gorm.DB) {
	intelDB = db

	internalChan := make(chan IngestRequest)

	go func() {
		var queue []IngestRequest
		for {
			if len(queue) == 0 {
				req, ok := <-IntelChannel
				if !ok {
					close(internalChan)
					return
				}
				queue = append(queue, req)
			}

			select {
			case req, ok := <-IntelChannel:
				if !ok {
					for _, r := range queue {
						internalChan <- r
					}
					close(internalChan)
					return
				}
				queue = append(queue, req)
			case internalChan <- queue[0]:
				queue = queue[1:]
			}
		}
	}()

	go func() {
		ctx := context.Background()
		slog.Info("p_seer_intel: async ingest worker started")

		var batch []IngestRequest
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		processBatch := func() {
			if len(batch) == 0 {
				return
			}
			slog.Info("p_seer_intel: processing batch", "size", len(batch))

			// Separate batch into IntelKinds
			var kinds []IntelKind
			var eventKinds []IntelEventKind
			var hasEventKinds bool
			for _, req := range batch {
				if req.Kind != nil {
					kinds = append(kinds, req.Kind)
					ek, ok := req.Kind.(IntelEventKind)
					if ok {
						eventKinds = append(eventKinds, ek)
						hasEventKinds = true
					} else {
						eventKinds = append(eventKinds, nil)
					}
				}
			}

			if len(kinds) == 0 {
				batch = nil
				return
			}

			// Filter out already existing intels before calling NewFromIntelKind
			var filteredKinds []IntelKind
			var filteredEventKinds []IntelEventKind
			for i, k := range kinds {
				kindLabel := k.Kind()
				id := k.IntelID()
				exists, err := IntelExistsForSource(ctx, intelDB, kindLabel, id)
				if err != nil {
					slog.Error("p_seer_intel: worker exists check error", "kind", kindLabel, "id", id, "error", err)
					continue
				}
				if exists {
					continue
				}
				filteredKinds = append(filteredKinds, k)
				filteredEventKinds = append(filteredEventKinds, eventKinds[i])
			}

			if len(filteredKinds) == 0 {
				batch = nil
				return
			}

			intels, err := NewFromIntelKind(ctx, filteredKinds)
			if err != nil {
				slog.Error("p_seer_intel: batch generate error", "error", err)
				batch = nil
				return
			}

			if hasEventKinds {
				var withEvents []Intel
				var withEventsKinds []IntelEventKind
				var withoutEvents []Intel

				for i, intel := range intels {
					if filteredEventKinds[i] != nil {
						withEvents = append(withEvents, intel)
						withEventsKinds = append(withEventsKinds, filteredEventKinds[i])
					} else {
						withoutEvents = append(withoutEvents, intel)
					}
				}

				if len(withEvents) > 0 {
					if err := CreateIntelWithEvent(ctx, intelDB, withEvents, withEventsKinds); err != nil {
						slog.Error("p_seer_intel: batch persist with events error", "error", err)
					}
				}
				if len(withoutEvents) > 0 {
					if err := CreateIntelAndEvent(ctx, intelDB, withoutEvents); err != nil {
						slog.Error("p_seer_intel: batch persist without events error", "error", err)
					}
				}
			} else {
				if err := CreateIntelAndEvent(ctx, intelDB, intels); err != nil {
					slog.Error("p_seer_intel: batch persist error", "error", err)
				}
			}

			batch = nil
		}

		for {
			select {
			case req, ok := <-internalChan:
				if !ok {
					processBatch()
					return
				}
				batch = append(batch, req)
				if len(batch) >= 60 {
					processBatch()
					ticker.Reset(1 * time.Minute)
				}
			case <-ticker.C:
				processBatch()
			}
		}
	}()
}
