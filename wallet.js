//we need to crearte wallet to hold jettons for rewards
import TonWeb from "tonweb";
import TonWebMnemonic from "tonweb-mnemonic";
import fs from "fs";
import dotenv from 'dotenv';
import * as bip39 from 'bip39';
import * as ed25519 from 'ed25519-hd-key';
import { exit } from "process";

dotenv.config();

const NODE_API_URL = 'https://toncenter.com/api/v2/jsonRPC';
const tonweb = new TonWeb(new TonWeb.HttpProvider(NODE_API_URL, {apiKey: process.env.TONCENTER_API_KEY}));

    // Generate a 12-word mnemonic
const mnemonic = bip39.generateMnemonic(128); // 128 bits of entropy
const keyPair = TonWeb.utils.keyPairFromSeed(await TonWebMnemonic.mnemonicToSeed(mnemonic.split(' ')))

const WalletClass = tonweb.wallet.all['v4R2'];
const wallet = new WalletClass(tonweb.provider, {
    publicKey: keyPair.publicKey,
    wc: 0
});

const walletAddress = await wallet.getAddress();
const addressString = walletAddress.toString(true, true, false);

console.log('Jettons wallet:', addressString);
console.log('Send jettons for rewards and TON for gas to this address');
const keyPairData = {
    publicKey: TonWeb.utils.bytesToHex(keyPair.publicKey),
    secretKey: TonWeb.utils.bytesToHex(keyPair.secretKey),
    mnemonic: mnemonic
};
fs.writeFileSync('jetton_keeper_wallet.json', JSON.stringify(keyPairData, null, 2));

