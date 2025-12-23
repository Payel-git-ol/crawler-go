package crawler

import (
	"Fyne-on/pkg/models"
	"Fyne-on/pkg/scraper"
	"Fyne-on/pkg/storage"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type TelegramJobCrawler struct {
	storage     *storage.StorageService
	htmlScraper *scraper.HTTPScraper
}

func NewTelegramJobCrawler(storage *storage.StorageService) *TelegramJobCrawler {
	return &TelegramJobCrawler{
		storage:     storage,
		htmlScraper: scraper.NewHTTPScraper(30),
	}
}

func (tjc *TelegramJobCrawler) CrawlTelegramChannelJobs(channelUsername string) error {
	log.Printf("[TelegramJobCrawler] 🔍 Starting job crawl for channel: @%s", channelUsername)

	channelUsername = strings.TrimPrefix(channelUsername, "@")
	url := fmt.Sprintf("https://t.me/s/%s", channelUsername)

	log.Printf("[TelegramJobCrawler] 📄 Fetching URL: %s", url)

	doc, err := tjc.htmlScraper.FetchDocument(url)
	if err != nil {
		return fmt.Errorf("failed to fetch Telegram channel %s: %w", channelUsername, err)
	}

	log.Printf("[TelegramJobCrawler] ✅ Successfully fetched page")

	pageTitle := doc.Find("title").Text()
	log.Printf("[TelegramJobCrawler] 📝 Page title: %s", pageTitle)

	var jobs []models.TelegramJob

	jobs = tjc.parseTelegramMessages(doc, channelUsername)

	if len(jobs) == 0 {
		log.Printf("[TelegramJobCrawler] ⚠️ No jobs found with main method, trying alternative...")
		jobs = tjc.parseAlternativeMethod(doc, channelUsername)
	}

	savedCount := 0
	for _, job := range jobs {
		saved, err := tjc.storage.SaveTelegramJob(job)
		if err != nil {
			log.Printf("[TelegramJobCrawler] ❌ Error saving job %s:%d: %v", job.Channel, job.MessageID, err)
		} else if saved {
			savedCount++
			log.Printf("[TelegramJobCrawler] ✅ Saved job #%d: %s (Type: %s, Payment: %s)",
				savedCount, truncateText(job.Title, 40), job.JobType, job.Payment)
		} else {
			log.Printf("[TelegramJobCrawler] ⚡ Job already exists: %s", truncateText(job.Title, 40))
		}
	}

	log.Printf("[TelegramJobCrawler] 🎉 Finished! Found %d jobs, saved %d new", len(jobs), savedCount)

	if len(jobs) == 0 {
		tjc.debugPageInfo(doc)
	}

	return nil
}

func (tjc *TelegramJobCrawler) parseTelegramMessages(doc *goquery.Document, channel string) []models.TelegramJob {
	var jobs []models.TelegramJob
	messageCount := 0

	selectors := []string{
		".tgme_widget_message_wrap",
		".tgme_widget_message",
		"[data-post]",
		".message",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			messageCount++
			job := tjc.parseSingleMessage(s, channel)
			if job != nil {
				jobs = append(jobs, *job)
			}
		})

		if len(jobs) > 0 {
			log.Printf("[TelegramJobCrawler] Found %d jobs using selector '%s'", len(jobs), selector)
			break
		}
	}

	log.Printf("[TelegramJobCrawler] Processed %d total messages", messageCount)
	return jobs
}

func (tjc *TelegramJobCrawler) parseSingleMessage(s *goquery.Selection, channel string) *models.TelegramJob {
	messageID := tjc.extractMessageID(s, channel)
	if messageID == 0 {
		return nil
	}

	text := tjc.extractMessageText(s)
	if text == "" {
		return nil
	}

	if !tjc.isJobPost(text) {
		return nil
	}

	job := &models.TelegramJob{
		ID:        fmt.Sprintf("%s:%d", channel, messageID),
		Channel:   channel,
		MessageID: messageID,
		URL:       fmt.Sprintf("https://t.me/%s/%d", channel, messageID),
		Timestamp: time.Now(),
		RawText:   text,
	}

	tjc.parseAllJobFields(job, text, s)

	return job
}

func (tjc *TelegramJobCrawler) extractMessageID(s *goquery.Selection, channel string) int {
	if dataPost, exists := s.Attr("data-post"); exists {
		parts := strings.Split(dataPost, "/")
		if len(parts) == 2 && parts[0] == channel {
			if msgID, err := strconv.Atoi(parts[1]); err == nil {
				return msgID
			}
		}
	}

	return time.Now().Nanosecond() % 1000000
}

func (tjc *TelegramJobCrawler) extractMessageText(s *goquery.Selection) string {
	textSelectors := []string{
		".tgme_widget_message_text",
		".js-message_text",
		".message_body",
		".text",
		".copyable-text",
	}

	for _, selector := range textSelectors {
		text := s.Find(selector).Text()
		if text != "" {
			return strings.TrimSpace(text)
		}
	}

	return strings.TrimSpace(s.Text())
}

