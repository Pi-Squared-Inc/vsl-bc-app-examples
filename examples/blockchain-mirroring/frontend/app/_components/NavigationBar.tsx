import Image from "next/image";
import Link from "next/link";
import logoImage from "../_assets/logo.png";

export function NavigationBar() {
  return (
    <nav className="flex justify-between items-center mb-10">
      <Link
        href="https://pi2.network"
        className="transition-opacity duration-200 hover:opacity-75"
      >
        <Image
          src={logoImage}
          alt="Pi Squared Logo"
          width={80}
          height={80}
          className="rounded"
        />
      </Link>

      <div className="flex gap-6">
        <Link
          href="/"
          className="text-white transition-opacity duration-200 hover:opacity-75"
        >
          Home
        </Link>
        <Link
          href="https://pi2.network/developer"
          target="_blank"
          rel="noopener noreferrer"
          className="text-white transition-opacity duration-200 hover:opacity-75"
        >
          Developer
        </Link>
        <Link
          href="https://blog.pi2.network"
          target="_blank"
          rel="noopener noreferrer"
          className="text-white transition-opacity duration-200 hover:opacity-75"
        >
          Blog
        </Link>
        <Link
          href="https://docs.pi2.network"
          target="_blank"
          rel="noopener noreferrer"
          className="text-white transition-opacity duration-200 hover:opacity-75"
        >
          Docs
        </Link>
      </div>
    </nav>
  );
}
