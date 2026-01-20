# go-launcher

`go-launcher` yra Go kalba parašytas paketas, skirtas paleisti mikroservisus kaip atskirus procesus. Jis leidžia lengvai konfigūruoti mikroserviso vykdymą, nustatyti įvesties duomenis ir apdoroti išvestį. Šis paketas yra naudingas orkestruojant mikroservisų veikimą ir užtikrinant sklandų duomenų perdavimą tarp jų.


Pavizdys kaip dirbti su failais:

```go
func main() {
    // 1. Sukuriame servisą (naudojame --std, nes failus valdo launcheris!)
    ms := launcher.Create("./bin/invoice-processor", "--std")

    // 2. Nustatome failus
    if err := ms.FromFile("data/input.json"); err != nil {
        log.Fatal(err)
    }
    if err := ms.ToFile("data/output.xml"); err != nil {
        log.Fatal(err)
    }

    // 3. Paleidžiame (be callback, nes viskas sukonfigūruota)
    log.Println("Apdorojami failai...")
    if err := ms.Run(nil); err != nil {
        log.Fatalf("Fail: %v", err)
    }
    log.Println("Baigta. Rezultatas: data/output.xml")
}
```

arba su stdio:

```go
func main() {
    ms := launcher.Create("./bin/invoice-processor", "--std")
    
    // Input iš atminties
    ms.FromMemory([]byte(`{"id": 999}`))

    // Output į callback funkciją
    err := ms.Run(func(r io.Reader) error {
        result, _ := io.ReadAll(r)
        fmt.Printf("Gautas atsakymas: %s\n", result)
        return nil
    })
    
    if err != nil {
        log.Fatal(err)
    }
}
```

arba iš failo ir pasiimti stdout

```go
func main() {
    ms := launcher.Create("./bin/invoice-processor", "--std")
    ms.FromFile("didelius_duomenys.json")

    ms.Run(func(r io.Reader) error {
        // Čia galime dekoduoti JSON streamą
        // ...
        return nil
    })
}
```