func (tjc *TelegramJobCrawler) isJobPost(text string) bool {
	jobPatterns := []string{
		`💼\s*Тип работы:`,
		`Тип работы:`,
		`👀\s*Задача:`,
		`Задача:`,
		`💰\s*Оплата:`,
		`Оплата:`,
		`ваканси`,
		`требуется`,
		`ищем`,
		`нужен`,
		`поиск`,
		`работа`,
		`фриланс`,
		`удален`,
	}

	textLower := strings.ToLower(text)
	for _, pattern := range jobPatterns {
		if matched, _ := regexp.MatchString(strings.ToLower(pattern), textLower); matched {
			return true
		}
	}

	return false
}

func (tjc *TelegramJobCrawler) parseAllJobFields(job *models.TelegramJob, text string, s *goquery.Selection) {
	lines := strings.Split(text, "\n")
	if len(lines) > 0 {
		job.Title = strings.TrimSpace(lines[0])
		if len(job.Title) > 100 {
			job.Title = job.Title[:100] + "..."
		}
	}

	job.JobType = tjc.extractJobField(text, []string{"💼 Тип работы:", "Тип работы:"})

	job.Task = tjc.extractMultilineField(text, "👀 Задача:", []string{"💰", "⏰", "Оплата:", "Сроки:"})
	if job.Task == "" {
		job.Task = tjc.extractMultilineField(text, "Задача:", []string{"Оплата:", "Сроки:"})
	}

	job.Payment = tjc.extractJobField(text, []string{"💰 Оплата:", "Оплата:"})

	job.Deadline = tjc.extractJobField(text, []string{"⏰ Сроки:", "Сроки:"})
	if job.Deadline == "" {
		job.Deadline = "не указано"
	}

	job.Views = tjc.extractViews(s)

	job.Date = tjc.extractDate(s)

	if job.JobType == "" {
		job.JobType = "не указан"
	}
	if job.Payment == "" {
		job.Payment = "по договорённости"
	}
	if job.Task == "" && len(text) > 0 {
		if len(text) > 500 {
			job.Task = text[:500] + "..."
		} else {
			job.Task = text
		}
	}
}

func (tjc *TelegramJobCrawler) extractJobField(text string, prefixes []string) string {
	for _, prefix := range prefixes {
		if idx := strings.Index(text, prefix); idx != -1 {
			start := idx + len(prefix)

			remaining := text[start:]
			end := len(remaining)

			nextSections := []string{"👀 Задача:", "💰 Оплата:", "⏰ Сроки:", "\n\n", "\n💰", "\n⏰"}
			for _, section := range nextSections {
				if pos := strings.Index(remaining, section); pos != -1 && pos < end {
					end = pos
				}
			}

			if newlinePos := strings.Index(remaining, "\n"); newlinePos != -1 && newlinePos < end {
				end = newlinePos
			}

			field := strings.TrimSpace(remaining[:end])

			field = strings.Trim(field, " \n\t\r")
			field = strings.ReplaceAll(field, "👀", "")
			field = strings.ReplaceAll(field, "💰", "")
			field = strings.ReplaceAll(field, "⏰", "")

			return strings.TrimSpace(field)
		}
	}
	return ""
}

func (tjc *TelegramJobCrawler) extractMultilineField(text, prefix string, nextPrefixes []string) string {
	if idx := strings.Index(text, prefix); idx != -1 {
		start := idx + len(prefix)
		end := len(text)

		for _, nextPrefix := range nextPrefixes {
			if nextIdx := strings.Index(text[start:], nextPrefix); nextIdx != -1 && nextIdx < end {
				end = nextIdx
			}
		}

		result := strings.TrimSpace(text[start : start+end])
		result = strings.ReplaceAll(result, "\n\n", "\n")
		return result
	}
	return ""
}

func (tjc *TelegramJobCrawler) extractViews(s *goquery.Selection) string {
	viewSelectors := []string{
		".tgme_widget_message_views",
		".message_views",
		".views",
		"[title*='view']",
		"[title*='просмотр']",
	}

	for _, selector := range viewSelectors {
		views := s.Find(selector).Text()
		if views != "" {
			return strings.TrimSpace(views)
		}
	}
	return ""
}

// extractDate извлекает дату
func (tjc *TelegramJobCrawler) extractDate(s *goquery.Selection) string {
	dateSelectors := []string{
		"time[datetime]",
		".datetime",
		".message_date",
		".tgme_widget_message_date",
	}

	for _, selector := range dateSelectors {
		elem := s.Find(selector)
		if datetime, exists := elem.Attr("datetime"); exists {
			return datetime
		}
		if dateText := elem.Text(); dateText != "" {
			return dateText
		}
	}
	return time.Now().Format("2006-01-02")
}

