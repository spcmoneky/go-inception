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

package session

import (
	"strings"
	"testing"

	"github.com/hanchuanchuan/goInception/ast"
	"github.com/hanchuanchuan/goInception/format"
	"github.com/hanchuanchuan/goInception/parser"
	"github.com/hanchuanchuan/goInception/sessionctx/variable"
)

func Test_buildInstantAlterSQL(t *testing.T) {
	parser := parser.New()
	sessionVars := variable.NewSessionVars()
	parser.SetSQLMode(sessionVars.SQLMode)

	tests := []struct {
		sql     string
		ok      bool
		want    string
		restore string
	}{
		{
			sql:     "alter table t1 add column c1 int;",
			ok:      true,
			want:    "ALTER TABLE `t1` ADD COLUMN `c1` INT, ALGORITHM = INSTANT",
			restore: "ALTER TABLE `t1` ADD COLUMN `c1` INT",
		},
		{
			sql:     "alter table t1 add column c1 int, algorithm=instant;",
			ok:      true,
			want:    "ALTER TABLE `t1` ADD COLUMN `c1` INT, ALGORITHM = INSTANT",
			restore: "ALTER TABLE `t1` ADD COLUMN `c1` INT, ALGORITHM = INSTANT",
		},
		{
			sql: "alter table t1 add column c1 int, algorithm=inplace;",
			ok:  false,
		},
	}

	for _, test := range tests {
		stmts, _, err := parser.Parse(test.sql, "", "")
		if err != nil {
			t.Fatal(err)
		}
		node := stmts[0].(*ast.AlterTableStmt)
		instantSQL, ok, err := buildInstantAlterSQL(node)
		if err != nil {
			t.Fatal(err)
		}
		if ok != test.ok {
			t.Fatalf("ok is %v, expected %v: %s", ok, test.ok, test.sql)
		}
		if instantSQL != test.want {
			t.Fatalf("sql is %s, expected %s", instantSQL, test.want)
		}
		if test.restore != "" {
			var builder strings.Builder
			err = node.Restore(format.NewRestoreCtx(format.DefaultRestoreFlags, &builder))
			if err != nil {
				t.Fatal(err)
			}
			if builder.String() != test.restore {
				t.Fatalf("restored sql is %s, expected %s", builder.String(), test.restore)
			}
		}
	}
}

func Test_checkDDLInstantMySQL(t *testing.T) {
	parser := parser.New()
	sessionVars := variable.NewSessionVars()
	parser.SetSQLMode(sessionVars.SQLMode)

	table := &TableInfo{Name: "t1",
		Fields: []FieldInfo{
			{Field: "c1", Extra: ""},
			{Field: "c2", Extra: "VIRTUAL"},
			{Field: "c3", Extra: "STORED"}}}

	tests := []struct {
		sql     string
		mysql57 bool
		mysql80 bool
	}{
		{"alter table t1 add column c1 int;", false, true},
		{"ALTER TABLE t1 ADD COLUMN c2 INT GENERATED ALWAYS AS (c1 + 1) VIRTUAL;", true, true},
		{"ALTER TABLE t1 ADD COLUMN c4 INT AFTER c1;", false, false},
		{"ALTER TABLE t1 ADD COLUMN c3 INT GENERATED ALWAYS AS (c1 + 1) STORED;", false, false},
	}

	for _, test := range tests {
		stmts, _, err := parser.Parse(test.sql, "", "")
		if err != nil {
			t.Error(err)
		}
		for _, stmtNode := range stmts {
			switch node := stmtNode.(type) {
			case *ast.AlterTableStmt:
				canInstant := checkDDLInstantMySQL57(node)
				if canInstant != test.mysql57 {
					t.Errorf("canInstant is %v, but excepted %v: sql: %v", canInstant, test.mysql57, test.sql)
				}

				canInstant = checkDDLInstantMySQL80(node, table, 80000)
				if canInstant != test.mysql80 {
					t.Errorf("canInstant is %v, but excepted %v: sql: %v", canInstant, test.mysql57, test.sql)
				}
			}
		}
	}
}
