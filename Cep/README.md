# Cep service

Objetivo: Recebe um CEP, consulta ViaCEP para encontrar a localização e retorna as temperaturas atuais em Celsius, Fahrenheit e Kelvin usando WeatherAPI.

Endpoints
- GET /zipcode/{cep}

Requisitos de ambiente
- WEATHER_API_KEY: chave da WeatherAPI

Como rodar

1. Copie `.env.example` para `.env` e defina `WEATHER_API_KEY`.
2. Docker Compose:

   docker-compose up --build

Testes

go test ./...

Deploy

Use o `Dockerfile` e as instruções do Google Cloud Run para subir a imagem:

gcloud builds submit --tag gcr.io/PROJECT/cep
gcloud run deploy cep --image gcr.io/PROJECT/cep --platform managed --region YOUR_REGION --allow-unauthenticated
