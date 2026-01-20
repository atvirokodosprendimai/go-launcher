package launcher

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

// Microservice struktūra laiko konfigūraciją apie tai, ką paleisti
type Microservice struct {
	Command string
	Args    []string
	Input   io.Reader // Ką paduosime į STDIN
	Env     []string  // Jei reiktų specifinių aplinkos kintamųjų
}

// Create sukuria naują instanciją
func Create(cmd string, args ...string) *Microservice {
	return &Microservice{
		Command: cmd,
		Args:    args,
		Input:   os.Stdin, // Default: paveldi tėvinio proceso stdin, nebent pakeisime
	}
}

// SetInput nustato, ką paduosime į pipe'ą (gali būti failas, bufferis ar kitas Readeris)
func (m *Microservice) SetInput(r io.Reader) *Microservice {
	m.Input = r
	return m
}

// SetInputBytes patogumui, jei turime tiesiog []byte
func (m *Microservice) SetInputBytes(data []byte) *Microservice {
	m.Input = bytes.NewReader(data)
	return m
}

// OutputHandler yra tavo callback funkcija.
// Ji gauna Readerį, iš kurio gali skaityti vaiko STDOUT.
type OutputHandler func(output io.Reader) error

// Run paleidžia procesą ir perduoda jo stdout į handlerį
func (m *Microservice) Run(handler OutputHandler) error {
	cmd := exec.Command(m.Command, m.Args...)

	// 0. procesų grupė
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// 1. STDIN: Prijungiame mūsų input srautą
	cmd.Stdin = m.Input

	// 2. STDERR: Labai svarbu! Prijungiame prie tėvinio proceso stderr.
	// Taip matysi visus logus/klaidas realiu laiku savo terminale.
	cmd.Stderr = os.Stderr

	// 3. STDOUT: Sukuriame pipe skaitymui
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("nepavyko sukurti stdout pipe: %w", err)
	}

	// 4. Startuojame procesą (asinchroniškai)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("nepavyko paleisti komandos '%s': %w", m.Command, err)
	}

	// 5. Kviečiame vartotojo callback'ą, kol procesas dirba
	// Čia vartotojas skaito duomenis.
	if err := handler(stdoutPipe); err != nil {
		// Jei callback'as grąžina klaidą, bandome žudyti procesų grupę
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		return fmt.Errorf("handlerio klaida: %w", err)
	}

	// 6. Laukiame pabaigos (Wait uždaro pipe'us)
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("procesas baigė darbą su klaida: %w", err)
	}

	return nil
}
