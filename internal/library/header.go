package library

func SetSharedComplete(path string, complete bool) error {
	header, records, err := readStream(path)
	if err != nil {
		return err
	}
	header["shared_complete"] = complete
	return writeStream(path, header, records)
}
