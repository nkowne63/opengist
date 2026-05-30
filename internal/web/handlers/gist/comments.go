package gist

import (
	"html/template"
	"strings"

	"github.com/thomiceli/opengist/internal/db"
	"github.com/thomiceli/opengist/internal/render"
	"github.com/thomiceli/opengist/internal/web/context"
)

type CommentView struct {
	*db.GistComment
	HTML template.HTML
}

func loadCommentViews(gistID uint) ([]CommentView, error) {
	comments, err := db.GetGistComments(gistID)
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
		})
	}

	return views, nil
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