func (tjc *TelegramJobCrawler) parseAlternativeMethod(doc *goquery.Document, channel string) []models.TelegramJob {
	var jobs []models.TelegramJob

	doc.Find("body").Each(func(i int, s *goquery.Selection) {
		html, _ := s.Html()

		if strings.Contains(html, "💼 Тип работы:") || strings.Contains(html, "Тип работы:") {
			parts := strings.Split(html, "\n")
			for _, part := range parts {
				if tjc.isJobPost(part) {
					job := &models.TelegramJob{
						ID:        fmt.Sprintf("%s:%d", channel, time.Now().Nanosecond()),
						Channel:   channel,
						MessageID: time.Now().Nanosecond() % 1000000,
						Title:     "Найденная вакансия",
						JobType:   "разовая/постоянная",
						Task:      truncateText(part, 300),
						Payment:   "по договорённости",
						Deadline:  "не указано",
						URL:       fmt.Sprintf("https://t.me/%s", channel),
						Timestamp: time.Now(),
						Views:     "N/A",
						Date:      time.Now().Format("2006-01-02"),
					}
					jobs = append(jobs, *job)
					break
				}
			}
		}
	})

	return jobs
}

func (tjc *TelegramJobCrawler) debugPageInfo(doc *goquery.Document) {
	log.Println("[TelegramJobCrawler] === DEBUG INFO ===")

	checks := []string{
		"body",
		".tgme_widget_message_wrap",
		".tgme_widget_message",
		".tgme_widget_message_text",
		"[data-post]",
	}

	for _, selector := range checks {
		count := doc.Find(selector).Length()
		log.Printf("[TelegramJobCrawler] Selector '%s': %d elements", selector, count)

		if count > 0 && (selector == ".tgme_widget_message_text" || selector == "[data-post]") {
			elem := doc.Find(selector).First()
			text := elem.Text()
			if len(text) > 0 {
				log.Printf("[TelegramJobCrawler] First element text (100 chars): %.100s", text)
			}
		}
	}

	html, _ := doc.Html()
	if strings.Contains(html, "💼 Тип работы:") {
		log.Println("[TelegramJobCrawler] ✅ Found '💼 Тип работы:' in HTML!")
	}
	if strings.Contains(html, "Тип работы:") {
		log.Println("[TelegramJobCrawler] ✅ Found 'Тип работы:' in HTML!")
	}
	if strings.Contains(html, "👀 Задача:") {
		log.Println("[TelegramJobCrawler] ✅ Found '👀 Задача:' in HTML!")
	}

	log.Println("[TelegramJobCrawler] === END DEBUG ===")
}

func truncateText(text string, length int) string {
	text = strings.TrimSpace(text)
	if len(text) > length {
		return text[:length] + "..."
	}
	return text
}

func (tjc *TelegramJobCrawler) CrawlTelegramMessage(channel string, messageID int) error {
	log.Printf("[TelegramJobCrawler] 📨 Parsing specific message: @%s/%d", channel, messageID)

	url := fmt.Sprintf("https://t.me/%s/%d", channel, messageID)

	doc, err := tjc.htmlScraper.FetchDocument(url)
	if err != nil {
		return fmt.Errorf("failed to fetch Telegram message %s: %w", url, err)
	}

	// Ищем сообщение на странице
	var job *models.TelegramJob

	selectors := []string{
		".tgme_widget_message_wrap",
		".tgme_widget_message",
		"[data-post]",
		".message",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			currentID := tjc.extractMessageID(s, channel)
			if currentID == messageID {
				text := tjc.extractMessageText(s)
				if text != "" && tjc.isJobPost(text) {
					job = &models.TelegramJob{
						ID:        fmt.Sprintf("%s:%d", channel, messageID),
						Channel:   channel,
						MessageID: messageID,
						URL:       url,
						Timestamp: time.Now(),
						RawText:   text,
					}
					tjc.parseAllJobFields(job, text, s)
				}
			}
		})

		if job != nil {
			break
		}
	}

	if job == nil {
		text := doc.Find(".tgme_widget_message_text").Text()
		if text == "" {
			text = doc.Find(".message_body").Text()
		}

		if text != "" && tjc.isJobPost(text) {
			job = &models.TelegramJob{
				ID:        fmt.Sprintf("%s:%d", channel, messageID),
				Channel:   channel,
				MessageID: messageID,
				URL:       url,
				Timestamp: time.Now(),
				RawText:   text,
			}
			tjc.parseAllJobFields(job, text, doc.Selection)
		}
	}

	if job != nil {
		saved, err := tjc.storage.SaveTelegramJob(*job)
		if err != nil {
			log.Printf("[TelegramJobCrawler] ❌ Error saving job: %v", err)
			return err
		}

		if saved {
			log.Printf("[TelegramJobCrawler] ✅ Saved specific message: %s", job.Title)
		} else {
			log.Printf("[TelegramJobCrawler] ⚡ Message already exists: %s", job.Title)
		}
	} else {
		log.Printf("[TelegramJobCrawler] ⚠️ No job found in message @%s/%d", channel, messageID)
	}

	return nil
}
