package v1

import (
	"strconv"
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

func canWriteComment(comment *db.GistComment, gist *db.Gist, user *db.User) bool {
	return user != nil && (comment.UserID == user.ID || gist.UserID == user.ID)
}

func lookupCommentByID(gist *db.Gist, commentID string) (*db.GistComment, *ErrorBody) {
	id, err := strconv.ParseUint(commentID, 10, 64)
	if err != nil {
		return nil, &ErrorBody{Code: "not_found", Message: "comment not found"}
	}

	comment, err := db.GetGistCommentByID(uint(id))
	if err != nil || comment.GistID != gist.ID {
		return nil, &ErrorBody{Code: "not_found", Message: "comment not found"}
	}
	return comment, nil
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

// UpdateComment handles PATCH /api/v1/gists/:uuid/comments/:comment_id
func UpdateComment(ctx *context.Context) error {
	g, errBody := lookupGistByUUID(ctx, ctx.Param("uuid"))
	if errBody != nil {
		return WriteJSONError(ctx, 404, errBody.Code, errBody.Message)
	}

	comment, errBody := lookupCommentByID(g, ctx.Param("comment_id"))
	if errBody != nil {
		return WriteJSONError(ctx, 404, errBody.Code, errBody.Message)
	}
	if !canWriteComment(comment, g, ctx.User) {
		return WriteJSONError(ctx, 404, "not_found", "comment not found")
	}

	var req UpdateCommentRequest
	if err := ctx.Bind(&req); err != nil {
		return WriteJSONError(ctx, 400, "validation_failed", "invalid request body")
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return WriteJSONError(ctx, 400, "validation_failed", "content is required")
	}

	comment.Content = content
	if err := comment.Update(); err != nil {
		return WriteJSONError(ctx, 500, "internal_error", "failed to update comment")
	}

	resp, err := commentToResponse(comment, ctx.User)
	if err != nil {
		return WriteJSONError(ctx, 500, "internal_error", "failed to serialize comment")
	}
	return ctx.JSON(200, resp)
}

// DeleteComment handles DELETE /api/v1/gists/:uuid/comments/:comment_id
func DeleteComment(ctx *context.Context) error {
	g, errBody := lookupGistByUUID(ctx, ctx.Param("uuid"))
	if errBody != nil {
		return WriteJSONError(ctx, 404, errBody.Code, errBody.Message)
	}

	comment, errBody := lookupCommentByID(g, ctx.Param("comment_id"))
	if errBody != nil {
		return WriteJSONError(ctx, 404, errBody.Code, errBody.Message)
	}
	if !canWriteComment(comment, g, ctx.User) {
		return WriteJSONError(ctx, 404, "not_found", "comment not found")
	}

	if err := comment.Delete(); err != nil {
		return WriteJSONError(ctx, 500, "internal_error", "failed to delete comment")
	}
	return ctx.NoContent(204)
}
