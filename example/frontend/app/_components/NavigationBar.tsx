"use client";

import Image from "next/image";
import Link from "next/link";
import logoImage from "../_assets/logo.png";
import WalletConnect from "./Account";

export function NavigationBar() {
  return (
    <nav className="flex justify-between items-center mb-10">
      <Link href="/" className="transition-opacity duration-200 hover:opacity-75">
        <Image src={logoImage} alt="Pi Squared Logo" width={80} height={80} className="rounded" />
      </Link>
      <WalletConnect />
    </nav>
  );
}
