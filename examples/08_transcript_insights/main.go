// Fetch a hosted video's plain-text transcript and print its insights.
//
// Every auto-generated caption track (subtitle operation "generate") also
// produces transcript renditions in json, txt and md. On Transcodely-managed
// storage each rendition carries a short-lived signed_url, re-issued on every
// single-resource read, so fetch the video right before downloading.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/transcodely/transcodely-go"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: 08_transcript_insights <video_id>")
	}
	ctx := context.Background()

	client, err := transcodely.New(os.Getenv("TRANSCODELY_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	video, err := client.Videos.Get(ctx, os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Transcript: the txt rendition of the first generated track.
	for _, track := range video.GetTextTracks() {
		for _, r := range track.GetTranscriptRenditions() {
			if r.GetFormat() != transcodely.TranscriptFormatTXT || r.GetSignedUrl() == "" {
				continue
			}
			text, err := download(ctx, r.GetSignedUrl())
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("--- transcript (%s) ---\n%s\n", track.GetLanguage(), text)
			break
		}
	}

	// Insights: present when the job enabled them on its generate track.
	for _, ins := range video.GetInsights() {
		switch ins.GetStatus() {
		case transcodely.InsightsStatusPending:
			fmt.Printf("insights (%s): still running\n", ins.GetLanguage())
		case transcodely.InsightsStatusFailed:
			fmt.Printf("insights (%s): failed: %s\n", ins.GetLanguage(), ins.GetErrorCode())
		case transcodely.InsightsStatusReady:
			fmt.Printf("--- insights (%s) ---\n%s\n", ins.GetLanguage(), ins.GetSummary())
			for _, p := range ins.GetKeyPoints() {
				fmt.Printf("  * %s\n", p)
			}
			for _, ch := range ins.GetChapters() {
				fmt.Printf("  %6.1fs  %s\n", ch.GetStartSeconds(), ch.GetTitle())
			}
		}
	}
}

func download(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download transcript: HTTP %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}
