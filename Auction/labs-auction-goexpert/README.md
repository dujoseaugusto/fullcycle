# labs-auction-goexpert — Auto-close Auction Feature

Este fork contém uma melhoria: fechamento automático do leilão (auction) após um intervalo configurável usando goroutines.

## Objetivo implementado

- Calcular o tempo do leilão a partir da variável de ambiente `AUCTION_INTERVAL`.
- Ao criar um leilão, iniciar uma goroutine que aguarda o término do intervalo e fecha o leilão (muda `status` para `Completed`).
- Teste de integração que valida o fechamento automático (`internal/infra/database/auction/create_auction_test.go`).

## Como rodar em ambiente de desenvolvimento

Pré-requisitos:
- Docker & docker-compose
- Go 1.20+

1. Subir os serviços necessários (MongoDB) via docker-compose:

```bash
cd Auction/labs-auction-goexpert
docker-compose up -d mongodb
```

2. Ajuste variáveis de ambiente conforme necessário.
   - O arquivo `cmd/auction/.env` já contém valores úteis para desenvolvimento:
     - `AUCTION_INTERVAL` — duração do leilão (ex.: `20s`)
     - `MONGODB_URL` — endereço do MongoDB (quando rodando com docker-compose pode ser `mongodb://admin:admin@mongodb:27017/auctions?authSource=admin`)

3. Rodar a aplicação localmente (conecta ao Mongo do docker-compose):

```bash
# opcional: export AUCTION_INTERVAL=20s
go run ./cmd/auction
```

ou usando docker-compose (constrói imagem e executa o binário):

```bash
docker-compose up --build
```

## Como rodar os testes (integração)

O teste responsável por validar o fechamento automático é um teste de integração e precisa de acesso ao MongoDB.

1. Inicie o MongoDB (conforme acima):

```bash
docker-compose up -d mongodb
```

2. No host (ou dentro de um ambiente que consiga acessar o Mongo), exporte as variáveis de ambiente para os testes (ajuste o host se necessário):

```bash
export MONGODB_URL="mongodb://admin:admin@localhost:27017/auctions?authSource=admin"
export MONGODB_DB="auctions"
export AUCTION_INTERVAL="1s"  # valor curto para testes
```

3. Rode os testes:

```bash
cd Auction/labs-auction-goexpert
go test ./... -v
```

O teste `TestAuctionAutoClose` é executado apenas se `MONGODB_URL` e `MONGODB_DB` estiverem definidos; caso contrário é ignorado.

## Observações

- A implementação adiciona uma goroutine por leilão criado que aguarda o tempo restante e fecha o leilão.
- Para garantir robustez em produção (por exemplo, em caso de reinício do serviço), uma melhoria adicional seria implementar um scanner periódico para fechar leilões vencidos ao iniciar o serviço.

---

Se quiser, eu implemento também o scanner periódico ou adapto a rotina para tornar o fechamento tolerante a reinícios (persistência do schedule).
