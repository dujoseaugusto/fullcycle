# Cep-climate

Sistema com dois serviços em Go (Service A e Service B) que, dado um CEP, retornam cidade e temperatura atual em Celsius, Fahrenheit e Kelvin.

Arquitetura:
- `service-a`: recebe JSON `{ "cep": "29902555" }`, valida e encaminha para `service-b`.
- `service-b`: consulta ViaCEP, faz geocoding e consulta Open-Meteo para obter temperatura.
- `otel-collector`: coleta traces via OTLP e exporta para Zipkin.
- `zipkin`: painel para visualizar traces (http://localhost:9411)


Como rodar (desenvolvimento):

Opção A — com Docker (recomendado):

1. Na raiz do projeto execute:

```bash
cd Cep-climate
docker compose up --build
```

2. Endpoints:
- Service A (entrada): `POST http://localhost:8080/zipcode` com body `{ "cep": "29902555" }`
- Service B (direto): `POST http://localhost:8081/zipcode`

3. Traces:
- Zipkin estará em `http://localhost:9411`.

Opção B — sem Docker (executar localmente):

1. Instale Go (1.20+). 2. No diretório de cada serviço execute:

```bash
cd Cep-climate/service-b
go mod download
go run .

# em outro terminal
cd Cep-climate/service-a
go mod download
go run .
```

2. Por padrão os serviços vão ouvir em `:8081` (service-b) e `:8080` (service-a). Você pode definir `SERVICE_B_URL` e `OTEL_EXPORTER_OTLP_ENDPOINT` via ambiente se precisar apontar para outro collector ou endereço.

Observações:
- O projeto inclui OTEL/collector + Zipkin no `docker-compose` para facilitar a visualização de traces; se executar localmente sem o collector, o tracing ainda será inicializado mas não será exportado.
- As chamadas de geocoding e clima usam APIs públicas sem chave: Open-Meteo e sua API de geocoding.

Exemplo de request (qualquer opção):

```bash
curl -X POST -H "Content-Type: application/json" -d '{"cep":"01001000"}' http://localhost:8080/zipcode
```

Variáveis de ambiente úteis (veja `.env.example`):
- `SERVICE_B_URL` — URL do service-b (padrão `http://service-b:8081` no docker)
- `OTEL_EXPORTER_OTLP_ENDPOINT` — endpoint OTLP do collector (padrão `http://otel-collector:4318` no docker)

Se for enviar para um repositório público, garanta que `go.sum` está incluído nos diretórios `service-a` e `service-b` (já incluídos aqui) para que quem baixar não precise rebuildar dependências.
