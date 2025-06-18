"use client";

import axios from "axios";
import { isEqual } from "es-toolkit/compat";
import { useEffect, useState } from "react";
import { useInterval } from "usehooks-ts";
import { useAccount } from "wagmi";
import { Tabs, TabsList, TabsTrigger } from "../components/ui/tabs";
import {
  backendAPIEndpoint,
  destChainTokenContractAddress,
  destinationChainWeb3,
  sourceChainTokenContractAddress,
  sourceChainWeb3,
} from "../config/constant";
import { balanceOf } from "../lib/calls";
import Pi2Logo from "./_assets/logo.svg";
import WormholeLogo from "./_assets/wormhole.svg";
import AppBar from "./_components/appbar";
import Footer from "./_components/footer";
import TransferCard from "./_components/transfer-card";
import TransferTransactionList from "./_components/transfer-transaction-list";

type Tab = "transfer" | "history";

export default function Home() {
  const { address } = useAccount();
  const [claims, setClaims] = useState<unknown[]>([]);
  const [tab, setTab] = useState<Tab>("transfer");
  const [sourceChainTokenBalance, setSourceChainTokenBalance] =
    useState<bigint>();
  const [destinationChainTokenBalance, setDestinationChainTokenBalance] =
    useState<bigint>();

  useEffect(() => {
    if (address) {
      fetchBalances();
      fetchClaims();
    }
  }, [address]);

  useInterval(() => {
    fetchClaims();
  }, 1000);

  useInterval(() => {
    fetchBalances();
  }, 3000);

  // Fetch balances from chains
  async function fetchBalances() {
    const sourceBalance = await balanceOf(
      sourceChainWeb3,
      sourceChainTokenContractAddress!,
      address!
    );
    const destinationBalance = await balanceOf(
      destinationChainWeb3,
      destChainTokenContractAddress!,
      address!
    );
    setSourceChainTokenBalance(sourceBalance);
    setDestinationChainTokenBalance(destinationBalance);
  }

  // Fetch claims from backend
  function fetchClaims() {
    axios
      .get(backendAPIEndpoint + "/claims", {
        params: {
          address,
        },
      })
      .then((response) => {
        if (isEqual(response.data, claims)) {
          return;
        }
        setClaims(response.data);
      });
  }

  return (
    <div className="h-full w-full flex flex-col font-clashGrotesk bg-gradient-radial space-y-20 items-center overflow-auto">
      <AppBar
        onFaucet={() => {
          fetchBalances();
        }}
      />
      <div className="text-4xl font-medium">Wormhole Multichain Demo</div>
      <div className="w-full flex flex-col items-center space-y-6">
        <Tabs defaultValue={tab}>
          <TabsList>
            <TabsTrigger value="transfer" onClick={() => setTab("transfer")}>
              Transfer
            </TabsTrigger>
            <TabsTrigger value="history" onClick={() => setTab("history")}>
              History
            </TabsTrigger>
          </TabsList>
        </Tabs>
        {tab === "transfer" && (
          <>
            <TransferCard
              sourceChainTokenBalance={sourceChainTokenBalance}
              destinationChainTokenBalance={destinationChainTokenBalance}
              onRefreshBalance={() => {
                fetchBalances();
              }}
              onTransferSuccess={() => {
                fetchBalances();
                // createTransaction(transactionHash, claimId);
              }}
            />
            <div className="flex flex-row items-center space-x-3">
              <div>Powered by</div>
              <Pi2Logo />
              <span>x</span>
              <WormholeLogo className="h-[14px]" />
            </div>
          </>
        )}
        {tab === "history" && (
          <TransferTransactionList
            claims={claims}
            onCompleteTransfer={() => {
              fetchBalances();
              fetchClaims();
            }}
          />
        )}
      </div>
      <Footer />
    </div>
  );
}
