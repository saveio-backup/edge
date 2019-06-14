/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2018-11-27
 */
package utils

import (
	"strings"

	"github.com/urfave/cli"
)

//GetFlagName deal with short flag, and return the flag name whether flag name have short name
func GetFlagName(flag cli.Flag) string {
	name := flag.GetName()
	if name == "" {
		return ""
	}

	return strings.TrimSpace(strings.Split(name, ",")[0])
}
