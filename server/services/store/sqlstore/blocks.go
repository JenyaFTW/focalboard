package sqlstore

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mattermost/focalboard/server/utils"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq" // postgres driver
	"github.com/mattermost/focalboard/server/model"
	"github.com/mattermost/focalboard/server/services/store"
	_ "github.com/mattn/go-sqlite3" // sqlite driver

	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

type RootIDNilError struct{}

func (re RootIDNilError) Error() string {
	return "rootId is nil"
}

type BlockNotFoundErr struct {
	blockID string
}

func (be BlockNotFoundErr) Error() string {
	return fmt.Sprintf("block not found (block id: %s", be.blockID)
}

func (s *SQLStore) blockFields() []string {
	return []string{
		"id",
		"parent_id",
		"root_id",
		"created_by",
		"modified_by",
		s.escapeField("schema"),
		"type",
		"title",
		"COALESCE(fields, '{}')",
		"create_at",
		"update_at",
		"delete_at",
		"COALESCE(workspace_id, '0')",
	}
}

func (s *SQLStore) getBlocksWithParentAndType(db sq.BaseRunner, c store.Container, parentID string, blockType string) ([]model.Block, error) {
	query := s.getQueryBuilder(db).
		Select(s.blockFields()...).
		From(s.tablePrefix + "blocks").
		Where(sq.Eq{"COALESCE(workspace_id, '0')": c.WorkspaceID}).
		Where(sq.Eq{"parent_id": parentID}).
		Where(sq.Eq{"type": blockType})

	rows, err := query.Query()
	if err != nil {
		s.logger.Error(`getBlocksWithParentAndType ERROR`, mlog.Err(err))

		return nil, err
	}
	defer s.CloseRows(rows)

	return s.blocksFromRows(rows)
}

func (s *SQLStore) getBlocksWithParent(db sq.BaseRunner, c store.Container, parentID string) ([]model.Block, error) {
	query := s.getQueryBuilder(db).
		Select(s.blockFields()...).
		From(s.tablePrefix + "blocks").
		Where(sq.Eq{"parent_id": parentID}).
		Where(sq.Eq{"coalesce(workspace_id, '0')": c.WorkspaceID})

	rows, err := query.Query()
	if err != nil {
		s.logger.Error(`getBlocksWithParent ERROR`, mlog.Err(err))

		return nil, err
	}
	defer s.CloseRows(rows)

	return s.blocksFromRows(rows)
}

func (s *SQLStore) getBlocksWithRootID(db sq.BaseRunner, c store.Container, rootID string) ([]model.Block, error) {
	query := s.getQueryBuilder(db).
		Select(s.blockFields()...).
		From(s.tablePrefix + "blocks").
		Where(sq.Eq{"root_id": rootID}).
		Where(sq.Eq{"coalesce(workspace_id, '0')": c.WorkspaceID})

	rows, err := query.Query()
	if err != nil {
		s.logger.Error(`GetBlocksWithRootID ERROR`, mlog.Err(err))

		return nil, err
	}
	defer s.CloseRows(rows)

	return s.blocksFromRows(rows)
}

func (s *SQLStore) getBlocksWithType(db sq.BaseRunner, c store.Container, blockType string) ([]model.Block, error) {
	query := s.getQueryBuilder(db).
		Select(s.blockFields()...).
		From(s.tablePrefix + "blocks").
		Where(sq.Eq{"type": blockType}).
		Where(sq.Eq{"coalesce(workspace_id, '0')": c.WorkspaceID})

	rows, err := query.Query()
	if err != nil {
		s.logger.Error(`getBlocksWithParentAndType ERROR`, mlog.Err(err))

		return nil, err
	}
	defer s.CloseRows(rows)

	return s.blocksFromRows(rows)
}

// GetSubTree2 returns blocks within 2 levels of the given blockID.
func (s *SQLStore) getSubTree2(db sq.BaseRunner, c store.Container, blockID string) ([]model.Block, error) {
	query := s.getQueryBuilder(db).
		Select(s.blockFields()...).
		From(s.tablePrefix + "blocks").
		Where(sq.Or{sq.Eq{"id": blockID}, sq.Eq{"parent_id": blockID}}).
		Where(sq.Eq{"coalesce(workspace_id, '0')": c.WorkspaceID})

	rows, err := query.Query()
	if err != nil {
		s.logger.Error(`getSubTree ERROR`, mlog.Err(err))

		return nil, err
	}
	defer s.CloseRows(rows)

	return s.blocksFromRows(rows)
}

