import { mdiLoading } from "@mdi/js";
import Icon from "@mdi/react";
import { useAppKit } from "@reown/appkit/react";
import { waitForTransactionReceipt } from "@wagmi/core";
import { useMemo, useState } from "react";
import { Hex, parseUnits } from "viem";
import { useAccount, useSendTransaction, useSwitchChain } from "wagmi";
import { Button } from "../../components/ui/button";
import { wagmiConfig } from "../../config/appkit";
import {
  mintPublicFunction,
  sourceChainId,
  sourceChainTokenContractAddress,
  sourceChainWeb3,
} from "../../config/constant";

type Props = {
  onFaucet: () => void;
};

export default function AppBar({ onFaucet }: Props) {
  const { address } = useAccount();
  const { open } = useAppKit();
  const { switchChainAsync } = useSwitchChain();
  const { sendTransactionAsync } = useSendTransaction();
  const [isMinting, setIsMinting] = useState(false);
  const addressString = useMemo(() => {
    if (!address) return "";
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
  }, [address]);

  async function faucet() {
    setIsMinting(true);
    try {
      await switchChainAsync({
        chainId: sourceChainId,
      });
      const mintTransactionHash = await sendTransactionAsync({
        to: sourceChainTokenContractAddress! as Hex,
        data: sourceChainWeb3.eth.abi.encodeFunctionCall(mintPublicFunction, [
          address!,
          parseUnits("100", 18),
        ]) as Hex,
      });

      await waitForTransactionReceipt(wagmiConfig, {
        hash: mintTransactionHash,
        chainId: sourceChainId,
      });
    } catch (error) {
      console.log(error);
    } finally {
      onFaucet();
      setIsMinting(false);
    }
  }

  return (
    <div className="flex flex-row justify-end w-full border-b p-4 space-x-4">
      {address ? (
        <>
          <Button variant="default" onClick={faucet} disabled={isMinting}>
            {isMinting ? (
              <>
                <Icon
                  className="animate-spin mr-2"
                  path={mdiLoading}
                  size={1}
                />
                Minting...
              </>
            ) : (
              "Faucet"
            )}
          </Button>
          <Button
            variant="outline"
            onClick={() => {
              open();
            }}
          >
            {addressString}
          </Button>
        </>
      ) : (
        <Button
          onClick={() => {
            open();
          }}
        >
          Connect Wallet
        </Button>
      )}
    </div>
  );
}
