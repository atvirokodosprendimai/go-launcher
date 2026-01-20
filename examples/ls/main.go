package main

import (
	"fmt"
	"io"
	"log"

	"github.com/atvirokodosprendimai/go-launcher"
)

func main() {
	// Duomenys, kuriuos siųsime į mikroservisą
	inputData := `{"invoice_id": "123", "amount": 500}`

	// 1. Konfigūruojame
	// Tarkime, tavo sukompiliuotas mikroservisas yra "./bin/invoice-processor"
	// Svarbu: Jei tas mikroservisas naudoja mūsų pirmąjį wrapperį, jam reikia "--std" flag'o!
	ms := launcher.Create("ls")

	// Nustatome inputą (gali būti failas, bet čia stringas)
	ms.FromMemory([]byte(inputData))

	log.Println("Orkestratorius: Paleidžiu mikroservisą...")

	// 2. Paleidžiame ir apdorojame rezultatą
	err := ms.Run(func(output io.Reader) error {
		// Čia mes gauname duomenis "streaming" būdu tiesiai iš mikroserviso

		// Pavyzdys: nuskaitome viską į atmintį (bet galėtume ir rašyti į failą)
		result, err := io.ReadAll(output)
		if err != nil {
			return err
		}

		fmt.Printf("\n--- REZULTATAS IŠ MICROSERVISO ---\n%s\n----------------------------------\n", string(result))
		return nil
	})

	if err != nil {
		log.Fatalf("Klaida vykdant mikroservisą: %v", err)
	}

	log.Println("Orkestratorius: Darbas baigtas.")
}
