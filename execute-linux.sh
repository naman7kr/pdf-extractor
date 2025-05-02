#!/bin/bash
# sudo apt install poppler-utils
# sudo apt install poppler-utils pdftk

# ./outputs/linux/pdf-extractor get chapters --file="Vol. 14. N. 2, 2022.pdf"
# ./outputs/linux/pdf-extractor extract --file="Vol. 14. N. 2, 2022.pdf"
# ./outputs/linux/pdf-extractor delete-pages --file="Vol. 14. N. 2, 2022.pdf" --from=1 --to=2 

# Read the last article from articles.txt
# last_article=$(tail -n 1 articles.txt)
# Replace spaces and special characters in the article title to create a valid filename
# file_name=$(echo "$last_article" | sed 's/ /_/g').pdf

# Run the delete-pages command with the dynamically generated file name
# ./outputs/linux/pdf-extractor delete-pages --file="extracted/$file_name" --starts-with="Guidelines for Contributors"

# ./outputs/linux/pdf-extractor undo-pdf --file="extracted/The_Role_of_MSMEs_in_the_Indian_Economy.pdf"