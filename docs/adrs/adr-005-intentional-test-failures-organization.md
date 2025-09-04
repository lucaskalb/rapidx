# ADR-005: Organização de Testes com Falha Intencional

## Status

**ACCEPTED** - 2024-12-19

## Context

O projeto rapidx contém diversos tipos de testes, incluindo testes que são projetados para falhar intencionalmente. Estes testes servem diferentes propósitos:

1. **Testes de Demonstração**: Mostram como o framework funciona quando propriedades falham
2. **Testes de Funcionalidade**: Verificam o comportamento correto do framework em cenários de falha
3. **Testes de Comparação**: Demonstram falhas esperadas em funções de comparação

Atualmente, estes testes estão espalhados por diferentes diretórios:
- `examples/` - contém testes de demonstração misturados com exemplos funcionais
- `prop/prop_test.go` - contém testes de framework com falhas intencionais
- `quick/quick_test.go` - contém testes de comparação com falhas esperadas

Esta organização cria confusão sobre quais testes devem passar e quais são projetados para falhar, dificultando a manutenção e execução seletiva de testes.

## Decision

Criar uma estrutura dedicada para testes com falha intencional, organizando-os em subpacotes específicos:

```
testfailures/
├── demo/              # Testes de demonstração
│   ├── shrinking_demo_test.go
│   ├── property_demo_test.go
│   └── comparison_demo_test.go
├── framework/         # Testes de funcionalidade do framework
│   ├── failure_behavior_test.go
│   ├── shrinking_failure_test.go
│   └── parallel_failure_test.go
└── integration/       # Testes de integração com falhas
    └── end_to_end_failure_test.go
```

### Características da Implementação:

1. **Build Tags**: Todos os testes com falha intencional usam `//go:build demo` para permitir execução seletiva
2. **Separação por Propósito**: Cada subpacote tem um propósito específico e bem documentado
3. **Documentação Clara**: Cada arquivo contém comentários explicando o propósito dos testes
4. **Preservação de Funcionalidade**: Os testes originais são movidos, não removidos

## Consequences

### Positivas:

- **Clareza**: Separação clara entre testes funcionais e de demonstração
- **Manutenibilidade**: Mais fácil encontrar e gerenciar testes específicos
- **CI/CD Friendly**: Possibilidade de executar apenas testes funcionais por padrão
- **Documentação**: Cada subpacote pode ter sua própria documentação
- **Execução Seletiva**: Uso de build tags permite controle granular

### Negativas:

- **Reorganização**: Requer movimentação de arquivos existentes
- **Build Tags**: Adiciona complexidade na execução de testes
- **Estrutura**: Aumenta a profundidade da estrutura de diretórios

### Riscos Mitigados:

- **Quebra de Funcionalidade**: Testes são movidos, não removidos
- **Confusão**: Documentação clara em cada arquivo explica o propósito
- **Execução**: Build tags permitem execução seletiva sem afetar testes funcionais

## Comandos de Execução

```bash
# Rodar apenas testes funcionais (padrão)
go test ./...

# Rodar testes de demonstração
go test -tags demo ./testfailures/demo/...

# Rodar testes de framework
go test -tags demo ./testfailures/framework/...

# Rodar todos os testes (incluindo falhas intencionais)
go test -tags demo ./...
```

## Alternativas Consideradas

1. **Manter Estrutura Atual**: Rejeitada por criar confusão sobre propósito dos testes
2. **Usar Sufixos**: Rejeitada por não resolver o problema de organização
3. **Diretório Único**: Rejeitada por não permitir categorização adequada
4. **Build Tags Sem Reorganização**: Rejeitada por não resolver o problema de clareza

## Implementação

A implementação foi realizada em 2024-12-19, incluindo:

1. Criação da estrutura de diretórios `testfailures/`
2. Movimentação dos testes de demonstração de `examples/` para `testfailures/demo/`
3. Movimentação dos testes de framework de `prop/prop_test.go` para `testfailures/framework/`
4. Movimentação dos testes de comparação de `quick/quick_test.go` para `testfailures/demo/`
5. Adição de build tags `//go:build demo` em todos os testes movidos
6. Atualização da documentação em cada arquivo movido

## Referências

- [ADR-001: Serial Shrinking Strategy](./adr-001-serial-shrinking.md)
- [ADR-002: Shrinking Strategies](./adr-002-shrinking-strategies.md)
- [ADR-003: Replay Command Line](./adr-003-replay-command-line.md)
- [ADR-004: Simple State Machine](./adr-004-simple-state-machine.md)
