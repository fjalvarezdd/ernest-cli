/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package manager

import (
	"encoding/json"
	"errors"

	"github.com/ernestio/ernest-cli/model"
)

// ListLoggers : Lists all active loggers
func (m *Manager) ListLoggers(token string) (loggers []model.Logger, err error) {
	body, resp, err := m.doRequest("/api/loggers/", "GET", []byte(""), token, "")
	if err != nil {
		if resp == nil {
			return nil, ErrConnectionRefused
		}
		return nil, err
	}
	err = json.Unmarshal([]byte(body), &loggers)
	if err != nil {
		return nil, err
	}
	return loggers, err
}

// SetLogger : Setup a specific loger
func (m *Manager) SetLogger(token string, logger model.Logger) (err error) {
	body, err := json.Marshal(logger)
	if err != nil {
		return err
	}
	if body, resp, err := m.doRequest("/api/loggers/", "POST", body, token, ""); err != nil {
		if resp == nil {
			return ErrConnectionRefused
		}
		if resp.StatusCode == 403 {
			return errors.New("You're not allowed to perform this action, please log in with an admin account")
		}

		return errors.New(string(body))
	}

	return nil
}

// DelLogger : Deletes a specific loger
func (m *Manager) DelLogger(token string, logger model.Logger) (err error) {
	body, err := json.Marshal(logger)
	if err != nil {
		return err
	}
	if body, resp, err := m.doRequest("/api/loggers/"+logger.Type, "DELETE", body, token, ""); err != nil {
		if resp == nil {
			return ErrConnectionRefused
		}
		if resp.StatusCode == 403 {
			return errors.New("You're not allowed to perform this action, please log in with an admin account")
		}
		return errors.New(string(body))
	}

	return nil
}
