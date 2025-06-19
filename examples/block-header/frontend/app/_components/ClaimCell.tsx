import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { Eye } from "lucide-react";

interface ClaimCellProps {
  claimId: string;
  onViewData: (
    title: string,
    claimId: string,
    contentType: "claim" | "verification_context"
  ) => void;
}

export function ClaimCell({ claimId, onViewData }: ClaimCellProps) {
  const handleViewClaim = () => {
    onViewData("Claim Details", claimId, "claim");
  };

  return (
    <TableCell className='px-4 py-2 text-center align-middle'>
      <div className='flex justify-center'>
        <Button
          variant='ghost'
          size='sm'
          className='h-7 px-2 text-xs'
          onClick={handleViewClaim}
        >
          <Eye className='h-3 w-3 mr-1' />
          View
        </Button>
      </div>
    </TableCell>
  );
}
