package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

func newSpotifyToken() (*oauth2.Token, error) {
	auth := spotify.NewAuthenticator("https://thisisasham.localhost",
		spotify.ScopePlaylistReadPrivate,
		spotify.ScopePlaylistModifyPrivate,
		spotify.ScopeUserLibraryRead,
	)
	u := auth.AuthURL("simpleman")
	fmt.Printf("Please visit: %s\n", u)

	fmt.Printf("Code: ")
	var code string
	if _, err := fmt.Scanf("%s", &code); err != nil {
		return nil, err
	}
	fmt.Printf("THis is the code: [%s]\n", code)

	token, err := auth.Exchange(code)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func cacheOrGetNewToken() (*oauth2.Token, error) {
	cdir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	fpath := filepath.Join(cdir, "spotify_token")

	if f, err := os.Open(fpath); err == nil {
		dec := gob.NewDecoder(f)

		var tok oauth2.Token
		if err := dec.Decode(&tok); err != nil {
			f.Close()
			return nil, err
		}
		f.Close()

		if tok.Valid() {
			return &tok, nil
		} else {
			return &tok, nil
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	tok, err := newSpotifyToken()
	if err != nil {
		return nil, err
	}

	fs, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer fs.Close()

	enc := gob.NewEncoder(fs)
	if err := enc.Encode(&tok); err != nil {
		return nil, err
	}

	return tok, nil
}

func getDiscoverWeeklyID(client *spotify.Client) (spotify.ID, error) {
	playlists, err := client.CurrentUsersPlaylists()
	if err != nil {
		return "", err
	}

	for _, playlist := range playlists.Playlists {
		if playlist.Name != "Discover Weekly" {
			continue
		}
		return playlist.ID, nil
	}

	return "", errors.New("Couldn't find the playlist!\n")
}

func getDiscoveryWeeklyArchiveID(client *spotify.Client) (spotify.ID, error) {
	playlists, err := client.CurrentUsersPlaylists()
	if err != nil {
		return "", err
	}

	for _, playlist := range playlists.Playlists {
		if playlist.Name != "Discover Weekly Archive" {
			continue
		}
		return playlist.ID, nil
	}

	// We must try to create
	us, err := client.CurrentUser()
	if err != nil {
		return "", err
	}
	pl, err := client.CreatePlaylistForUser(us.ID, "Discover Weekly Archive", "Archive of the weekly songs", false)
	if err != nil {
		return "", err
	}

	return pl.ID, nil
}

func appendSongs(client *spotify.Client, from, to spotify.ID) error {
	fpl, err := client.GetPlaylist(from)
	if err != nil {
		return err
	}
	tpl, err := client.GetPlaylist(to)
	if err != nil {
		return err
	}

	var toAdd []spotify.ID

	for i := 0; i < len(fpl.Tracks.Tracks); i++ {
		fr := fpl.Tracks.Tracks[i]
		found := false
		for j := 0; j < len(tpl.Tracks.Tracks); j++ {
			if tpl.Tracks.Tracks[j].Track.ID == fr.Track.ID {
				found = true
				break
			}
		}
		if !found {
			toAdd = append(toAdd, fr.Track.ID)
		}
	}
	if len(toAdd) > 0 {
		snapList, err := client.AddTracksToPlaylist(to, toAdd...)
		if err != nil {
			return err
		}

		// We also now need to reorder so the last tracks are first
		if _, err := client.ReorderPlaylistTracks(to, spotify.PlaylistReorderOptions{
			RangeStart:   len(tpl.Tracks.Tracks),
			RangeLength:  len(toAdd),
			InsertBefore: 0,
			SnapshotID: snapList,
		}); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	tok, err := cacheOrGetNewToken()
	if err != nil {
		log.Fatalf("Couldn't get spotify token: %s\n", err.Error())
	}

	auth := spotify.NewAuthenticator("https://thisisasham.localhost",
		spotify.ScopePlaylistReadPrivate,
		spotify.ScopePlaylistModifyPrivate,
		spotify.ScopeUserLibraryRead,
	)

	client := auth.NewClient(tok)

	wid, err := getDiscoverWeeklyID(&client)
	if err != nil {
		log.Fatalf("We couldn't find the playlist id: %s\n", err.Error())
	}
	aid, err := getDiscoveryWeeklyArchiveID(&client)
	if err != nil {
		log.Fatalf("We could not find the archive playlist id: %s\n", err.Error())
	}

	if err := appendSongs(&client, wid, aid); err != nil {
		log.Fatalf("We couldn't append the songs: %s\n", err.Error())
	}
}
