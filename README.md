# encrypted-paper

Utility to compress, encrypt and make data printable

## Usage

```bash
# Encode
podman run -it -v"$(pwd):/host:z" --workdir /host docker.io/jenswbe/encrypted-paper encode --title "Very important file" -o secret.pdf secret.png

# Decode
# Assuming PDF was rescanned into multiple *.jpg files
podman run -it -v"$(pwd):/host:z" --workdir /host docker.io/jenswbe/encrypted-paper decode -o secret.png scan-*.jpg
```
