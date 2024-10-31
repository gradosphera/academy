
const wallet = getWallet("EQAbRLmFnI7y5BUmSVvxsz3X4Ejy50uMkgvXPmevthl5K3n9");

const jetton = new TonWeb.token.jetton.JettonMinter(tonweb.provider, { address: process.env.JETTON_WALLET_ADDRESS });
const jettonWalletAddress = await jetton.getJettonWalletAddress(wallet.address);
console.log(jettonWalletAddress.toString(true, true, false));
const jettonWallet = new TonWeb.token.jetton.JettonWallet(tonweb.provider, { address: jettonWalletAddress });

const jettonBalance = (await jettonWallet.getData()).balance;
console.log(jettonBalance.toString());



sendJettons('UQCjNf6y_RhVATipbgKpCBAa8h5z6mwIXv3oDY7UZRyv01Aj');