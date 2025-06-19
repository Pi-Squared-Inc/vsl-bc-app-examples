import { Toaster } from "@/components/ui/sonner";
import { GeistMono } from "geist/font/mono";
import type { Metadata } from "next";
import getConfig from "next/config";
import { headers } from "next/headers";
import { cookieToInitialState } from "wagmi";
import "./globals.css";
import { Providers } from "./providers";

export const metadata: Metadata = {
  title: "Trusted Execution Environment (TEE) Attestation",
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const initialState = cookieToInitialState(
    getConfig(),
    (await headers()).get("cookie")
  );

  return (
    <html lang="en" className="dark">
      <body
        className={`${GeistMono.variable} font-sans bg-background text-foreground`}
      >
        <Providers initialState={initialState}>
          {children}
          <Toaster />
        </Providers>
      </body>
    </html>
  );
}
