/*
SPDX-FileCopyrightText: Maurice Bleuel <mandrakey@litir.de>
SPDX-License-Identifier: BSD-3-Clause
*/

package config

import (
	"fmt"
	"regexp"

	"gopkg.in/macaron.v1"
)

var (
	rxAddress = regexp.MustCompile("(\\S+):\\d+$")
	allowed   = make([]*regexp.Regexp, 0)
	denied    = make([]*regexp.Regexp, 0)
)

func IpFilterer(c *AppConfig) macaron.Handler {
	log := Logger()

	for _, rule := range c.AccessPolicy.Rules {
		rx, err := regexp.Compile(rule.Origin)
		if err != nil {
			log.Warningf("Failed to parse access rule '%s'; skipping.", rule.Origin)
			continue
		}

		switch rule.Policy {
		case AccessAllow:
			allowed = append(allowed, rx)
			break

		case AccessDeny:
			denied = append(denied, rx)
			break

		default:
			log.Warningf("Invalid access policy for origin '%s': %d; skipping.", rule.Origin, rule.Policy)
		}
	}

	return func(ctx *macaron.Context) error {
		m := rxAddress.FindStringSubmatch(ctx.Req.RemoteAddr)
		if m == nil {
			return fmt.Errorf("address %s needs to match address:port layout", ctx.Req.RemoteAddr)
		}

		if isValidIp(m[1], &c.AccessPolicy) {
			return nil
		} else {
			return fmt.Errorf("origin address '%s' is not allowed", m[1])
		}
	}
}

func isValidIp(address string, ap *AccessPolicy) bool {
	log := Logger()

	for _, rx := range denied {
		log.Debugf("IPFILTER checking %s against %s", address, rx)
		if rx.MatchString(address) {
			log.Debug("... matched as denied")
			return false
		}
	}

	for _, rx := range allowed {
		log.Debugf("IPFILTER checking %s against %s", address, rx)
		if rx.MatchString(address) {
			log.Debug("... matched as allowed")
			return true
		}
	}

	log.Debugf("IPFILTER return default policy %d", ap.Default)
	return ap.Default == AccessAllow
}
