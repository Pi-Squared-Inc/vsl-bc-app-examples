import { expect, test } from "../fixtures";

test("Test whole flow", async ({ page, context, extensionId }) => {
  // Unlock the MetaMask wallet
  console.log("Unlocking MetaMask wallet");
  await page.goto(`chrome-extension://${extensionId}/home.html#unlock`);
  await page.waitForTimeout(5000);
  await page.bringToFront();
  await page.waitForLoadState("domcontentloaded");

  // Fill password
  console.log("Filling password");
  const passwordInput = page.locator("xpath=//*[@id='password']");
  await passwordInput.waitFor({
    state: "visible",
  });
  await passwordInput.fill(process.env.METAMASK_PASSWORD!);

  // Click unlock button
  console.log("Clicking unlock button");
  const unlockButton = page.locator(
    "xpath=//*[@id='app-content']/div/div[2]/div/div/button"
  );
  await unlockButton.waitFor({
    state: "visible",
  });
  await unlockButton.click();

  // Click the "Got it" button if it exists
  try {
    console.log("Clicking got it button if it exists");
    const gotItButton = page.getByRole("button", { name: "Got it" });
    await gotItButton
      .waitFor({
        state: "visible",
        timeout: 5000,
      })
      .catch(() => {
        console.log("Not found the got it button");
      });
    await gotItButton.click();
  } catch (error) {}

  // Switch to Sepolia network
  console.log("Switching to Sepolia network");
  const networkSwitchButton = page.locator(
    "xpath=//*[@id='app-content']/div/div[2]/div/div[1]/button"
  );
  await networkSwitchButton.waitFor({
    state: "visible",
  });
  await networkSwitchButton.click();

  const sepoliaNetworkButton = page.locator(
    "xpath=/html/body/div[3]/div[3]/div/section/div[1]/div[3]/div[5]/div[1]"
  );
  await sepoliaNetworkButton.waitFor({
    state: "visible",
  });
  await sepoliaNetworkButton.click();

  // Setup USL page and connect wallet
  console.log("Setup USL page and connect wallet");
  const uslPage = await context.newPage();
  await uslPage.goto(process.env.WORMHOLE_DEMO_URL!);
  await uslPage.waitForLoadState("domcontentloaded");
  await uslPage.bringToFront();
  await uslPage.waitForTimeout(10000);

  const connectWalletButton = uslPage.locator(
    "xpath=/html/body/div[2]/div[1]/button"
  );
  await connectWalletButton.waitFor({
    state: "visible",
  });
  await connectWalletButton.click();

  const connectMetamaskButton = uslPage.getByRole("button", {
    name: "MetaMask MetaMask installed",
  });
  await connectMetamaskButton.waitFor({
    state: "visible",
  });

  const [connectMetamaskPopup] = await Promise.all([
    context.waitForEvent("page"),
    connectMetamaskButton.click(),
  ]);

  await connectMetamaskPopup.waitForLoadState("domcontentloaded");
  await connectMetamaskPopup.bringToFront();

  const connectButton = connectMetamaskPopup.getByRole("button", {
    name: "Connect",
  });
  await connectButton.waitFor({
    state: "visible",
  });
  await connectButton.click();

  await uslPage.waitForTimeout(5000);
  await uslPage.bringToFront();

  // Check Sepolia token balance
  console.log("Checking Sepolia token balance");
  const sepoliaTokenBalance = uslPage.locator(
    "xpath=/html/body/div[2]/div[3]/div[2]/div[2]/div/div[1]/div[2]/div/div[2]/span[2]"
  );
  await sepoliaTokenBalance.waitFor({
    state: "visible",
  });
  const sepoliaTokenBalanceText = await sepoliaTokenBalance.textContent();
  expect(sepoliaTokenBalanceText).toContain("PT");
  const initialSepoliaTokenBalance = parseFloat(
    sepoliaTokenBalanceText!.replace(" PT", "")
  );

  const arbitrumTokenBalance = uslPage.locator(
    "xpath=/html/body/div[2]/div[3]/div[2]/div[2]/div/div[3]/div[2]/div/div[2]/span[2]"
  );
  await arbitrumTokenBalance.waitFor({
    state: "visible",
  });
  const arbitrumTokenBalanceText = await arbitrumTokenBalance.textContent();
  expect(arbitrumTokenBalanceText).toContain("PT");
  const initialArbitrumTokenBalance = parseFloat(
    arbitrumTokenBalanceText!.replace(" PT", "")
  );
  console.log("Initial Sepolia token balance:", initialSepoliaTokenBalance);
  console.log("Initial Arbitrum token balance:", initialArbitrumTokenBalance);

  // Faucet Sepolia tokens
  console.log("Fauceting Sepolia tokens");
  const faucetButton = uslPage.getByRole("button", { name: "Faucet" });
  await faucetButton.waitFor({
    state: "visible",
  });
  const [faucetPopup] = await Promise.all([
    context.waitForEvent("page"),
    faucetButton.click(),
  ]);

  await faucetPopup.waitForLoadState("domcontentloaded");
  await faucetPopup.bringToFront();

  const confirmFaucetButton = faucetPopup.getByRole("button", {
    name: "Confirm",
  });
  await confirmFaucetButton.waitFor({
    state: "visible",
  });
  await confirmFaucetButton.click();

  await uslPage.bringToFront();

  // Check Sepolia token balance after faucet
  console.log("Checking Sepolia token balance after faucet");
  await expect
    .poll(
      async () => {
        const newSepoliaBalance = uslPage.locator(
          "xpath=/html/body/div[2]/div[3]/div[2]/div[2]/div/div[1]/div[2]/div/div[2]/span[2]"
        );
        const newSepoliaBalanceText = await newSepoliaBalance.textContent();
        return (
          parseFloat(newSepoliaBalanceText!.replace(" PT", "")) ===
          initialSepoliaTokenBalance + 100
        );
      },
      { timeout: 60000 }
    )
    .toBe(true);

  // Approve transfer
  console.log("Approving transfer");
  const amountInput = uslPage.locator(
    "xpath=/html/body/div[2]/div[3]/div[2]/div[2]/form/div[1]/div/input"
  );
  await amountInput.waitFor({
    state: "visible",
  });
  await amountInput.fill("100");

  const approveButton = uslPage.getByRole("button", { name: "Approve" });
  await approveButton.waitFor({
    state: "visible",
  });

  const [approvePopup] = await Promise.all([
    context.waitForEvent("page"),
    approveButton.click(),
  ]);

  await approvePopup.waitForLoadState("domcontentloaded");
  await approvePopup.bringToFront();

  const approveConfirmButton = approvePopup.getByRole("button", {
    name: "Confirm",
  });
  await approveConfirmButton.waitFor({
    state: "visible",
  });
  await approveConfirmButton.click();

  await uslPage.waitForTimeout(5000);
  await uslPage.bringToFront();

  // Transfer tokens
  console.log("Transferring tokens");
  const transferButton = uslPage.getByRole("button", { name: "Transfer" });
  await transferButton.waitFor({
    state: "visible",
  });

  await expect
    .poll(
      async () => {
        return await transferButton.isDisabled();
      },
      { timeout: 60000 }
    )
    .toBe(false);
  const [transferPopup] = await Promise.all([
    context.waitForEvent("page"),
    transferButton.click(),
  ]);

  await transferPopup.waitForLoadState("domcontentloaded");
  await transferPopup.bringToFront();

  const confirmTransferButton = transferPopup.getByRole("button", {
    name: "Confirm",
  });
  await confirmTransferButton.waitFor({
    state: "visible",
  });
  await confirmTransferButton.click();

  await uslPage.waitForTimeout(5000);
  await uslPage.bringToFront();

  await expect
    .poll(
      async () => {
        return await transferButton.isDisabled();
      },
      { timeout: 60000 }
    )
    .toBe(false);

  // Check Arbitrum token balance after transfer
  console.log("Checking Arbitrum token balance after transfer");
  const transferTab = uslPage.getByRole("tab", { name: "Transfer" });
  await transferTab.waitFor({
    state: "visible",
  });
  await transferTab.click();
  await expect
    .poll(
      async () => {
        const newArbitrumBalance = uslPage.locator(
          "xpath=/html/body/div[2]/div[3]/div[2]/div[2]/div/div[3]/div[2]/div/div[2]/span[2]"
        );
        const newArbitrumBalanceText = await newArbitrumBalance.textContent();
        return (
          parseFloat(newArbitrumBalanceText!.replace(" PT", "")) ===
          initialArbitrumTokenBalance + 100
        );
      },
      { timeout: 60000 }
    )
    .toBe(true);
});
