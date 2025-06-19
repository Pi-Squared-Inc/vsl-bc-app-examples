import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Footer } from "../_components/Footer";
import { NavigationBar } from "../_components/NavigationBar";
import Ethereum from "./_components/Ethereum";
import Bitcoin from "./_components/Bitcoin";

export default function DashboardPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <NavigationBar />
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-5xl font-medium bg-gradient-to-r from-pi2-purple-500 to-pi2-accent-white text-transparent bg-clip-text">Blockchain Mirroring</h1>
      </div>
      <Tabs defaultValue="eth">
        <TabsList>
          <TabsTrigger value="eth">Ethereum</TabsTrigger>
          <TabsTrigger value="btc">Bitcoin</TabsTrigger>
        </TabsList>
        <TabsContent value="eth">
          <Ethereum />
        </TabsContent>
        <TabsContent value="btc">
          <Bitcoin />
        </TabsContent>
      </Tabs>
      <Footer />
    </div>
  );
}
