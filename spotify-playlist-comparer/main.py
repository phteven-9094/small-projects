import spotipy
from spotipy.oauth2 import SpotifyOAuth
import os
from dotenv import load_dotenv
from typing import List, Dict

# Load environment variables from .env file
load_dotenv()

# Set up Spotify API credentials
CLIENT_ID: str = os.getenv("CLIENT_ID")
CLIENT_SECRET: str = os.getenv("CLIENT_SECRET")
MASTER_PLAYLIST: str = os.getenv("MASTER_PLAYLIST")
ARCHIVE_PLAYLIST: str = os.getenv("ARCHIVE_PLAYLIST")

SCOPE: str = "user-library-read playlist-modify-public"
REDIRECT_URI: str = "http://localhost/"

# Authenticate with Spotify
sp: spotipy.Spotify = spotipy.Spotify(
    auth_manager=SpotifyOAuth(
        scope=SCOPE,
        redirect_uri=REDIRECT_URI,
        client_id=CLIENT_ID,
        client_secret=CLIENT_SECRET,
        show_dialog=True,
        cache_path='token.txt')
)
user_id: str = sp.current_user()['id']


def fetch_playlist_tracks(playlist_id: str) -> List[str]:
    """
    Fetches track URIs from a given playlist.

    Args:
        playlist_id (str): The Spotify playlist ID.

    Returns:
        List[str]: List of track URIs from the playlist.
    """
    tracks: Dict = sp.playlist_tracks(
        playlist_id=playlist_id,
        fields=None,
        limit=100,
        offset=0
    )
    return [item['track']['uri'] for item in tracks['items']]


def combine_tracks(master_uris: List[str], archive_uris: List[str]) -> List[str]:
    """
    Combines tracks from the master playlist and archive playlist, excluding duplicates.

    Args:
        master_uris (List[str]): List of track URIs from the master playlist.
        archive_uris (List[str]): List of track URIs from the archive playlist.

    Returns:
        List[str]: List of unique track URIs to be added to the archive playlist.
    """
    return [uri for uri in master_uris if uri not in archive_uris]


def add_tracks_to_playlist(playlist_id: str, track_uris: List[str]) -> None:
    """
    Adds tracks to a playlist.

    Args:
        playlist_id (str): The Spotify playlist ID.
        track_uris (List[str]): List of track URIs to be added.
    """
    sp.playlist_add_items(playlist_id=playlist_id, items=track_uris)


def remove_tracks_from_playlist(playlist_id: str, track_uris: List[str]) -> None:
    """
    Removes tracks from a playlist.

    Args:
        playlist_id (str): The Spotify playlist ID.
        track_uris (List[str]): List of track URIs to be removed.
    """
    sp.playlist_remove_all_occurrences_of_items(playlist_id=playlist_id, items=track_uris)


def main() -> None:
    """
    Main function to move tracks from the master playlist to the archive playlist.
    """
    print("Fetching tracks from the master playlist...")
    master_uris: List[str] = fetch_playlist_tracks(MASTER_PLAYLIST)

    print("Fetching tracks from the archive playlist...")
    archive_uris: List[str] = fetch_playlist_tracks(ARCHIVE_PLAYLIST)

    combined_uris: List[str] = combine_tracks(master_uris, archive_uris)

    if combined_uris:
        print("Adding tracks to the archive playlist...")
        add_tracks_to_playlist(ARCHIVE_PLAYLIST, combined_uris)

        print("Removing tracks from the master playlist...")
        remove_tracks_from_playlist(MASTER_PLAYLIST, master_uris)

        print("Tracks have been successfully moved.")
    else:
        print("No new tracks to move.")


if __name__ == "__main__":
    main()

