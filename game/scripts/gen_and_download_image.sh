#!/usr/bin/env zsh
set -euo pipefail

# Ensure server is up before proceeding (optional quick check)
if ! curl -sS http://localhost:8080/api/state >/dev/null 2>&1; then
  echo "Server not responding at http://localhost:8080. Start it first: go run . web 8080" >&2
  exit 1
fi

# 1) Start + create an event
curl -sS -X POST http://localhost:8080/api/start -H 'Content-Type: application/json' -d '{}' >/dev/null
curl -sS -X POST http://localhost:8080/api/new-round -H 'Content-Type: application/json' -d '{}' >/dev/null

# 2) Generate image
IMG_JSON=$(curl -sS -X POST http://localhost:8080/api/generate-image -H 'Content-Type: application/json' -d '{}')

# 3) Extract imageUrl
URL=$(python3 - <<'PY' <<<"$IMG_JSON"
import json,sys
try:
    d=json.loads(sys.stdin.read())
    print(d.get('imageUrl') or d.get('image_url') or '')
except Exception:
    print('')
PY
)

if [[ -z "$URL" ]]; then
  echo "No imageUrl in response:" >&2
  echo "$IMG_JSON" >&2
  exit 1
fi

OUTDIR="/Users/lap/AgentStageFramework/game"
mkdir -p "$OUTDIR"
TS=$(date +%s)
OUT="$OUTDIR/generated_image_${TS}"

# 4) Download with wget or decode data URL
if [[ "$URL" == data:* ]]; then
  python3 - "$OUT" <<'PY' <<<"$URL"
import sys,base64,re
path=sys.argv[1]; url=sys.stdin.read().strip()
m=re.match(r'data:(.*?);base64,(.*)', url)
mime=m.group(1) if m else 'image/png'
b64=m.group(2) if m else ''
ext='png' if 'png' in mime else ('jpg' if 'jpeg' in mime or 'jpg' in mime else ('webp' if 'webp' in mime else 'bin'))
with open(path+'.'+ext,'wb') as f: f.write(base64.b64decode(b64))
print(path+'.'+ext)
PY
else
  EXT="${URL##*.}"; [[ "$EXT" == "$URL" ]] && EXT="webp"
  FILE="${OUT}.${EXT}"
  if command -v wget >/dev/null 2>&1; then
    wget -q -O "$FILE" "$URL"
  else
    echo "wget not found; using curl"
    curl -sSL "$URL" -o "$FILE"
  fi
  echo "$FILE"
fi
