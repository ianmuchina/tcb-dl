#!/bin/bash
file="projects"
uri="https://onepiecechapters.com/projects"
curl -s -o "$file" -z "$file" "$uri"

# select series to download
SERIES=$(
    cat projects |
        grep 'mb-3 text-white text-lg font-bold' |
        grep -o -P '(?<=href=).*(?=<)' |
        tr -d '"' |
        fzf -q "one" -d ">" --with-nth=2
)

# Extract series url & title
URL=$(echo "$SERIES" | cut -f1 -d ">")
URL="https://onepiecechapters.com$URL"

TITLE=$(echo "$SERIES" | cut -f2 -d ">")

# series name without spaces
SERIES_2=$(echo "$TITLE" | tr -s ' ' | tr ' ' '_')

LINKS=$(curl -s "$URL" | grep href | grep -o -P '(?<=href=).*(?= class)' | tr -d '"')

# Interactive chapter select
chapter1=$(echo "$LINKS" | fzf -d "/" --with-nth=4)
chapter=$(echo "https://onepiecechapters.com$chapter1")

# Get image list
imgs=$(curl -s $chapter | grep -o -P '(?<=src=).*(?= />)')

# Download images using alt text as filename
echo "$imgs" | while read -r l; do
    DIR=$(echo $chapter1 | cut -f4 -d"/")
    DST=$(echo "$l" | cut -f2 -d"=" | tr -s ' ' | tr ' ' '_' | xargs -I{} echo {}.png)
    SRC=$(echo "$l" | cut -f1 -d"=" | tr -d '"' | cut -f1 -d" " )
    DST="$SERIES_2/$DIR/$DST"

    if ! [ -f "$DST" ]; then
        curl --progress-bar "$SRC" --create-dirs -o "$DST"
    fi
done
