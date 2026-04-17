#!/usr/bin/env python3
"""
Scaffold a per-resource or per-datasource documentation template.

Usage:
    python3 scripts/scaffold_templates.py resource <name>
    python3 scripts/scaffold_templates.py data-source <name>

Examples:
    python3 scripts/scaffold_templates.py resource my_new_resource
    python3 scripts/scaffold_templates.py data-source my_new_datasource

Generates a .md.tmpl file in the appropriate templates/ subdirectory with
placeholder values for subcategory, API docs URL, and provider source URL.
The contributor then fills in the actual values.

Existing templates are never overwritten.
"""

import os
import sys

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
TEMPLATES_DIR = os.path.join(REPO_ROOT, "templates")


RESOURCE_TEMPLATE = """\
---
page_title: "{{{{.Type}}}}: {{{{.Name}}}}"
subcategory: "TODO"
description: |-
{{{{ if .Description }}}}
{{{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}}}
{{{{ else }}}}
  Terraform {{{{.Type}}}} for {{{{.Name}}}}.
{{{{ end }}}}
---

# {{{{.Type}}}}: {{{{.Name}}}}

{{{{ if .Description }}}}
{{{{ .Description | trimspace }}}}
{{{{ else }}}}
Terraform {{{{.Type}}}} for {{{{.Name}}}}.
{{{{ end }}}}

## Links

- [Okta API docs](TODO)
- [Provider source](https://github.com/okta/terraform-provider-okta/blob/master/okta/services/idaas/resource_okta_{name}.go)

{{{{ if .HasExample -}}}}

## Example Usage

{{{{ tffile .ExampleFile }}}}

{{{{- end }}}}

{{{{ .SchemaMarkdown | trimspace }}}}

{{{{ if .HasImport -}}}}

## Import

Import is supported using the following syntax:

{{{{ codefile "shell" .ImportFile }}}}

{{{{- end }}}}
"""

DATASOURCE_TEMPLATE = """\
---
page_title: "{{{{.Type}}}}: {{{{.Name}}}}"
subcategory: "TODO"
description: |-
{{{{ if .Description }}}}
{{{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}}}
{{{{ else }}}}
  Terraform {{{{.Type}}}} for {{{{.Name}}}}.
{{{{ end }}}}
---

# {{{{.Type}}}}: {{{{.Name}}}}

{{{{ if .Description }}}}
{{{{ .Description | trimspace }}}}
{{{{ else }}}}
Terraform {{{{.Type}}}} for {{{{.Name}}}}.
{{{{ end }}}}

## Links

- [Okta API docs](TODO)
- [Provider source](https://github.com/okta/terraform-provider-okta/blob/master/okta/services/idaas/data_source_okta_{name}.go)

{{{{ if .HasExample -}}}}

## Example Usage

{{{{ tffile .ExampleFile }}}}

{{{{- end }}}}

{{{{ .SchemaMarkdown | trimspace }}}}
"""


def main():
    if len(sys.argv) != 3 or sys.argv[1] not in ("resource", "data-source"):
        print(__doc__.strip())
        sys.exit(1)

    doc_type = sys.argv[1]
    name = sys.argv[2]

    if doc_type == "resource":
        tmpl_dir = os.path.join(TEMPLATES_DIR, "resources")
        content = RESOURCE_TEMPLATE.format(name=name)
    else:
        tmpl_dir = os.path.join(TEMPLATES_DIR, "data-sources")
        content = DATASOURCE_TEMPLATE.format(name=name)

    tmpl_path = os.path.join(tmpl_dir, f"{name}.md.tmpl")

    if os.path.exists(tmpl_path):
        print(f"Template already exists: {tmpl_path}")
        sys.exit(1)

    os.makedirs(tmpl_dir, exist_ok=True)
    with open(tmpl_path, "w") as f:
        f.write(content)

    print(f"Created: {tmpl_path}")
    print()
    print("Next steps:")
    print(f"  1. Replace TODO values in {tmpl_path}")
    print(f"  2. Run: go generate ./...")


if __name__ == "__main__":
    main()
