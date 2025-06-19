import { TableCell, TableRow } from "@/components/ui/table";
import { ValidationRecord } from "@/types";
import Link from "next/link";
import { ClientClaimCell } from "./ClientClaimCell";
import { ClientValidationCell } from "./ClientValidationCell";

interface EthereumMirroringRecordRowProps {
  claim: ValidationRecord;
  onViewError: (title: string, content: string, executionClient: string) => void;
}

export default function EthereumMirroringRecordRow({ claim, onViewError }: EthereumMirroringRecordRowProps) {
  return (
    <TableRow id={`claim-${claim.block_number}`} suppressHydrationWarning className="transition-colors duration-150 hover:bg-white/15 h-[53px] [&_td]:border-r [&_.claim]:border-r-0">
      <TableCell className="font-medium px-4 whitespace-nowrap align-middle first:border-r-2">
        <Link
          href={`https://etherscan.io/block/${claim.block_number}`}
          target="_blank"
          rel="noopener noreferrer"
          className="hover:underline text-pi2-purple-300"
        >
          {claim.block_number}
        </Link>
      </TableCell>
      <ClientClaimCell className="claim" details={claim.claim_details} executionClient={"MirroringReth"} />
      <ClientValidationCell details={claim.claim_details} executionClient={"MirroringReth"} onViewError={onViewError} />
      <ClientClaimCell className="claim" details={claim.claim_details} executionClient={"MirroringGeth"} />
      <ClientValidationCell details={claim.claim_details} executionClient={"MirroringGeth"} onViewError={onViewError} />
    </TableRow>
  );
}
