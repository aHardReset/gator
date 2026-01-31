package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
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
	if len(cmd.args) < 1 {
		return fmt.Errorf("A string that represents a duration is needed")
	}
	ctx := context.Background()

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %s\n\n", timeBetweenRequests.String())
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s, ctx); err != nil {
			break
		}
	}
	return err
}

func tryParse(date string, now time.Time) time.Time {
	if detected, err := time.Parse(time.Layout, date); err != nil {
		return detected
	}

	if detected, err := time.Parse(time.ANSIC, date); err != nil {
		return detected
	}

	if detected, err := time.Parse(time.UnixDate, date); err != nil {
		return detected
	}

	if detected, err := time.Parse(time.RubyDate, date); err != nil {
		return detected
	}

	if detected, err := time.Parse(time.RFC822, date); err != nil {
		return detected
	}

	if detected, err := time.Parse(time.RFC822Z, date); err != nil {
		return detected
	}

	if detected, err := time.Parse(time.RFC850, date); err != nil {
		return detected
	}

	if detected, err := time.Parse(time.RFC1123, date); err != nil {
		return detected
	}

	if detected, err := time.Parse(time.RFC1123Z, date); err != nil {
		return detected
	}

	if detected, err := time.Parse(time.RFC3339, date); err != nil {
		return detected
	}

	if detected, err := time.Parse(time.RFC3339Nano, date); err != nil {
		return detected
	}

	return now
}

func scrapeFeeds(s *state, ctx context.Context) error {
	storedFeed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	s.db.MarkFeedFetched(ctx, database.MarkFeedFetchedParams{
		ID:            storedFeed.ID,
		LastFetchedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt:     now,
	})
	feed, err := fetchFeed(context.Background(), storedFeed.Url)
	fmt.Printf("Fetching Feed '%s': %s\n", storedFeed.Name, feed.Channel.Title)
	if len(feed.Channel.Item) == 0 {
		fmt.Println("  With No feeds")
		return nil
	}
	for _, item := range feed.Channel.Item {
		newUUID, err := uuid.NewUUID()
		if err != nil {
			return err
		}
		now := time.Now()
		_, err = s.db.CreatePost(ctx, database.CreatePostParams{
			ID:          newUUID,
			CreatedAt:   now,
			UpdatedAt:   now,
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description},
			PublishedAt: tryParse(item.PubDate, now),
			FeedID:      storedFeed.ID,
		})

		if err != nil {
			if !strings.Contains(err.Error(), "duplicate key value") {
				fmt.Println(err)
			}
		}
	}
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

func handleBrowse(s *state, cmd command, user database.User) error {
	limit := int32(2)
	if len(cmd.args) > 0 {
		userLimit, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return err
		}
		limit = int32(userLimit)

	}

	ctx := context.Background()
	posts, err := s.db.GetPostsByUser(ctx, database.GetPostsByUserParams{
		Limit:  limit,
		UserID: user.ID,
	})
	if err != nil {
		return err
	}

	feedsFromPosts := map[uuid.UUID]database.Feed{}
	fmt.Println("===== Gator Post ======")
	for _, post := range posts {
		if _, ok := feedsFromPosts[post.FeedID]; !ok {
			feedFromDb, err := s.db.GetFeedByID(ctx, post.FeedID)
			if err != nil {
				return err
			}
			feedsFromPosts[post.FeedID] = feedFromDb
		}

		feed := feedsFromPosts[post.FeedID]

		fmt.Printf("'%s' Published '%s' at '%s'\n", feed.Name, post.Title, post.PublishedAt)
		desc := post.Description
		if desc.Valid {
			fmt.Printf("  - %s\n", desc.String)
		}
		fmt.Printf("  - See More: %s\n", post.Url)
		fmt.Println("-----------------------")

	}
	fmt.Println("=====+++++++++++++=====")

	return nil
}
