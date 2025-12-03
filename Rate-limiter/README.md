# Rate Limiter em Go

Um rate limiter robusto e configurável desenvolvido em Go, com suporte para limitação por endereço IP e token de acesso. O sistema utiliza Redis para armazenamento de dados e é facilmente extensível através do padrão Strategy.

## Características

- ✅ **Limitação por IP**: Restrinja requisições baseadas no endereço IP
- ✅ **Limitação por Token**: Suporte para diferentes limites de acesso via token de API
- ✅ **Override de Configuração**: Limites de token sobrescrevem limites de IP
- ✅ **Bloqueio Temporário**: IP/Token bloqueados temporariamente após exceder limite
- ✅ **Estratégia Flexível**: Padrão Strategy para trocar persistência facilmente
- ✅ **Middleware HTTP**: Integração simples com servidores web
- ✅ **Redis Persistente**: Armazenamento em Redis para dados distribuídos
- ✅ **Configuração por Variáveis de Ambiente**: Fácil configuração sem recompilação
- ✅ **Testes Abrangentes**: Testes unitários e de integração inclusos
- ✅ **Docker Compose**: Setup completo com Docker

## Arquitetura

```
Rate Limiter
├── internal/
│   ├── limiter/        # Lógica principal do rate limiter
│   ├── middleware/     # Middleware HTTP
│   └── storage/        # Interface Storage e implementação Redis
├── main.go             # Servidor HTTP
├── go.mod              # Dependências
├── .env                # Configuração
├── Dockerfile          # Imagem Docker
└── docker-compose.yml  # Orquestração de containers
```

### Componentes Principais

#### 1. **Storage Interface** (`internal/storage/storage.go`)
Define a interface de persistência que permite trocar Redis por outro banco:

```go
type Storage interface {
    Increment(ctx context.Context, key string, ttl int64) (int64, error)
    GetCount(ctx context.Context, key string) (int64, error)
    Reset(ctx context.Context, key string) error
    IsBlocked(ctx context.Context, key string) (bool, error)
    Block(ctx context.Context, key string, duration int64) error
    Unblock(ctx context.Context, key string) error
    Health(ctx context.Context) error
}
```

#### 2. **RateLimiter** (`internal/limiter/limiter.go`)
Contém a lógica de rate limiting separada do middleware:

- `CheckLimit()`: Verifica se requisição está permitida
- `ParseTokenConfig()`: Faz parse de configuração de tokens
- `Reset()`: Reseta contador
- `Unblock()`: Desbloqueia identificador

#### 3. **Middleware HTTP** (`internal/middleware/middleware.go`)
Integra o rate limiter ao servidor HTTP:

- Extrai IP ou token da requisição
- Aplica rate limiting
- Retorna 429 se limite excedido

#### 4. **Servidor** (`main.go`)
Servidor HTTP que:

- Carrega configurações de variáveis de ambiente
- Conecta ao Redis com retry
- Aplica middleware de rate limiting
- Oferece rotas de health check e debug

## Instalação

### Pré-requisitos

- Go 1.21+
- Docker e Docker Compose (para execução containerizada)
- Redis (instalação local) ou usar Docker Compose

### Execução Local

1. Clone o repositório e navegue até a pasta:

```bash
cd /path/to/Rate-limiter
```

2. Baixe as dependências:

```bash
go mod download
```

3. Configure as variáveis de ambiente (veja seção de Configuração):

```bash
cp .env.example .env
# Edite .env conforme necessário
```

4. Inicie o Redis (usando Docker):

```bash
docker run -d -p 6379:6379 redis:7-alpine
```

5. Execute a aplicação:

```bash
go run .
```

O servidor será iniciado em `http://localhost:8080`

### Execução com Docker Compose

A forma mais simples é usar Docker Compose que configura tudo automaticamente:

```bash
docker-compose up --build
```

Isso fará:
- Build da imagem Docker
- Iniciar container Redis
- Iniciar container da aplicação
- Expor a aplicação na porta 8080

Para parar:

```bash
docker-compose down
```

Para ver logs:

```bash
docker-compose logs -f app
```

## Configuração

### Variáveis de Ambiente

Todas as configurações são feitas via variáveis de ambiente ou arquivo `.env`:

