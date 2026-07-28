# Release signing

Provider release archives and checksum files are signed with the project release key.

- Algorithm: RSA 4096
- Fingerprint: `00C9 E5DD DF64 8781 7E9A BF3E D40B 28DB 454A 028F`
- Public key: [`provider-signing-key.asc`](provider-signing-key.asc)

The private key is not stored in the repository. GitHub Actions imports it from the `GPG_PRIVATE_KEY` repository secret and reads its passphrase from `PASSPHRASE`.
