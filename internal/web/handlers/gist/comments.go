package gist

import (
	"html/template"
	"strconv"
	"strings"

	"github.com/thomiceli/opengist/internal/db"
	"github.com/thomiceli/opengist/internal/render"
	"github.com/thomiceli/opengist/internal/web/context"
)

type CommentView struct {
	*db.GistComment
	HTML    template.HTML
	CanEdit bool
}

func loadCommentViews(gist *db.Gist, user *db.User) ([]CommentView, error) {
	comments, err := db.GetGistComments(gist.ID)
	if err != nil {
		return nil, err
	}

	views := make([]CommentView, 0, len(comments))
	for _, comment := range comments {
		html, err := render.MarkdownString(comment.Content)
		if err != nil {
			return nil, err
		}
		views = append(views, CommentView{
			GistComment: comment,
			HTML:        template.HTML(html),
			CanEdit:     user != nil && (comment.UserID == user.ID || gist.UserID == user.ID),
		})
	}

	return views, nil
}

func lookupGistComment(gist *db.Gist, commentID string) (*db.GistComment, error) {
	id, err := strconv.ParseUint(commentID, 10, 64)
	if err != nil {
		return nil, nil
	}

	comment, err := db.GetGistCommentByID(uint(id))
	if err != nil || comment.GistID != gist.ID {
		return nil, nil
	}
	return comment, nil
}

func AddComment(ctx *context.Context) error {
	gist := ctx.GetData("gist").(*db.Gist)
	content := strings.TrimSpace(ctx.FormValue("content"))
	if content == "" {
		return ctx.ErrorRes(400, "comment content cannot be empty", nil)
	}

	comment := &db.GistComment{
		GistID:  gist.ID,
		UserID:  ctx.User.ID,
		Content: content,
	}
	if err := comment.Create(); err != nil {
		return ctx.ErrorRes(500, "Error creating comment", err)
	}

	return ctx.RedirectTo("/" + gist.User.Username + "/" + gist.Identifier() + "#comments")
}

func EditComment(ctx *context.Context) error {
	gist := ctx.GetData("gist").(*db.Gist)
	comment, err := lookupGistComment(gist, ctx.Param("commentid"))
	if err != nil || comment == nil {
		return ctx.NotFound("Comment not found")
	}
	if ctx.User == nil || (comment.UserID != ctx.User.ID && gist.UserID != ctx.User.ID) {
		return ctx.NotFound("Comment not found")
	}

	content := strings.TrimSpace(ctx.FormValue("content"))
	if content == "" {
		return ctx.ErrorRes(400, "comment content cannot be empty", nil)
	}

	comment.Content = content
	if err := comment.Update(); err != nil {
		return ctx.ErrorRes(500, "Error updating comment", err)
	}

	return ctx.RedirectTo("/" + gist.User.Username + "/" + gist.Identifier() + "#comments")
}

func DeleteComment(ctx *context.Context) error {
	gist := ctx.GetData("gist").(*db.Gist)
	comment, err := lookupGistComment(gist, ctx.Param("commentid"))
	if err != nil || comment == nil {
		return ctx.NotFound("Comment not found")
	}
	if ctx.User == nil || (comment.UserID != ctx.User.ID && gist.UserID != ctx.User.ID) {
		return ctx.NotFound("Comment not found")
	}

	if err := comment.Delete(); err != nil {
		return ctx.ErrorRes(500, "Error deleting comment", err)
	}

	return ctx.RedirectTo("/" + gist.User.Username + "/" + gist.Identifier() + "#comments")
}
