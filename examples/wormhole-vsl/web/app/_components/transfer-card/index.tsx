"use client";

import { mdiArrowDown, mdiRefresh } from "@mdi/js";
import Icon from "@mdi/react";
import { ReloadIcon } from "@radix-ui/react-icons";
import { useAppKit } from "@reown/appkit/react";
import { FunctionComponent, ReactNode } from "react";
import { formatUnits } from "viem";
import { useAccount } from "wagmi";
import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import TransferForm from "./form";

interface TransferCardProps {
  sourceChainTokenBalance?: bigint;
  destinationChainTokenBalance?: bigint;
  onRefreshBalance: () => void;
  onTransferSuccess: () => void;
}

const TransferCard: FunctionComponent<TransferCardProps> = ({
  sourceChainTokenBalance,
  destinationChainTokenBalance,
  onRefreshBalance,
  onTransferSuccess,
}) => {
  const { open } = useAppKit();
  const { address } = useAccount();

  function chainBalance(chainName: ReactNode, balance?: BigInt) {
    return (
      <Card className="w-full">
        <CardHeader>
          <CardTitle className="flex flex-row items-center space-x-2">{chainName}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col space-y-2">
            <div className="flex flex-row items-center space-x-2"></div>
            <div className="flex flex-row space-x-2 items-center">
              <span>Balance:</span>
              {!address ? (
                <span>Connect wallet to view balance</span>
              ) : balance !== undefined ? (
                <>
                  <span className="font-semibold">{parseFloat(formatUnits(balance.valueOf(), 18)).toFixed(4) + " PT"}</span>
                  <Button variant="outline" size="icon" onClick={onRefreshBalance}>
                    <Icon path={mdiRefresh} size={0.8} />
                  </Button>
                </>
              ) : (
                <ReloadIcon className="animate-spin" />
              )}
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  function transferForm() {
    if (!address) {
      return <Button onClick={() => open()}>Connect Wallet</Button>;
    }
    if (sourceChainTokenBalance === undefined || destinationChainTokenBalance === undefined) {
      return null;
    }
    return (
      <TransferForm
        address={address}
        sourceChainTokenBalance={sourceChainTokenBalance}
        onTransferSuccess={onTransferSuccess}
      />
    );
  }

  return (
    <Card className="min-w-[500px] max-w-[500px] bg-gradient-radial">
      <CardHeader>
        <CardTitle>Transfer</CardTitle>
        <CardDescription>Transfer token between chains</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col space-y-6">
        <div className="flex flex-col space-y-4 items-center">
          {chainBalance("Sepolia", sourceChainTokenBalance)}
          <div className="shadow-gradient-radial rounded-full size-10 flex flex-row items-center justify-center">
            <Icon path={mdiArrowDown} size={1} />
          </div>
          {chainBalance("Arbitrum Sepolia", destinationChainTokenBalance)}
        </div>
        {transferForm()}
      </CardContent>
    </Card>
  );
};

export default TransferCard;
