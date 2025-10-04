package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/app/realesrgan/local"
	"github.com/kmulvey/realesrgan-scheduler/pkg/realesrgan"
	log "github.com/sirupsen/logrus"

	progress "github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const promNamespace = "realesrgan_scheduler"

type imageStatus struct {
	Name     string
	Size     string
	Progress float64 // 0.0 - 1.0
	Bar      progress.Model
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
		barWidth := m.width / 2
		for _, img := range m.Processing {
			img.Bar.Width = barWidth // Only update the width, don't recreate the bar
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case addImageMsg:
		barWidth := m.width / 2
		if barWidth <= 0 {
			barWidth = 40 // fallback
		}
		m.Processing[msg.Name] = &imageStatus{
			Name: msg.Name,
			Size: msg.Size,
			Bar:  progress.New(progress.WithDefaultGradient(), progress.WithWidth(barWidth)),
		}
	case progressMsg:
		if img, ok := m.Processing[msg.Name]; ok {
			log.Printf("UPDATE HANDLER: %s progress=%.2f", msg.Name, msg.Progress)
			progress := msg.Progress
			if progress > 1.0 {
				progress = 1.0
			}
			if progress < 0.0 {
				progress = 0.0
			}
			img.Progress = progress
			cmd := img.Bar.SetPercent(progress)
			if progress >= 1.0 {
				m.ImagesRemaining--
			}
			return m, cmd
		} else {
			log.Printf("progressMsg for unknown image: %s", msg.Name)
		}
	case progress.FrameMsg:
		cmds := make([]tea.Cmd, 0, len(m.Processing))
		for _, img := range m.Processing {
			barModel, cmd := img.Bar.Update(msg)
			if bm, ok := barModel.(progress.Model); ok {
				img.Bar = bm
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Images Remaining: %d\n\n", m.ImagesRemaining)

	// Responsive width
	w := m.width
	if w <= 0 {
		w = 80
	}
	barWidth := w / 2
	infoWidth := w - barWidth - 2 // -2 for spacing

	// Convert map to slice and sort by progress descending
	var sorted []*imageStatus
	for _, v := range m.Processing {
		sorted = append(sorted, v)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Progress > sorted[j].Progress
	})

	for _, img := range sorted {
		bar := img.Bar.View()
		info := fmt.Sprintf("%s (%s)", img.Name, img.Size)
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

func main() {
	// Open log file
	logFile, err := os.OpenFile("scheduler.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Could not open log file:", err)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	skipDirs, err := makeSkipMap("./skip.txt")
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

	images, err := findFilesToUpsize(upsizedDirs, "/home/kmulvey/empyrean/backup/retirees", skipDirs, skipImages)
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
					progressVal := 0.0
					fmt.Sscanf(pct, "%f%%", &progressVal)
					log.Printf("GOT progress from realesrgan for %s: raw=%q parsed=%.2f%%", f.SourceFile, pct, progressVal)
					p.Send(progressMsg{Name: f.SourceFile, Progress: progressVal / 100.0})
				}
				log.Printf("SENDING FINAL progressMsg for %s: 1.00", f.SourceFile)
				p.Send(progressMsg{Name: f.SourceFile, Progress: 1.0})
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
