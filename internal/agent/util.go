package agent

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// CreateInstanceFromCore membuat runner instance baru dengan symlink langsung ke core
// func CreateInstanceFromCore(coreDir, instanceDir string) error {
// 	if err := os.MkdirAll(instanceDir, 0755); err != nil {
// 		return fmt.Errorf("failed to create instance dir: %v", err)
// 	}

// 	// 🪄 Symlink folder bin & externals dari core (langsung, bukan bin.2.329.0)
// 	for _, d := range []string{"bin", "externals"} {
// 		src := filepath.Join(coreDir, d)
// 		dst := filepath.Join(instanceDir, d)
// 		if _, err := os.Lstat(dst); err == nil {
// 			_ = os.Remove(dst) // hapus kalau udah ada
// 		}
// 		if err := os.Symlink(src, dst); err != nil {
// 			return fmt.Errorf("failed to symlink %s → %s: %v", src, dst, err)
// 		}
// 		log.Printf("🔗 Linked %s → %s", dst, src)
// 	}

// 	// 🧩 Copy file ringan dari core
// 	filesToCopy := []string{"config.sh", "run.sh", "svc.sh", "env.sh", "safe_sleep.sh"}
// 	for _, f := range filesToCopy {
// 		src := filepath.Join(coreDir, f)
// 		dst := filepath.Join(instanceDir, f)
// 		if err := copyFile(src, dst); err != nil {
// 			log.Printf("⚠️ Skipped missing %s (non-fatal)", f)
// 			continue
// 		}
// 		if filepath.Ext(dst) == ".sh" {
// 			_ = os.Chmod(dst, 0755)
// 		}
// 	}

// 	return nil
// }

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func CreateInstanceFromCore(coreDir, instanceDir string) error {
	filesToCopy := []string{
		"config.sh",
		"env.sh",
		"run.sh",
		"run-helper.sh.template",
		"run-helper.sh",
		"run-helper.cmd.template",
		"safe_sleep.sh",
		"svc.sh",
		".path",
		".env",
	}

	// Pastikan direktori instance ada
	if err := os.MkdirAll(instanceDir, 0755); err != nil {
		return fmt.Errorf("failed to create instance dir: %v", err)
	}

	// 1️⃣ Buat symlink ke folder besar
	for _, folder := range []string{"bin", "externals"} {
		src := filepath.Join(coreDir, folder)
		dst := filepath.Join(instanceDir, folder)

		_ = os.RemoveAll(dst)
		if err := os.Symlink(src, dst); err != nil {
			return fmt.Errorf("failed to symlink %s: %v", folder, err)
		}
		log.Printf("🔗 Linked %s → %s", dst, src)
	}

	// 2️⃣ Copy file penting dari core
	for _, f := range filesToCopy {
		src := filepath.Join(coreDir, f)
		dst := filepath.Join(instanceDir, f)

		data, err := os.ReadFile(src)
		if err != nil {
			// file mungkin tidak ada — skip saja
			continue
		}
		if err := os.WriteFile(dst, data, 0755); err != nil {
			return fmt.Errorf("copy %s: %v", f, err)
		}
	}

	log.Printf("✅ Instance created at %s (linked to core)", instanceDir)
	return nil
}

// package agent

// import (
// 	"archive/tar"
// 	"compress/gzip"
// 	"fmt"
// 	"io"
// 	"os"
// 	"path/filepath"
// 	"strings"
// )

// func CopyDirContents(src, dst string) error {
// 	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		// ⚙️ Skip folder "instances"
// 		if info.IsDir() && filepath.Base(path) == "instances" {
// 			return filepath.SkipDir
// 		}

// 		// Relatif & target
// 		relPath, err := filepath.Rel(src, path)
// 		if err != nil {
// 			return err
// 		}
// 		targetPath := filepath.Join(dst, relPath)

// 		if info.IsDir() {
// 			// 🧱 Folder → buat aja
// 			return os.MkdirAll(targetPath, info.Mode())
// 		}

// 		// 🧱 Pastikan tidak overwrite folder
// 		if fi, err := os.Stat(targetPath); err == nil && fi.IsDir() {
// 			return nil // skip kalau target sudah folder
// 		}

// 		// 🧱 Skip file non-regular (misal socket, symlink)
// 		if !info.Mode().IsRegular() {
// 			return nil
// 		}

// 		// Pastikan folder parent ada
// 		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
// 			return err
// 		}

// 		// Copy isi file
// 		srcFile, err := os.Open(path)
// 		if err != nil {
// 			return err
// 		}
// 		defer srcFile.Close()

// 		dstFile, err := os.Create(targetPath)
// 		if err != nil {
// 			return err
// 		}
// 		defer dstFile.Close()

// 		if _, err := io.Copy(dstFile, srcFile); err != nil {
// 			return err
// 		}

// 		// Jadikan executable kalau script shell
// 		if filepath.Ext(targetPath) == ".sh" {
// 			_ = os.Chmod(targetPath, 0755)
// 		}

// 		return nil
// 	})
// }

// func ExtractTarGz(r io.Reader, dest string) error {
// 	gzr, err := gzip.NewReader(r)
// 	if err != nil {
// 		return err
// 	}
// 	defer gzr.Close()
// 	tr := tar.NewReader(gzr)

// 	for {
// 		hdr, err := tr.Next()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			return err
// 		}

// 		target := filepath.Join(dest, hdr.Name)
// 		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
// 			return fmt.Errorf("illegal file path: %s", target)
// 		}

// 		switch hdr.Typeflag {
// 		case tar.TypeDir:
// 			os.MkdirAll(target, os.FileMode(hdr.Mode))
// 		case tar.TypeReg:
// 			os.MkdirAll(filepath.Dir(target), 0755)
// 			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(hdr.Mode))
// 			if err != nil {
// 				return err
// 			}
// 			if _, err := io.Copy(f, tr); err != nil {
// 				f.Close()
// 				return err
// 			}
// 			f.Close()
// 		}
// 	}
// 	return nil
// }
