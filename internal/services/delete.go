package services

import "pdf-extractor/internal/utils"

func Delete(file string, backupPath string, backupFlag bool) error {
	err := utils.CheckFileExists(file)
	if err != nil {
		return err
	}
	err = utils.CheckFileExists(backupPath)
	if err != nil {
		return err
	}
	if backupFlag {
		// delete file "file" move it to the backup path
		err = utils.CreateBackup(file, backupPath)
		if err != nil {
			return err
		}
	}
	// delete the file file
	err = utils.DeleteFile(file)
	if err != nil {
		return err
	}
	return nil
}
