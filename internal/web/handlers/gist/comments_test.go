package gist_test

import (
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thomiceli/opengist/internal/db"
	webtest "github.com/thomiceli/opengist/internal/web/test"
)

func TestGistCommentsWebUI(t *testing.T) {
	s := webtest.Setup(t)
	defer webtest.Teardown(t)
	s.Register(t, "admin")
	s.Logout()
	s.Register(t, "thomas")
	s.Register(t, "alice")

	s.Login(t, "thomas")
	_, gist, username, identifier := s.CreateGistAs(t, "thomas", "0")
	require.Equal(t, "thomas", username)

	s.Login(t, "alice")
	resp := s.Request(t, "GET", "/"+username+"/"+identifier, nil, 200)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), `name="content"`)
	require.Contains(t, string(body), "Post comment")

	resp = s.Request(t, "POST", "/"+username+"/"+identifier+"/comments", url.Values{"content": {"Nice post!"}}, 302)
	require.Contains(t, resp.Header.Get("Location"), "#comments")

	comments, err := db.GetGistComments(gist.ID)
	require.NoError(t, err)
	require.Len(t, comments, 1)
	require.Equal(t, "Nice post!", comments[0].Content)
	require.Equal(t, "alice", comments[0].User.Username)

	s.Logout()
	resp = s.Request(t, "GET", "/"+username+"/"+identifier, nil, 200)
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Nice post!")
	require.Contains(t, string(body), "alice")
}

func TestGistCommentsRejectAnonymousPost(t *testing.T) {
	s := webtest.Setup(t)
	defer webtest.Teardown(t)
	s.Register(t, "admin")
	s.Logout()
	s.Register(t, "thomas")
	_, _, username, identifier := s.CreateGistAs(t, "thomas", "0")
	s.Logout()

	resp := s.Request(t, "POST", "/"+username+"/"+identifier+"/comments", url.Values{"content": {"Hi"}}, 302)
	require.Contains(t, resp.Header.Get("Location"), "/all")
	require.False(t, strings.Contains(resp.Header.Get("Location"), "#comments"))
}
