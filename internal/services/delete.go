package services

import "pdf-extractor/internal/utils"

func Delete(file string, backupPath string) error {
	err := utils.CheckFileExists(file)
	if err != nil {
		return err
	}
	err = utils.CheckFileExists(backupPath)
	if err != nil {
		return err
	}
	err = utils.CreateBackup(file, backupPath)
	// delete file "file" and move it to the backup path
	if err != nil {
		return err
	}
	// delete the file file
	err = utils.DeleteFile(file)
	if err != nil {
		return err
	}
	return nil
}
