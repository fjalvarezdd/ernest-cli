/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package model

import (
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEvent(t *testing.T) {
	var d Definition

	templated, _ := ioutil.ReadFile("../internal/definitions/aws-template1-completed.yml")

	Convey("When loading a definition with supporting files", t, func() {
		p, err := ioutil.ReadFile("../internal/definitions/aws-template1.yml")
		So(err, ShouldBeNil)

		err = d.Load(p)
		So(err, ShouldBeNil)

		Convey("And I load file imports", func() {
			errimports := d.LoadFileImports()
			So(errimports, ShouldBeNil)

			Convey("It should have loaded the supporting file", func() {
				output, err := d.Save()
				So(err, ShouldBeNil)
				So(string(output), ShouldEqual, string(templated))
			})
		})
	})

}
