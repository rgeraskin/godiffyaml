# godiffyaml

Human-readable diffs for yamls with multiple documents.

![Demo](./demo.gif)

**godiffyaml** is a command-line tool for generating pretty diffs of YAML files, with special support for Kubernetes manifests and sorting YAML documents. It leverages [difftastic](https://github.com/Wilfred/difftastic) for visually rich and accurate diffs.

It can split multi-document YAML files into documents and diff them document-by-document using custom keys. This makes the diff to be human-readable and easier to understand.

## Features

- **Diff YAML files**: Split multi-document YAML files and diff them document-by-document using custom keys.
- **Kubernetes-aware diff**: Diff Kubernetes manifests using `apiVersion`, `kind`, `metadata.namespace`, and `metadata.name` as document keys.
- **Sort YAML documents**: Output YAML files with documents sorted by specified fields.
- **Rich diff output**: Uses difftastic for side-by-side, inline, or JSON diff views.

## Installation

1. Install [difftastic](https://github.com/Wilfred/difftastic) and ensure `difft` is in your `$PATH`.
2. Build godiffyaml:

   ```sh
   git clone https://github.com/rgeraskin/godiffyaml.git
   cd godiffyaml
   go build -o godiffyaml ./cmd/godiffyaml
   ```

## Usage

```
godiffyaml <subcommand> [options] <file(s)>
```

### Subcommands

#### `diff`

Diff two YAML files by splitting them into documents and comparing each document by custom keys.

```
godiffyaml diff --paths=key1,key2 <file1.yaml> <file2.yaml>
```

- `--paths`: Comma-separated list of YAML paths to use as document keys (required).
- `--display`: Output format (`side-by-side`, `side-by-side-show-both`, `inline`, `json`). Default: `side-by-side-show-both`.

#### `k8s`

Diff two Kubernetes YAML files using `apiVersion,kind,metadata.namespace,metadata.name` as keys.

```
godiffyaml k8s <file1.yaml> <file2.yaml>
```

- `--display`: Output format (see above).

#### `sort`

Sort documents in a YAML file by specified fields and print to stdout.

```
godiffyaml sort --order=field1,field2 <file.yaml>
```

- `--order`: Comma-separated list of fields to sort by.

### Passing difftastic options

All unrecognized flags are passed directly to difftastic (except `--display`). Use `--flag=value` notation.

## Example

```sh
godiffyaml diff --paths=kind,metadata.name a.yaml b.yaml
godiffyaml k8s a.yaml b.yaml
godiffyaml sort --order=kind,metadata.name a.yaml
```
