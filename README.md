# API de Gestão de Clínica — Capim

<!-- TODO: um parágrafo curto descrevendo o projeto -->

Capim Hub - encontre seu sorriso!

## Sumário

- [Visão Geral](#visão-geral)
- [Arquitetura](#arquitetura)
- [Pré-requisitos](#pré-requisitos)
- [Como executar](#como-executar)
  - [Backend (clinic-api)](#backend-clinic-api)
  - [Frontend](#frontend)
- [Como rodar os testes](#como-rodar-os-testes)
- [Documentação da API](#documentação-da-api)
- [Justificativa Técnica](#justificativa-técnica)
  - [1. Decisões técnicas mais importantes](#1-decisões-técnicas-mais-importantes)
  - [2. O que faria diferente com mais tempo](#2-o-que-faria-diferente-com-mais-tempo)
  - [3. Uso de IA](#3-uso-de-ia)

## Visão Geral

O desafio consiste em uma api golang para gerenciamento de clinicas e dentistas

### Capim Hub
- Como principal diferencial criei um frontend para simular um marketplace de clinicas, onde pode buscar clinicas, ver perfil da clinica e perfil de dentistas
- tambem é possivel gerenciar suas clinicas e dentistas

## Arquitetura

<!-- TODO: estrutura de pastas, camadas (handler/service/repository), storage in-memory, etc. -->

### Backend
A arquitetura do backend é baseada em APp-styled archteture(ou feature archteture)

```
clinic-api
| cmd
| docs
| internal
| |- clinic
| |- <dominio>
| mocks
| pkg
| .mockery.yml
```

a ideia é que cada dominio seja desacoplado e que converse com outros dominio via DI
estou fazendo o uso de http std como lib http e chi como lib de router

### frontend


## Pré-requisitos

```
go 1.26.6
node 24
```

Opcional:
```
Make
```

<!-- TODO: versões de Go, Node, ferramentas -->

## Como executar

### Backend (clinic-api)

rodar com seed+make:
```
make seed
```

rodar sem seed:
```
make run
```

rodar com seed (sem make):
```
@SEED_DATA=true go run ./cmd
```

rodar sem seed:
```
go run ./cmd
```

<!-- TODO: comandos para rodar a API, variáveis de ambiente, seed de dados -->

### Frontend

```
npm run dev
```

<!-- TODO: comandos para rodar o frontend -->

## Como rodar os testes

<!-- TODO: comandos de teste unitário e integração -->

## Documentação da API

### Swagger

[documentação da api](http://localhost:8080/swagger/index.html)

<!-- TODO: link/instruções do Swagger -->

## Justificativa Técnica

### 1. Decisões técnicas mais importantes

<!-- TODO: 3 decisões técnicas mais importantes e por quê -->

### 2. O que faria diferente com mais tempo

<!-- TODO -->

### 3. Uso de IA

<!-- TODO: como a IA ajudou e onde você optou por fazer diferente do que ela sugeriu -->
