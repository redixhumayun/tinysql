// Copyright 2015 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package ast

import (
	"github.com/pingcap/tidb/parser/model"
	"github.com/pingcap/tidb/parser/mysql"
)

var (
	_ DMLNode = &DeleteStmt{}
	_ DMLNode = &InsertStmt{}
	_ DMLNode = &SelectStmt{}
	_ DMLNode = &ShowStmt{}

	_ Node = &Assignment{}
	_ Node = &ByItem{}
	_ Node = &FieldList{}
	_ Node = &GroupByClause{}
	_ Node = &HavingClause{}
	_ Node = &Join{}
	_ Node = &Limit{}
	_ Node = &OnCondition{}
	_ Node = &OrderByClause{}
	_ Node = &SelectField{}
	_ Node = &TableName{}
	_ Node = &TableRefsClause{}
	_ Node = &TableSource{}
	_ Node = &WildCardField{}
)

// JoinType is join type, including cross/left/right/full.
type JoinType int

const (
	// CrossJoin is cross join type.
	CrossJoin JoinType = iota + 1
	// LeftJoin is left Join type.
	LeftJoin
	// RightJoin is right Join type.
	RightJoin
	// InnerJoin
	InnerJoin
)

// Join represents table join.
type Join struct {
	node

	// Left table can be TableSource or JoinNode.
	Left ResultSetNode
	// Right table can be TableSource or JoinNode or nil.
	Right ResultSetNode
	// Tp represents join type.
	Tp JoinType
	// On represents join on condition.
	On *OnCondition
}

// Accept implements Node Accept interface.
func (n *Join) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*Join)
	node, ok := n.Left.Accept(v)
	if !ok {
		return n, false
	}
	n.Left = node.(ResultSetNode)
	if n.Right != nil {
		node, ok = n.Right.Accept(v)
		if !ok {
			return n, false
		}
		n.Right = node.(ResultSetNode)
	}
	if n.On != nil {
		node, ok = n.On.Accept(v)
		if !ok {
			return n, false
		}
		n.On = node.(*OnCondition)
	}
	return v.Leave(n)
}

// TableName represents a table name.
type TableName struct {
	node

	Schema model.CIStr
	Name   model.CIStr

	DBInfo    *model.DBInfo
	TableInfo *model.TableInfo

	IndexHints     []*IndexHint
	PartitionNames []model.CIStr
}

// IndexHintType is the type for index hint use, ignore or force.
type IndexHintType int

// IndexHintUseType values.
const (
	HintUse    IndexHintType = 1
	HintIgnore IndexHintType = 2
	HintForce  IndexHintType = 3
)

// IndexHintScope is the type for index hint for join, order by or group by.
type IndexHintScope int

// Index hint scopes.
const (
	HintForScan    IndexHintScope = 1
	HintForJoin    IndexHintScope = 2
	HintForOrderBy IndexHintScope = 3
	HintForGroupBy IndexHintScope = 4
)

// IndexHint represents a hint for optimizer to use/ignore/force for join/order by/group by.
type IndexHint struct {
	IndexNames []model.CIStr
	HintType   IndexHintType
	HintScope  IndexHintScope
}

// Accept implements Node Accept interface.
func (n *TableName) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*TableName)
	return v.Leave(n)
}

// OnCondition represents JOIN on condition.
type OnCondition struct {
	node

	Expr ExprNode
}

// Accept implements Node Accept interface.
func (n *OnCondition) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*OnCondition)
	node, ok := n.Expr.Accept(v)
	if !ok {
		return n, false
	}
	n.Expr = node.(ExprNode)
	return v.Leave(n)
}

// TableSource represents table source with a name.
type TableSource struct {
	node

	// Source is the source of the data, can be a TableName,
	// a SelectStmt, a UnionStmt, or a JoinNode.
	Source ResultSetNode

	// AsName is the alias name of the table source.
	AsName model.CIStr
}

