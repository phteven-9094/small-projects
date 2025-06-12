package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify"
)

var (
	clientID        string
	clientSecret    string
	masterPlaylist  string
	archivePlaylist string
	auth            spotify.Authenticator
	client          spotify.Client
)

func init() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	clientID = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
	masterPlaylist = os.Getenv("MASTER_PLAYLIST")
	archivePlaylist = os.Getenv("ARCHIVE_PLAYLIST")

	// Set up Spotify authenticator
	auth = spotify.NewAuthenticator("http://localhost:8080/callback", spotify.ScopePlaylistModifyPublic, spotify.ScopePlaylistReadPrivate)
	auth.SetAuthInfo(clientID, clientSecret)
}

func authenticate() error {
	// Authenticate with Spotify
	ch := make(chan *spotify.Client)
	go func() {
		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			token, err := auth.Token(r.URL.Query().Get("state"), r)
			if err != nil {
				http.Error(w, "Couldn't get token", http.StatusForbidden)
				log.Println("Error getting token:", err)
				return
			}
			client = auth.NewClient(token)
			ch <- &client
		})
		log.Println("Please log in to Spotify by visiting the following page in your browser:", auth.AuthURL("state"))
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Println("Error starting server:", err)
		}
	}()
	client = *<-ch
	return nil
}

func fetchPlaylistTracks(playlistID spotify.ID) ([]spotify.ID, error) {
	// Fetch track URIs from a given playlist
	playlist, err := client.GetPlaylistTracks(playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch playlist tracks: %w", err)
	}

	var trackURIs []spotify.ID
	for _, item := range playlist.Tracks {
		trackURIs = append(trackURIs, item.Track.ID)
	}
	return trackURIs, nil
}

func combineTracks(masterURIs, archiveURIs []spotify.ID) []spotify.ID {
	// Combine tracks from master and archive playlists, excluding duplicates
	archiveSet := make(map[spotify.ID]bool)
	for _, uri := range archiveURIs {
		archiveSet[uri] = true
	}

	var combinedURIs []spotify.ID
	for _, uri := range masterURIs {
		if !archiveSet[uri] {
			combinedURIs = append(combinedURIs, uri)
		}
	}
	return combinedURIs
}

func addTracksToPlaylist(playlistID spotify.ID, trackURIs []spotify.ID) error {
	// Add tracks to a playlist
	_, err := client.AddTracksToPlaylist(playlistID, trackURIs...)
	if err != nil {
		return fmt.Errorf("failed to add tracks to playlist: %w", err)
	}
	return nil
}

func removeTracksFromPlaylist(playlistID spotify.ID, trackURIs []spotify.ID) error {
	// Remove tracks from a playlist
	_, err := client.RemoveTracksFromPlaylist(playlistID, trackURIs...)
	if err != nil {
		return fmt.Errorf("failed to remove tracks from playlist: %w", err)
	}
	return nil
}

func main() {
	if err := authenticate(); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("Fetching tracks from the master playlist...")
	masterURIs, err := fetchPlaylistTracks(spotify.ID(masterPlaylist))
	if err != nil {
		log.Fatalf("Error fetching master playlist tracks: %v", err)
	}

	fmt.Println("Fetching tracks from the archive playlist...")
	archiveURIs, err := fetchPlaylistTracks(spotify.ID(archivePlaylist))
	if err != nil {
		log.Fatalf("Error fetching archive playlist tracks: %v", err)
	}

	combinedURIs := combineTracks(masterURIs, archiveURIs)

	if len(combinedURIs) > 0 {
		fmt.Println("Adding tracks to the archive playlist...")
		if err := addTracksToPlaylist(spotify.ID(archivePlaylist), combinedURIs); err != nil {
			log.Fatalf("Error adding tracks to archive playlist: %v", err)
		}

		fmt.Println("Removing tracks from the master playlist...")
		if err := removeTracksFromPlaylist(spotify.ID(masterPlaylist), masterURIs); err != nil {
			log.Fatalf("Error removing tracks from master playlist: %v", err)
		}

		fmt.Println("Tracks have been successfully moved.")
	} else {
		fmt.Println("No new tracks to move.")
	}
}
