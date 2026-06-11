# Go-Gallery

Go-Gallery é uma plataforma robusta desenvolvida em Go para o gerenciamento de portfólios fotográficos, permitindo que fotógrafos organizem álbuns, gerenciem clientes e disponibilizem fotografias para visualização e venda.

## 🚀 Tecnologias Utilizadas

- **Linguagem:** Go (v1.25.0)
- **Banco de Dados:** PostgreSQL (v15)
- **Object Storage:** MinIO (compatível com AWS S3)
- **Ferramentas de Banco:** 
  - [sqlc](https://sqlc.dev/): Geração de código Go tipo-seguro a partir de SQL.
  - [pgx](https://github.com/jackc/pgx): Driver e toolkit para PostgreSQL.
- **Infraestrutura:** Docker e Docker Compose.

## 📋 Funcionalidades Principais (Em desenvolvimento)

- **Gestão de Usuários:** Diferenciação entre fotógrafos, administradores e clientes.
- **Álbuns Fotográficos:** Criação de álbuns vinculados a eventos, com suporte a precificação por álbum ou por fotografia individual.
- **Galeria de Fotos:** Gerenciamento de fotos em alta e baixa resolução (armazenadas no MinIO).
- **Associações:** Vinculação de fotógrafos a álbuns e liberação de acesso para clientes específicos.

## 📁 Estrutura do Projeto

```text
├── cmd/
│   └── migrations/        # Lógica para execução de migrações do banco de dados
├── db/
│   ├── queries/           # Definições de consultas SQL para o sqlc
│   └── schema/            # Scripts DDL (Data Definition Language) para o esquema do banco
├── internal/
│   ├── db/                # Código Go gerado pelo sqlc e lógica de conexão
│   └── s3/                # Cliente Singleton para integração com S3/MinIO
├── main.go                # Ponto de entrada da aplicação
├── docker-compose.yaml    # Orquestração do PostgreSQL e MinIO
└── sqlc.yaml              # Configuração do gerador sqlc
```

## 🛠️ Como Executar

### Pré-requisitos
- Docker e Docker Compose instalados.
- Go instalado (para execução local).

### Passo 1: Subir a infraestrutura
Inicie os serviços de banco de dados e storage:
```bash
docker-compose up -d
```
- **PostgreSQL:** Acessível em `localhost:5432`
- **MinIO Console:** Acessível em `http://localhost:9001` (Credenciais: `admin` / senha no `docker-compose.yaml`)

### Passo 2: Configurar o Banco de Dados
Execute as migrações para criar as tabelas necessárias:
```bash
go run cmd/migrations/run.go up
```
Para resetar o banco de dados (CUIDADO: apaga todos os dados):
```bash
go run cmd/migrations/run.go reset
```

### Passo 3: Executar a aplicação
```bash
go run main.go
```

## ⚙️ Configuração do sqlc
Para gerar o código do banco de dados após alterar os arquivos em `db/queries/` ou `db/schema/`, execute:
```bash
sqlc generate
```
