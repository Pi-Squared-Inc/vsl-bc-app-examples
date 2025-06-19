// config/index.tsx

import { WagmiAdapter } from "@reown/appkit-adapter-wagmi";
import { cookieStorage, createStorage } from "@wagmi/core";
import { networks } from "./constant";

// Your WalletConnect Cloud project ID
export const projectId = "f12eec2f410761a0eae3eb54480eccdb";

export const wagmiAdapter = new WagmiAdapter({
  storage: createStorage({
    storage: cookieStorage,
  }),
  ssr: true,
  projectId,
  networks: networks,
});

export const wagmiConfig = wagmiAdapter.wagmiConfig;
