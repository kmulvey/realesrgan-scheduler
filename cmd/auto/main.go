package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/app/realesrgan/local"
	"github.com/kmulvey/realesrgan-scheduler/pkg/realesrgan"
	log "github.com/sirupsen/logrus"

	tea "github.com/charmbracelet/bubbletea"
)

const promNamespace = "realesrgan_scheduler"

type imageStatus struct {
	Name     string
	Size     string
	Progress float64 // 0.0 - 1.0
}

type progressMsg struct {
	Name     string
	Progress float64
}

type addImageMsg struct {
	Name string
	Size string
}

type model struct {
	ImagesRemaining int
	Processing      map[string]*imageStatus
	width           int
}

func initialModel(total int) model {
	return model{
		ImagesRemaining: total,
		Processing:      make(map[string]*imageStatus),
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case addImageMsg:
		m.Processing[msg.Name] = &imageStatus{
			Name: msg.Name,
			Size: msg.Size,
		}
	case progressMsg:
		if img, ok := m.Processing[msg.Name]; ok {
			img.Progress = msg.Progress
			if msg.Progress >= 1.0 {
				m.ImagesRemaining--
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Images Remaining: %d\n\n", m.ImagesRemaining)

	// Default width if not set yet
	w := m.width
	if w <= 0 {
		w = 80
	}
	barWidth := w / 2
	infoWidth := w - barWidth - 2 // -2 for spacing

	for _, img := range m.Processing {
		bar := progressBar(img.Progress, barWidth)
		info := fmt.Sprintf("%3.0f%%  %s (%s)", img.Progress*100, img.Name, img.Size)
		// Pad or trim info to fit
		if len(info) > infoWidth {
			info = info[:infoWidth]
		} else {
			info = info + strings.Repeat(" ", infoWidth-len(info))
		}
		fmt.Fprintf(&b, "%s  %s\n", bar, info)
	}
	b.WriteString("\nPress q to quit.\n")
	return b.String()
}

func progressBar(p float64, width int) string {
	filled := int(p * float64(width))
	return "[" + strings.Repeat("█", filled) + strings.Repeat("-", width-filled) + "]"
}

func main() {
	var skipDirs, err = makeSkipMap("./skip.txt")
	if err != nil {
		log.Fatal(err)
	}

	skipImages, err := getSkipFiles("../auto/skipcache")
	if err != nil {
		log.Fatal(err)
	}

	upsizedDirs, err := path.List("/home/kmulvey/empyrean/backup/upscayl", 2, false, path.NewDirEntitiesFilter())
	if err != nil {
		log.Fatal(err)
	}

	images, err := findFilesToUpsize(upsizedDirs, "/home/kmulvey/Documents", skipDirs, skipImages)
	if err != nil {
		log.Fatal(err)
	}

	//////////////////
	var files = make(chan *realesrgan.ImageConfig)
	p := tea.NewProgram(initialModel(len(images)))

	go func() {
		for f := range files {
			// Get file size (optional)
			fi, _ := os.Stat(f.SourceFile)
			size := "?"
			if fi != nil {
				size = fmt.Sprintf("%.1fMB", float64(fi.Size())/1024/1024)
			}
			p.Send(addImageMsg{Name: f.SourceFile, Size: size})

			go func(f *realesrgan.ImageConfig) {
				for pct := range f.Progress {
					// Parse "3.5%" to float
					progressVal := 0.0
					fmt.Sscanf(pct, "%f%%", &progressVal)
					p.Send(progressMsg{Name: f.SourceFile, Progress: progressVal / 100.0})
				}
			}(f)
		}
	}()

	rl, err := local.NewRealesrganLocal(promNamespace, "/home/kmulvey/src/realesrgan-ncnn-vulkan-20220424-ubuntu/realesrgan-ncnn-vulkan", "realesrgan-x4plus", 2, true, files)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err = rl.Run(images...)
		if err != nil {
			log.Fatal(err)
		}
	}()

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
