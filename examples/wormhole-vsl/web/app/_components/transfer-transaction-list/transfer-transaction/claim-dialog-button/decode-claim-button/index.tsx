"use client";

import { get } from "es-toolkit/compat";
import React, { useMemo, useState } from "react";
import {
  bytesToBigInt,
  bytesToHex,
  bytesToNumber,
  Hex,
  hexToBytes,
  hexToNumber,
} from "viem";
import LabelField from "../../../../../../components/common/label-field";
import { Button } from "../../../../../../components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../../../../../../components/ui/dialog";
import { sourceChainWeb3 } from "../../../../../../config/constant";
import { Claim } from "../../../../../../lib/models";

interface DecodeClaimButtonProps {
  claimJSON: string;
}

const DecodeClaimButton: React.FC<DecodeClaimButtonProps> = ({ claimJSON }) => {
  const [showDecodedInputContent, setShowDecodedInputContent] = useState(false);
  const [showDecodedOutputContent, setShowDecodedOutputContent] =
    useState(false);
  const claim = useMemo<Claim | null>(() => {
    try {
      return JSON.parse(claimJSON);
    } catch (error) {
      return null;
    }
  }, [claimJSON]);
  const [claimInputHex, destChainId, nonce] = useMemo(() => {
    if (!claim) {
      return [null, null, null];
    }
    const claimInputHex = bytesToHex(
      new Uint8Array(Buffer.from(get(claim, "action.input")!, "base64"))
    );
    const parametersSignature = claimInputHex.slice(10);
    const parameters = sourceChainWeb3.eth.abi.decodeParameters(
      ["uint256", "address"],
      "0x" + parametersSignature
    );
    return [claimInputHex, parameters[0] as bigint, parameters[1] as Hex];
  }, [claim]);
  const [
    claimOutputHex,
    relayMessageSrcChainId,
    relayMessageDestChainId,
    relayMessageSenderTransceiver,
    relayMessageReceiverTransceiver,
    relayMessageTransceiverMessage,
  ] = useMemo(() => {
    if (!claim) {
      return [null, null, null, null, null];
    }
    const claimOutputHex = bytesToHex(
      new Uint8Array(Buffer.from(get(claim, "result")!, "base64"))
    );
    const decodedOutput = sourceChainWeb3.eth.abi.decodeParameters(
      ["bytes"],
      claimOutputHex
    );

    const relayMessage = sourceChainWeb3.eth.abi.decodeParameter(
      {
        RelayMessage: {
          srcChainid: "uint16",
          destChainid: "uint16",
          senderUslTransciever: "address",
          receiverUslTransciever: "address",
          transceiverMessage: "bytes",
        },
      },
      decodedOutput[0] as Hex
    );

    return [
      claimOutputHex,
      get(relayMessage, 0) as bigint,
      get(relayMessage, 1) as bigint,
      get(relayMessage, 2) as Hex,
      get(relayMessage, 3) as Hex,
      get(relayMessage, 4) as Hex,
    ];
  }, [claim]);
  const assumptions = useMemo(() => {
    if (!claim) {
      return null;
    }
    return JSON.stringify(get(claim, "assumptions"), null, 2);
  }, [claim]);
  const metadata = useMemo(() => {
    if (!claim) {
      return null;
    }
    return JSON.stringify(get(claim, "metadata"), null, 2);
  }, [claim]);
  const [
    transceiverMessagePrefixHex,
    transceiverMessageSourceNttManagerAddressHex,
    transceiverMessageRecipientNttManagerAddressHex,
    transceiverMessageNttManagerPayloadHex,
  ] = useMemo(() => {
    if (!relayMessageTransceiverMessage) {
      return [null, null, null, null];
    }
    let offset = 0;
    const transceiverMessageBytes = hexToBytes(relayMessageTransceiverMessage);
    const prefixHex = bytesToHex(
      transceiverMessageBytes.slice(offset, offset + 4)
    );
    offset += 4;
    const sourceNttManagerAddressHex = bytesToHex(
      transceiverMessageBytes.slice(offset, offset + 32)
    );
    offset += 32;
    const recipientNttManagerAddressHex = bytesToHex(
      transceiverMessageBytes.slice(offset, offset + 32)
    );
    offset += 32;
    const nttManagerPayloadLength = hexToNumber(
      bytesToHex(transceiverMessageBytes.slice(offset, offset + 2))
    );
    offset += 2;
    const nttManagerPayloadHex = bytesToHex(
      transceiverMessageBytes.slice(offset, offset + nttManagerPayloadLength)
    );
    return [
      prefixHex,
      sourceNttManagerAddressHex,
      recipientNttManagerAddressHex,
      nttManagerPayloadHex,
    ];
  }, [relayMessageTransceiverMessage]);
  const [
    nttManagerPayloadIdHex,
    nttManagerPayloadSenderHex,
    nttManagerPayloadPayloadHex,
  ] = useMemo(() => {
    if (!transceiverMessageNttManagerPayloadHex) {
      return [null, null, null];
    }
    let offset = 0;
    const nttManagerPayloadBytes = hexToBytes(
      transceiverMessageNttManagerPayloadHex
    );
    const idHex = bytesToHex(nttManagerPayloadBytes.slice(offset, offset + 32));
    offset += 32;
    const senderHex = bytesToHex(
      nttManagerPayloadBytes.slice(offset, offset + 32)
    );
    offset += 32;
    const payloadLength = bytesToNumber(
      nttManagerPayloadBytes.slice(offset, offset + 2)
    );
    offset += 2;
    const payloadHex = bytesToHex(
      nttManagerPayloadBytes.slice(offset, offset + payloadLength)
    );
    return [idHex, senderHex, payloadHex];
  }, [transceiverMessageNttManagerPayloadHex]);
  const [
    nativeTokenTransferPrefixHex,
    nativeTokenTransferNumDecimals,
    nativeTokenTransferAmountHex,
    nativeTokenTransferSourceTokenHex,
    nativeTokenTransferToHex,
    nativeTokenTransferToChain,
  ] = useMemo(() => {
    if (!nttManagerPayloadPayloadHex) {
      return [null, null, null];
    }
    let offset = 0;
    const nttManagerPayloadBytes = hexToBytes(nttManagerPayloadPayloadHex);
    const prefix = bytesToHex(nttManagerPayloadBytes.slice(offset, offset + 4));
    offset += 4;
    const numDecimals = bytesToNumber(
      nttManagerPayloadBytes.slice(offset, offset + 1)
    );
    offset += 1;
    const amountHex = bytesToBigInt(
      nttManagerPayloadBytes.slice(offset, offset + 8)
    );
    offset += 8;
    const sourceTokenHex = bytesToHex(
      nttManagerPayloadBytes.slice(offset, offset + 32)
    );
    offset += 32;
    const toHex = bytesToHex(nttManagerPayloadBytes.slice(offset, offset + 32));
    offset += 32;
    const toChain = bytesToNumber(
      nttManagerPayloadBytes.slice(offset, offset + 2)
    );
    return [prefix, numDecimals, amountHex, sourceTokenHex, toHex, toChain];
  }, [nttManagerPayloadPayloadHex]);

  function inputContent(decode?: boolean) {
    let content = `Signature: relays(uint16 destChainId, uint nonce)\n`;
    if (decode) {
      content += `\n[destChainId]: ${
        destChainId?.toString() ?? "N/A"
      }\n[nonce]: ${hexToNumber(nonce ?? "0x") ?? "N/A"}`;
    } else {
      content += `\nHex: ${claimInputHex}`;
    }
    return content;
  }

  function outputContent(decode?: boolean) {
    let content = `Signature: RelayMessage { uint16 srcChainid, uint16 destChainid, address senderUslTransceiver, address receiverUslTransceiver, bytes transceiverMessage }\n`;
    if (decode) {
      content += `\n[srcChainId]: ${
        relayMessageSrcChainId?.toString() ?? "N/A"
      }`;
      content += `\n[destChainId]: ${
        relayMessageDestChainId?.toString() ?? "N/A"
      }`;
      content += `\n[senderUslTransceiver]: ${
        relayMessageSenderTransceiver?.toString() ?? "N/A"
      }`;
      content += `\n[receiverUslTransceiver]: ${
        relayMessageReceiverTransceiver?.toString() ?? "N/A"
      }`;
      content += `\n[transceiverMessage]:`;
      content += `\n  [prefix]: ${transceiverMessagePrefixHex ?? "N/A"}`;
      content += `\n  [sourceNttManagerAddress]: ${
        transceiverMessageSourceNttManagerAddressHex ?? "N/A"
      }`;
      content += `\n  [recipientNttManagerAddress]: ${
        transceiverMessageRecipientNttManagerAddressHex ?? "N/A"
      }`;
      content += `\n  [nttManagerPayload]:`;
      content += `\n    [id]: ${nttManagerPayloadIdHex ?? "N/A"}`;
      content += `\n    [sender]: ${nttManagerPayloadSenderHex ?? "N/A"}`;
      content += `\n    [payload]:`;
      content += `\n      [prefix]: ${nativeTokenTransferPrefixHex ?? "N/A"}`;
      content += `\n      [numDecimals]: ${
        nativeTokenTransferNumDecimals ?? "N/A"
      }`;
      content += `\n      [amount]: ${nativeTokenTransferAmountHex ?? "N/A"}`;
      content += `\n      [sourceToken]: ${
        nativeTokenTransferSourceTokenHex ?? "N/A"
      }`;
      content += `\n      [to]: ${nativeTokenTransferToHex ?? "N/A"}`;
      content += `\n      [toChain]: ${nativeTokenTransferToChain ?? "N/A"}`;
    } else {
      content += `\nHex: ${claimOutputHex}`;
    }
    return content;
  }

  return (
    <Dialog>
      <DialogTrigger>
        <Button>Decode Claim</Button>
      </DialogTrigger>
      <DialogContent className="overflow-scroll h-full max-h-[900px] max-w-[1000px]">
        <DialogHeader>
          <DialogTitle>Decode Claim</DialogTitle>
        </DialogHeader>
        {claim && (
          <>
            <LabelField label="Type" value={get(claim, "type") ?? "N/A"} />
            <LabelField
              valueClassName="max-h-[300px]"
              label="Assumptions"
              value={assumptions}
            />
            <LabelField label="Metadata" value={metadata} />
            <LabelField
              label="Action"
              valueType="normal"
              valueClassName="p-4 flex flex-col gap-4 max-h-max"
              value={
                <>
                  <LabelField label="From" value={get(claim, "action.from")} />
                  <LabelField label="To" value={get(claim, "action.to")} />
                  <LabelField
                    label={
                      <div className="flex flex-row items-center space-x-2">
                        <div>Input</div>
                        <Button
                          size="sm"
                          variant="secondary"
                          onClick={() =>
                            setShowDecodedInputContent(!showDecodedInputContent)
                          }
                        >
                          {showDecodedInputContent
                            ? "Hide decode result"
                            : "Show decode result"}
                        </Button>
                      </div>
                    }
                    value={inputContent(showDecodedInputContent)}
                  />
                </>
              }
            />
            <LabelField
              valueClassName="max-h-[400px]"
              label={
                <div className="flex flex-row items-center space-x-2">
                  <div>Result</div>
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() =>
                      setShowDecodedOutputContent(!showDecodedOutputContent)
                    }
                  >
                    {showDecodedOutputContent
                      ? "Hide decode result"
                      : "Show decode result"}
                  </Button>
                </div>
              }
              value={outputContent(showDecodedOutputContent)}
            />
          </>
        )}
      </DialogContent>
    </Dialog>
  );
};

export default DecodeClaimButton;
