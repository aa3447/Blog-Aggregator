package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
)

// Rss was generated 2025-08-25 18:20:38 by https://xml-to-go.github.io/ in Ukraine.
type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Atom    string   `xml:"atom,attr"`
	Channel struct {
		Title string `xml:"title"`
		Link  struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
		} `xml:"link"`
		Description   string `xml:"description"`
		Generator     string `xml:"generator"`
		Language      string `xml:"language"`
		LastBuildDate string `xml:"lastBuildDate"`
		Item          []FeedItem `xml:"item"`
	} `xml:"channel"`
} 

type FeedItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Guid        string `xml:"guid"`
	Description string `xml:"description"`
}

func (rf *RSSFeed) String() string{
	return fmt.Sprintf("Title: %s, Link: %s, Description: %s, Items: %d", rf.Channel.Title, rf.Channel.Link.Href, rf.Channel.Description, len(rf.Channel.Item))
}

func (fi *FeedItem) String() string{
	return fmt.Sprintf("Title: %s, Link: %s, PubDate: %s, Guid: %s, Description: %s", fi.Title, fi.Link, fi.PubDate, fi.Guid, fi.Description)
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error){
	request, err := http.NewRequestWithContext(ctx,http.MethodGet ,feedURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	response.Header.Set("User-Agent", "gator")
	
	var rss RSSFeed
	err = xml.NewDecoder(response.Body).Decode(&rss)
	if err != nil {
		return nil, err
	}

	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
	for i := range rss.Channel.Item {
		rss.Channel.Item[i].Title = html.UnescapeString(rss.Channel.Item[i].Title)
		rss.Channel.Item[i].Description = html.UnescapeString(rss.Channel.Item[i].Description)
	}	

	return &rss, nil
}