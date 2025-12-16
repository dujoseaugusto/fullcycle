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

Aplicação em produção

Disponível em: https://fullcycle-549134865612.europe-west1.run.app/  


CI (GitHub Actions)

Este repositório inclui um workflow leve em `.github/workflows/ci.yml` que executa verificações de formatação (`gofmt`), `go vet`, `golangci-lint` e os testes em pushes/PRs para a branch `main`. Ele **não** faz deploy automático — use o console do GCP ou um workflow de deploy separado caso queira automatizar o deploy.

Se você ainda desejar um deploy automático via GitHub Actions, mantenha ou crie um workflow separado para deploy e configure os segredos do GCP conforme mostrado abaixo.

Criar service account mínima (exemplo):

```bash
gcloud iam service-accounts create github-actions-cd --display-name "GitHub Actions CD"
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID --member="serviceAccount:github-actions-cd@YOUR_PROJECT_ID.iam.gserviceaccount.com" --role="roles/run.admin"
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID --member="serviceAccount:github-actions-cd@YOUR_PROJECT_ID.iam.gserviceaccount.com" --role="roles/iam.serviceAccountUser"
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID --member="serviceAccount:github-actions-cd@YOUR_PROJECT_ID.iam.gserviceaccount.com" --role="roles/cloudbuild.builds.editor"

# Create key and add to GitHub secrets (GCP_SA_KEY)
gcloud iam service-accounts keys create key.json --iam-account=github-actions-cd@YOUR_PROJECT_ID.iam.gserviceaccount.com
```

Depois, cole o conteúdo de `key.json` como o segredo `GCP_SA_KEY` no repositório GitHub (Settings → Secrets & variables → Actions).

