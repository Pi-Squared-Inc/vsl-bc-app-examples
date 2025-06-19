# TEE Attester

**IMPORTANT NOTE:**    
The attester has to be run on a Google Cloud Confidential VM (CVM), which  supports SEV-SNP (Secure Encrypted Virtualization - Secure Nested Paging). You can follow the guide as described [here](https://github.com/samcamwilliams/HyperBEAM/blob/main/GCP-notes.md#create-an-amd-sev-snp-instance).

## Prerequisites

- [Go](https://go.dev/dl/): v1.24.2 or higher
- Run [prerequisite.sh](inference/src/prerequisite.sh) to install necessary libraries for image classification and LLM prompt generation computations
- Run [dl-inference-models.sh](inference/src/dl-inference-models.sh) to download the necessary 
inference models.


## Getting Started

Start the attester server:

```bash
sudo go run main.go start
```

Once the server is running, it will be listening fo computation requests for the following tasks:
- Image classification (`img_class`): For image classification tasks
- LLM prompt generation (`llama`): For generating prompts using large language models (LLMs)
- [Available to run upon request]* Block processing using KEVM+Reth (`block_processing_kreth`): For processing Ethereum blocks using KEVM + Reth

***Note:** To run the block processing using KEVM + Reth, you have to run the [`setup_kreth.sh`](block_processing/kreth/setup_kreth.sh) script 
to generate a new `block_processing_kreth` binary executable to replace the existing one in the repository. However, as building it requires 
dependencies from private repositories, if you are interested in running this, please contact us via [Discord](https://discord.gg/vYpVVTKx) 
to request access to the private repositories. Once you have been granted access, please follow the instructions below to set it up.

1. Set up SSH key to access your Github account:
    - Check for existing SSH keys:
      ```bash
      ls -al ~/.ssh
      ```

    - If you do not have an SSH key, generate one:
      ```bash
      ssh-keygen -t ed25519 -C "your_email@example.com"
      ```
      (If your system does not support `ed25519`, use `rsa` instead.)
      Press Enter to accept the default file location and set a passphrase if you want.

    - Add your SSH key to the SSH agent:
      ```bash
      eval "$(ssh-agent -s)"
      ssh-add ~/.ssh/id_ed25519 # or any file name that you used to generate the key
      ```

    - Copy the SSH public key to your clipboard:
      ```bash
      cat ~/.ssh/id_ed25519.pub # or any file name that you used to generate the key
      ```

    - Add the SSH key to your GitHub account:
      - Go to your GitHub account settings
      - Navigate to "SSH and GPG keys"
      - Click "New SSH key"
      - Paste the copied public key into the "Key" field
      - Give it a title and click "Add SSH key"

2. Run the `setup_kreth.sh` script to generate the `block_processing_kreth` binary:
   ```bash
   cd <repo-root>/example/common/attester/
   block_processing/kreth/setup_kreth.sh
   ```
   Note that you will be prompted at several points during the setup process to ensure the set up is done correctly. Follow the instructions carefully.

3. Once you have generated a new `block_processing_kreth` binary executable to replace the existing one in the repository, 
you have to log a hash of the newly generated `block_processing_kreth` binary in the TEE attestation reports, which is compared 
against a reference hash. Since we don't guarantee deterministic builds, you may need to update the `BlockProcessingKReth` in 
`<repo-root>/generation/pkg/generation/claim_generation.go` reference hashes to match the hashes of your build. You can obtain 
the hashes of your build by running `go run main.go file-hashes --file <file-name>`.
