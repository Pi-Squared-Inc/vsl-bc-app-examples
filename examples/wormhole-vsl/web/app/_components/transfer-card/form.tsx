"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { ReloadIcon } from "@radix-ui/react-icons";
import { waitForTransactionReceipt } from "@wagmi/core";
import axios from "axios";
import { FunctionComponent, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { formatUnits, Hex, parseUnits } from "viem";
import { useAccount, useSendTransaction, useSwitchChain } from "wagmi";
import Web3 from "web3";
import { z } from "zod";
import { Button } from "../../../components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../../../components/ui/form";
import { Input } from "../../../components/ui/input";
import { wagmiConfig } from "../../../config/appkit";
import {
  approveFunction,
  backendAPIEndpoint,
  destinationWormholeId,
  sourceChainId,
  sourceChainManagerContractAddress,
  sourceChainTokenContractAddress,
  sourceChainWeb3,
  transferFunction,
} from "../../../config/constant";

interface TransferFormProps {
  address: string;
  sourceChainTokenBalance: bigint;
  onTransferSuccess?: () => void;
}

const TransferForm: FunctionComponent<TransferFormProps> = ({
  address,
  sourceChainTokenBalance,
  onTransferSuccess,
}) => {
  const { address: accountAddress } = useAccount();
  const { switchChainAsync } = useSwitchChain();
  const { sendTransactionAsync } = useSendTransaction();
  const transferFormSchema = useMemo(() => {
    return z.object({
      amount: z
        .string({
          required_error: "Amount is required",
        })
        .regex(/^(0|[1-9]\d*)(\.\d{1,18})?$/g, "Amount must be a valid number")
        .refine((value) => {
          try {
            const bigIntValue = parseUnits(value, 18);
            return bigIntValue > 0 && bigIntValue <= sourceChainTokenBalance;
          } catch (error) {
            return false;
          }
        }, "Amount must be greater than 0 and less than or equal to source chain balance"),
    });
  }, [sourceChainTokenBalance]);
  const form = useForm<z.infer<typeof transferFormSchema>>({
    mode: "onSubmit",
    resolver: zodResolver(transferFormSchema),
  });
  const [isTransfering, setIsTransfering] = useState(false);
  const [isApproving, setIsApproving] = useState(false);

  async function onApprove() {
    const isValid = await form.trigger();
    if (!isValid) {
      return;
    }
    const amount = parseUnits(form.getValues("amount"), 18);
    setIsApproving(true);
    try {
      await switchChainAsync({
        chainId: sourceChainId,
      });

      // Execute the approve transaction
      const approveTransactionHash = await sendTransactionAsync({
        to: sourceChainTokenContractAddress! as Hex,
        data: sourceChainWeb3.eth.abi.encodeFunctionCall(approveFunction, [
          sourceChainManagerContractAddress!,
          amount,
        ]) as Hex,
      });
      await waitForTransactionReceipt(wagmiConfig, {
        hash: approveTransactionHash,
        chainId: sourceChainId,
      });
    } catch (error) {
      console.log(error);
    } finally {
      setIsApproving(false);
    }
  }

  async function onSubmit(values: z.infer<typeof transferFormSchema>) {
    const amount = parseUnits(values.amount, 18);
    setIsTransfering(true);
    try {
      await switchChainAsync({
        chainId: sourceChainId,
      });

      // Execute the bridge transaction
      const bridgeTransaction = await sendTransactionAsync({
        to: sourceChainManagerContractAddress! as Hex,
        data: sourceChainWeb3.eth.abi.encodeFunctionCall(transferFunction, [
          amount,
          destinationWormholeId,
          Web3.utils.padLeft(address!, 64),
        ]) as Hex,
      });
      await waitForTransactionReceipt(wagmiConfig, {
        hash: bridgeTransaction,
        chainId: sourceChainId,
      });

      // Upsert the claim to the backend with the source transaction hash
      await axios.put(
        backendAPIEndpoint + "/claim-by-source-tx/" + bridgeTransaction,
        {
          source_transaction_hash: bridgeTransaction,
          address: accountAddress?.toLowerCase(),
        }
      );

      // // Find the transfer event
      // const transferEvent = bridgeTransactionReceipt.logs.find(
      //   (log) =>
      //     log.address?.toLowerCase() ==
      //     sourceChainVSLContractAddress!.toLowerCase()
      // );
      // if (!transferEvent) {
      //   throw new Error("Transfer event not found");
      // }

      // const decodedLog = sourceChainWeb3.eth.abi.decodeLog(
      //   genStateQueryClaimEvent.inputs,
      //   Web3.utils.bytesToHex(transferEvent.data!),
      //   transferEvent.topics!.map((topic) => Web3.utils.bytesToHex(topic))
      // );

      // const blockNumber = toHex(decodedLog.blockNumber as bigint);
      // const contractAddress = decodedLog.contractAddress as Hex;
      // const viewFunctionEncoding = decodedLog.viewFunctionEncoding as Hex;

      // // Construct the claim id
      // const rawClaimId = `${numberToHex(
      //   sourceChainId
      // )}:${blockNumber}:${contractAddress}:${viewFunctionEncoding}` as Hex;
      // const claimId = sha256(rawClaimId).slice(2);

      form.setValue("amount", "");
      onTransferSuccess && onTransferSuccess();
    } catch (error) {
      console.log(error);
    } finally {
      setIsTransfering(false);
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="flex flex-col space-y-6"
      >
        <FormField
          control={form.control}
          name="amount"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Amount</FormLabel>
              <div className="flex flex-row space-x-2">
                <FormControl>
                  <Input
                    {...field}
                    className="bg-card"
                    defaultValue={"0"}
                    type="text"
                    inputMode="decimal"
                    placeholder="0"
                    disabled={isTransfering}
                  />
                </FormControl>
                <Button
                  type="button"
                  disabled={isTransfering}
                  onClick={() => {
                    form.setValue(
                      "amount",
                      formatUnits(sourceChainTokenBalance, 18)
                    );
                  }}
                >
                  Max
                </Button>
              </div>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className="flex flex-col space-y-2">
          <Button
            type="button"
            variant="outline"
            disabled={isApproving || isTransfering}
            onClick={onApprove}
          >
            {isApproving ? (
              <>
                <ReloadIcon className="mr-2 h-4 w-4 animate-spin" />
                Approving...
              </>
            ) : (
              "Approve"
            )}
          </Button>
          <Button type="submit" disabled={isApproving || isTransfering}>
            {isTransfering ? (
              <>
                <ReloadIcon className="mr-2 h-4 w-4 animate-spin" />
                Transfering...
              </>
            ) : (
              "Transfer"
            )}
          </Button>
        </div>
      </form>
    </Form>
  );
};

export default TransferForm;
