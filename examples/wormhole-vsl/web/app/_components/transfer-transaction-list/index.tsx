import { get } from "es-toolkit/compat";
import { useAccount } from "wagmi";
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "../../../components/ui/table";
import TransferTransaction from "./transfer-transaction";

export type TransferTransactionListProps = {
  claims: unknown[];
  onCompleteTransfer: () => void;
};

export default function TransferTransactionList({
  claims,
  onCompleteTransfer,
}: TransferTransactionListProps) {
  const { address } = useAccount();

  function content() {
    if (!address) {
      return (
        <div className="text-lg font-semibold h-[200px] content-center text-center text-gray-400 border rounded-md w-full">
          Connect wallet to view transactions history
        </div>
      );
    }

    if (claims.length === 0) {
      return (
        <div className="text-lg font-semibold h-[200px] content-center text-center text-gray-400 border rounded-md w-full">
          No transactions history
        </div>
      );
    }

    return (
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>From</TableHead>
            <TableHead>To</TableHead>
            <TableHead>Claim</TableHead>
            <TableHead>View Claim</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created At</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {claims.map((claim) => (
            <TransferTransaction
              key={get(claim, "ID")}
              initialClaim={claim}
              onCompleteTransfer={onCompleteTransfer}
            />
          ))}
        </TableBody>
      </Table>
    );
  }

  return (
    <div className="w-full px-8 flex flex-col items-center space-y-3">
      {content()}
    </div>
  );
}
