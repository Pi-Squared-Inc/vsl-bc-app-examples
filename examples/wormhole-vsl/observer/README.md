# Observer

The observer is the service that observes the source chain transfer events to generate claims and submit them to the backend.

## How to run

1. Install dependencies:
   ```bash
    go mod download
   ```
2. Create `.env` file:
   ```bash
    cp sample.env .env
   ```
3. Start the observer:
   ```bash
    go run main.go
   ```

## Example

Please refer to the [example](../README.md) to see how to use the observer.

## Mode

You can set the mode to `auto` or `manual` on the `.env` file.

- `auto`: The observer will automatically generate claims by monitoring the source chain events
- `manual`: The observer will generate claims manually by calling the `/generate_claim` API

### Manual Mode

You can using following command to call the `/generate_claim` API:

```bash
curl --request POST \
  --url http://localhost:10001/generate_claim \
  --header 'Content-Type: application/json' \
  --data '{
	"transaction_hash": "0x..."
}'
```
