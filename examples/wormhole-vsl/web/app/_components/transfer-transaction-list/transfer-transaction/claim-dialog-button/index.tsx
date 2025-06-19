"use client";

import { useClipboard } from "use-clipboard-copy";
import { Button } from "../../../../../components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../../../../../components/ui/dialog";
import { Label } from "../../../../../components/ui/label";
import { useToast } from "../../../../../components/ui/use-toast";
import DecodeClaimButton from "./decode-claim-button/index";

interface ClaimDialogButtonProps {
  claimId: string;
  claim: string;
  claimHex: string;
}

const ClaimDialogButton = ({
  claimId,
  claim,
  claimHex,
}: ClaimDialogButtonProps) => {
  const { toast } = useToast();
  const { copy } = useClipboard();

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="secondary">View</Button>
      </DialogTrigger>
      <DialogContent className="max-w-[800px]">
        <DialogHeader>
          <DialogTitle>View Claim</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col items-center space-y-4">
          <div className="flex flex-col items-start space-y-3 w-full">
            <Label className="font-semibold">Claim ID</Label>
            <pre className="w-full bg-card border p-4 rounded-md text-baes whitespace-break-spaces break-all max-h-[200px] overflow-y-auto">
              {claimId}
            </pre>
            <Label className="font-semibold">Claim</Label>
            <pre className="w-full bg-card border p-4 rounded-md text-base whitespace-break-spaces break-all max-h-[400px] overflow-y-auto">
              {claimHex}
            </pre>
            <DialogFooter className="w-full">
              <Button
                type="button"
                onClick={() => {
                  copy(claimId);
                  toast({
                    title: "Copied",
                  });
                }}
              >
                Copy Claim ID
              </Button>
              <Button
                type="button"
                onClick={() => {
                  copy(claim);
                  toast({
                    title: "Copied",
                  });
                }}
              >
                Copy Claim
              </Button>
              <DecodeClaimButton claimJSON={claim} />
            </DialogFooter>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default ClaimDialogButton;
