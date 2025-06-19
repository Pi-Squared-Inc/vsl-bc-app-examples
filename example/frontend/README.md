# Hello World TEE Demo Frontend

## Prerequisites

- Node.js (v22.14.0 or higher)
- [Bun](https://bun.sh/docs/installation)

## Getting Started

1. [If necessary] Prepare the environment file.

    ```bash
    cp sample.env .env
    ```
    
    Update the appropriate backend API endpoint via this environment variable, `NEXT_PUBLIC_API_URL`.

3. Start the frontend.

    ```bash
    bun install
    bun run dev # for development
    # bun run build && bun run start # comment the line above and uncomment this line for production
    ```