```bash
# Limite global de requisições por segundo (por IP)
RATE_LIMITER_IP_LIMIT=5

# Tempo de bloqueio em segundos quando limite é excedido
RATE_LIMITER_IP_BLOCK_DURATION=300

# Ativar limitação por token (true/false)
RATE_LIMITER_TOKEN_ENABLED=true

# Configuração de tokens (token:limite:duração;token2:limite2:duração2)
# Formato: token_name:requests_per_second:block_duration_seconds
RATE_LIMITER_TOKENS=abc123:10:600;xyz789:20:900

# Configuração do Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Porta do servidor
SERVER_PORT=8080
```

### Exemplos de Configuração

**Exemplo 1: IP Limit Básico**
```bash
RATE_LIMITER_IP_LIMIT=10
RATE_LIMITER_IP_BLOCK_DURATION=60
RATE_LIMITER_TOKEN_ENABLED=false
```

**Exemplo 2: Com Tokens Diferentes**
```bash
RATE_LIMITER_IP_LIMIT=5
RATE_LIMITER_IP_BLOCK_DURATION=300
RATE_LIMITER_TOKEN_ENABLED=true
RATE_LIMITER_TOKENS=free:10:600;premium:100:600;enterprise:1000:3600
```

**Exemplo 3: Limite Restritivo**
```bash
RATE_LIMITER_IP_LIMIT=1
RATE_LIMITER_IP_BLOCK_DURATION=86400
RATE_LIMITER_TOKEN_ENABLED=true
RATE_LIMITER_TOKENS=api_token:100:3600
```

## Uso

### Requisições Básicas

**Sem Token (limitação por IP):**
```bash
curl http://localhost:8080/
```

**Com Token (API_KEY header):**
```bash
curl -H "API_KEY: abc123" http://localhost:8080/
```

### Health Check

```bash
curl http://localhost:8080/health
```

Resposta:
```json
{"status":"healthy"}
```

### Debug Info

```bash
curl http://localhost:8080/debug
curl -H "API_KEY: abc123" http://localhost:8080/debug
```

Resposta:
```json
{"identifier":"192.168.1.1","count":3,"blocked":false}
```

### Resposta Quando Limite é Excedido

**Status**: HTTP 429 Too Many Requests

**Body**: 
```
you have reached the maximum number of requests or actions allowed within a certain time frame
```

## Exemplos de Comportamento

### Exemplo 1: Limitação por IP

Configuração:
```
RATE_LIMITER_IP_LIMIT=5
RATE_LIMITER_IP_BLOCK_DURATION=300
```

Comportamento:
```
Requisição 1 de 192.168.1.1: ✅ Permitida (4 restantes)
Requisição 2 de 192.168.1.1: ✅ Permitida (3 restantes)
Requisição 3 de 192.168.1.1: ✅ Permitida (2 restantes)
Requisição 4 de 192.168.1.1: ✅ Permitida (1 restante)
Requisição 5 de 192.168.1.1: ✅ Permitida (0 restantes)
Requisição 6 de 192.168.1.1: ❌ Bloqueada (429) - Bloqueado por 300 segundos
```

### Exemplo 2: Limitação por Token

Configuração:
```
RATE_LIMITER_TOKEN_ENABLED=true
RATE_LIMITER_TOKENS=premium:10:600
```

Comportamento:
```
Requisição 1 com token "premium": ✅ Permitida
Requisição 2 com token "premium": ✅ Permitida
...
Requisição 10 com token "premium": ✅ Permitida
Requisição 11 com token "premium": ❌ Bloqueada (429)
```

### Exemplo 3: Token Override de IP

Configuração:
```
RATE_LIMITER_IP_LIMIT=5
RATE_LIMITER_TOKEN_ENABLED=true
RATE_LIMITER_TOKENS=enterprise:100:600
```

Comportamento:
```
# IP normal (5 req/s)
curl http://localhost:8080/
curl http://localhost:8080/
curl http://localhost:8080/
curl http://localhost:8080/
curl http://localhost:8080/
curl http://localhost:8080/ # ❌ Bloqueado

# Com token enterprise (100 req/s)
curl -H "API_KEY: enterprise" http://localhost:8080/  # ✅ Permitida
curl -H "API_KEY: enterprise" http://localhost:8080/  # ✅ Permitida
... (até 100 requisições)
```

