/*
Copyright 2020 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package planbuilder

import (
	vtrpcpb "vitess.io/vitess/go/vt/proto/vtrpc"
	"vitess.io/vitess/go/vt/sqlparser"
	"vitess.io/vitess/go/vt/vterrors"
	"vitess.io/vitess/go/vt/vtgate/engine"
	"vitess.io/vitess/go/vt/vtgate/planbuilder/plancontext"
	"vitess.io/vitess/go/vt/vtgate/semantics"
)

var _ logicalPlan = (*sqlCalcFoundRows)(nil)

type sqlCalcFoundRows struct {
	LimitQuery, CountQuery logicalPlan

	// only used by WireUp for V3
	ljt, cjt *jointab
}

// Wireup implements the logicalPlan interface
func (s *sqlCalcFoundRows) Wireup(logicalPlan, *jointab) error {
	err := s.LimitQuery.Wireup(s.LimitQuery, s.ljt)
	if err != nil {
		return err
	}
	return s.CountQuery.Wireup(s.CountQuery, s.cjt)
}

// WireupGen4 implements the logicalPlan interface
func (s *sqlCalcFoundRows) WireupGen4(ctx *plancontext.PlanningContext) error {
	err := s.LimitQuery.WireupGen4(ctx)
	if err != nil {
		return err
	}
	return s.CountQuery.WireupGen4(ctx)
}

// ContainsTables implements the logicalPlan interface
func (s *sqlCalcFoundRows) ContainsTables() semantics.TableSet {
	return s.LimitQuery.ContainsTables()
}

// Primitive implements the logicalPlan interface
func (s *sqlCalcFoundRows) Primitive() engine.Primitive {
	countPrim := s.CountQuery.Primitive()
	rb, ok := countPrim.(*engine.Route)
	if ok {
		// if our count query is an aggregation, we want the no-match result to still return a zero
		rb.NoRoutesSpecialHandling = true
	}
	return engine.SQLCalcFoundRows{
		LimitPrimitive: s.LimitQuery.Primitive(),
		CountPrimitive: countPrim,
	}
}

// All the methods below are not implemented. They should not be called on a sqlCalcFoundRows plan

// Order implements the logicalPlan interface
func (s *sqlCalcFoundRows) Order() int {
	return s.LimitQuery.Order()
}

// ResultColumns implements the logicalPlan interface
func (s *sqlCalcFoundRows) ResultColumns() []*resultColumn {
	return s.LimitQuery.ResultColumns()
}

// Reorder implements the logicalPlan interface
func (s *sqlCalcFoundRows) Reorder(order int) {
	s.LimitQuery.Reorder(order)
}

// SupplyVar implements the logicalPlan interface
func (s *sqlCalcFoundRows) SupplyVar(from, to int, col *sqlparser.ColName, varname string) {
	s.LimitQuery.SupplyVar(from, to, col, varname)
}

// SupplyCol implements the logicalPlan interface
func (s *sqlCalcFoundRows) SupplyCol(col *sqlparser.ColName) (*resultColumn, int) {
	return s.LimitQuery.SupplyCol(col)
}

// SupplyWeightString implements the logicalPlan interface
func (s *sqlCalcFoundRows) SupplyWeightString(int, bool) (weightcolNumber int, err error) {
	return 0, UnsupportedSupplyWeightString{Type: "sqlCalcFoundRows"}
}

// Rewrite implements the logicalPlan interface
func (s *sqlCalcFoundRows) Rewrite(inputs ...logicalPlan) error {
	if len(inputs) != 2 {
		return vterrors.Errorf(vtrpcpb.Code_INTERNAL, "[BUG] wrong number of inputs for SQL_CALC_FOUND_ROWS: %d", len(inputs))
	}
	s.LimitQuery = inputs[0]
	s.CountQuery = inputs[1]
	return nil
}

// Inputs implements the logicalPlan interface
func (s *sqlCalcFoundRows) Inputs() []logicalPlan {
	return []logicalPlan{s.LimitQuery, s.CountQuery}
}

// OutputColumns implements the logicalPlan interface
func (s *sqlCalcFoundRows) OutputColumns() []sqlparser.SelectExpr {
	return s.LimitQuery.OutputColumns()
}
