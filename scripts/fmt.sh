#!/usr/bin/env bash
# Format & lint local — replica o pipeline do .github/workflows/format.yml,
# pulando tmp/ (clone do upstream) e ntgcalls/ (binding cgo gerado).
#
# Uso:  bash scripts/fmt.sh        # formata + lint
#       bash scripts/fmt.sh --check # só checa (CI-style), não escreve
set -euo pipefail

cd "$(dirname "$0")/.."
export PATH="$(go env GOPATH)/bin:$PATH"

CHECK=0
[[ "${1:-}" == "--check" ]] && CHECK=1

need() { command -v "$1" >/dev/null 2>&1; }

if ! need goimports || ! need gofumpt || ! need gci || ! need golines || ! need staticcheck; then
	echo "▶ instalando toolchain..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/daixiang0/gci@latest
	go install github.com/segmentio/golines@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
fi

mapfile -t FILES < <(find . -name '*.go' -not -path './tmp/*' -not -path './ntgcalls/*')

if [[ $CHECK -eq 1 ]]; then
	echo "▶ gofmt -l (check)"
	out=$(gofmt -l "${FILES[@]}")
	if [[ -n "$out" ]]; then echo "Arquivos não formatados:"; echo "$out"; exit 1; fi
	echo "▶ go vet"; go vet ./...
	echo "▶ staticcheck"; staticcheck ./...
	echo "✓ tudo limpo"
	exit 0
fi

echo "▶ go mod tidy";  go mod tidy
echo "▶ gofmt";        gofmt -w "${FILES[@]}"
echo "▶ goimports";    goimports -w "${FILES[@]}"
echo "▶ gci";          gci write -s standard -s default -s "prefix(main)" "${FILES[@]}"
echo "▶ gofumpt";      gofumpt -w -extra "${FILES[@]}"
echo "▶ golines";      golines -w --max-len=90 "${FILES[@]}"
echo "▶ go mod tidy";  go mod tidy
echo "▶ go vet";       go vet ./... || true
echo "▶ staticcheck";  staticcheck ./... || true
echo "✓ done"
