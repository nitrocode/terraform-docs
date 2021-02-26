/*
Copyright 2021 The terraform-docs Authors.

Licensed under the MIT license (the "License"); you may not
use this file except in compliance with the License.

You may obtain a copy of the License at the LICENSE file in
the root directory of this source tree.
*/

package template

import (
	"fmt"

	"github.com/terraform-docs/terraform-docs/internal/print"
)

// createAnchor
func createAnchor(s string, t string, settings *print.Settings) string {
	var anchor string

	if settings.ShowAnchors {
		anchorName := fmt.Sprintf("%s_%s", t, s)
		anchor = fmt.Sprintf("<a name=\"%s\"></a> [%s](#%s)", anchorName, s, anchorName)
	} else {
		anchor = s
	}

	return anchor
}
