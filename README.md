# API de Gestão de Clínica — Capim

**Capim Hub — encontre seu sorriso!**

API em Go para gerenciamento de clínicas e dentistas, acompanhada de uma interface web que simula um marketplace de clínicas odontológicas.

## Sumário

* [Visão Geral](#visão-geral)
* [Arquitetura](#arquitetura)

  * [Backend](#backend)
  * [Frontend](#frontend)
* [Pré-requisitos](#pré-requisitos)
* [Como executar](#como-executar)

  * [Backend](#backend-clinic-api)
  * [Frontend](#frontend-1)
* [Como rodar os testes](#como-rodar-os-testes)
* [Documentação da API](#documentação-da-api)
* [Justificativa Técnica](#justificativa-técnica)

  * [1. Decisões técnicas mais importantes](#1-decisões-técnicas-mais-importantes)
  * [2. O que faria diferente com mais tempo](#2-o-que-faria-diferente-com-mais-tempo)
  * [3. Uso de IA](#3-uso-de-ia)


## Visão Geral

O projeto consiste em uma API REST desenvolvida em **Go** para gerenciamento de clínicas odontológicas e dentistas.

Além da API, foi desenvolvido o **Capim Hub**, uma interface web que explora uma possível experiência de produto para a plataforma.

### Capim Hub

O frontend possui duas áreas principais:

**Área pública**

* Busca e descoberta de clínicas.
* Listagem de clínicas.
* Perfil público da clínica.
* Informações sobre a clínica.
* Listagem dos dentistas associados.
* Perfil dos dentistas.

**Área de gestão**

* Criação e gerenciamento de clínicas.
* Cadastro e gerenciamento de dentistas.
* Visualização das informações básicas relacionadas à clínica.

A ideia é representar, de forma simples, tanto o lado de **descoberta de clínicas** quanto o lado de **gestão da operação**.


## Arquitetura

### Backend

O backend utiliza uma **Feature-based Architecture**, organizando o código por domínio de negócio em vez de agrupá-lo apenas por tipo técnico.

```text
clinic-api/
├── cmd/                # entrypoint (main.go) e seed de dados
├── docs/                # documentação Swagger/OpenAPI gerada (swaggo)
├── internal/
│   ├── clinic/          # handler, service, repository, dto, validate, errors
│   ├── dentist/         # mesma organização interna
│   ├── payment/         # mesma organização interna
│   ├── config/          # carregamento de configuração via env vars
│   └── server/          # bootstrap do servidor HTTP e rotas
├── mocks/               # mocks gerados via mockery para as interfaces de repositório
├── pkg/
│   ├── pix/             # cliente Pix simulado
│   ├── problem/         # respostas de erro no formato RFC 9457 (problem+json)
│   └── storage/         # store genérico in-memory, thread-safe
├── .mockery.yml
└── Dockerfile
```

Cada domínio (`clinic`, `dentist`, `payment`) segue a mesma organização interna e possui suas próprias responsabilidades e interfaces, buscando reduzir o acoplamento entre as diferentes partes da aplicação.

A comunicação entre os domínios acontece por meio de **injeção de dependências manual, via construtores** (sem container/framework de DI), e de contratos (interfaces) bem definidos, evitando que um domínio precise conhecer detalhes internos de outro.

Dois detalhes de implementação valem destaque:

* **Erros padronizados (RFC 9457):** todos os handlers retornam erros no formato `application/problem+json` (`pkg/problem`), com um contrato único de erro em toda a API.
* **Idempotência e simulação assíncrona em `payment`:** `POST /payments` exige um header `Idempotency-Key` para evitar cobranças duplicadas, e a transição `pending -> approved` roda em uma goroutine com atraso aleatório (2–5s) — o atraso é injetável via functional option, o que permite testar o fluxo de forma determinística sem `time.Sleep` real nos testes.
 
De forma simplificada, a estrutura segue o fluxo:

```text
HTTP Request
     │
  Handler
     │
  Service
     │
 Repository
     │
   Storage
```

Essa separação permite manter as regras de negócio independentes da camada HTTP e facilita a criação de testes unitários utilizando mocks.

### HTTP

A aplicação utiliza principalmente componentes da biblioteca padrão do Go para HTTP, mantendo a implementação simples e próxima das primitivas nativas da linguagem.

O **Chi** é utilizado como router, fornecendo recursos de roteamento e middleware sem introduzir um framework HTTP pesado.

### Storage

Para o desafio, foi utilizado **storage in-memory**, evitando a necessidade de infraestrutura externa para executar o projeto localmente.

## Frontend

O frontend foi desenvolvido como uma camada complementar à API e tem como objetivo demonstrar uma possível experiência de produto para o ecossistema Capim.

A aplicação apresenta uma experiência semelhante a um marketplace:

```text
                 Capim Hub
                    │
          ┌─────────┴─────────┐
          │                   │
       Descoberta          Gestão
          │                   │
     ┌────┴────┐         ┌────┴────┐
     │         │         │         │
 Clínicas  Dentistas  Clínicas  Dentistas
```

A área pública prioriza descoberta e apresentação das informações, enquanto a área privada concentra as operações de gerenciamento.


## Pré-requisitos

* Go 1.26+ (desenvolvido com 1.26.6)
* Node.js 20+ (desenvolvido com 24)

Opcional:

* GNU Make


## Como executar

### Backend (clinic-api)

Para iniciar a API com dados de exemplo:

```bash
make seed
```

Para iniciar a API sem dados de exemplo:

```bash
make run
```

Também é possível executar diretamente com Go.

Com seed:

```bash
SEED_DATA=true go run ./cmd
```

Sem seed:

```bash
go run ./cmd
```

Por padrão, a API estará disponível em:

```text
http://localhost:8080
```

#### Variáveis de ambiente

| Variável    | Padrão  | Descrição                                                                    |
|-------------|---------|-------------------------------------------------------------------------------|
| `HTTP_PORT` | `8080`  | Porta HTTP do servidor                                                       |
| `LOG_LEVEL` | `info`  | Nível de log (`debug`, `info`, `warn`, `error`)                              |
| `SEED_DATA` | `false` | Se `true`, popula o storage in-memory com dados de exemplo na inicialização  |

#### Docker

Também é possível executar a API em um container (imagem baseada em `distroless`, expõe a porta `8080`):

```bash
make docker
docker run -p 8080:8080 clinic-api
```

### Frontend

Instale as dependências:

```bash
npm install
```

Inicie o ambiente de desenvolvimento:

```bash
npm run dev
```

O frontend fica disponível em `http://localhost:5173` e faz proxy das chamadas `/api/*` para `http://localhost:8080` (configurado em `vite.config.ts`) — **o backend precisa estar rodando** para a interface carregar dados reais.


## Como rodar os testes

Para executar todos os testes do backend:

```bash
go test ./...
```

Para executar os testes com informações de cobertura:

```bash
go test ./... -cover
```

Para executar os testes com race detector:

```bash
go test -race ./...
```

Atalhos equivalentes via Makefile: `make test` (com `-race`) e `make cover` (gera e abre o relatório de cobertura em HTML).

### Lint e segurança

O projeto também conta com verificação estática e de vulnerabilidades:

```bash
make lint   # golangci-lint
make vuln   # gosec + govulncheck
```


## Documentação da API

A API é documentada via Swagger/OpenAPI, gerado a partir de anotações nos handlers (`swaggo/swag`). Cada endpoint tem descrição, parâmetros, exemplos de payload/resposta e os códigos de erro possíveis.

Com o backend em execução, acesse:

**Swagger UI**

```text
http://localhost:8080/swagger/index.html
```

Para regenerar a documentação após alterar as anotações:

```bash
make swagger
```

### Convenções

* Todos os endpoints de negócio ficam sob o prefixo `/api` — por exemplo, `GET http://localhost:8080/api/clinics`. O único endpoint fora desse prefixo é o health check, em `/health`.
* Todo erro é retornado no formato [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457) (`application/problem+json`), com um campo `code` estável (ex.: `VALIDATION_ERROR`, `CLINIC_NOT_ACTIVE`) para tratamento programático, além do `status` HTTP.
* `POST /payments` exige o header `Idempotency-Key`: reenviar a mesma chave com o mesmo corpo retorna o pagamento já criado (200); com um corpo diferente, retorna `409 IDEMPOTENCY_KEY_CONFLICT`.

### Endpoints

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/health` | Health check |
| `POST` | `/api/clinics` | Cria uma clínica (status inicial `pending`) |
| `GET` | `/api/clinics` | Lista/busca clínicas (paginado; filtros `q`, `city`) |
| `GET` | `/api/clinics/{id}` | Detalhe de uma clínica |
| `PUT` | `/api/clinics/{id}` | Atualiza uma clínica (parcial; `document` é imutável) |
| `DELETE` | `/api/clinics/{id}` | Remove (soft delete) uma clínica |
| `POST` | `/api/clinics/{clinic_id}/dentists` | Cria um dentista na clínica |
| `GET` | `/api/clinics/{clinic_id}/dentists` | Lista dentistas da clínica (paginado; filtros por papel) |
| `GET` | `/api/clinics/{clinic_id}/dentists/{id}` | Detalhe de um dentista |
| `PUT` | `/api/clinics/{clinic_id}/dentists/{id}` | Atualiza um dentista (parcial) |
| `PATCH` | `/api/clinics/{clinic_id}/dentists/{id}/roles` | Altera as flags de administrador/responsável legal |
| `DELETE` | `/api/clinics/{clinic_id}/dentists/{id}` | Remove (soft delete) um dentista |
| `POST` | `/api/payments` | Cria um pagamento Pix (requer `Idempotency-Key`) |
| `GET` | `/api/payments` | Lista pagamentos de uma clínica (paginado; filtro `status`) |
| `GET` | `/api/payments/{id}` | Detalhe de um pagamento (poll para ver `pending → approved`) |

A lista completa de parâmetros, exemplos e todos os códigos de erro possíveis por endpoint está no Swagger UI.


## Justificativa Técnica

### 1. Decisões técnicas mais importantes

#### Feature-based Architecture

A organização por domínio foi escolhida para manter as responsabilidades de negócio próximas umas das outras.

Em vez de uma estrutura como:

```text
handlers/
services/
repositories/
models/
```

o projeto organiza as funcionalidades por domínio:

```text
internal/
├── clinic/
├── dentist/
└── ...
```

Isso facilita a evolução da aplicação, reduz o acoplamento e torna mais simples localizar o código relacionado a uma determinada funcionalidade.

#### Uso da biblioteca padrão do Go

Para a camada HTTP, foi priorizado o uso da biblioteca padrão do Go, utilizando o Chi apenas como router.

A escolha evita adicionar abstrações desnecessárias e mantém o comportamento da aplicação próximo das primitivas nativas da linguagem.

Além disso, facilita a compreensão do fluxo da aplicação por qualquer desenvolvedor familiarizado com Go.

#### Dependency Injection e interfaces

Os componentes dependem de abstrações em vez de implementações concretas.

Isso permite, por exemplo, substituir o storage in-memory por uma implementação baseada em banco de dados sem precisar alterar as regras de negócio.

A mesma estratégia facilita a criação de mocks para testes unitários.


### 2. O que faria diferente com mais tempo

O projeto foi deliberadamente mantido simples para o contexto do desafio. Em um ambiente de produção, algumas evoluções seriam importantes.

#### Persistência

Substituiria o storage in-memory por PostgreSQL, adicionando:

* migrations
* índices adequados
* constraints
* transações
* controle de concorrência
* estratégia de backup e recuperação.

#### Autenticação e autorização

Adicionaria autenticação para separar claramente:

* usuários públicos
* administradores da clínica
* dentistas
* outros possíveis perfis de acesso.

Também implementaria autorização baseada nas permissões do usuário e na clínica à qual ele pertence.

#### Observabilidade

Adicionaria:

* structured logging
* métricas
* tracing distribuído
* correlation/request IDs
* health checks
* métricas de latência e erro.

Isso permitiria acompanhar melhor o comportamento da API em produção.

#### Infraestrutura

O projeto já conta com `Dockerfile`, `golangci-lint` e verificação de vulnerabilidades (`gosec` + `govulncheck`) via Makefile, mas rodados apenas localmente. Com mais tempo, formalizaria isso em pipeline, adicionando:

* CI/CD (build, lint, testes, `gosec`/`govulncheck` e cobertura em cada PR)
* Docker Compose para subir backend e frontend juntos em ambiente local

### 3. Uso de IA

A IA foi utilizada como ferramenta de apoio durante o desenvolvimento, principalmente para:

* geração de codigo (revisado por mim)
* explorar alternativas de arquitetura
* revisar decisões técnicas
* identificar possíveis edge cases
* auxiliar na criação de testes
* revisar documentação
* gerar ideias para melhorias do produto.

usei a skill coder (criada por mim) e seus agents para geração de codigo, debate(através do comiter), debug etc

usei a extrategia de usar o ai.md em que se resume em uma "ordem petra" da aplicação, onde toda a discursão, discovery deve respeita-lo
