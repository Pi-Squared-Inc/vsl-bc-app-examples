import { TableCell } from "@/components/ui/table";
import { ExternalLink } from "lucide-react";
import { Button } from "@/components/ui/button";

interface BlockNumberCellProps {
  blockNumber: string | number;
  explorerUrl?: string | null;
  onExplorerClick?: () => void;
}

export function BlockNumberCell({
  blockNumber,
  explorerUrl,
  onExplorerClick,
}: BlockNumberCellProps) {
  const showClickableButton = onExplorerClick && explorerUrl !== null;

  return (
    <TableCell className='px-4 py-2 text-center align-middle'>
      <div className='flex items-center justify-center gap-2'>
        <span className='font-mono text-sm'>{blockNumber}</span>
        {showClickableButton && (
          <Button
            variant='ghost'
            size='sm'
            onClick={onExplorerClick}
            className='h-8 w-8 p-0'
            title='View on blockchain explorer'
          >
            <ExternalLink className='h-4 w-4' />
          </Button>
        )}
      </div>
    </TableCell>
  );
}
