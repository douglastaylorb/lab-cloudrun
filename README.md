### Acessos
Executar `docker compose up -d` para testar via docker
Endereço para testar via cloudrun: https://lab-cloudrun-905306807446.us-central1.run.app/weather?cep=29108790

#### Observação
Substituir `weatherAPIKey` por uma API KEY válida para testes locais

### Endpoint
`GET /weather?cep={CEP}`

### Exemplos de Uso

#### Sucesso (CEP válido)
```bash
curl "http://localhost:8080/weather?cep=01310930"
```

**Resposta:**
```json
{
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

#### CEP inválido (formato incorreto)
```bash
curl "http://localhost:8080/weather?cep=123"
```

**Resposta:**
```json
{
  "message": "invalid zipcode"
}
```
**Status:** 422

#### CEP não encontrado
```bash
curl "http://localhost:8080/weather?cep=99999999"
```