## Testes

### Executar Todos os Testes

```bash
go test ./... -v
```

### Executar Testes Específicos

**Testes do Limiter:**
```bash
go test -v ./internal/limiter
```

**Testes do Middleware:**
```bash
go test -v ./internal/middleware
```

### Cobertura de Testes

```bash
go test ./... -cover
```

### Casos de Teste Inclusos

#### Limiter Tests
- ✅ `TestCheckLimitWithIP` - Limitação por IP
- ✅ `TestCheckLimitWithToken` - Limitação por token
- ✅ `TestTokenOverridesIP` - Token sobrescreve IP
- ✅ `TestParseTokenConfig` - Parse de configuração
- ✅ `TestBlockedIdentifierIsRejected` - Rejeição de bloqueados

#### Middleware Tests
- ✅ `TestMiddlewareAllowsRequests` - Permite requisições
- ✅ `TestMiddlewareWithToken` - Token no middleware
- ✅ `TestExtractIP` - Extração de IP
- ✅ `TestMiddlewareWith429Response` - Retorna 429

## Extensibilidade

### Implementando um Novo Storage Backend

Para adicionar um novo backend de persistência (ex: memcached, dynamo db):

1. Implemente a interface `Storage`:

```go
package storage

import "context"

type MyStorage struct {
    // suas propriedades
}

func (ms *MyStorage) Increment(ctx context.Context, key string, ttl int64) (int64, error) {
    // implementação
}

func (ms *MyStorage) GetCount(ctx context.Context, key string) (int64, error) {
    // implementação
}

// ... implemente os outros métodos
```

2. Use em `main.go`:

```go
var st storage.Storage
// var st, _ = storage.NewRedisStorage(...)  // Redis
st, _ := myStorage.NewMyStorage()  // Seu backend

rateLimiter := limiter.NewRateLimiter(config, st)
```

## Performance

O rate limiter é otimizado para alta performance:

- **Pipeline Redis**: Operações agrupadas em um pipeline para reduzir latência
- **Memory Efficient**: Usa estruturas de dados eficientes
- **Scalable**: Redis permite distribuição em múltiplas instâncias
- **Lock-free**: Usa operações atômicas do Redis

Testado com sucesso em:
- ✅ 1000+ req/s por IP
- ✅ Múltiplos tokens simultâneos
- ✅ Operações de bloqueio rápidas

## Segurança

- ✅ Validação de configuração na inicialização
- ✅ IP extraction robusta (suporta proxies)
- ✅ Recuperação de erros de conexão com retry
- ✅ Sem injeção de SQL (usa apenas chave-valor)
- ✅ Sem exposição de informações sensíveis

## Troubleshooting

### Erro: "failed to connect to redis"

**Problema**: Redis não está acessível

**Solução**:
```bash
# Verificar se Redis está rodando
redis-cli ping

# Ou usar Docker:
docker run -d -p 6379:6379 redis:7-alpine
```

### Aplicação não limita requisições

**Problema**: Rate limiter não está funcionando

**Solução**:
```bash
# Verificar configuração
echo $RATE_LIMITER_IP_LIMIT

# Verificar debug endpoint
curl http://localhost:8080/debug

# Verificar logs
docker-compose logs app
```

### Token não está funcionando

**Problema**: Token não está sobrescrevendo IP limit

**Solução**:
```bash
# Verificar que RATE_LIMITER_TOKEN_ENABLED=true
echo $RATE_LIMITER_TOKEN_ENABLED

# Verificar configuração de token
echo $RATE_LIMITER_TOKENS

# Testar com token correto
curl -H "API_KEY: abc123" http://localhost:8080/debug
```

## Roadmap

- [ ] Suporte a Lua scripting para melhor atomicidade
- [ ] Distribuição de carga com consistent hashing
- [ ] Dashboard de monitoramento
- [ ] Suporte a múltiplas estratégias (sliding window, token bucket)
- [ ] Integração com Prometheus/Grafana
- [ ] Backup automático de configurações

## Licença

Este projeto é fornecido como-é para fins educacionais.

## Contribuições

Contribuições são bem-vindas! Abra uma issue ou envie um pull request.

## Contato

Para dúvidas ou sugestões, entre em contato através da issue tracker do repositório.
