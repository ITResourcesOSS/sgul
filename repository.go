// Copyright 2019 Luca Stasio <joshuagame@gmail.com>
// Copyright 2019 IT Resources s.r.l.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
//
// DoInTransaction (C) vertazzar on a comment at https://github.com/jinzhu/gorm/issues/2089
// Thanks for your useful tip.

// Package sgul defines common structures and functionalities for applications.
// repository.go defines commons for a Gorm based Repository structure.
package sgul

import "github.com/jinzhu/gorm"

// GormRepository defines the base repository structure form gorm based db access
type GormRepository struct {
	DB *gorm.DB
}

// InTransaction defines the func type to be executed in a gorm transaction.
type InTransaction func(tx *gorm.DB) error

// DoInTransaction executes the fn() callback in a gorm transaction.
func (r GormRepository) DoInTransaction(fn InTransaction) error {
	tx := r.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := fn(tx); err != nil {
		xerr := tx.Rollback().Error
		if xerr != nil {
			return err
		}
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// NewRepository returns a new Repository instance
func NewRepository(db *gorm.DB) GormRepository {
	return GormRepository{DB: db}
}
