import { defineStore } from "pinia"
import { ref } from "vue";
import { TonConnectUI } from '@tonconnect/ui';
import { Address, toNano, Cell } from '@ton/core';
import { TonClient, JettonMaster } from '@ton/ton';
import { JettonWallet } from '../helpers/jettonWallet.js'
import { useProductsStore } from "./productsStore.js";
import { useToastStore } from "./toastStore.js";
import { useMiniAppStore } from "./miniAppStore.js";
import { TON_CENTER_API, VITE_CONTRACT, VITE_TON_RPC } from "../constants/const.js";

export const useWalletStore = defineStore("wallet", () => {
  const tonConnect = ref(null);
  const tonClient = ref(null);
  const walletAddress = ref(null);
  const balance = ref({ ton: 0, blg: 0 });

  const miniAppStore = useMiniAppStore();
  const productStore = useProductsStore();
  const toastStore = useToastStore();

  const BLG_MASTER_ADDRESS = Address.parse(VITE_CONTRACT);
  const JETTON_TRANSFER_GAS_FEES = toNano('0.038');

  const initConnect = () => {
    if (!tonConnect.value) {
      tonConnect.value = new TonConnectUI({
        manifestUrl: 'https://raw.githubusercontent.com/gradosphera/brand-assets/refs/heads/main/academy/tonconnect-manifest.json'
      });
    }

    return tonConnect.value;
  }

  const initClient = async () => {
    if (!tonClient.value) {
      tonClient.value = new TonClient({ endpoint: VITE_TON_RPC });
    }

    return tonClient.value;
  }

  const connectWallet = async () => {
    try {
      let wallet;

      const tonConnect = initConnect();

      if (!tonConnect.connected) {
        wallet = await tonConnect.connectWallet();
      } else {
        wallet = tonConnect.wallet;
      }

      walletAddress.value = wallet?.account?.address ? Address.parse(wallet.account.address) : undefined;

      return walletAddress.value;
    } catch (error) {
      console.error('connectWallet:', error);
    }
  };

  const disconnectWallet = async () => {
    try {
      const tonConnect = initConnect();

      if (tonConnect.connected) await tonConnect.disconnect();

      walletAddress.value = null;
    } catch (error) {
      console.error('disconnectWallet:', error);
    }
  }

  const clearLocalStorage = () => {
    localStorage.removeItem('ton-connect-ui_last-selected-wallet-info');
    localStorage.removeItem('ton-connect-storage_bridge-connection');
    localStorage.removeItem('ton-connect-storage_http-bridge-gateway::https://tonconnectbridge.mytonwallet.org/bridge/');
    localStorage.removeItem('ton-connect-ui_wallet-info');
    localStorage.removeItem('ton-connect-ui_preferred-wallet');
  }

  const transferBLG = async (invoice) => {
    if (!invoice) return;

    const invoice_address = Address.parse(invoice.url);

    if (!productStore.selectedPaymentTariff) return;

    try {
      await disconnectWallet();

      clearLocalStorage();

      await connectWallet();

      const tonClient = await initClient();

      if (!walletAddress.value) return;

      const jettonMaster = tonClient.open(JettonMaster.create(BLG_MASTER_ADDRESS));
      const usersBlgAddress = await jettonMaster.getWalletAddress(walletAddress.value);
      const jettonWallet = tonClient.open(JettonWallet.createFromAddress(usersBlgAddress));


      // Fetch TON balance
      const response = await fetch(`https://toncenter.com/api/v2/getAddressInformation?address=${walletAddress.value}&api_key=${TON_CENTER_API}`);
      const data = await response.json();
      balance.value.ton = data.result.balance / 1e9;

      // Fetch BLG balance
      const walletData = await jettonWallet.getWalletData();
      balance.value.blg = Number(walletData.balance) / 1e0;

      if (balance.value.blg < invoice.amount_blg && balance.value.ton < 0.1) {
        toastStore.error({ text: t('general.toast_notifications.not_enough_money') })

        return;
      }

      let transactionResult = null;

      const resp = await jettonWallet.sendTransfer(
        {
          send: async (args) => {
            // Capture the result from sendTransaction
            transactionResult = await tonConnect.value.sendTransaction({
              messages: [
                {
                  address: args.to.toString(),
                  amount: args.value.toString(),
                  payload: args.body?.toBoc()?.toString('base64'),
                },
              ],
              validUntil: Date.now() + 5 * 60 * 1000, // 5 min
            });
          },
          address: walletAddress.value,
        },
        {
          fwdAmount: 1n,
          comment: invoice.id,
          jettonAmount: BigInt(invoice.amount_blg * 10 ** 0), // blago,
          toAddress: invoice_address,
          value: JETTON_TRANSFER_GAS_FEES,
        }
      );

      miniAppStore.afterPaymentData = { status: 'paid', product_id: productStore.selectedPaymentTariff?.product_id || '', service: 'ton' }

    } catch (error) {
      console.error('transferBLG:', error);
    }
  };

  return {
    walletAddress,

    connectWallet,
    disconnectWallet,
    transferBLG,
  }
})
