# PDF Extractor

PDF Extractor is a CLI tool developed using Go and Cobra. It provides an easy way to manipulate PDF files, allowing you to extract specific content or modify the structure of your documents.

## Table of Contents

1. [Features](#features)
2. [Installation](#installation)
   - [Download Prebuilt Binaries](#download-prebuilt-binaries)
   - [Build from Source](#build-from-source)
   - [Prerequisites](#prerequisites)
3. [Usage](#usage)
   - [Extract Index](#extract-index)
   - [Generate PDFs for Chapters or Articles](#generate-pdfs-for-chapters-or-articles)
   - [Delete Pages from a PDF](#delete-pages-from-a-pdf)
   - [Delete PDF file](#delete-pdf)
   - [Undo Delete Operation](#undo-delete-operation)
4. [Contributing](#contributing)
5. [Contact](#contact)

## Features

- **Extract Index**: Extract authors and titles from a PDF file to config.yaml.
- **Create Chapters PDF**: Create separate PDF files for each chapter. 
- **Delete Pages**: Remove specific pages or a range of pages from a PDF.
- **Delete PDF File**: Delete an entire PDF file with optional backup
- **Undo Delete Operation**:  Restore deleted pages or files using the undo functionality.

## Installation

### Download Prebuilt Binaries
You can download the prebuilt binaries for Windows, Mac, and Linux from the [GitHub Releases](https://github.com/naman7kr/pdf-extractor/releases).

### Build from Source
To generate the binary yourself, follow these steps:
1. Clone the repository:
   ```bash
   git clone https://github.com/naman7kr/pdf-extractor.git
   cd pdf-extractor
   ```
2. Run the build script:
   ```bash
   chmod +x ./build.sh
   ./build.sh
   ```
### Prerequisites
Ensure the following dependencies are installed and set to your system's PATH:
#### For Linux and Mac:
Install pdftk and poppler-utils:
```bash
sudo apt install poppler-utils
sudo apt install poppler-utils pdftk
```
#### For Windows:
1. ***Install*** `poppler-utils`
- Download the Poppler binaries for Windows from [Poppler for Windows](https://github.com/oschwartz10612/poppler-windows).
Extract the downloaded ZIP file to a directory (e.g., C:\poppler-utils).
- Add the bin folder inside the extracted directory to your system's PATH:
    1. Open ***System Properties → Advanced → Environment Variables.***
    2. Under ***System Variables***, find the `Path` variable and click Edit.
    3. Add the path to the bin folder (e.g., `C:\poppler-utils\bin`).
2. ***Install*** `pdftk`
- Download the `pdftk` installer for Windows from [pdflabs.com](https://www.pdflabs.com/tools/pdftk-the-pdf-toolkit/)
- Run the installer and follow the instructions.
- Ensure the installation directory (e.g., `C:\Program Files (x86)\PDFtk\bin`) is added to your system's PATH.

After completing these steps, both `poppler-utils` and `pdftk` should be available for use in the command line on Windows.


## Usage

### Extract Index
The following command generates separate PDF files for all the chapters or articles in the specified PDF file:

```bash
pdf-extractor extract-index --file=$pdfFile --output-path=$outputPath
```
- ***Description***: This command scans the content page of the PDF to identify all the chapters or articles and their authors. The extracted information is saved to `config.yaml`

- ***Options***:
  - `--output-path`: Specify the directory where the output files (`config.yaml`) will be saved. Defaults to `./`.
  - `--file`(***Required***): Specify the file path which will be used for the extraction process

- ***Note:*** Currently, this command only scans the page where the content is located. If the content page spans more than one page, it will not scan the additional pages.

### Generate PDFs for Chapters or Articles
The following command generates separate PDF files for all the chapters or articles in the specified PDF file:

```bash
pdf-extractor extract --file=$pdfFile --output-path="$outputPath" --config-path="$configPath"
```

- ***Description:*** This command uses the `config.yaml` file present in `$configPath`. It scans through all the pages of the PDF `$pdfFile`, searches for the titles, and generates separate PDF files for each chapter or article.

- ***Options***:
  - `--file` (***Required***): Specify the 
  - `--output-path`: Specify the directory where the generated PDFs will be saved. Defaults to `./extracted`.
  - `--config-path`: Specify the directory containing the `articles.txt` file. Defaults to `./configs`. The file name must always be `articles.txt`.
  - `--ends-with`: Specify the text to find the page where the last article ends. If found, the last PDF will end before the page containing this text.
  - `from`: Specify the page number to start the extraction process
  - `to`: Specify the page number to end the extraction process
  - `article-title`: Enter the title of the article 

***Note: The from, to and article-title options are used to extract pdf using page range***
```bash
    ./outputs/linux/pdf-extractor extract --file=$pdfFile --output-path="$outputPath" --from=$from --to=$to --article-title="$articleTitle"
```

### Delete Pages from a PDF
The following command allows you to delete specific pages, a range of pages, or pages based on their content from a PDF file:

```bash
pdf-extractor delete-pages --file="<pdf-file>" [options]
```
***Options:***
1. ***Delete a Specific Page:***
    - Use the `--at` flag to specify the page number to delete.
    - Example:
    ```bash
    pdf-extractor delete-pages --file="example.pdf" --at=5
    ```
2. ***Delete a Range of Pages***
    - Use the `--from` and `--to` flags to specify the starting and ending page numbers.
    - If `--to` is not provided, all pages from the starting page to the end of the document will be deleted.
    - Examples:
    ```bash
    pdf-extractor delete-pages --file="example.pdf" --from=3 --to=7
    
    pdf-extractor delete-pages --file="example.pdf" --from=10
    ```
3. ***Delete Pages Based on Content:***
    - Use the `--starts-with` flag to delete pages where the content starts with the specified string.
    - If `--to` is not provided, it will delete all pages starting from the matching page to the end of the document.
    - Example:
    ```bash
    pdf-extractor delete-pages --file="example.pdf" --starts-with="Introduction"
    ```

4. ***Backup Path***:
    - Use the `--backup-path` flag to specify the directory where backups will be stored. Defaults to `./backup`.
    - Use `--no-backup` flag to skip the backup

***Constraints***
- You cannot combine the following flags in a single command:
    - `--at` with `--from`/`--to` or `--starts-with`.
    - `--from`/`--to` with `--starts-with`.
- The `--file` flag is required to specify the PDF file to operate on.

***Notes:***
- A backup of the original PDF is created before performing the delete operation.
- Ensure that the specified flags are used correctly to avoid errors.

### Delete PDF file
The following command deletes an entire PDF file:
```bash
    pdf-extractor delete --file="<pdf-file>" [options]
```
- **Options**
1. `--file`: Path to the PDF file (required)
2. `--backup-path`: Specify the directory where backups will be stored. Defaults to `./backup`
3. `--no-backup`: Skip creating a backup before deleting the file.

### Undo Delete Operation
The following command restores the previous state of a PDF file by using the backup stored in the backup folder:

```bash
pdf-extractor undo-pdf --file="<pdf-file>" --backup-path="<backup-directory>"
```
- ***Description:*** This command reverts the PDF file to its state before the last delete operation by using the backup stored in the backup folder.

- ***Options***:
  - `--backup-path`: Specify the directory where backups are stored. Defaults to `./backup`.

- ***Backup Management:***
    - The tool automatically creates a backup of the PDF before performing any delete operation.
    - You can set the `BACKUP_CAPACITY` environment variable to limit the total number of backup PDFs stored in the backup folder.
    - Example:
    ```bash
    export BACKUP_CAPACITY=5
    ```
- ***Note:*** Ensure that the backup folder contains the required backup file for the undo operation to succeed.


## Contributing

Contributions are welcome! If you'd like to contribute to this project, please follow these steps:

1. Fork the repository on GitHub.
2. Create a new branch for your feature or bug fix:
   ```bash
   git checkout -b feature-name
   ```
3. Make your changes and commit them with clear and concise messages.
4. Push your changes to your forked repository:
    ```bash
    git push origin feature-name
    ```
5. Open a pull request to the main repository, describing your changes in detail.

Please ensure your code adheres to the project's coding standards and includes appropriate tests.

## Contact

For any questions, feedback, or support, feel free to reach out to the maintainers:

- [Kumar Naman](https://github.com/naman7kr)
- [Kumar Aman](https://github.com/aman54kumar)