"use client";
import { useState } from "react";
import { Footer } from "./_components/Footer";
import { NavigationBar } from "./_components/NavigationBar";
import { BlockHeaderSettlement } from "./demos/BlockHeaderSettlement";

const tabs = [
  { id: "ethereum", label: "Ethereum", chain: "Ethereum" },
  { id: "bitcoin", label: "Bitcoin", chain: "Bitcoin" },
];

export default function Home() {
  const [activeTab, setActiveTab] = useState("ethereum");

  const renderTabContent = () => {
    switch (activeTab) {
      case "ethereum":
        return (
          <BlockHeaderSettlement
            apiBaseUrl={process.env.NEXT_PUBLIC_API_URL}
            defaultChain='Ethereum'
            className='w-full'
          />
        );
      case "bitcoin":
        return (
          <BlockHeaderSettlement
            apiBaseUrl={process.env.NEXT_PUBLIC_API_URL}
            defaultChain='Bitcoin'
            className='w-full'
          />
        );
      default:
        return null;
    }
  };

  return (
    <div className='container mx-auto px-4 py-8 flex flex-col min-h-screen'>
      <NavigationBar />
      <main className='flex-grow flex flex-col'>
        <div className='text-center mb-16'>
          <h1 className='text-5xl font-medium bg-gradient-to-r from-pi2-purple-500 to-pi2-accent-white text-transparent bg-clip-text'>
            Block Header Settlement{" "}
          </h1>
        </div>
        <div className='w-full mb-8'>
          <div className='border-b border-border'>
            <nav className='flex space-x-8'>
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                    activeTab === tab.id
                      ? "border-pi2-purple-500 text-pi2-purple-500"
                      : "border-transparent text-muted-foreground hover:text-foreground hover:border-gray-300"
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </nav>
          </div>
        </div>
        <div className='flex-grow'>{renderTabContent()}</div>
      </main>
      <Footer />
    </div>
  );
}
