package launcher

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type Microservice struct {
	Command string
	Args    []string

	// Vidiniai laukai srautų valdymui
	stdinSource io.Reader
	stdoutDest  io.Writer

	// Saugome failus, kuriuos reikės uždaryti po Run()
	closers []io.Closer
}

func Create(cmd string, args ...string) *Microservice {
	return &Microservice{
		Command:     cmd,
		Args:        args,
		stdinSource: os.Stdin,  // Default
		stdoutDest:  os.Stdout, // Default
	}
}

// --- INPUT Metodai ---

// FromMemory - nustatome input iš baitų masyvo
func (m *Microservice) FromMemory(data []byte) *Microservice {
	m.stdinSource = bytes.NewReader(data)
	return m
}

// FromReader - nustatome input iš bet kokio Readerio (pvz., kito pipe)
func (m *Microservice) FromReader(r io.Reader) *Microservice {
	m.stdinSource = r
	return m
}

// FromFile - atidarome failą skaitymui ir nukreipiame į stdin
func (m *Microservice) FromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("nepavyko atidaryti input failo: %w", err)
	}
	m.stdinSource = f
	m.closers = append(m.closers, f) // Užregistruojame uždarymui
	return nil
}

// --- OUTPUT Metodai ---

// ToWriter - nustatome output į bet kokį Writerį (pvz., bufferį)
func (m *Microservice) ToWriter(w io.Writer) *Microservice {
	m.stdoutDest = w
	return m
}

// ToFile - sukuriame failą ir nukreipiame stdout ten
func (m *Microservice) ToFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("nepavyko sukurti output failo: %w", err)
	}
	m.stdoutDest = f
	m.closers = append(m.closers, f) // Užregistruojame uždarymui
	return nil
}

// --- Vykdymas ---

// HandlerFunc neprivalomas - naudojamas tik jei norime perimti srautą kode
type HandlerFunc func(r io.Reader) error

// Run vykdo procesą.
// Jei handler == nil, duomenys teka į nustatytą stdoutDest (pvz., failą).
// Jei handler != nil, duomenys teka į handlerį (stdoutDest ignoruojamas).
func (m *Microservice) Run(handler HandlerFunc) error {
	// 1. Visada uždarome atidarytus failus funkcijos pabaigoje
	defer m.closeAll()

	cmd := exec.Command(m.Command, m.Args...)

	// procesų grupė
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Nustatome Stderr į tėvinį terminalą (kad matytume logus)
	cmd.Stderr = os.Stderr

	// Nustatome Input
	cmd.Stdin = m.stdinSource

	// 2. Output logika
	if handler != nil {
		// Scenarijus A: Vartotojas nori apdoroti srautą kode (Callback)
		pipe, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start error: %w", err)
		}

		// Kviečiame callback
		if err := handler(pipe); err != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
			return err
		}
	} else {
		// Scenarijus B: Duomenys eina tiesiai į failą arba stdout (nėra callback)
		cmd.Stdout = m.stdoutDest

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start error: %w", err)
		}
	}

	// 3. Laukiame pabaigos
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("proceso klaida: %w", err)
	}

	return nil
}

func (m *Microservice) closeAll() {
	for _, c := range m.closers {
		_ = c.Close()
	}
}
