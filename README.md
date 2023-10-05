# GitOps

`gitops` is a cli for generating and pushing git patches to a git
repository.

```bash
# Edit file using sed
gitops format-patch https://example.com/somerepo.git#path/to/file.txt \
  --sed 's|^version=".*"$|version="v2"|' \
  --message 'update version to v2' \
  --out file.patch

# Edit file using yq
gitops format-patch https://example.com/somerepo.git#path/to/file.yaml \
  --yq '.a.b.c = "v2"' \
  --message 'update version to v2' \
  --out file.patch

# Edit file using jq
gitops format-patch https://example.com/somerepo.git#path/to/file.json \
  --jq '.a.b.c = "v2"' \
  --message 'update version to v2' \
  --out file.patch

# Override file with input file
gitops format-patch https://example.com/somerepo.git#path/to/file.txt \
  --file file.txt \
  --message 'update version to v2' \
  --out file.patch

# Push patch file to remote repo
gitops push-patch https://example.com/somerepo.git \
  --patch file.patch
```
