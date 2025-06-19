#! /bin/bash

pnpm install
npx playwright install --with-deps
unzip mock.zip -d ./
cp sample.env .env
xvfb-run npx playwright test
