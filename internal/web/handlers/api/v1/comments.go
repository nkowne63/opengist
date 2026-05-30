package v1

import (
	"strings"
	"time"

	"github.com/thomiceli/opengist/internal/db"
	"github.com/thomiceli/opengist/internal/render"
	"github.com/thomiceli/opengist/internal/web/context"
)

func commentToResponse(c *db.GistComment, fallbackUser *db.User) (Comment, error) {
	html, err := render.MarkdownString(c.Content)
	if err != nil {
		return Comment{}, err
	}

	user := CommentUser{}
	if c.User != nil {
		user.ID = c.User.ID
		user.Username = c.User.Username
	} else if fallbackUser != nil {
		user.ID = fallbackUser.ID
		user.Username = fallbackUser.Username
	}

	return Comment{
		ID:        c.ID,
		Content:   c.Content,
		HTML:      html,
		CreatedAt: time.Unix(c.CreatedAt, 0),
		UpdatedAt: time.Unix(c.UpdatedAt, 0),
		User:      user,
	}, nil
}

func commentsToResponse(comments []*db.GistComment) ([]Comment, error) {
	resp := make([]Comment, 0, len(comments))
	for _, comment := range comments {
		item, err := commentToResponse(comment, nil)
		if err != nil {
			return nil, err
		}
		resp = append(resp, item)
	}
	return resp, nil
}

// ListComments handles GET /api/v1/gists/:uuid/comments
func ListComments(ctx *context.Context) error {
	g, errBody := lookupGistByUUID(ctx, ctx.Param("uuid"))
	if errBody != nil {
		return WriteJSONError(ctx, 404, errBody.Code, errBody.Message)
	}

	comments, err := db.GetGistComments(g.ID)
	if err != nil {
		return WriteJSONError(ctx, 500, "internal_error", "failed to load comments")
	}

	resp, err := commentsToResponse(comments)
	if err != nil {
		return WriteJSONError(ctx, 500, "internal_error", "failed to serialize comments")
	}
	return ctx.JSON(200, resp)
}

// CreateComment handles POST /api/v1/gists/:uuid/comments
func CreateComment(ctx *context.Context) error {
	g, errBody := lookupGistByUUID(ctx, ctx.Param("uuid"))
	if errBody != nil {
		return WriteJSONError(ctx, 404, errBody.Code, errBody.Message)
	}

	var req CreateCommentRequest
	if err := ctx.Bind(&req); err != nil {
		return WriteJSONError(ctx, 400, "validation_failed", "invalid request body")
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return WriteJSONError(ctx, 400, "validation_failed", "content is required")
	}

	comment := &db.GistComment{
		GistID:  g.ID,
		UserID:  ctx.User.ID,
		Content: content,
	}
	if err := comment.Create(); err != nil {
		return WriteJSONError(ctx, 500, "internal_error", "failed to create comment")
	}

	resp, err := commentToResponse(comment, ctx.User)
	if err != nil {
		return WriteJSONError(ctx, 500, "internal_error", "failed to serialize comment")
	}
	return ctx.JSON(201, resp)
}
