# Cotação App

Este projeto consiste em um cliente e um servidor em Go que interagem para obter e registrar a cotação do dólar em relação ao real.

## Estrutura do Projeto

```
cotacao-app
├── client.go      # Implementa um cliente HTTP que solicita a cotação do dólar.
├── server.go      # Implementa um servidor HTTP que fornece a cotação do dólar.
├── go.mod         # Definição do módulo Go e suas dependências.
└── README.md      # Documentação do projeto.
```

## Como Executar

### Pré-requisitos

- Go instalado na sua máquina. Você pode baixar a versão mais recente em [golang.org](https://golang.org/dl/).
- SQLite para persistência de dados.

### Passos para Rodar o Servidor

1. Navegue até o diretório do projeto:
   ```
   cd cotacao-app
   ```

2. Execute o servidor:
   ```
   go run server.go
   ```

O servidor irá escutar na porta 8080 e estará disponível no endpoint `/cotacao`.

### Passos para Rodar o Cliente

1. Em um novo terminal, navegue até o diretório do projeto:
   ```
   cd cotacao-app
   ```

2. Execute o cliente:
   ```
   go run client.go
   ```

O cliente fará uma requisição ao servidor e salvará a cotação atual do dólar em um arquivo chamado `cotacao.txt` no formato: `Dólar: {valor}`.

## Observações

- O servidor possui um timeout de 200ms para a requisição da API de câmbio e 10ms para a operação de persistência no banco de dados.
- O cliente possui um timeout de 300ms para receber a resposta do servidor.
- Erros de timeout serão registrados nos logs.

## Contribuições

Sinta-se à vontade para contribuir com melhorias ou correções.