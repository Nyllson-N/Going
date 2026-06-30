# Docker Manager

> Backend e biblioteca pública para gerenciamento de servidores Docker locais ou remotos.

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Build](https://img.shields.io/badge/build-passing-brightgreen.svg)]()
[![Version](https://img.shields.io/badge/version-1.0.0-orange.svg)]()

## Visão geral

Docker Manager é uma solução composta por uma **API backend** e uma **biblioteca cliente pública** que permitem gerenciar containers, imagens, volumes e redes Docker em hosts locais ou remotos a partir de uma interface unificada. Conecte-se a múltiplos daemons Docker simultaneamente, execute operações em lote e monitore recursos em tempo real.

## Recursos

- **Multi-host** — gerencie vários servidores Docker (local, SSH, TCP/TLS) a partir de um único ponto.
- **Containers** — criar, iniciar, parar, reiniciar, remover, inspecionar e acompanhar logs em streaming.
- **Imagens** — pull, push, build, tag e remoção, com suposto suporte a registries privados.
- **Volumes e redes** — criação, listagem e limpeza de recursos órfãos.
- **Monitoramento** — estatísticas de CPU, memória, rede e I/O em tempo real via WebSocket.
- **Autenticação** — controle de acesso por token (JWT) e por host.
- **Biblioteca cliente** — SDK público para integrar o backend em outras aplicações.

## Arquitetura

```
┌─────────────┐      ┌──────────────┐      ┌─────────────────┐
│   Client    │─────▶│   Backend    │─────▶│  Docker Daemon  │
│   (Lib/SDK) │ HTTP │   (API/WS)   │ API  │  (local/remoto) │
└─────────────┘      └──────────────┘      └─────────────────┘
```

O backend expõe uma API REST + WebSocket e se comunica com cada daemon Docker através do socket Unix local ou de conexões remotas (SSH / TCP+TLS). A biblioteca cliente abstrai essas chamadas.

## Instalação

### Backend

```bash
git clone https://github.com/seu-usuario/docker-manager.git
cd docker-manager
npm install
npm run build
```

### Biblioteca cliente

```bash
npm install docker-manager-client
```

## Configuração

Crie um arquivo `.env` na raiz do backend:

```env
PORT=8080
JWT_SECRET=sua-chave-secreta
# Hosts Docker (separados por vírgula)
DOCKER_HOSTS=unix:///var/run/docker.sock,tcp://192.168.0.10:2376
# Caminho dos certificados TLS para hosts remotos (opcional)
DOCKER_TLS_CERT_PATH=/etc/docker/certs
```

## Uso

### Iniciando o backend

```bash
npm start
# Servidor disponível em http://localhost:8080
```

### Usando a biblioteca cliente

```javascript
import { DockerManager } from "docker-manager-client";

const manager = new DockerManager({
  baseUrl: "http://localhost:8080",
  token: "seu-jwt-token",
});

// Conectar a um host remoto
await manager.hosts.add({
  name: "producao",
  url: "tcp://192.168.0.10:2376",
  tls: true,
});

// Listar containers
const containers = await manager.containers.list({ host: "producao" });

// Iniciar um container
await manager.containers.start("producao", "meu-container");

// Logs em streaming
manager.containers.logs("producao", "meu-container").on("data", (line) => {
  console.log(line);
});
```

## API REST

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/hosts` | Lista os hosts conectados |
| `POST` | `/hosts` | Adiciona um novo host |
| `DELETE` | `/hosts/:id` | Remove um host |
| `GET` | `/hosts/:id/containers` | Lista containers do host |
| `POST` | `/hosts/:id/containers/:name/start` | Inicia um container |
| `POST` | `/hosts/:id/containers/:name/stop` | Para um container |
| `GET` | `/hosts/:id/images` | Lista imagens |
| `GET` | `/hosts/:id/volumes` | Lista volumes |
| `GET` | `/hosts/:id/networks` | Lista redes |

Documentação completa disponível em `/docs` (Swagger) após iniciar o backend.

## Conectando hosts remotos

**Via TCP + TLS** (recomendado para produção):

```bash
docker -H tcp://0.0.0.0:2376 --tlsverify \
  --tlscacert=ca.pem --tlscert=server-cert.pem --tlskey=server-key.pem
```

**Via SSH:**

```env
DOCKER_HOSTS=ssh://usuario@192.168.0.10
```

## Requisitos

- Node.js 18+
- Docker Engine 20.10+
- Acesso ao socket Docker ou a um daemon remoto exposto

## Desenvolvimento

```bash
npm run dev      # modo watch
npm test         # executa os testes
npm run lint     # verifica o código
```

## Contribuindo

Contribuições são bem-vindas. Abra uma issue para discutir mudanças significativas antes de enviar um pull request. Siga o padrão de commits e garanta que os testes passem.

## Licença

Distribuído sob a licença MIT. Veja [LICENSE](LICENSE) para mais detalhes.
