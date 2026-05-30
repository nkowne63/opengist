package v1_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thomiceli/opengist/internal/db"
	v1 "github.com/thomiceli/opengist/internal/web/handlers/api/v1"
	webtest "github.com/thomiceli/opengist/internal/web/test"
)

func TestCreateAndListComments(t *testing.T) {
	s, tok := setupAPIUser(t)
	_, gist, _, _ := s.CreateGistAs(t, "thomas", "0")

	req := v1.CreateCommentRequest{Content: "hello **world**"}
	var created v1.Comment
	s.APIRequestUnmarshal(t, "POST", "/api/v1/gists/"+gist.Uuid+"/comments", tok, req, &created, 201)
	require.Equal(t, "hello **world**", created.Content)
	require.Equal(t, "thomas", created.User.Username)
	require.Contains(t, created.HTML, "<strong>world</strong>")

	var comments []v1.Comment
	s.APIRequestUnmarshal(t, "GET", "/api/v1/gists/"+gist.Uuid+"/comments", tok, nil, &comments, 200)
	require.Len(t, comments, 1)
	require.Equal(t, created.Content, comments[0].Content)
	require.Equal(t, created.User.Username, comments[0].User.Username)
}

func TestCreateComment_NoWriteScope(t *testing.T) {
	s := webtest.Setup(t)
	defer webtest.Teardown(t)
	s.Register(t, "admin")
	s.Logout()
	s.Register(t, "thomas")
	s.Login(t, "thomas")
	tok := s.CreateAccessToken(t, "ro", db.ReadPermission, db.NoPermission)
	require.NoError(t, db.UpdateSetting(db.SettingApiEnabled, "1"))
	_, gist, _, _ := s.CreateGistAs(t, "thomas", "0")

	var body map[string]string
	s.APIRequestUnmarshal(t, "POST", "/api/v1/gists/"+gist.Uuid+"/comments", tok, v1.CreateCommentRequest{Content: "test"}, &body, 403)
	require.Equal(t, "forbidden", body["code"])
}

func TestUpdateAndDeleteComment(t *testing.T) {
	s, tok := setupAPIUser(t)
	_, gist, _, _ := s.CreateGistAs(t, "thomas", "0")

	req := v1.CreateCommentRequest{Content: "hello **world**"}
	var created v1.Comment
	s.APIRequestUnmarshal(t, "POST", "/api/v1/gists/"+gist.Uuid+"/comments", tok, req, &created, 201)
	require.Equal(t, "hello **world**", created.Content)

	var updated v1.Comment
	s.APIRequestUnmarshal(t, "PATCH", fmt.Sprintf("/api/v1/gists/%s/comments/%d", gist.Uuid, created.ID), tok, v1.UpdateCommentRequest{Content: "edited **world**"}, &updated, 200)
	require.Equal(t, created.ID, updated.ID)
	require.Equal(t, "edited **world**", updated.Content)
	require.Contains(t, updated.HTML, "<strong>world</strong>")

	s.APIRequest(t, "DELETE", fmt.Sprintf("/api/v1/gists/%s/comments/%d", gist.Uuid, created.ID), tok, nil, 204)

	var comments []v1.Comment
	s.APIRequestUnmarshal(t, "GET", "/api/v1/gists/"+gist.Uuid+"/comments", tok, nil, &comments, 200)
	require.Empty(t, comments)
}

func TestUpdateComment_NoWriteScope(t *testing.T) {
	s, tok := setupAPIUser(t)
	_, gist, _, _ := s.CreateGistAs(t, "thomas", "0")

	var created v1.Comment
	s.APIRequestUnmarshal(t, "POST", "/api/v1/gists/"+gist.Uuid+"/comments", tok, v1.CreateCommentRequest{Content: "hello"}, &created, 201)

	ro := s.CreateAccessToken(t, "ro", db.ReadPermission, db.NoPermission)
	require.NoError(t, db.UpdateSetting(db.SettingApiEnabled, "1"))

	var body map[string]string
	s.APIRequestUnmarshal(t, "PATCH", fmt.Sprintf("/api/v1/gists/%s/comments/%d", gist.Uuid, created.ID), ro, v1.UpdateCommentRequest{Content: "edited"}, &body, 403)
	require.Equal(t, "forbidden", body["code"])

	body = map[string]string{}
	s.APIRequestUnmarshal(t, "DELETE", fmt.Sprintf("/api/v1/gists/%s/comments/%d", gist.Uuid, created.ID), ro, nil, &body, 403)
	require.Equal(t, "forbidden", body["code"])
}
