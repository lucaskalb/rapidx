# Testfailures Package

Este pacote contém testes que são projetados para falhar intencionalmente. Estes testes servem diferentes propósitos educacionais e de verificação do framework.

## Estrutura

```
testfailures/
├── demo/              # Testes de demonstração
│   ├── shrinking_demo_test.go    # Demonstra shrinking com propriedades falsas
│   ├── property_demo_test.go     # Demonstra falhas em propriedades
│   └── comparison_demo_test.go   # Demonstra falhas em comparações
├── framework/         # Testes de funcionalidade do framework
│   ├── failure_behavior_test.go  # Testa comportamento de falha sequencial
│   ├── shrinking_failure_test.go # Testa shrinking com falhas
│   └── parallel_failure_test.go  # Testa comportamento de falha paralela
└── integration/       # Testes de integração com falhas
    └── (futuro)
```

## Como Executar

### Rodar Apenas Testes Funcionais (Padrão)
```bash
go test ./...
```

### Rodar Testes de Demonstração
```bash
go test -tags demo ./testfailures/demo/...
```

### Rodar Testes de Framework
```bash
go test -tags demo ./testfailures/framework/...
```

### Rodar Todos os Testes (Incluindo Falhas Intencionais)
```bash
go test -tags demo ./...
```

## Build Tags

Todos os testes neste pacote usam a build tag `demo` para permitir execução seletiva:

```go
//go:build demo
// +build demo
```

## Propósito dos Testes

### Demo Tests
- **Demonstram** como o framework funciona quando propriedades falham
- **Mostram** o mecanismo de shrinking em ação
- **Educam** sobre property-based testing
- **Falham intencionalmente** para fins de demonstração

### Framework Tests
- **Verificam** o comportamento correto do framework em cenários de falha
- **Testam** caminhos de código de falha (sequencial e paralelo)
- **Validam** o mecanismo de shrinking
- **Garantem** que o framework funciona corretamente quando propriedades falham

## Exemplos de Uso

### Para Desenvolvedores
Use os testes de demonstração para entender como o rapidx funciona:

```bash
# Ver shrinking em ação
go test -tags demo -run Test_Slice_SomaNaoNegativa ./testfailures/demo/...

# Ver falha de propriedade
go test -tags demo -run Test_String_FalsaRegra ./testfailures/demo/...
```

### Para CI/CD
Configure seu pipeline para rodar apenas testes funcionais:

```bash
# Em CI, rode apenas testes que devem passar
go test ./...
```

### Para Desenvolvimento Local
Quando desenvolvendo o framework, rode todos os testes:

```bash
# Rode todos os testes para verificar comportamento completo
go test -tags demo ./...
```

## Notas Importantes

1. **Estes testes FALHAM intencionalmente** - isso é esperado
2. **Não inclua em CI/CD** a menos que queira verificar comportamento de falha
3. **Use build tags** para controlar quando executar
4. **Documentação** em cada arquivo explica o propósito específico