/**
 * OpenBmclAPI (Golang Edition)
 * Copyright (C) 2023 Kevin Z <zyxkad@gmail.com>
 * All rights reserved
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published
 *  by the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	_ "embed"
	"fmt"
)

const cliHint = `
Go-OpenMCIM based on Go-OpenBMCLAPI, edited by WetemCloud <master@wetem.cn>
Go-OpenBmclAPI  Copyright (C) 2023 Kevin Z <zyxkad@gmail.com>

This program comes with ABSOLUTELY NO WARRANTY;
This is free software, and you are welcome to redistribute it under certain conditions;
Use subcommand 'license' for more information

`

func printShortLicense() {
	fmt.Print(cliHint)
}

//go:embed LICENSE
var fullLicense string

func printLongLicense() {
	fmt.Println(fullLicense)
}
