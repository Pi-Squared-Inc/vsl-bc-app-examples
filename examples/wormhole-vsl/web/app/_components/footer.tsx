import { cn } from "../../lib/utils";

type FooterProps = {
  className?: string;
};

export default function Footer({ className }: FooterProps) {
  const year = new Date().getFullYear();

  return (
    <div className={cn("flex flex-row justify-center w-full p-6", className)}>
      <div>&copy; {year} Pi Squared, Inc. All rights reserved.</div>
    </div>
  );
}
