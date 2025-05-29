#!/bin/bash

# Find all .go files and replace BLS12-377 with BN254 in imports
find . -name "*.go" -type f -exec sed -i 's|github.com/consensys/gnark-crypto/ecc/bls12-377/fr|github.com/consensys/gnark-crypto/ecc/bn254/fr|g' {} +

# Update comments referencing BLS12-377 to BN254
find . -name "*.go" -type f -exec sed -i 's/BLS12-377/BN254/g' {} + 