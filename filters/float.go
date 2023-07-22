package filters

import (
	"github.com/cloudmatelabs/gorm-gqlgen-relay/query"
	"gorm.io/gorm"
)

type FloatFilter struct {
	Not                *FloatFilter   `json:"not,omitempty"`
	And                *[]FloatFilter `json:"and,omitempty"`
	Or                 *[]FloatFilter `json:"or,omitempty"`
	Equal              *float64       `json:"equal,omitempty"`
	NotEqual           *float64       `json:"notEqual,omitempty"`
	In                 *[]float64     `json:"in,omitempty"`
	NotIn              *[]float64     `json:"notIn,omitempty"`
	GreaterThan        *float64       `json:"gt,omitempty"`
	GreaterThanOrEqual *float64       `json:"gte,omitempty"`
	LessThan           *float64       `json:"lt,omitempty"`
	LessThanOrEqual    *float64       `json:"lte,omitempty"`
	IsNull             *bool          `json:"isNull,omitempty"`
	IsNotNull          *bool          `json:"isNotNull,omitempty"`
}

func Float(db *gorm.DB, field string, input interface{}) (*gorm.DB, error) {
	var filter Filter[float64]
	if err := filter.Parse(input); err != nil {
		return db, err
	}

	db = db.Scopes(
		query.Equal(field, filter.Equal),
		query.NotEqual(field, filter.NotEqual),
		query.In(field, filter.In),
		query.NotIn(field, filter.NotIn),
		query.GreaterThan(field, filter.GreaterThan),
		query.GreaterThanOrEqual(field, filter.GreaterThanOrEqual),
		query.LessThan(field, filter.LessThan),
		query.LessThanOrEqual(field, filter.LessThanOrEqual),
		query.IsNull(field, filter.IsNull),
		query.IsNotNull(field, filter.IsNotNull),
	)

	if filter.Not != nil {
		db = db.Scopes(func(d *gorm.DB) *gorm.DB {
			return d.Not(Float(d, field, *filter.Not))
		})
	}

	if filter.And != nil {
		for _, and := range *filter.And {
			_filter := and

			db = db.Scopes(func(d *gorm.DB) *gorm.DB {
				return d.Where(Float(d, field, _filter))
			})
		}
	}

	if filter.Or != nil {
		for _, or := range *filter.Or {
			_filter := or

			db = db.Scopes(func(d *gorm.DB) *gorm.DB {
				return d.Or(Float(d, field, _filter))
			})
		}
	}

	return db, nil
}