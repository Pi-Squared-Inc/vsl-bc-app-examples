"use client";

import { Button } from "@/components/ui/button";
import Image from "next/image";
import Link from "next/link";
import heroImage from "./_assets/hero.png";
import { Footer } from "./_components/Footer";
import { NavigationBar } from "./_components/NavigationBar";

export default function Home() {
  return (
    <div className="container mx-auto px-4 py-8">
      <NavigationBar />

      <main className="py-20">
        <div className="max-w-4xl mx-auto text-center">
          <Image
            src={heroImage}
            alt="Hero Image"
            width={960}
            height={0}
            className="w-full h-auto rounded-lg"
            priority
          />

          <h1 className="mt-8 text-5xl font-medium bg-gradient-to-r from-pi2-purple-500 to-pi2-accent-white text-transparent bg-clip-text">
            Blockchain Mirroring
          </h1>

          <Link href="/dashboard" className="mt-8 inline-block">
            <Button
              size="lg"
              className="bg-pi2-accent-white text-pi2-accent-black hover:bg-pi2-accent-white/90 rounded-full font-semibold"
            >
              Explore
            </Button>
          </Link>

          <div className="mt-20 text-left">
            <p className="text-xl text-pi2-accent-white mb-6">
              The Pi Squared Verifiable Settlement Layer (VSL) is a
              decentralized network for all users and applications in all
              ecosystems to submit, verify, settle, query, and use claims about
              their transactions, computation, vetted information, oracle
              values, and more. Everything that can be proved is a claim, and
              every claim with a valid proof can be settled on VSL. VSL accepts
              claims, verifies their proofs, and settles them for good. Verify
              once, use everywhere!
            </p>
            <p className="text-xl text-pi2-accent-white mb-6">
              We are excited to announce the first VSL client, which is a
              mirroring of the Ethereum network. Each block on Ethereum becomes
              a VSL claim, together with a state footprint witness as its proof,
              independently verifiable.
            </p>
            <p className="text-xl text-pi2-accent-white">
              In the future, we plan extensions in two ways. Vertically,
              we&apos;ll enable mathematical proofs that are directly based on a
              formal semantics of Ethereum, as an alternative to the state
              footprint witness proofs, because mathematical proofs are known to
              enjoy a much smaller trust base. Horizontally, we&apos;ll enable
              mirroring of more L1s and L2s as well as applications, including
              but not limited to CEXes and AI agents. Stay tuned!
            </p>
          </div>
        </div>
      </main>

      <Footer />
    </div>
  );
}
