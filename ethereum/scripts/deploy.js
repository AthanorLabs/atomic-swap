async function main() {
  const SwapFactory = await ethers.getContractFactory("SwapFactory");
  const contract = await SwapFactory.deploy();

  console.log("SwapFactory deployed to:", contract.address);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });