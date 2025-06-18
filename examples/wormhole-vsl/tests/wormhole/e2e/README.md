# E2E(End-to-End) Tests for Wormhole

## Prerequisites

- [Playwright](https://playwright.dev/docs/intro)
- [Node.js (LTS version)](https://nodejs.org)
- [pnpm](https://pnpm.io/installation)
- unzip

## Run the tests with mock data

### Setup

```bash
pnpm install
unzip mock.zip -d ./
cp sample.env .env
```

### Update the .env file

- METAMASK_PASSWORD: The password of the Metamask account, 12345678 is used for the mock data
- WORMHOLE_DEMO_URL: The URL of the Wormhole demo page

## Run the tests

```bash
pnpm exec playwright test
```

## Run the tests on headless mode (without GUI)

```bash
xvfb-run npx playwright test
```

## How to prepare the user-data folder

TBD