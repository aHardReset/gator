package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"time"

	"github.com/aHardReset/gator/internal/database"
	"github.com/google/uuid"
)

const appName string = "gator"

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", appName)
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("Response have %d status code", res.StatusCode)
	}
	var rssFeed RSSFeed
	xml.NewDecoder(res.Body).Decode(&rssFeed)

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for _, item := range rssFeed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}

	return &rssFeed, nil
}

func handleAgg(s *state, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Println(feed)
	return nil
}

func handleAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("This command needs feed name and url")
	}

	ctx := context.Background()
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        newUUID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	})

	newFeedFollowUUID, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	_, err = s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        newFeedFollowUUID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})

	if err != nil {
		return err
	}

	fmt.Printf("feed '%s' with url '%s' created\n", feed.Name, feed.Url)

	return nil

}

func handleListFeeds(s *state, cmd command) error {

	feeds, err := s.db.ListFeeds(context.Background())
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		fmt.Println("There are no feeds")
		return nil
	}

	users := map[uuid.UUID]string{}

	for _, feed := range feeds {

		if _, ok := users[feed.UserID]; !ok {
			u, err := s.db.GetUserByID(context.Background(), feed.UserID)
			if err != nil {
				return err
			}
			users[feed.UserID] = u.Name
		}
		fmt.Printf("Feed name: %s | Feed url: %s | Created By: %s \n", feed.Name, feed.Url, users[feed.UserID])
	}
	return nil
}

func handleFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("This command needs a feed url")
	}
	ctx := context.Background()
	feed, err := s.db.GetFeedByUrl(ctx, cmd.args[0])
	if err != nil {
		return err
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	feed_follow, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})

	if err != nil {
		return err
	}

	fmt.Printf("user '%s' now follows '%s'", feed_follow.FeedName, feed_follow.UserName)
	return nil
}

func handleUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("URL to unfollow required")
	}
	ctx := context.Background()

	feed, err := s.db.GetFeedByUrl(ctx, cmd.args[0])
	if err != nil {
		return err
	}

	err = s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		FeedID: feed.ID,
		UserID: user.ID,
	})

	if err != nil {
		return err
	}
	fmt.Printf("successfully unfollowed:\n  %s[%s]\n", feed.Name, feed.Url)

	return nil
}

func handleListFollow(s *state, cmd command, user database.User) error {
	ctx := context.Background()
	feeds, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		fmt.Print("You are not following any feed")
		return nil
	}

	fmt.Printf("user '%s' follows:\n", user.Name)
	for _, feed := range feeds {
		fmt.Printf("  - %s\n", feed.FeedName)
	}
	return nil
}
