#!/bin/bash
SERIES=$(
    cat projects |
        grep 'mb-3 text-white text-lg font-bold' |
        grep -o -P '(?<=href=).*(?=<)' |
        tr -d '"' |
        fzf -q "one" -d ">" --with-nth=2
)
URL=$(echo "$SERIES" | cut -f1 -d ">")
TITLE=$(echo "$SERIES" | cut -f2 -d ">")

URL="https://onepiecechapters.com$URL"

LINKS=$(curl -q $URL | grep href | grep -o -P '(?<=href=).*(?= class)')

echo "$LINKS" | fzf -d "/" --with-nth=4