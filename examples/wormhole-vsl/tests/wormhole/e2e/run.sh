#! /bin/bash

pnpm install
npx playwright install --with-deps
unzip mock.zip -d ./
cp sample.env .env
pnpm exec playwright test
