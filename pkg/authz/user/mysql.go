package authz

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

type MySQLRepository struct {
	database.SQLRepository
}

func NewMySQLRepository(db *database.MySQL) MySQLRepository {
	return MySQLRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo MySQLRepository) Create(ctx context.Context, model Model) (int64, error) {
	result, err := repo.DB.ExecContext(
		ctx,
		`
			INSERT INTO user (
				userId,
				objectId,
				email
			) VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE
				objectId = ?,
				email = ?,
				createdAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
		`,
		model.GetUserId(),
		model.GetObjectId(),
		model.GetEmail(),
		model.GetObjectId(),
		model.GetEmail(),
	)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to create user")
	}

	newUserId, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "Unable to create user")
	}

	return newUserId, nil
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var user User
	err := repo.DB.GetContext(
		ctx,
		&user,
		`
			SELECT id, objectId, userId, email, createdAt, updatedAt, deletedAt
			FROM user
			WHERE
				id = ? AND
				deletedAt IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("User", id)
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (repo MySQLRepository) GetByUserId(ctx context.Context, userId string) (Model, error) {
	var user User
	err := repo.DB.GetContext(
		ctx,
		&user,
		`
			SELECT id, objectId, userId, email, createdAt, updatedAt, deletedAt
			FROM user
			WHERE
				userId = ? AND
				deletedAt IS NULL
		`,
		userId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("User", userId)
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (repo MySQLRepository) List(ctx context.Context, listParams middleware.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	users := make([]User, 0)
	query := `
		SELECT id, objectId, userId, email, createdAt, updatedAt, deletedAt
		FROM user
		WHERE
			deletedAt IS NULL
	`
	replacements := []interface{}{}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (userId LIKE ? OR email LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.AfterId != "" {
		if listParams.AfterValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s > ? OR (userId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s < ? OR (userId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND userId > ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND userId < ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.BeforeId != "" {
		if listParams.BeforeValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s < ? OR (userId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s > ? OR (userId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND userId < ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND userId > ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.SortBy != "userId" {
		query = fmt.Sprintf("%s ORDER BY %s %s, userId %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY userId %s LIMIT ?", query, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	}

	err := repo.DB.SelectContext(
		ctx,
		&users,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return nil, err
		}
	}

	for i := range users {
		models = append(models, &users[i])
	}

	return models, nil
}

func (repo MySQLRepository) UpdateByUserId(ctx context.Context, userId string, model Model) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE user
			SET
				email = ?
			WHERE
				userId = ? AND
				deletedAt IS NULL
		`,
		model.GetEmail(),
		model.GetUserId(),
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating user %d", model.GetID()))
	}

	return nil
}

func (repo MySQLRepository) DeleteByUserId(ctx context.Context, userId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE user
			SET deletedAt = ?
			WHERE
				userId = ? AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		userId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("User", userId)
		default:
			return err
		}
	}

	return nil
}
