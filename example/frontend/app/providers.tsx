"use client";

import { wagmiConfig } from "@/config/wallet";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Provider as JotaiProvider } from "jotai";
import { type State, WagmiProvider } from "wagmi";

// Set up queryClient
const queryClient = new QueryClient();

export function Providers({
  children,
  initialState,
}: {
  children: React.ReactNode;
  initialState?: State;
}) {
  return (
    <WagmiProvider config={wagmiConfig} initialState={initialState}>
      <QueryClientProvider client={queryClient}>
        <JotaiProvider>{children}</JotaiProvider>
      </QueryClientProvider>
    </WagmiProvider>
  );
}
