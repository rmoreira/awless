/*
Copyright 2017 WALLIX

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

package ast

import (
	"reflect"
	"testing"
)

func TestCloneAST(t *testing.T) {
	tree := &AST{}

	tree.Statements = append(tree.Statements, &Statement{Node: &DeclarationNode{
		Ident: "myvar",
		Expr: &CommandNode{
			Action: "create", Entity: "vpc",
			Params: map[string]CompositeValue{"count": &interfaceValue{val: 1}, "myname": &referenceValue{ref: "name"}},
		}}}, &Statement{Node: &DeclarationNode{
		Ident: "myothervar",
		Expr: &CommandNode{
			Action: "create", Entity: "subnet",
			Params: map[string]CompositeValue{"vpc": &holeValue{hole: "myvar"}},
		}}}, &Statement{Node: &CommandNode{
		Action: "create", Entity: "instance",
		Params: map[string]CompositeValue{"subnet": &holeValue{hole: "myothervar"}},
	}},
	)

	clone := tree.Clone()

	if got, want := clone, tree; !reflect.DeepEqual(got, want) {
		t.Fatalf("\ngot %#v\n\nwant %#v", got, want)
	}

	clone.Statements[0].Node.(*DeclarationNode).Expr.(*CommandNode).Params["new"] = &interfaceValue{"mynode"}

	if got, want := clone.Statements, tree.Statements; reflect.DeepEqual(got, want) {
		t.Fatalf("\ngot %s\n\nwant %s", got, want)
	}
}
