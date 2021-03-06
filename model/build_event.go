/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package model

// BuildEvent represents an event of type build from an SSE stream
type BuildEvent struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Subject string           `json:"_subject"`
	Changes []ComponentEvent `json:"changes"`
}
