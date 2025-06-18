import { useState } from "react";
import { TableRow } from "@/components/ui/table";
import { BlockNumberCell } from "../BlockNumberCell";
import { ClaimIdCell } from "../ClaimIdCell";
import { ClaimCell } from "../ClaimCell";
import { CreatedAtCell } from "../CreatedAtCell";
import { VerificationTimeCell } from "../VerificationTimeCell";
import { Dialog } from "../Dialog";
import { BlockHeaderRecord } from "@/types/blockHeader";
import {
  fetchClaimDetails,
  fetchVerificationContext,
} from "@/services/blockHeaderClaimsService";

const EXPLORER_URLS = {
  ethereum: "https://etherscan.io/block/",
  bitcoin: "https://mempool.space/block/",
} as const;

interface BlockHeaderRecordRowProps {
  blockHeaderRecord: BlockHeaderRecord;
}

export function BlockHeaderRecordRow({
  blockHeaderRecord,
}: BlockHeaderRecordRowProps) {
  const [dialogState, setDialogState] = useState({
    isOpen: false,
    title: "",
    content: "",
    isJson: false,
    isLoading: false,
  });

  const { block_number, claim_details } = blockHeaderRecord;
  const { claim_id, chain, created_at, verification_time } = claim_details;

  const copyToClipboard = async (text: string): Promise<boolean> => {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch (error) {
      console.error("Failed to copy to clipboard:", error);
      return false;
    }
  };

  const getBitcoinBlockHash = async (
    blockNumber: string | number
  ): Promise<string | null> => {
    try {
      const response = await fetch(
        `https://mempool.space/api/v1/blocks/${blockNumber}`
      );
      if (!response.ok) return null;
      const blocks = await response.json();
      return Array.isArray(blocks) && blocks.length > 0 && blocks[0].id
        ? blocks[0].id
        : null;
    } catch (error) {
      console.error("Failed to fetch Bitcoin block hash:", error);
      return null;
    }
  };

  const handleBlockExplorerClick = async (
    blockNumber: string | number,
    chain: string
  ) => {
    const normalizedChain = chain
      .toLowerCase()
      .trim() as keyof typeof EXPLORER_URLS;
    const baseUrl = EXPLORER_URLS[normalizedChain];
    if (!baseUrl) return;

    if (normalizedChain === "bitcoin") {
      try {
        const blockHash = await getBitcoinBlockHash(blockNumber);
        if (!blockHash) {
          console.error("Failed to get Bitcoin block hash");
          return;
        }
        const url = `${baseUrl}${blockHash}`;
        window.open(url, "_blank");
      } catch (error) {
        console.error("Error opening Bitcoin block explorer:", error);
      }
    } else {
      const url = `${baseUrl}${blockNumber}`;
      window.open(url, "_blank");
    }
  };

  const getBlockExplorerUrl = (
    blockNumber: string | number,
    chain: string
  ): string | null => {
    const normalizedChain = chain
      .toLowerCase()
      .trim() as keyof typeof EXPLORER_URLS;
    const baseUrl = EXPLORER_URLS[normalizedChain];
    if (!baseUrl) return null;

    return `${baseUrl}${blockNumber}`;
  };

  const getClaimExplorerUrl = (claimId: string): string | null => {
    return `${process.env.NEXT_PUBLIC_CLAIM_EXPLORER_URL}/claim/${claimId}`;
  };

  const handleViewData = async (
    title: string,
    claimId: string,
    contentType: "claim" | "verification_context"
  ) => {
    setDialogState({
      isOpen: true,
      title,
      content: `Loading ${contentType} details...`,
      isJson: false,
      isLoading: true,
    });

    try {
      let contentData;
      if (contentType === "claim") {
        contentData = await fetchClaimDetails(claimId);
      } else {
        contentData = await fetchVerificationContext(claimId);
      }

      if (contentData) {
        let finalContent = "";
        let isFinalContentJson = false;

        if (typeof contentData === "string") {
          try {
            const parsedData = JSON.parse(contentData);
            finalContent = JSON.stringify(parsedData, null, 2);
            isFinalContentJson = true;
          } catch (parseError) {
            finalContent = contentData;
            isFinalContentJson = false;
          }
        } else if (typeof contentData === "object" && contentData !== null) {
          finalContent = JSON.stringify(contentData, null, 2);
          isFinalContentJson = true;
        } else {
          finalContent = String(contentData);
          isFinalContentJson = false;
        }

        setDialogState({
          isOpen: true,
          title,
          content: finalContent,
          isJson: isFinalContentJson,
          isLoading: false,
        });
      } else {
        setDialogState({
          isOpen: true,
          title,
          content: `No ${contentType} data is available for this record`,
          isJson: false,
          isLoading: false,
        });
      }
    } catch (error) {
      setDialogState({
        isOpen: true,
        title,
        content:
          `Error loading ${contentType} data: ` +
          (error instanceof Error ? error.message : String(error)),
        isJson: false,
        isLoading: false,
      });
    }
  };

  const handleViewError = (title: string, content: string) => {
    setDialogState({
      isOpen: true,
      title,
      content,
      isJson: false,
      isLoading: false,
    });
  };

  const closeDialog = () => {
    setDialogState((prev) => ({ ...prev, isOpen: false }));
  };

  const blockExplorerUrl = getBlockExplorerUrl(block_number, chain);
  const claimExplorerUrl = getClaimExplorerUrl(claim_id);

  return (
    <>
      <Dialog
        isOpen={dialogState.isOpen}
        onClose={closeDialog}
        title={dialogState.title}
        content={dialogState.content}
        isJson={dialogState.isJson}
        isLoading={dialogState.isLoading}
      />

      <TableRow
        id={`record-${claim_id}`}
        suppressHydrationWarning
        className='transition-colors duration-150 hover:bg-white/15 h-[53px]'
      >
        <BlockNumberCell
          blockNumber={block_number}
          explorerUrl={blockExplorerUrl}
          onExplorerClick={() => handleBlockExplorerClick(block_number, chain)}
        />
        <ClaimIdCell
          claimId={claim_id}
          onCopyClaimId={copyToClipboard}
          explorerUrl={claimExplorerUrl}
        />
        <ClaimCell claimId={claim_id} onViewData={handleViewData} />
        <CreatedAtCell createdAt={created_at} />
        <VerificationTimeCell verificationTime={verification_time} />
      </TableRow>
    </>
  );
}