// GetSubTree3 returns blocks within 3 levels of the given blockID.
func (s *SQLStore) getSubTree3(db sq.BaseRunner, c store.Container, blockID string) ([]model.Block, error) {
	// This first subquery returns repeated blocks
	query := s.getQueryBuilder(db).Select(
		"l3.id",
		"l3.parent_id",
		"l3.root_id",
		"l3.created_by",
		"l3.modified_by",
		"l3."+s.escapeField("schema"),
		"l3.type",
		"l3.title",
		"l3.fields",
		"l3.create_at",
		"l3.update_at",
		"l3.delete_at",
		"COALESCE(l3.workspace_id, '0')",
	).
		From(s.tablePrefix + "blocks as l1").
		Join(s.tablePrefix + "blocks as l2 on l2.parent_id = l1.id or l2.id = l1.id").
		Join(s.tablePrefix + "blocks as l3 on l3.parent_id = l2.id or l3.id = l2.id").
		Where(sq.Eq{"l1.id": blockID}).
		Where(sq.Eq{"COALESCE(l3.workspace_id, '0')": c.WorkspaceID})

	if s.dbType == postgresDBType {
		query = query.Options("DISTINCT ON (l3.id)")
	} else {
		query = query.Distinct()
	}

	rows, err := query.Query()
	if err != nil {
		s.logger.Error(`getSubTree3 ERROR`, mlog.Err(err))

		return nil, err
	}
	defer s.CloseRows(rows)

	return s.blocksFromRows(rows)
}

func (s *SQLStore) getAllBlocks(db sq.BaseRunner, c store.Container) ([]model.Block, error) {
	query := s.getQueryBuilder(db).
		Select(s.blockFields()...).
		From(s.tablePrefix + "blocks").
		Where(sq.Eq{"coalesce(workspace_id, '0')": c.WorkspaceID})

	rows, err := query.Query()
	if err != nil {
		s.logger.Error(`getAllBlocks ERROR`, mlog.Err(err))

		return nil, err
	}
	defer s.CloseRows(rows)

	return s.blocksFromRows(rows)
}

func (s *SQLStore) blocksFromRows(rows *sql.Rows) ([]model.Block, error) {
	results := []model.Block{}

	for rows.Next() {
		var block model.Block
		var fieldsJSON string
		var modifiedBy sql.NullString

		err := rows.Scan(
			&block.ID,
			&block.ParentID,
			&block.RootID,
			&block.CreatedBy,
			&modifiedBy,
			&block.Schema,
			&block.Type,
			&block.Title,
			&fieldsJSON,
			&block.CreateAt,
			&block.UpdateAt,
			&block.DeleteAt,
			&block.WorkspaceID)
		if err != nil {
			// handle this error
			s.logger.Error(`ERROR blocksFromRows`, mlog.Err(err))

			return nil, err
		}

		if modifiedBy.Valid {
			block.ModifiedBy = modifiedBy.String
		}

		err = json.Unmarshal([]byte(fieldsJSON), &block.Fields)
		if err != nil {
			// handle this error
			s.logger.Error(`ERROR blocksFromRows fields`, mlog.Err(err))

			return nil, err
		}

		results = append(results, block)
	}

	return results, nil
}

func (s *SQLStore) getRootID(db sq.BaseRunner, c store.Container, blockID string) (string, error) {
	query := s.getQueryBuilder(db).Select("root_id").
		From(s.tablePrefix + "blocks").
		Where(sq.Eq{"id": blockID}).
		Where(sq.Eq{"coalesce(workspace_id, '0')": c.WorkspaceID})

	row := query.QueryRow()

	var rootID string

	err := row.Scan(&rootID)
	if err != nil {
		return "", err
	}

	return rootID, nil
}

func (s *SQLStore) getParentID(db sq.BaseRunner, c store.Container, blockID string) (string, error) {
	query := s.getQueryBuilder(db).Select("parent_id").
		From(s.tablePrefix + "blocks").
		Where(sq.Eq{"id": blockID}).
		Where(sq.Eq{"coalesce(workspace_id, '0')": c.WorkspaceID})

	row := query.QueryRow()

	var parentID string

	err := row.Scan(&parentID)
	if err != nil {
		return "", err
	}

	return parentID, nil
}

