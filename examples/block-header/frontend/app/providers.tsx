"use client";

import { projectId, wagmiAdapter, wagmiConfig } from "@/config/wallet";
import { createAppKit } from "@reown/appkit/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactNode } from "react";
import { mainnet } from "viem/chains";
import { WagmiProvider } from "wagmi";

// Set up queryClient
const queryClient = new QueryClient();

// App metadata
const metadata = {
  name: "VSL AI Agents",
  description: "VSL AI Agents Demo",
  url: "https://vsl-ai-agents.example.com",
  icons: ["https://example.com/logo.png"],
};

// Initialize AppKit with mainnet
createAppKit({
  adapters: [wagmiAdapter],
  projectId: projectId,
  networks: [mainnet],
  defaultNetwork: mainnet,
  metadata,
  features: {
    analytics: true,
  },
});

export function Providers({ children }: { children: ReactNode }) {
  return (
    <WagmiProvider config={wagmiConfig}>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </WagmiProvider>
  );
}
