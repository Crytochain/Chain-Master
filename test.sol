pragma solidity ^0.4.0;

contract addab {
  uint public count;

  function add(uint a, uint b) returns(uint) {
    count++;
    return a + b;
  }

  function getCount() returns(uint){
    return count;
  }
}

st, _ := sc.State()
    msg := libtypes.NewMessage(args.Sender, &args.SubChainAddr, 0, big.NewInt(0), big.NewInt(0), big.NewInt(0),
        []byte{}, false, false, false, big.NewInt(0), common.Address{}, 0)
    context := blockchain.NewEVMContext(msg, sc.CurrentHeader(), sc, nil, nil)
    evm := vm.NewEVM(context, st, params.AllProtocolChanges, vm.Config{EnableJit: false, ForceJit: false})

    abiobj, err := abi.JSON(strings.NewReader(dappAbi))
    if err != nil {
        return err
    }

    var data []byte
    var retJson string
    if paramsLen == 1 {
        data, err = abiobj.PackRemote(args.Params[0], nil)
    } else {
        data, err = abiobj.PackRemote(args.Params[0], args.Params[1:])
    }
    if err != nil {
        return err
    }
    ret, _, err := evm.Call(vm.AccountRef(args.Sender), 
    sc.DappAddr, data, params.SCSTxGasLimit, big.NewInt(0))
    if err != nil {
        return err
    }
    retJson, err = abiobj.UnpackToJson(args.Params[0], ret)
    if err != nil {
        return err
    }
    *reply = retJson
