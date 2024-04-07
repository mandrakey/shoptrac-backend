/*
SPDX-FileCopyrightText: Maurice Bleuel <mandrakey@litir.de>
SPDX-License-Identifier: BSD-3-Clause
*/

package handler

import (
	"github.com/mandrakey/shoptrac/config"
)

func GetVersion() (int, string) {
	return 200, SuccessResponse(map[string]interface{}{"version": config.AppVersion})
}
