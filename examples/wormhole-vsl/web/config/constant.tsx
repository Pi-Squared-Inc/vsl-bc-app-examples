import {
  arbitrumSepolia,
  sepolia,
  type AppKitNetwork,
} from "@reown/appkit/networks";
import { Web3 } from "web3";

export const sourceChainId = parseInt(
  process.env.NEXT_PUBLIC_SOURCE_CHAIN_ID as string
);
export const destinationChainId = parseInt(
  process.env.NEXT_PUBLIC_DEST_CHAIN_ID as string
);
export const sourceWormholeId = parseInt(
  process.env.NEXT_PUBLIC_SOURCE_WORMHOLE_ID as string
);
export const destinationWormholeId = parseInt(
  process.env.NEXT_PUBLIC_DEST_WORMHOLE_ID as string
);
export const sourceChainTokenContractAddress =
  process.env.NEXT_PUBLIC_SOURCE_TOKEN_ADDRESS;
export const sourceChainManagerContractAddress =
  process.env.NEXT_PUBLIC_SOURCE_MANAGER_ADDRESS;
export const sourceChainVSLContractAddress =
  process.env.NEXT_PUBLIC_SOURCE_VSL_ADDRESS;
export const destChainTokenContractAddress =
  process.env.NEXT_PUBLIC_DEST_TOKEN_ADDRESS;
export const destChainVSLContractAddress =
  process.env.NEXT_PUBLIC_DEST_VSL_ADDRESS;
export const explorerUrl = process.env.NEXT_PUBLIC_VSL_EXPLORER_URL;

export const networks: AppKitNetwork[] = [sepolia, arbitrumSepolia];

export const sourceChainWeb3 = new Web3(process.env.NEXT_PUBLIC_SOURCE_API);
export const destinationChainWeb3 = new Web3(process.env.NEXT_PUBLIC_DEST_API);
export const backendAPIEndpoint = process.env.NEXT_PUBLIC_BACKEND_API;

export const balanceOfFunction = {
  name: "balanceOf",
  type: "function",
  inputs: [
    {
      type: "address",
      name: "account",
    },
  ],
};

export const mintPublicFunction = {
  name: "mintPublic",
  type: "function",
  inputs: [
    {
      type: "address",
      name: "account",
    },
    {
      type: "uint256",
      name: "amount",
    },
  ],
};

export const approveFunction = {
  name: "approve",
  type: "function",
  inputs: [
    {
      type: "address",
      name: "spender",
    },
    {
      type: "uint256",
      name: "value",
    },
  ],
};

export const transferFunction = {
  name: "transfer",
  type: "function",
  inputs: [
    {
      type: "uint256",
      name: "amount",
    },
    {
      type: "uint16",
      name: "recipientChain",
    },
    {
      type: "bytes32",
      name: "recipient",
    },
  ],
};

export const genStateQueryClaimEvent = {
  name: "genStateQueryClaimEvent",
  type: "event",
  inputs: [
    {
      type: "uint16",
      name: "srcChainid",
    },
    {
      type: "uint16",
      name: "destChainid",
    },
    {
      type: "uint",
      name: "blockNumber",
    },
    {
      type: "address",
      name: "contractAddress",
    },
    {
      type: "bytes",
      name: "viewFunctionEncoding",
    },
  ],
};

export const deliverClaimFunction = {
  name: "deliverClaim",
  type: "function",
  inputs: [
    {
      type: "bytes",
      name: "claim",
    },
  ],
};

export const executeTransferFunction = {
  name: "executeTransfer",
  type: "function",
  inputs: [
    {
      type: "bytes",
      name: "claimInput",
    },
  ],
};
