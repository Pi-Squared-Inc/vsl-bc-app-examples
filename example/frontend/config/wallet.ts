import { WagmiAdapter } from "@reown/appkit-adapter-wagmi";
import { mainnet } from "wagmi/chains";

// Project ID for wallet connection
export const projectId = "7df920667d2a614a48e0477ee260175b";

// Support only Ethereum mainnet
export const networks = [mainnet];

if (!process.env.NEXT_PUBLIC_REOWN_PROJECT_ID) {
  console.warn("Reown Project ID is not defined. Using placeholder ID for development. Some features may be limited.");
}

// Set up the Wagmi Adapter
export const wagmiAdapter = new WagmiAdapter({
  networks,
  projectId,
  ssr: true,
});

export const wagmiConfig = wagmiAdapter.wagmiConfig;
