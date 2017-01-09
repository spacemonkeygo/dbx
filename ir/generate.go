// Copyright (C) 2016 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ir

type GenerateOptions struct {
	Insert             bool
	RawInsert          bool
	SelectAll          bool
	SelectByPrimaryKey bool
	SelectByUnique     bool
	DeleteByPrimaryKey bool
	DeleteByUnique     bool
	UpdateByPrimaryKey bool
	UpdateByUnique     bool
	SelectCount        bool
	SelectPaged        bool
}

// func generateBasicInserts(model *Model, options GenerateOptions) (
// 	inserts []*Insert) {

// 	if options.Insert {
// 		inserts = append(inserts, &Insert{
// 			Model: model,
// 			Raw:   false,
// 		})
// 	}
// 	if options.RawInsert {
// 		inserts = append(inserts, &Insert{
// 			Model: model,
// 			Raw:   true,
// 		})
// 	}
// 	return inserts
// }

// func generateBasicSelects(model *Model, options GenerateOptions) (
// 	selects []*Select) {

// 	// Select by unique tuples
// 	if options.SelectAll {
// 		selects = append(selects, &Select{
// 			Fields: []Selectable{model},
// 			From:   model,
// 		})
// 	}

// 	// Select by primary key
// 	if options.UpdateByPrimaryKey && model.PrimaryKey != nil {
// 		selects = append(selects, &Select{
// 			Fields: []Selectable{model},
// 			From:   model,
// 			Where:  WhereFieldsEquals(model.PrimaryKey...),
// 		})
// 	}

// 	// Select by unique tuples
// 	if options.UpdateByUnique {
// 		for _, unique := range model.Unique {
// 			selects = append(selects, &Select{
// 				Fields: []Selectable{model},
// 				From:   model,
// 				Where:  WhereFieldsEquals(unique...),
// 			})
// 		}
// 	}
// 	return selects
// }

// func generateBasicDeletes(model *Model, options GenerateOptions) (
// 	deletes []*Delete) {

// 	// Always generate delete alls
// 	deletes = append(deletes, &Delete{
// 		Model: model,
// 	})

// 	// Delete by primary key
// 	if options.DeleteByPrimaryKey && model.PrimaryKey != nil {
// 		deletes = append(deletes, &Delete{
// 			Model: model,
// 			Where: WhereFieldsEquals(model.PrimaryKey...),
// 		})
// 	}

// 	// Delete by unique tuples
// 	if options.DeleteByUnique {
// 		for _, unique := range model.Unique {
// 			deletes = append(deletes, &Delete{
// 				Model: model,
// 				Where: WhereFieldsEquals(unique...),
// 			})
// 		}
// 	}
// 	return deletes
// }

// func generateBasicUpdates(model *Model, options GenerateOptions) (
// 	updates []*Update) {

// 	if len(model.UpdatableFields()) == 0 {
// 		return nil
// 	}

// 	// Update by primary key
// 	if options.UpdateByPrimaryKey && model.PrimaryKey != nil {
// 		updates = append(updates, &Update{
// 			Model: model,
// 			Where: WhereFieldsEquals(model.PrimaryKey...),
// 		})
// 	}

// 	// Update by unique tuples
// 	if options.UpdateByUnique {
// 		for _, unique := range model.Unique {
// 			updates = append(updates, &Update{
// 				Model: model,
// 				Where: WhereFieldsEquals(unique...),
// 			})
// 		}
// 	}
// 	return updates
// }
