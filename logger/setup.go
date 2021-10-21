/*
 * Copyright (c) 2021, salesforce.com, inc.
 * All rights reserved.
 * SPDX-License-Identifier: BSD-3-Clause
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
 */

package logger

var environ string

// Setup takes a DSN for sentry.io and creates the logging client
func Setup(env string) error {

	environ = env

	return nil
}
