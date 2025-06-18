import clsx from "clsx";
import type { Metadata } from "next";
import localFont from "next/font/local";
import { headers } from "next/headers";
import ContextProvider from "../components/common/context-provider";
import { Toaster } from "../components/ui/toaster";
import "./globals.css";

const clashGrotesk = localFont({
  src: "./_fonts/ClashGrotesk-Variable.woff2",
  display: "swap",
  variable: "--font-clash-grotesk",
});

export const metadata: Metadata = {
  title: "Wormhole Multichain Demo",
  description: "Wormhole Multichain Demo",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const cookies = headers().get("cookie");

  return (
    <html lang="en">
      <body
        className={clsx(
          clashGrotesk.variable,
          "antialiased dark h-screen w-screen"
        )}
      >
        <Toaster />
        <ContextProvider cookies={cookies}>{children}</ContextProvider>
      </body>
    </html>
  );
}
