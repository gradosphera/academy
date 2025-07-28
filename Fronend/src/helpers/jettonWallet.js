import {beginCell} from '@ton/core';

export class JettonWallet {
  static OPCODES = {
    TRANSFER: 0xf8a7ea5,
  };

  constructor(address) {
    this.address = address;
  }

  static createFromAddress(address) {
    return new JettonWallet(address);
  }

  async getWalletData(provider) {
    const {stack} = await provider.get('get_wallet_data', []);

    return {
      balance: stack.readBigNumber(),
      ownerAddress: stack.readAddress(),
      jettonMasterAddress: stack.readAddress(),
      jettonWalletCode: stack.readCell(),
    };
  }

  async sendTransfer(provider, via, opts) {
    const builder = beginCell()
      .storeUint(JettonWallet.OPCODES.TRANSFER, 32)
      .storeUint(opts.queryId ?? 0, 64)
      .storeCoins(opts.jettonAmount)
      .storeAddress(opts.toAddress)
      .storeAddress(via.address)
      .storeUint(0, 1)
      .storeCoins(opts.fwdAmount);

    if ('comment' in opts) {
      const commentPayload = beginCell()
        .storeUint(0, 32)
        .storeStringTail(opts.comment)
        .endCell();

      builder.storeBit(1);
      builder.storeRef(commentPayload);
    } else {
      if (opts.forwardPayload && opts.forwardPayload.beginParse) {
        builder.storeBit(0);
        builder.storeSlice(opts.forwardPayload);
      } else if (opts.forwardPayload) {
        builder.storeBit(1);
        builder.storeRef(opts.forwardPayload);
      } else {
        builder.storeBit(0);
      }
    }

    await provider.internal(via, {
      value: opts.value,
      sendMode: 1, // SendMode.PAY_GAS_SEPARATELY
      body: builder.endCell(),
    });
  }
}
