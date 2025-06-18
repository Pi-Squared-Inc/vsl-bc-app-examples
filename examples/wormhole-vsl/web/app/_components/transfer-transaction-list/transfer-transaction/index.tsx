"use client";

import { mdiCheck } from "@mdi/js";
import Icon from "@mdi/react";
import { ReloadIcon } from "@radix-ui/react-icons";
import axios from "axios";
import { formatDistance } from "date-fns";
import { get, isEqual } from "es-toolkit/compat";
import Link from "next/link";
import { FunctionComponent, useMemo, useState } from "react";
import { useInterval } from "usehooks-ts";
import { TableCell, TableRow } from "../../../../components/ui/table";
import { backendAPIEndpoint, explorerUrl } from "../../../../config/constant";
import ClaimDialogButton from "./claim-dialog-button";

interface TransferTransactionProps {
  initialClaim?: unknown;
  onCompleteTransfer?: () => void;
}

const TransferTransaction: FunctionComponent<TransferTransactionProps> = ({
  initialClaim,
  onCompleteTransfer,
}) => {
  const [claim, setClaim] = useState<unknown>(initialClaim);
  const claimJSON = useMemo<string | undefined>(() => {
    return get(claim, "claim_json");
  }, [claim]);
  const claimId = useMemo<string | undefined>(() => {
    return get(claim, "claim_id");
  }, [claim]);
  const claimHex = useMemo<string | undefined>(() => {
    return get(claim, "claim");
  }, [claim]);
  const destinationTransactionHash = useMemo(() => {
    return get(claim, "destination_transaction_hash");
  }, [claim]);

  useInterval(
    () => {
      fetchClaim();
    },
    destinationTransactionHash ? null : 1000
  );

  function fetchClaim() {
    const claimId = get(claim, "claim_id");

    if (claimId) {
      axios.get(backendAPIEndpoint + "/claim/" + claimId).then((response) => {
        if (isEqual(response.data, claim)) {
          return;
        }
        setClaim(response.data);
      });
      return;
    }

    const sourceTransactionHash = get(claim, "source_transaction_hash");
    if (sourceTransactionHash) {
      axios
        .get(
          backendAPIEndpoint + "/claim-by-source-tx/" + sourceTransactionHash
        )
        .then((response) => {
          if (isEqual(response.data, claim)) {
            return;
          }
          setClaim(response.data);
        });
      return;
    }
  }

  function status() {
    if (destinationTransactionHash) {
      return (
        <div className="flex flex-row items-center text-green-500">
          <Icon className="mr-2" path={mdiCheck} size={0.8} />
          <span>Success</span>
        </div>
      );
    }
    return (
      <div className="flex flex-row items-center">
        <ReloadIcon className="mr-2 h-3 w-3 animate-spin" />
        Claim is settling
      </div>
    );
  }

  if (!claim) {
    return (
      <TableRow>
        <TableCell colSpan={4}>
          <div className="flex flex-row justify-center h-12 items-center">
            <ReloadIcon className="h-4 w-4 animate-spin" />
          </div>
        </TableCell>
      </TableRow>
    );
  }

  return (
    <TableRow>
      <TableCell>
        <Link
          className="underline"
          target="_blank"
          href={`https://sepolia.etherscan.io/tx/${get(
            claim,
            "source_transaction_hash"
          )}`}
        >
          Sepolia
        </Link>
      </TableCell>
      <TableCell>
        {destinationTransactionHash && (
          <Link
            className="underline"
            target="_blank"
            href={`https://sepolia.arbiscan.io/tx/${get(
              claim,
              "destination_transaction_hash"
            )}`}
          >
            Arbitrum Sepolia
          </Link>
        )}
      </TableCell>
      <TableCell>
        <Link
          className="text-primary font-medium underline"
          target="_blank"
          href={`${explorerUrl}/claim/${claimId}`}
        >
          {claimId}
        </Link>
      </TableCell>
      <TableCell>
        {claimJSON && (
          <ClaimDialogButton
            claimId={claimId!}
            claim={claimJSON!}
            claimHex={claimHex!}
          />
        )}
      </TableCell>
      <TableCell>{status()}</TableCell>
      <TableCell>
        {formatDistance(
          new Date(get(claim, "CreatedAt")! as string),
          new Date(),
          {
            addSuffix: true,
          }
        )}
      </TableCell>
    </TableRow>
  );
};

export default TransferTransaction;
