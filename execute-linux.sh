#!/bin/bash
# sudo apt install poppler-utils
# sudo apt install poppler-utils pdftk

# all available commands
# help command
# ./outputs/linux/pdf-extractor --help

# version command
# ./outputs/linux/pdf-extractor -v
# or
# ./outputs/linux/pdf-extractor --version

# extract-index command
# pdfFile=<file-name>
# outputPath=<output-path>
# configPath=<config-path>

# default output path is ./
# ./outputs/linux/pdf-extractor extract-index --file=$pdfFile


# you can manually specify the output path
# ./outputs/linux/pdf-extractor extract-index --file=$pdfFile --output-path=$outputPath

# extract pdf command to extract using pdf config file
# ./outputs/linux/pdf-extractor extract --file=$pdfFile --output-path="$outputPath" --config-path="$configPath"

# you can specify the prefix of the page of the last article where it ends
# endsWith="Guidelines for Contributors"
# ./outputs/linux/pdf-extractor extract --file=$pdfFile --output-path="$outputPath" --config-path="$configPath" --ends-with="$endsWith"


# you can specify the range of pages to extract the pdf, make sure to specify the article title
# from=11
# to=13
# articleTitle="PERFORMANCE APPRAISAL OF UTTAR PRADESH POWER CORPORATION LIMITED"
# ./outputs/linux/pdf-extractor extract --file=$pdfFile --output-path="$outputPath" --from=$from --to=$to --article-title="$articleTitle"


# delete-pages command to delete pages from the pdf
# you can specify the range of pages to delete
# from=1
# to=2
# ./outputs/linux/pdf-extractor delete-pages --file=$pdfFile --from=$from --to=$to

# you can specify a specific page to delete
# at=2
# ./outputs/linux/pdf-extractor delete-pages --file=$pdfFile --at=$at

# you can specify starts-with to delete pages from the start of a content in pdf
# startsWith="Guidelines for Contributors"
# ./outputs/linux/pdf-extractor delete-pages --file=$pdfFile --starts-with="$startsWith"

# you can also specify 'to' flag to delete pages from the start of a content to 'to' page 
# to=2
# ./outputs/linux/pdf-extractor delete-pages --file=$pdfFile --starts-with="$startsWith" --to=$to
# you can specify backup path to save the original pdf before deleting pages. if you don't specify backup path, it will default to ./backup
# backupPath="./mybackups"
# ./outputs/linux/pdf-extractor delete-pages --file=$pdfFile --starts-with="$startsWith" --backup-path="$backupPath"

# you can skip the backup using --no-backup flag, Note this flag is valid for all delete-pages commands
# ./outputs/linux/pdf-extractor delete-pages --file=$pdfFile --starts-with="$startsWith" --no-backup

# you can use delete command to delete pdf file 
# ./outputs/linux/pdf-extractor delete --file=$pdfFile

# you can specify backup path to save the original pdf before deleting the file. if you don't specify backup path, it will default to ./backup
# backupPath="./mybackups"
# ./outputs/linux/pdf-extractor delete --file=$pdfFile --backup-path="$backupPath"

# you can skip the backup using --no-backup flag, Note this flag is valid for all delete commands
# ./outputs/linux/pdf-extractor delete --file=$pdfFile --no-backup


# you can undo the delete and delete pages command using undo
# ./outputs/linux/pdf-extractor undo --file=$pdfFile

# you can specify backup path to restore the original pdf before deleting the file or pages. if you don't specify backup path, it will default to ./backup
# backupPath="./mybackups"
# ./outputs/linux/pdf-extractor undo --file=$pdfFile --backup-path="$backupPath"