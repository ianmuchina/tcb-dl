#!/bin/bash
git config user.name "Github Actions"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

git add projects.* &&
    git commit -F "Update" &&
    git push origin v2 &&
    echo "Done" || echo "No Changes"