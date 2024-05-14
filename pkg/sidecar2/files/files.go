package files

import "github.com/direktiv/direktiv/pkg/refactor/engine"

func WriteFiles(location string, files []engine.FunctionFileDefinition) error {
	// // Create the target directory with appropriate permissions.
	// if err := os.MkdirAll(location, 0o750); err != nil {
	// 	return err
	// }

	// // Process each file definition.
	// for _, f := range files {
	// 	path := filepath.Join(location, f.Key)
	// 	data, err := base64.StdEncoding.DecodeString(f.Content)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// Handle different file types.
	// 	switch f.Type {
	// 	case "plain":
	// 		// Write plain text content.
	// 		if err := os.WriteFile(path, data, 0o640); err != nil {
	// 			return err
	// 		}

	// 	case "base64":
	// 		if err := os.WriteFile(path, data, 0o640); err != nil {
	// 			return err
	// 		}

	// 	case "tar":
	// 		// Extract the contents of a TAR archive.
	// 		buf := bytes.NewBuffer(data)
	// 		err := untar(location, f.Permissions, buf)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	case "tar.gz":
	// 		// Call wrapper function for gzip decompression and untar.
	// 		err := decompressAndUntar(location, f.Permissions, data)
	// 		if err != nil {
	// 			return err
	// 		}

	// 	default:
	// 		return fmt.Errorf("unsupported file type: %s", f.Type)
	// 	}

	// 	if f.Permissions != "" {
	// 		p, err := strconv.ParseUint(f.Permissions, 8, 32)
	// 		if err != nil {
	// 			return fmt.Errorf("failed to parse file permissions: %w", err)
	// 		}

	// 		err = os.Chmod(path, os.FileMode(uint32(p)))
	// 		if err != nil {
	// 			return fmt.Errorf("failed to apply file permissions: %w", err)
	// 		}
	// 	}
	// }

	// return nil
	panic("")
}

func decompressAndUntar(location string, perms string, encodedData []byte) error {
	// gr, err := gzip.NewReader(bytes.NewBuffer(encodedData))
	// if err != nil {
	// 	return err
	// }
	// defer gr.Close()

	// // Untar directly from the gzip reader
	// return untar(location, perms, gr)
	panic("")
}