// Accept implements Node Accept interface.
func (n *TableSource) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*TableSource)
	node, ok := n.Source.Accept(v)
	if !ok {
		return n, false
	}
	n.Source = node.(ResultSetNode)
	return v.Leave(n)
}

// WildCardField is a special type of select field content.
type WildCardField struct {
	node

	Table  model.CIStr
	Schema model.CIStr
}

// Accept implements Node Accept interface.
func (n *WildCardField) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*WildCardField)
	return v.Leave(n)
}

// SelectField represents fields in select statement.
// There are two type of select field: wildcard
// and expression with optional alias name.
type SelectField struct {
	node

	// Offset is used to get original text.
	Offset int
	// WildCard is not nil, Expr will be nil.
	WildCard *WildCardField
	// Expr is not nil, WildCard will be nil.
	Expr ExprNode
	// AsName is alias name for Expr.
	AsName model.CIStr
	// Auxiliary stands for if this field is auxiliary.
	// When we add a Field into SelectField list which is used for having/orderby clause but the field is not in select clause,
	// we should set its Auxiliary to true. Then the TrimExec will trim the field.
	Auxiliary bool
}

// Accept implements Node Accept interface.
func (n *SelectField) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*SelectField)
	if n.Expr != nil {
		node, ok := n.Expr.Accept(v)
		if !ok {
			return n, false
		}
		n.Expr = node.(ExprNode)
	}
	return v.Leave(n)
}

// FieldList represents field list in select statement.
type FieldList struct {
	node

	Fields []*SelectField
}

// Accept implements Node Accept interface.
func (n *FieldList) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*FieldList)
	for i, val := range n.Fields {
		node, ok := val.Accept(v)
		if !ok {
			return n, false
		}
		n.Fields[i] = node.(*SelectField)
	}
	return v.Leave(n)
}

// TableRefsClause represents table references clause in dml statement.
type TableRefsClause struct {
	node

	TableRefs *Join
}

// Accept implements Node Accept interface.
func (n *TableRefsClause) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*TableRefsClause)
	node, ok := n.TableRefs.Accept(v)
	if !ok {
		return n, false
	}
	n.TableRefs = node.(*Join)
	return v.Leave(n)
}

// ByItem represents an item in order by or group by.
type ByItem struct {
	node

	Expr ExprNode
	Desc bool
}

// Accept implements Node Accept interface.
func (n *ByItem) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*ByItem)
	node, ok := n.Expr.Accept(v)
	if !ok {
		return n, false
	}
	n.Expr = node.(ExprNode)
	return v.Leave(n)
}

// GroupByClause represents group by clause.
type GroupByClause struct {
	node
	Items []*ByItem
}

// Accept implements Node Accept interface.
func (n *GroupByClause) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*GroupByClause)
	for i, val := range n.Items {
		node, ok := val.Accept(v)
		if !ok {
			return n, false
		}
		n.Items[i] = node.(*ByItem)
	}
	return v.Leave(n)
}

// HavingClause represents having clause.
type HavingClause struct {
	node
	Expr ExprNode
}

// Accept implements Node Accept interface.
func (n *HavingClause) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*HavingClause)
	node, ok := n.Expr.Accept(v)
	if !ok {
		return n, false
	}
	n.Expr = node.(ExprNode)
	return v.Leave(n)
}

// OrderByClause represents order by clause.
type OrderByClause struct {
	node
	Items    []*ByItem
	ForUnion bool
}

// Accept implements Node Accept interface.
func (n *OrderByClause) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*OrderByClause)
	for i, val := range n.Items {
		node, ok := val.Accept(v)
		if !ok {
			return n, false
		}
		n.Items[i] = node.(*ByItem)
	}
	return v.Leave(n)
}

