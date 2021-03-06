/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package manager

import (
	"encoding/json"
)

// FindComponents ...
func (m *Manager) FindComponents(token, project, component, service string) (components []interface{}, err error) {
	body, resp, err := m.doRequest("/api/components/"+component+"/?project="+project+"&service="+service, "GET", []byte(""), token, "")
	if err != nil {
		if resp == nil {
			return nil, ErrConnectionRefused
		}
		return nil, err
	}
	err = json.Unmarshal([]byte(body), &components)
	if err != nil {
		return nil, err
	}
	return components, err
}
