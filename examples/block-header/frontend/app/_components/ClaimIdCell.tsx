import { TableCell } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Copy, Check, ExternalLink } from "lucide-react";
import { useState } from "react";

interface ClaimIdCellProps {
  claimId: string;
  onCopyClaimId?: (claimId: string) => Promise<boolean>;
  explorerUrl?: string | null;
}

export function ClaimIdCell({
  claimId,
  onCopyClaimId,
  explorerUrl,
}: ClaimIdCellProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    if (!onCopyClaimId || copied) return;

    try {
      const success = await onCopyClaimId(claimId);
      if (success) {
        setCopied(true);
        setTimeout(() => setCopied(false), 1000);
      }
    } catch (error) {
      console.error("Copy failed:", error);
    }
  };

  const displayClaimId =
    claimId.length > 12
      ? `${claimId.slice(0, 6)}...${claimId.slice(-6)}`
      : claimId;

  return (
    <TableCell className='px-4 py-2 text-center align-middle'>
      <div className='w-full flex justify-center'>
        <div className='flex items-center gap-2'>
          <span className='font-mono text-sm' title={claimId}>
            {displayClaimId}
          </span>
          <div className='flex items-center gap-1'>
            {onCopyClaimId && (
              <Button
                variant='ghost'
                size='sm'
                onClick={handleCopy}
                className='h-6 w-6 p-0 hover:bg-gray-100'
                title={copied ? "Copied!" : "Copy claim ID"}
              >
                {copied ? (
                  <Check className='h-3 w-3 text-green-600' />
                ) : (
                  <Copy className='h-3 w-3' />
                )}
              </Button>
            )}
            {explorerUrl && (
              <Button
                variant='ghost'
                size='sm'
                asChild
                className='h-6 w-6 p-0 hover:bg-gray-100'
              >
                <a
                  href={explorerUrl}
                  target='_blank'
                  rel='noopener noreferrer'
                  title='View claim on blockchain explorer'
                >
                  <ExternalLink className='h-3 w-3' />
                </a>
              </Button>
            )}
          </div>
        </div>
      </div>
    </TableCell>
  );
}
