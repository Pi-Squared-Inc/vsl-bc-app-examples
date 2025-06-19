import { TableCell, TableRow } from "@/components/ui/table";
import { ValidationRecord } from "@/types";
import Link from "next/link";
import { ClientClaimCell } from "./ClientClaimCell";
import { ClientValidationCell } from "./ClientValidationCell";

interface BitcoinMirroringRecordRowProps {
  claim: ValidationRecord;
  onViewError: (
    title: string,
    content: string,
    executionClient: string
  ) => void;
}

export default function BitcoinMirroringRecordRow({
  claim,
  onViewError,
}: BitcoinMirroringRecordRowProps) {
  return (
    <TableRow
      id={`claim-${claim.block_number}`}
      suppressHydrationWarning
      className="transition-colors duration-150 hover:bg-white/15 h-[53px] [&_td]:border-r"
    >
      <TableCell className="font-medium px-4 whitespace-nowrap align-middle">
        <Link
          href={`https://www.blockchain.com/explorer/blocks/btc/${claim.block_number}`}
          target="_blank"
          rel="noopener noreferrer"
          className="hover:underline text-pi2-purple-300"
        >
          {claim.block_number}
        </Link>
      </TableCell>
      <ClientClaimCell
        details={claim.claim_details}
        executionClient={"BitcoinBlock"}
      />
      <ClientValidationCell
        details={claim.claim_details}
        executionClient={"BitcoinBlock"}
        onViewError={onViewError}
      />
    </TableRow>
  );
}
