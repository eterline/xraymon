#!/bin/bash

# Copyright text
read -r -d '' COPYRIGHT <<'EOF'
// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.

EOF

process_file() {
    local file="$1"

    if grep -q "Copyright (c) 2025 EterLine" "$file"; then
        echo "Skipping $file (already contains copyright)"
        return
    fi

    echo "Processing $file"

    if head -n1 "$file" | grep -q '^#!'; then
        { head -n1 "$file"; echo "$COPYRIGHT"; tail -n +2 "$file"; } > "$file.tmp" && mv "$file.tmp" "$file"
    else
        { echo "$COPYRIGHT"; cat "$file"; } > "$file.tmp" && mv "$file.tmp" "$file"
    fi
}

find . -type f -name "*.go" | while read -r file; do
    process_file "$file"
done