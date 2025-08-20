package bot

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func LogError(logger *log.Logger, err error, msg string) {
	logger.Printf("❌ %s: %v", msg, err)
}

func DownloadFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func ConvertOggToWav(oggPath, wavPath string) error {
	cmd := exec.Command("ffmpeg", "-i", oggPath, "-acodec", "pcm_s16le", "-ar", "16000", wavPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ Ошибка обработки FFmpeg: %v, результат: %s", err, string(output))
		return err
	}
	log.Printf("✅ Конец конвертации %s в %s", oggPath, wavPath)
	return nil
}
