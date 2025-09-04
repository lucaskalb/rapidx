# Domain-Specific Generators

This package contains domain-specific generators for property-based testing. These generators produce valid data according to specific business rules and validation algorithms.

## Available Generators

### CPF (Brazilian Tax ID)

The CPF generator produces valid Brazilian CPF (Cadastro de Pessoas Físicas) numbers.

#### Functions

- `CPF(masked bool) Generator[string]` - Generates valid CPF numbers
  - `masked=true`: Returns formatted CPF (e.g., "123.456.789-01")
  - `masked=false`: Returns raw CPF (e.g., "12345678901")

- `CPFAny() Generator[string]` - Generates CPF with random masking (50/50 chance)

#### Validation and Utilities

- `ValidCPF(s string) bool` - Validates if a string is a valid CPF
- `MaskCPF(raw string) string` - Formats a raw CPF with dots and dashes
- `UnmaskCPF(s string) string` - Removes formatting from a CPF string

#### Example Usage

```go
import "github.com/lucaskalb/rapidx/gen/domain"

// Generate unmasked CPF
prop.ForAll(t, cfg, domain.CPF(false))(func(t *testing.T, cpf string) {
    if !domain.ValidCPF(cpf) {
        t.Fatalf("invalid CPF generated: %q", cpf)
    }
})

// Generate masked CPF
prop.ForAll(t, cfg, domain.CPF(true))(func(t *testing.T, masked string) {
    raw := domain.UnmaskCPF(masked)
    if !domain.ValidCPF(raw) {
        t.Fatalf("invalid CPF after unmasking: %q", raw)
    }
})
```

## Future Generators

This package is designed to accommodate additional domain-specific generators:

- **CNPJ** - Brazilian company tax ID
- **Email** - Valid email addresses
- **Phone** - Phone numbers with country-specific formats
- **Credit Card** - Valid credit card numbers
- **UUID** - Universally unique identifiers

## Design Principles

1. **Validation**: All generators produce valid data according to domain rules
2. **Shrinking**: Generators include intelligent shrinking for minimal counterexamples
3. **Formatting**: Support for both raw and formatted output when applicable
4. **Performance**: Optimized for property-based testing scenarios
5. **Extensibility**: Easy to add new domain-specific generators

## Contributing

When adding new domain-specific generators:

1. Follow the existing patterns for validation and shrinking
2. Include comprehensive tests
3. Update this README with documentation
4. Ensure the generator produces valid data according to domain rules