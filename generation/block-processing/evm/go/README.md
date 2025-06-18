
# Block processing claim generation Go package for EVM

> [!IMPORTANT]
> The claim generation package depends on the `debug_executionWitness` API, so you may need to run your own Geth node if you couldn't find a node that supports this API. Please use our [forked Geth](https://github.com/Pi-Squared-Inc/geth-measurement), it already implements this API. You can compile it from the source code on your own. Note that you also need to [enable the debug API](https://geth.ethereum.org/docs/interacting-with-geth/rpc#:~:text=geth%20%2D%2Dhttp%20%2D%2Dhttp.api%20eth%2Cnet%2Cweb3).

This package includes and exports all the necessary functions to generate the block processing claim for the Geth execution client.

## License

Private