func (s *SQLStore) insertBlock(db sq.BaseRunner, c store.Container, block *model.Block, userID string) error {
	if block.RootID == "" {
		return RootIDNilError{}
	}

	fieldsJSON, err := json.Marshal(block.Fields)
	if err != nil {
		return err
	}

	existingBlock, err := s.getBlock(db, c, block.ID)
	if err != nil {
		return err
	}

	insertQuery := s.getQueryBuilder(db).Insert("").
		Columns(
			"workspace_id",
			"id",
			"parent_id",
			"root_id",
			"created_by",
			"modified_by",
			s.escapeField("schema"),
			"type",
			"title",
			"fields",
			"create_at",
			"update_at",
			"delete_at",
		)

	insertQueryValues := map[string]interface{}{
		"workspace_id":          c.WorkspaceID,
		"id":                    block.ID,
		"parent_id":             block.ParentID,
		"root_id":               block.RootID,
		s.escapeField("schema"): block.Schema,
		"type":                  block.Type,
		"title":                 block.Title,
		"fields":                fieldsJSON,
		"delete_at":             block.DeleteAt,
		"created_by":            block.CreatedBy,
		"modified_by":           block.ModifiedBy,
		"create_at":             block.CreateAt,
		"update_at":             block.UpdateAt,
	}

	block.UpdateAt = utils.GetMillis()
	block.ModifiedBy = userID

	if existingBlock != nil {
		// block with ID exists, so this is an update operation
		query := s.getQueryBuilder(db).Update(s.tablePrefix+"blocks").
			Where(sq.Eq{"id": block.ID}).
			Where(sq.Eq{"COALESCE(workspace_id, '0')": c.WorkspaceID}).
			Set("parent_id", block.ParentID).
			Set("root_id", block.RootID).
			Set("modified_by", block.ModifiedBy).
			Set(s.escapeField("schema"), block.Schema).
			Set("type", block.Type).
			Set("title", block.Title).
			Set("fields", fieldsJSON).
			Set("update_at", block.UpdateAt).
			Set("delete_at", block.DeleteAt)

		if _, err := query.Exec(); err != nil {
			s.logger.Error(`InsertBlock error occurred while updating existing block`, mlog.String("blockID", block.ID), mlog.Err(err))
			return err
		}
	} else {
		block.CreatedBy = userID
		block.CreateAt = utils.GetMillis()
		block.ModifiedBy = userID
		block.UpdateAt = utils.GetMillis()

		insertQueryValues["created_by"] = block.CreatedBy
		insertQueryValues["create_at"] = block.CreateAt
		insertQueryValues["update_at"] = block.UpdateAt
		insertQueryValues["modified_by"] = block.ModifiedBy

		query := insertQuery.SetMap(insertQueryValues).Into(s.tablePrefix + "blocks")
		if _, err := query.Exec(); err != nil {
			return err
		}
	}

	// writing block history
	query := insertQuery.SetMap(insertQueryValues).Into(s.tablePrefix + "blocks_history")
	if _, err := query.Exec(); err != nil {
		return err
	}

	return nil
}

func (s *SQLStore) patchBlock(db sq.BaseRunner, c store.Container, blockID string, blockPatch *model.BlockPatch, userID string) error {
	existingBlock, err := s.getBlock(db, c, blockID)
	if err != nil {
		return err
	}
	if existingBlock == nil {
		return BlockNotFoundErr{blockID}
	}

	block := blockPatch.Patch(existingBlock)
	return s.insertBlock(db, c, block, userID)
}

func (s *SQLStore) deleteBlock(db sq.BaseRunner, c store.Container, blockID string, modifiedBy string) error {
	now := utils.GetMillis()
	insertQuery := s.getQueryBuilder(db).Insert(s.tablePrefix+"blocks_history").
		Columns(
			"workspace_id",
			"id",
			"modified_by",
			"update_at",
			"delete_at",
		).
		Values(
			c.WorkspaceID,
			blockID,
			modifiedBy,
			now,
			now,
		)

	if _, err := insertQuery.Exec(); err != nil {
		return err
	}

	deleteQuery := s.getQueryBuilder(db).
		Delete(s.tablePrefix + "blocks").
		Where(sq.Eq{"id": blockID}).
		Where(sq.Eq{"COALESCE(workspace_id, '0')": c.WorkspaceID})

	if _, err := deleteQuery.Exec(); err != nil {
		return err
	}

	return nil
}

func (s *SQLStore) getBlockCountsByType(db sq.BaseRunner) (map[string]int64, error) {
	query := s.getQueryBuilder(db).
		Select(
			"type",
			"COUNT(*) AS count",
		).
		From(s.tablePrefix + "blocks").
		GroupBy("type")

	rows, err := query.Query()
	if err != nil {
		s.logger.Error(`GetBlockCountsByType ERROR`, mlog.Err(err))

		return nil, err
	}
	defer s.CloseRows(rows)

	m := make(map[string]int64)

	for rows.Next() {
		var blockType string
		var count int64

		err := rows.Scan(&blockType, &count)
		if err != nil {
			s.logger.Error("Failed to fetch block count", mlog.Err(err))
			return nil, err
		}
		m[blockType] = count
	}
	return m, nil
}

func (s *SQLStore) getBlock(db sq.BaseRunner, c store.Container, blockID string) (*model.Block, error) {
	query := s.getQueryBuilder(db).
		Select(s.blockFields()...).
		From(s.tablePrefix + "blocks").
		Where(sq.Eq{"id": blockID}).
		Where(sq.Eq{"coalesce(workspace_id, '0')": c.WorkspaceID})

	rows, err := query.Query()
	if err != nil {
		s.logger.Error(`GetBlock ERROR`, mlog.Err(err))
		return nil, err
	}

	blocks, err := s.blocksFromRows(rows)
	if err != nil {
		return nil, err
	}

	if len(blocks) == 0 {
		return nil, nil
	}

	return &blocks[0], nil
}
