import { Web3 } from "web3";
import {
  approveFunction,
  balanceOfFunction,
  sourceChainManagerContractAddress,
  transferFunction,
} from "../config/constant";

export async function balanceOf(
  chain: Web3,
  tokenContractAddress: string,
  address: string
) {
  const result = await chain.eth.call({
    to: tokenContractAddress!,
    data: chain.eth.abi.encodeFunctionCall(balanceOfFunction, [address]),
  });
  return chain.eth.abi.decodeParameter("uint256", result) as bigint;
}

export async function approve(
  chain: Web3,
  tokenContractAddress: string,
  address: string,
  amount: bigint
) {
  return await chain.eth.sendTransaction({
    to: tokenContractAddress!,
    data: chain.eth.abi.encodeFunctionCall(approveFunction, [address, amount]),
  });
}

export async function transfer(
  chain: Web3,
  amount: bigint,
  recipientChain: number,
  recipient: string
) {
  return await chain.eth.sendTransaction({
    to: sourceChainManagerContractAddress!,
    data: chain.eth.abi.encodeFunctionCall(transferFunction, [
      amount,
      recipientChain,
      Web3.utils.padLeft(recipient, 24),
    ]),
  });
}
