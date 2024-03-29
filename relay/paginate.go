package relay

import (
	"github.com/cloudmatelabs/gorm-gqlgen-relay/order"
	"github.com/cloudmatelabs/gorm-gqlgen-relay/utils"
	"github.com/cloudmatelabs/gorm-gqlgen-relay/where"
	"gorm.io/gorm"
)

type PaginateOption struct {
	First      *int
	Last       *int
	After      *string
	Before     *string
	Prefix     *string
	Table      string
	Tables     *map[string]string
	PrimaryKey string
}

func Paginate[Model any](db *gorm.DB, _where any, _orderBy any, option PaginateOption) (*Connection[Model], error) {
	if err := validation(option.First, option.Last, option.After, option.Before); err != nil {
		return nil, err
	}

	w, err := where.Do(db.Dialector.Name(), option.Table, option.Tables, option.Prefix, _where)
	if err != nil {
		return nil, err
	}

	stmt := where.Traverse(db, w)

	totalCount, err := getTotalCount[Model](db, w)
	if err != nil {
		return nil, err
	}

	orderBy, err := utils.ConvertToMap(_orderBy)
	if err != nil {
		return nil, err
	}

	orders, err := order.By(option.Table, option.Tables, _orderBy, option.Last != nil)
	if err != nil {
		return nil, err
	}

	for _, order := range orders {
		stmt = stmt.Order(order)
	}

	stmt, err = setAfter(stmt, option.After, orderBy, option.PrimaryKey)
	if err != nil {
		return nil, err
	}

	// remaining count is derived
	// before limit and setBefore are applied
	// after order by and setAfter are applied
	var remainingCount int64
	var model Model
	err = stmt.Model(&model).Count(&remainingCount).Error
	if err != nil {
		return nil, err
	}

	stmt, err = setBefore(stmt, option.Before, orderBy, option.PrimaryKey)
	if err != nil {
		return nil, err
	}

	stmt = limit(stmt, option.First, option.Last)

	var rows []*Model
	if err := stmt.Find(&rows).Error; err != nil {
		return nil, err
	}

	edges, err := convertToEdge(rows, utils.Keys(orderBy), option.PrimaryKey)
	if err != nil {
		return nil, err
	}

	pageInfo := &PageInfo{}
	_totalCount := int(totalCount)
	_remainingCount := int(remainingCount)
	edgesLen := len(edges)
	pageInfo.SetHasPreviousPage(_totalCount, edgesLen, option.After)
	pageInfo.SetHasNextPage(_remainingCount, edgesLen, option.First, option.Last, option.Before, option.After)

	if edgesLen > 0 {
		pageInfo.StartCursor = &edges[0].Cursor
		pageInfo.EndCursor = &edges[edgesLen-1].Cursor
	}

	return &Connection[Model]{
		TotalCount: totalCount,
		Edges:      edges,
		PageInfo:   pageInfo,
	}, nil
}
