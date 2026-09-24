package agent

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractLimits limita a extracao de arquivos (defesa contra zip-slip e zip bomb).
type ExtractLimits struct {
	MaxFiles      int   // numero maximo de entradas
	MaxFileBytes  int64 // tamanho descomprimido maximo por arquivo
	MaxTotalBytes int64 // tamanho descomprimido total maximo
	MaxRatio      int64 // razao descomprimido/comprimido maxima (guarda zip bomb)
}

// DefaultExtractLimits sao limites conservadores usados quando nao especificados.
func DefaultExtractLimits() ExtractLimits {
	return ExtractLimits{
		MaxFiles:      10000,
		MaxFileBytes:  256 << 20, // 256 MiB/arquivo
		MaxTotalBytes: 1 << 30,   // 1 GiB total
		MaxRatio:      200,       // 200x
	}
}

// ExtractedEntry e o status por-arquivo (nunca ha skip silencioso).
type ExtractedEntry struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "extracted" | "skipped" | "error"
	Reason string `json:"reason,omitempty"`
	Bytes  int64  `json:"bytes"`
}

// ExtractReport resume a extracao.
type ExtractReport struct {
	Entries    []ExtractedEntry `json:"entries"`
	Files      int              `json:"files"`
	TotalBytes int64            `json:"total_bytes"`
}

var (
	ErrArchiveTooManyFiles = errors.New("archive exceeds file-count limit")
	ErrArchivePathEscapes  = errors.New("archive entry path escapes destination (zip-slip)")
	ErrArchiveSymlink      = errors.New("archive entry is a symlink")
	ErrArchiveFileTooBig   = errors.New("archive entry exceeds per-file size limit")
	ErrArchiveTotalTooBig  = errors.New("archive exceeds total size limit")
	ErrArchiveRatio        = errors.New("archive compression ratio exceeds limit (possible zip bomb)")
)

// safeJoin garante que o caminho relativo fica dentro de destDir.
func safeJoin(destDir, name string) (string, error) {
	slashed := strings.ReplaceAll(name, "\\", "/")
	// Rejeita caminho absoluto explicitamente (unix "/x", windows "C:/x").
	if strings.HasPrefix(slashed, "/") || filepath.IsAbs(name) ||
		(len(slashed) >= 2 && slashed[1] == ':') {
		return "", ErrArchivePathEscapes
	}
	// Politica estrita: qualquer segmento ".." e rejeitado (mesmo se colapsaria
	// para dentro do dest) — nao aceitamos zip-slip nem em forma canonicalizavel.
	for _, seg := range strings.Split(slashed, "/") {
		if seg == ".." {
			return "", ErrArchivePathEscapes
		}
	}
	clean := filepath.ToSlash(filepath.Clean("/" + slashed))
	// clean comeca com "/"; remove para juntar como relativo.
	rel := strings.TrimPrefix(clean, "/")
	if rel == "" || strings.HasPrefix(rel, "../") || strings.Contains(rel, "/../") || rel == ".." {
		return "", ErrArchivePathEscapes
	}
	target := filepath.Join(destDir, filepath.FromSlash(rel))
	// Garantia final: target precisa estar sob destDir.
	rp, err := filepath.Rel(destDir, target)
	if err != nil || rp == ".." || strings.HasPrefix(rp, ".."+string(os.PathSeparator)) {
		return "", ErrArchivePathEscapes
	}
	return target, nil
}

// SafeExtractZip extrai zipPath em destDir com protecoes de zip-slip/symlink/
// contagem/tamanho/ratio, retornando um relatorio por-arquivo. Falha fechada:
// qualquer entrada perigosa aborta a extracao (nao extrai parcialmente conteudo
// malicioso), com o motivo no relatorio.
func SafeExtractZip(zipPath, destDir string, limits ExtractLimits) (ExtractReport, error) {
	if limits.MaxFiles <= 0 {
		limits = DefaultExtractLimits()
	}
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return ExtractReport{}, err
	}
	defer reader.Close()

	if len(reader.File) > limits.MaxFiles {
		return ExtractReport{}, ErrArchiveTooManyFiles
	}
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return ExtractReport{}, err
	}

	report := ExtractReport{}
	var total int64
	for _, entry := range reader.File {
		name := entry.Name
		mode := entry.Mode()

		// Symlink: rejeita.
		if mode&os.ModeSymlink != 0 {
			report.Entries = append(report.Entries, ExtractedEntry{Name: name, Status: "error", Reason: ErrArchiveSymlink.Error()})
			return report, ErrArchiveSymlink
		}

		target, joinErr := safeJoin(destDir, name)
		if joinErr != nil {
			report.Entries = append(report.Entries, ExtractedEntry{Name: name, Status: "error", Reason: joinErr.Error()})
			return report, joinErr
		}

		// Diretorio.
		if strings.HasSuffix(name, "/") || mode.IsDir() {
			if err := os.MkdirAll(target, 0o700); err != nil {
				return report, err
			}
			report.Entries = append(report.Entries, ExtractedEntry{Name: name, Status: "skipped", Reason: "directory"})
			continue
		}

		// Ratio (zip bomb): descomprimido/comprimido.
		if entry.CompressedSize64 > 0 && limits.MaxRatio > 0 {
			ratio := int64(entry.UncompressedSize64 / entry.CompressedSize64)
			if ratio > limits.MaxRatio {
				report.Entries = append(report.Entries, ExtractedEntry{Name: name, Status: "error", Reason: ErrArchiveRatio.Error()})
				return report, ErrArchiveRatio
			}
		}
		// Tamanho declarado por arquivo.
		if int64(entry.UncompressedSize64) > limits.MaxFileBytes {
			report.Entries = append(report.Entries, ExtractedEntry{Name: name, Status: "error", Reason: ErrArchiveFileTooBig.Error()})
			return report, ErrArchiveFileTooBig
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return report, err
		}
		written, wErr := writeZipEntry(entry, target, limits.MaxFileBytes, limits.MaxTotalBytes-total)
		if wErr != nil {
			report.Entries = append(report.Entries, ExtractedEntry{Name: name, Status: "error", Reason: wErr.Error(), Bytes: written})
			return report, wErr
		}
		total += written
		report.Files++
		report.TotalBytes = total
		report.Entries = append(report.Entries, ExtractedEntry{Name: name, Status: "extracted", Bytes: written})
	}
	return report, nil
}

// writeZipEntry copia com teto real (nao confia no UncompressedSize declarado).
func writeZipEntry(entry *zip.File, target string, maxFile, remainingTotal int64) (int64, error) {
	rc, err := entry.Open()
	if err != nil {
		return 0, err
	}
	defer rc.Close()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	perFileCap := maxFile
	if remainingTotal < perFileCap {
		perFileCap = remainingTotal
	}
	// LimitReader em cap+1 para detectar estouro real do declarado/limite.
	written, err := io.Copy(out, io.LimitReader(rc, perFileCap+1))
	if err != nil {
		return written, err
	}
	if written > maxFile {
		return written, ErrArchiveFileTooBig
	}
	if written > remainingTotal {
		return written, ErrArchiveTotalTooBig
	}
	return written, nil
}
