/*
SPDX-FileCopyrightText: Maurice Bleuel <mandrakey@litir.de>
SPDX-License-Identifier: BSD-3-Clause
*/

package handler

import (
	"testing"
)

func TestFormatSum(t *testing.T) {
	res, err := FormatSum("999.99")
	if err != nil {
		t.Error("999.99 returns error")
	}
	if res != "999.99" {
		t.Errorf("999.99 returns '%s' instead of expected '999.99'", res)
	}

	res, err = FormatSum("2")
	if err != nil {
		t.Error("'2' returns error")
	}
	if res != "2.00" {
		t.Errorf("'2' returns '%s' instead of expected '2.00'", res)
	}

	res, err = FormatSum("156.4")
	if err != nil {
		t.Error("'156.4' returns error")
	}
	if res != "156.40" {
		t.Errorf("'156.4' returns '%s' instead of expected '156.40'", res)
	}

	res, err = FormatSum("no number")
	if err == nil {
		t.Error("'no number' returns no error")
	}
}