// SelectStmt represents the select query node.
// See https://dev.mysql.com/doc/refman/5.7/en/select.html
type SelectStmt struct {
	dmlNode

	// SelectStmtOpts wraps around select hints and switches.
	*SelectStmtOpts
	// Distinct represents whether the select has distinct option.
	Distinct bool
	// From is the from clause of the query.
	From *TableRefsClause
	// Where is the where clause in select statement.
	Where ExprNode
	// Fields is the select expression list.
	Fields *FieldList
	// GroupBy is the group by expression list.
	GroupBy *GroupByClause
	// Having is the having condition.
	Having *HavingClause
	// OrderBy is the ordering expression list.
	OrderBy *OrderByClause
	// Limit is the limit clause.
	Limit *Limit
	// TableHints represents the table level Optimizer Hint for join type
	TableHints []*TableOptimizerHint
	// IsInBraces indicates whether it's a stmt in brace.
	IsInBraces bool
}

// Accept implements Node Accept interface.
func (n *SelectStmt) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}

	n = newNode.(*SelectStmt)
	if n.TableHints != nil && len(n.TableHints) != 0 {
		newHints := make([]*TableOptimizerHint, len(n.TableHints))
		for i, hint := range n.TableHints {
			node, ok := hint.Accept(v)
			if !ok {
				return n, false
			}
			newHints[i] = node.(*TableOptimizerHint)
		}
		n.TableHints = newHints
	}

	if n.Fields != nil {
		node, ok := n.Fields.Accept(v)
		if !ok {
			return n, false
		}
		n.Fields = node.(*FieldList)
	}

	if n.From != nil {
		node, ok := n.From.Accept(v)
		if !ok {
			return n, false
		}
		n.From = node.(*TableRefsClause)
	}

	if n.Where != nil {
		node, ok := n.Where.Accept(v)
		if !ok {
			return n, false
		}
		n.Where = node.(ExprNode)
	}

	if n.GroupBy != nil {
		node, ok := n.GroupBy.Accept(v)
		if !ok {
			return n, false
		}
		n.GroupBy = node.(*GroupByClause)
	}

	if n.Having != nil {
		node, ok := n.Having.Accept(v)
		if !ok {
			return n, false
		}
		n.Having = node.(*HavingClause)
	}

	if n.OrderBy != nil {
		node, ok := n.OrderBy.Accept(v)
		if !ok {
			return n, false
		}
		n.OrderBy = node.(*OrderByClause)
	}

	if n.Limit != nil {
		node, ok := n.Limit.Accept(v)
		if !ok {
			return n, false
		}
		n.Limit = node.(*Limit)
	}

	return v.Leave(n)
}

// Assignment is the expression for assignment, like a = 1.
type Assignment struct {
	node
	// Column is the column name to be assigned.
	Column *ColumnName
	// Expr is the expression assigning to ColName.
	Expr ExprNode
}

// Accept implements Node Accept interface.
func (n *Assignment) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*Assignment)
	node, ok := n.Column.Accept(v)
	if !ok {
		return n, false
	}
	n.Column = node.(*ColumnName)
	node, ok = n.Expr.Accept(v)
	if !ok {
		return n, false
	}
	n.Expr = node.(ExprNode)
	return v.Leave(n)
}

type ColumnNameOrUserVar struct {
	ColumnName *ColumnName
	UserVar    *VariableExpr
}

// InsertStmt is a statement to insert new rows into an existing table.
// See https://dev.mysql.com/doc/refman/5.7/en/insert.html
type InsertStmt struct {
	dmlNode

	IsReplace bool
	Table     *TableRefsClause
	Columns   []*ColumnName
	Lists     [][]ExprNode
	Setlist   []*Assignment
	Priority  mysql.PriorityEnum
	Select    ResultSetNode
}

// Accept implements Node Accept interface.
func (n *InsertStmt) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}

	n = newNode.(*InsertStmt)
	if n.Select != nil {
		node, ok := n.Select.Accept(v)
		if !ok {
			return n, false
		}
		n.Select = node.(ResultSetNode)
	}

	node, ok := n.Table.Accept(v)
	if !ok {
		return n, false
	}
	n.Table = node.(*TableRefsClause)

	for i, val := range n.Columns {
		node, ok := val.Accept(v)
		if !ok {
			return n, false
		}
		n.Columns[i] = node.(*ColumnName)
	}
	for i, list := range n.Lists {
		for j, val := range list {
			node, ok := val.Accept(v)
			if !ok {
				return n, false
			}
			n.Lists[i][j] = node.(ExprNode)
		}
	}
	for i, val := range n.Setlist {
		node, ok := val.Accept(v)
		if !ok {
			return n, false
		}
		n.Setlist[i] = node.(*Assignment)
	}
	return v.Leave(n)
}

