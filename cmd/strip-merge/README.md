# Strip merge() from inputs in a terragrunt.hcl

This tool is a small command-line utility written that **simplifies HCL/Terragrunt configuration files** by **rewriting `merge(...)` expressions** into their final object form when possible.

It is designed for **safe, syntax-aware transformation of HCL**, without relying on regexes or string hacks.

## What This Tool Does

The program reads an HCL file from **stdin**, analyzes it using the official **HashiCorp HCL v2 syntax AST**, and prints the following transformation to **stdout**.

### Supported Transformation

If it encounters a root-level attribute defined as:

```hcl
inputs = merge(
  expr1,
  expr2,
  {
    key1 = value1
    key2 = value2
  }
)
```

...it rewrites it into:

```hcl
inputs = {
  key1 = value1
  key2 = value2
}
```

In other words:

- it removes the merge(...) call
- selects an object literal argument of the largest size (by source range), regardless of its position
- rewrites the attribute only if at least one object literal is present

### Safety Guarantees

- If an attribute is not a merge(...) call, it is left unchanged
- If a merge(...) call contains no object literals, no rewrite happens
- If an attribute value is already an object, the file remains untouched
- All transformations are syntax-aware, not text-based

### Formatting

The output is always formatted using HashiCorp’s canonical HCL formatter:

- consistent indentation
- aligned =
- normalized braces and spacing

This guarantees the same formatting style as:

- terraform fmt
- terragrunt hcl fmt

### Example

#### Input

```hcl
inputs = merge(
  yamldecode(file(find_in_parent_folders("env.yaml"))),
  { a = 1 },
  {
    dei_backend_address = local.vars.global.esb.dev05.ip
    dei_backend_port    = "8080"
  },
  yamldecode(file("${get_terragrunt_dir()}/../service.yaml")),
)
```

#### Output

```hcl
inputs = {
  dei_backend_address = local.vars.global.esb.dev05.ip
  dei_backend_port    = "8080"
}
```

### Usage

```sh
strip-merge < terragrunt.hcl > stripped.hcl
```

### Current Limitations

- Only root-level attributes are processed
- Nested blocks (locals, dependency, etc.) are not traversed
- Only merge(...) expressions are analyzed

These are intentional design choices to keep the tool small, predictable, and safe.

### Intended Use Case

```sh
strip-merge < terragrunt.hcl | hcl2json | jq --raw-output '.inputs | keys | join("\n")'
```
