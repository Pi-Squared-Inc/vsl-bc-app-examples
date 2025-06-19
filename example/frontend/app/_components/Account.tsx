"use client";

import { Button } from "@/components/ui/button";
import { shortenAddress } from "@/utils/format";
import { useAtom, useAtomValue } from "jotai";
import { useEffect } from "react";
import { useAccount, useConnect, useDisconnect } from "wagmi";
import {
  balanceAtom,
  fetchBalanceAtom,
  isFetchingBalanceAtom,
} from "../store/balance";

function BalanceDisplay() {
  const balance = useAtomValue(balanceAtom);
  const isFetching = useAtomValue(isFetchingBalanceAtom);

  if (isFetching && balance === null) {
    return (
      <span className="text-xs font-semibold text-primary">Fetching...</span>
    );
  }

  return (
    <span className="text-xs font-semibold text-primary">
      {Intl.NumberFormat('en-us', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format((balance ?? 0) / (10**18))} VSL
    </span>
  );
}

export default function WalletConnect() {
  const { address, isConnected, isConnecting } = useAccount();
  const { disconnect } = useDisconnect();
  const { connectors, connect } = useConnect();
  const [, fetchBalance] = useAtom(fetchBalanceAtom);

  useEffect(() => {
    if (isConnected && address) {
      fetchBalance(address);
    }
  }, [isConnected, address, fetchBalance]);

  const handleConnect = () => {
    const metaMaskConnector = connectors.find((c) => c.id === "io.metamask");
    if (metaMaskConnector) {
      connect({ connector: metaMaskConnector });
    }
  };

  if (isConnected) {
    return (
      <div className="flex items-center gap-4 rounded-lg bg-gray-900 p-3">
        <div className="flex flex-col text-left">
          <span className="text-sm font-medium text-white">
            {shortenAddress(address!)}
          </span>
          <BalanceDisplay />
        </div>
        <Button size="sm" variant="outline" onClick={() => disconnect()}>
          Disconnect
        </Button>
      </div>
    );
  }

  const metaMaskConnector = connectors.find((c) => c.id === "io.metamask");

  if (!metaMaskConnector) {
    return (
      <div className="flex flex-col items-center gap-2">
        <Button variant="outline" disabled>
          Connect
        </Button>
        <p className="text-sm">
          Please install{" "}
          <a
            className="text-primary underline"
            href="https://metamask.io/download"
            target="_blank"
            rel="noopener noreferrer"
          >
            MetaMask
          </a>{" "}
          extension first.
        </p>
      </div>
    );
  }

  return (
    <Button onClick={handleConnect} disabled={isConnecting}>
      {isConnecting ? "Connecting..." : "Connect"}
    </Button>
  );
}