// DeleteStmt is a statement to delete rows from table.
// See https://dev.mysql.com/doc/refman/5.7/en/delete.html
type DeleteStmt struct {
	dmlNode

	// TableRefs is used in both single table and multiple table delete statement.
	TableRefs *TableRefsClause
	Where     ExprNode
	Order     *OrderByClause
	Limit     *Limit
	Priority  mysql.PriorityEnum
	Quick     bool
}

// Accept implements Node Accept interface.
func (n *DeleteStmt) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}

	n = newNode.(*DeleteStmt)
	node, ok := n.TableRefs.Accept(v)
	if !ok {
		return n, false
	}
	n.TableRefs = node.(*TableRefsClause)

	if n.Where != nil {
		node, ok = n.Where.Accept(v)
		if !ok {
			return n, false
		}
		n.Where = node.(ExprNode)
	}
	if n.Order != nil {
		node, ok = n.Order.Accept(v)
		if !ok {
			return n, false
		}
		n.Order = node.(*OrderByClause)
	}
	if n.Limit != nil {
		node, ok = n.Limit.Accept(v)
		if !ok {
			return n, false
		}
		n.Limit = node.(*Limit)
	}
	return v.Leave(n)
}

// Limit is the limit clause.
type Limit struct {
	node

	Count  ExprNode
	Offset ExprNode
}

// Accept implements Node Accept interface.
func (n *Limit) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	if n.Count != nil {
		node, ok := n.Count.Accept(v)
		if !ok {
			return n, false
		}
		n.Count = node.(ExprNode)
	}
	if n.Offset != nil {
		node, ok := n.Offset.Accept(v)
		if !ok {
			return n, false
		}
		n.Offset = node.(ExprNode)
	}

	n = newNode.(*Limit)
	return v.Leave(n)
}

// ShowStmtType is the type for SHOW statement.
type ShowStmtType int

// Show statement types.
const (
	ShowNone = iota
	ShowDatabases
	ShowTables
	ShowWarnings
	ShowVariables
	ShowCreateTable
	ShowProcessList
	ShowCreateDatabase
	ShowErrors
)

// ShowStmt is a statement to provide information about databases, tables, columns and so on.
// See https://dev.mysql.com/doc/refman/5.7/en/show.html
type ShowStmt struct {
	dmlNode

	Tp          ShowStmtType // Databases/Tables/Columns/....
	DBName      string
	Table       *TableName  // Used for showing columns.
	Column      *ColumnName // Used for `desc table column`.
	IndexName   model.CIStr
	Flag        int // Some flag parsed from sql, such as FULL.
	Full        bool
	IfNotExists bool // Used for `show create database if not exists`
	Extended    bool // Used for `show extended columns from ...`

	// GlobalScope is used by `show variables` and `show bindings`
	GlobalScope bool
	Where       ExprNode
}

// Accept implements Node Accept interface.
func (n *ShowStmt) Accept(v Visitor) (Node, bool) {
	newNode, skipChildren := v.Enter(n)
	if skipChildren {
		return v.Leave(newNode)
	}
	n = newNode.(*ShowStmt)
	if n.Table != nil {
		node, ok := n.Table.Accept(v)
		if !ok {
			return n, false
		}
		n.Table = node.(*TableName)
	}
	if n.Column != nil {
		node, ok := n.Column.Accept(v)
		if !ok {
			return n, false
		}
		n.Column = node.(*ColumnName)
	}

	if n.Where != nil {
		node, ok := n.Where.Accept(v)
		if !ok {
			return n, false
		}
		n.Where = node.(ExprNode)
	}
	return v.Leave(n)
}
