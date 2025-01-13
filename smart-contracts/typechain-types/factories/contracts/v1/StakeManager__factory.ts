/* Autogenerated file. Do not edit manually. */
/* tslint:disable */
/* eslint-disable */
import {
  Contract,
  ContractFactory,
  ContractTransactionResponse,
  Interface,
} from "ethers";
import type { Signer, ContractDeployTransaction, ContractRunner } from "ethers";
import type { NonPayableOverrides } from "../../../common";
import type {
  StakeManager,
  StakeManagerInterface,
} from "../../../contracts/v1/StakeManager";

const _abi = [
  {
    inputs: [],
    stateMutability: "nonpayable",
    type: "constructor",
  },
  {
    inputs: [
      {
        internalType: "address",
        name: "target",
        type: "address",
      },
    ],
    name: "AddressEmptyCode",
    type: "error",
  },
  {
    inputs: [
      {
        internalType: "address",
        name: "implementation",
        type: "address",
      },
    ],
    name: "ERC1967InvalidImplementation",
    type: "error",
  },
  {
    inputs: [],
    name: "ERC1967NonPayable",
    type: "error",
  },
  {
    inputs: [],
    name: "FailedCall",
    type: "error",
  },
  {
    inputs: [],
    name: "InsufficientStakedBalance",
    type: "error",
  },
  {
    inputs: [],
    name: "InvalidAddress",
    type: "error",
  },
  {
    inputs: [],
    name: "InvalidInitialization",
    type: "error",
  },
  {
    inputs: [],
    name: "NotAdmin",
    type: "error",
  },
  {
    inputs: [],
    name: "NotBridge",
    type: "error",
  },
  {
    inputs: [],
    name: "NotInitializing",
    type: "error",
  },
  {
    inputs: [
      {
        internalType: "address",
        name: "owner",
        type: "address",
      },
    ],
    name: "OwnableInvalidOwner",
    type: "error",
  },
  {
    inputs: [
      {
        internalType: "address",
        name: "account",
        type: "address",
      },
    ],
    name: "OwnableUnauthorizedAccount",
    type: "error",
  },
  {
    inputs: [],
    name: "ReentrancyGuardReentrantCall",
    type: "error",
  },
  {
    inputs: [],
    name: "RewardRecipientAlreadyExists",
    type: "error",
  },
  {
    inputs: [],
    name: "RewardRecipientNotFound",
    type: "error",
  },
  {
    inputs: [],
    name: "StakingFailed",
    type: "error",
  },
  {
    inputs: [],
    name: "UUPSUnauthorizedCallContext",
    type: "error",
  },
  {
    inputs: [
      {
        internalType: "bytes32",
        name: "slot",
        type: "bytes32",
      },
    ],
    name: "UUPSUnsupportedProxiableUUID",
    type: "error",
  },
  {
    inputs: [],
    name: "UnstakingFailed",
    type: "error",
  },
  {
    anonymous: false,
    inputs: [
      {
        indexed: false,
        internalType: "uint256",
        name: "amount",
        type: "uint256",
      },
      {
        indexed: false,
        internalType: "uint64",
        name: "lastStakingBlock",
        type: "uint64",
      },
      {
        indexed: false,
        internalType: "uint64",
        name: "stakingEpochId",
        type: "uint64",
      },
    ],
    name: "FundsStakedOnPrecompile",
    type: "event",
  },
  {
    anonymous: false,
    inputs: [
      {
        indexed: false,
        internalType: "uint64",
        name: "version",
        type: "uint64",
      },
    ],
    name: "Initialized",
    type: "event",
  },
  {
    anonymous: false,
    inputs: [
      {
        indexed: true,
        internalType: "address",
        name: "previousOwner",
        type: "address",
      },
      {
        indexed: true,
        internalType: "address",
        name: "newOwner",
        type: "address",
      },
    ],
    name: "OwnershipTransferred",
    type: "event",
  },
  {
    anonymous: false,
    inputs: [
      {
        indexed: true,
        internalType: "address",
        name: "user",
        type: "address",
      },
      {
        indexed: false,
        internalType: "uint256",
        name: "amount",
        type: "uint256",
      },
    ],
    name: "TokensStaked",
    type: "event",
  },
  {
    anonymous: false,
    inputs: [
      {
        indexed: true,
        internalType: "address",
        name: "user",
        type: "address",
      },
      {
        indexed: false,
        internalType: "uint256",
        name: "amount",
        type: "uint256",
      },
    ],
    name: "TokensUnstaked",
    type: "event",
  },
  {
    anonymous: false,
    inputs: [
      {
        indexed: true,
        internalType: "address",
        name: "implementation",
        type: "address",
      },
    ],
    name: "Upgraded",
    type: "event",
  },
  {
    inputs: [],
    name: "NORMALIZER",
    outputs: [
      {
        internalType: "uint64",
        name: "",
        type: "uint64",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [],
    name: "ROOT_NET_UID",
    outputs: [
      {
        internalType: "uint16",
        name: "",
        type: "uint16",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [],
    name: "STAKE_INTERVAL",
    outputs: [
      {
        internalType: "uint64",
        name: "",
        type: "uint64",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [],
    name: "UPGRADE_INTERFACE_VERSION",
    outputs: [
      {
        internalType: "string",
        name: "",
        type: "string",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [
      {
        internalType: "address",
        name: "rewardAddress",
        type: "address",
      },
    ],
    name: "addBridgeParticipantReward",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [],
    name: "bridge",
    outputs: [
      {
        internalType: "contract IBridge",
        name: "",
        type: "address",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [],
    name: "daturaStakingHotkey",
    outputs: [
      {
        internalType: "bytes32",
        name: "",
        type: "bytes32",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [
      {
        internalType: "contract IBridge",
        name: "_bridge",
        type: "address",
      },
      {
        internalType: "uint256",
        name: "_rewardRate",
        type: "uint256",
      },
      {
        internalType: "bytes32",
        name: "_daturaHotkey",
        type: "bytes32",
      },
      {
        internalType: "address",
        name: "_stakePrecompile",
        type: "address",
      },
    ],
    name: "initialize",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [],
    name: "lastStakingBlock",
    outputs: [
      {
        internalType: "uint64",
        name: "",
        type: "uint64",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [],
    name: "nextStakingEpochId",
    outputs: [
      {
        internalType: "uint64",
        name: "",
        type: "uint64",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [],
    name: "owner",
    outputs: [
      {
        internalType: "address",
        name: "",
        type: "address",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [],
    name: "proxiableUUID",
    outputs: [
      {
        internalType: "bytes32",
        name: "",
        type: "bytes32",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [
      {
        internalType: "address",
        name: "rewardAddress",
        type: "address",
      },
    ],
    name: "removeBridgeParticipantReward",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [],
    name: "renounceOwnership",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [],
    name: "rewardRate",
    outputs: [
      {
        internalType: "uint256",
        name: "",
        type: "uint256",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [
      {
        internalType: "uint256",
        name: "amount",
        type: "uint256",
      },
      {
        internalType: "address",
        name: "user",
        type: "address",
      },
    ],
    name: "stake",
    outputs: [],
    stateMutability: "payable",
    type: "function",
  },
  {
    inputs: [],
    name: "stakePrecompile",
    outputs: [
      {
        internalType: "address",
        name: "",
        type: "address",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [],
    name: "stakeToPrecompile",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [
      {
        internalType: "address",
        name: "",
        type: "address",
      },
    ],
    name: "stakes",
    outputs: [
      {
        internalType: "uint256",
        name: "amount",
        type: "uint256",
      },
      {
        internalType: "uint64",
        name: "stakingEpochId",
        type: "uint64",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [
      {
        internalType: "uint64",
        name: "",
        type: "uint64",
      },
    ],
    name: "stakingEpochIdToLastStakingBlock",
    outputs: [
      {
        internalType: "uint64",
        name: "",
        type: "uint64",
      },
    ],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [
      {
        internalType: "address",
        name: "newOwner",
        type: "address",
      },
    ],
    name: "transferOwnership",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [
      {
        internalType: "uint256",
        name: "amount",
        type: "uint256",
      },
      {
        internalType: "address",
        name: "user",
        type: "address",
      },
    ],
    name: "unstake",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  {
    inputs: [
      {
        internalType: "address",
        name: "newImplementation",
        type: "address",
      },
      {
        internalType: "bytes",
        name: "data",
        type: "bytes",
      },
    ],
    name: "upgradeToAndCall",
    outputs: [],
    stateMutability: "payable",
    type: "function",
  },
  {
    stateMutability: "payable",
    type: "receive",
  },
] as const;

const _bytecode =
  "0x60a06040523060805234801561001457600080fd5b507ff0c57e16840df040f15088dc2f81fe391c3923bec73e23a9662efc9c229c6a00805468010000000000000000810460ff1615906001600160401b031660008115801561005f5750825b90506000826001600160401b0316600114801561007b5750303b155b905081158015610089575080155b156100a75760405163f92ee8a960e01b815260040160405180910390fd5b84546001600160401b031916600117855583156100d557845460ff60401b1916680100000000000000001785555b831561011b57845460ff60401b19168555604051600181527fc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d29060200160405180910390a15b5050505050608051611bb461014960003960008181610fa301528181610fcc015261110d0152611bb46000f3fe6080604052600436106101445760003560e01c80638381e182116100b6578063d544f6741161006f578063d544f6741461040d578063e78cea9214610423578063ec8b70d114610443578063ee8aae3014610459578063f2fde38b1461046e578063f82b45be1461048e57600080fd5b80638381e182146102f75780638da5cb5b14610317578063ad3cb1cc14610368578063af03528b146103a6578063b12f7998146103c6578063c347e4e8146103e657600080fd5b80634f1ef286116101085780634f1ef2861461024d57806352d1902d1461026057806365f5f67014610283578063715018a6146102b95780637acb7757146102ce5780637b0a47ee146102e157600080fd5b806307e72f451461015057806316934fc414610189578063230cde7b146101e3578063294e8d0e1461020b57806344f989901461022d57600080fd5b3661014b57005b600080fd5b34801561015c57600080fd5b5061016c670de0b6b3a764000081565b6040516001600160401b0390911681526020015b60405180910390f35b34801561019557600080fd5b506101c66101a4366004611835565b600460205260009081526040902080546001909101546001600160401b031682565b604080519283526001600160401b03909116602083015201610180565b3480156101ef57600080fd5b506101f8600081565b60405161ffff9091168152602001610180565b34801561021757600080fd5b5061022b610226366004611852565b6104ae565b005b34801561023957600080fd5b5060075461016c906001600160401b031681565b61022b61025b3660046118b2565b6106c0565b34801561026c57600080fd5b506102756106df565b604051908152602001610180565b34801561028f57600080fd5b5061016c61029e366004611975565b6008602052600090815260409020546001600160401b031681565b3480156102c557600080fd5b5061022b6106fc565b61022b6102dc36600461199e565b610710565b3480156102ed57600080fd5b5061027560025481565b34801561030357600080fd5b5061022b61031236600461199e565b6108f3565b34801561032357600080fd5b507f9016d09d72d40fdae2fd8ceac6b6234c7706214fd39c1cd1e609a0528c199300546001600160a01b03165b6040516001600160a01b039091168152602001610180565b34801561037457600080fd5b50610399604051806040016040528060058152602001640352e302e360dc1b81525081565b60405161018091906119f2565b3480156103b257600080fd5b50600054610350906001600160a01b031681565b3480156103d257600080fd5b5061022b6103e1366004611835565b610d82565b3480156103f257600080fd5b5060035461016c90600160a01b90046001600160401b031681565b34801561041957600080fd5b5061027560015481565b34801561042f57600080fd5b50600354610350906001600160a01b031681565b34801561044f57600080fd5b5061016c61016881565b34801561046557600080fd5b5061022b610de5565b34801561047a57600080fd5b5061022b610489366004611835565b610ef4565b34801561049a57600080fd5b5061022b6104a9366004611835565b610f34565b7ff0c57e16840df040f15088dc2f81fe391c3923bec73e23a9662efc9c229c6a008054600160401b810460ff1615906001600160401b03166000811580156104f35750825b90506000826001600160401b0316600114801561050f5750303b155b90508115801561051d575080155b1561053b5760405163f92ee8a960e01b815260040160405180910390fd5b845467ffffffffffffffff19166001178555831561056557845460ff60401b1916600160401b1785555b600380546001600160a01b0319166001600160a01b038b16179055600288905561058d610f6f565b610595610f77565b600187905560035460408051636e9960c360e01b815290516000926001600160a01b031691636e9960c39160048083019260209291908290030181865afa1580156105e4573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106089190611a25565b905061061381610f87565b600080546001600160a01b0319166001600160a01b0389161790556007805467ffffffffffffffff1916436001600160401b031617905560038054600160a01b67ffffffffffffffff60a01b1990911617905561066e610f77565b5083156106b557845460ff60401b19168555604051600181527fc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d29060200160405180910390a15b505050505050505050565b6106c8610f98565b6106d18261103d565b6106db8282611045565b5050565b60006106e9611102565b50600080516020611b5f83398151915290565b61070461114b565b61070e60006111a6565b565b6003546001600160a01b0316331461073b57604051637fea9dc560e01b815260040160405180910390fd5b600754600090610754906001600160401b031643611a58565b90506101686001600160401b03821610610847574761077281611217565b6007805467ffffffffffffffff1916436001600160401b039081169190911790915560038054600160a01b900490911690819060146107b083611a78565b82546101009290920a6001600160401b0381810219909316918316021790915560078054848316600081815260086020908152604091829020805494871667ffffffffffffffff199095169490941790935592548351888152941691840191909152908201527fcb35838c2d3b529d682bc65aac48be02411f8eb3b9146368db6feead7e1f6e2c915060600160405180910390a150505b6001600160a01b0382166000908152600460205260408120805485929061086f908490611a9e565b90915550506003546001600160a01b0383166000818152600460209081526040918290206001018054600160a01b9095046001600160401b031667ffffffffffffffff19909516949094179093555185815290917fb539ca1e5c8d398ddf1c41c30166f33404941683be4683319b57669a93dad4ef910160405180910390a2505050565b6003546001600160a01b0316331461091e57604051637fea9dc560e01b815260040160405180910390fd5b600754600090610937906001600160401b031643611a58565b90506101686001600160401b03821610610a2a574761095581611217565b6007805467ffffffffffffffff1916436001600160401b039081169190911790915560038054600160a01b9004909116908190601461099383611a78565b82546101009290920a6001600160401b0381810219909316918316021790915560078054848316600081815260086020908152604091829020805494871667ffffffffffffffff199095169490941790935592548351888152941691840191909152908201527fcb35838c2d3b529d682bc65aac48be02411f8eb3b9146368db6feead7e1f6e2c915060600160405180910390a150505b610a326112dd565b6001600160a01b03821660009081526004602052604090208054841115610a6c576040516345a5c39960e11b815260040160405180910390fd5b60035460018201546000916001600160401b03600160a01b909104811691161015610b14576001828101546001600160401b03908116600090815260086020526040902054670de0b6b3a76400009291610ac7911643611a58565b610ad19190611a58565b6001600160401b03166002548460000154610aec9190611ab1565b610af69190611ab1565b610b009190611ac8565b9050848111610b0f5780610b11565b845b90505b600354600183018054600160a01b9092046001600160401b031667ffffffffffffffff19909216919091179055815485908390600090610b55908490611aea565b90915550506040518581526001600160a01b038516907f9845e367b683334e5c0b12d7b81721ac518e649376fa65e3d68324e8f34f26799060200160405180910390a26000818015801590610bb357506000610bb16005611327565b115b15610c0457610bc26005611327565b610bcd600285611ac8565b610bd79085611aea565b610be19190611ac8565b9150610bed6005611327565b610bf79083611ab1565b610c019084611aea565b90505b8215610c1c57610c1c610c178489611a9e565b611337565b60006001600160a01b038716610c32838a611a9e565b604051600081818185875af1925050503d8060008114610c6e576040519150601f19603f3d011682016040523d82523d6000602084013e610c73565b606091505b5050905080610c9557604051635d7c508f60e11b815260040160405180910390fd5b600084118015610cae57506000610cac6005611327565b115b15610d4f5760005b610cc06005611327565b811015610d4d57610cd2600582611401565b6001600160a01b03168460405160006040518083038185875af1925050503d8060008114610d1c576040519150601f19603f3d011682016040523d82523d6000602084013e610d21565b606091505b50508092505081610d4557604051635d7c508f60e11b815260040160405180910390fd5b600101610cb6565b505b5050505050610d7d60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055565b505050565b610d8a61114b565b6001600160a01b038116610db15760405163e6c4247b60e01b815260040160405180910390fd5b610dbc60058261143a565b15610dda57604051632bc30e8360e11b815260040160405180910390fd5b6106db60058261145c565b600754600090610dfe906001600160401b031643611a58565b90506101686001600160401b03821610610ef15747610e1c81611217565b6007805467ffffffffffffffff1916436001600160401b039081169190911790915560038054600160a01b90049091169081906014610e5a83611a78565b82546101009290920a6001600160401b0381810219909316918316021790915560078054848316600081815260086020908152604091829020805494871667ffffffffffffffff199095169490941790935592548351888152941691840191909152908201527fcb35838c2d3b529d682bc65aac48be02411f8eb3b9146368db6feead7e1f6e2c915060600160405180910390a150505b50565b610efc61114b565b6001600160a01b038116610f2b57604051631e4fbdf760e01b8152600060048201526024015b60405180910390fd5b610ef1816111a6565b610f3c61114b565b610f4760058261143a565b610f6457604051633568a07760e11b815260040160405180910390fd5b6106db600582611471565b61070e611486565b610f7f611486565b61070e6114cf565b610f8f611486565b610ef1816114d7565b306001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016148061101f57507f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316611013600080516020611b5f833981519152546001600160a01b031690565b6001600160a01b031614155b1561070e5760405163703e46dd60e11b815260040160405180910390fd5b610ef161114b565b816001600160a01b03166352d1902d6040518163ffffffff1660e01b8152600401602060405180830381865afa92505050801561109f575060408051601f3d908101601f1916820190925261109c91810190611afd565b60015b6110c757604051634c9c8ce360e01b81526001600160a01b0383166004820152602401610f22565b600080516020611b5f83398151915281146110f857604051632a87526960e21b815260048101829052602401610f22565b610d7d83836114df565b306001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000161461070e5760405163703e46dd60e11b815260040160405180910390fd5b3361117d7f9016d09d72d40fdae2fd8ceac6b6234c7706214fd39c1cd1e609a0528c199300546001600160a01b031690565b6001600160a01b03161461070e5760405163118cdaa760e01b8152336004820152602401610f22565b7f9016d09d72d40fdae2fd8ceac6b6234c7706214fd39c1cd1e609a0528c19930080546001600160a01b031981166001600160a01b03848116918217845560405192169182907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a3505050565b60008054600154604080516024810192909252604480830185905281518084039091018152606490920181526020820180516001600160e01b031663b6b11efd60e01b179052516001600160a01b039092169291839185916112799190611b16565b60006040518083038185875af1925050503d80600081146112b6576040519150601f19603f3d011682016040523d82523d6000602084013e6112bb565b606091505b5050905080610d7d5760405163a437293760e01b815260040160405180910390fd5b7f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0080546001190161132157604051633ee5aeb560e01b815260040160405180910390fd5b60029055565b6000611331825490565b92915050565b6000805460015460408051602481019290925260448201859052606480830185905281518084039091018152608490920181526020820180516001600160e01b0316630636c28b60e01b179052516001600160a01b039092169291839161139d91611b16565b6000604051808303816000865af19150503d80600081146113da576040519150601f19603f3d011682016040523d82523d6000602084013e6113df565b606091505b5050905080610d7d57604051635d7c508f60e11b815260040160405180910390fd5b600061140d8383611535565b9392505050565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055565b6001600160a01b0381166000908152600183016020526040812054151561140d565b600061140d836001600160a01b03841661155f565b600061140d836001600160a01b0384166115ae565b7ff0c57e16840df040f15088dc2f81fe391c3923bec73e23a9662efc9c229c6a0054600160401b900460ff1661070e57604051631afcd79f60e31b815260040160405180910390fd5b611414611486565b610efc611486565b6114e8826116a8565b6040516001600160a01b038316907fbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b90600090a280511561152d57610d7d828261170d565b6106db611783565b600082600001828154811061154c5761154c611b32565b9060005260206000200154905092915050565b60008181526001830160205260408120546115a657508154600181810184556000848152602080822090930184905584548482528286019093526040902091909155611331565b506000611331565b600081815260018301602052604081205480156116975760006115d2600183611aea565b85549091506000906115e690600190611aea565b905080821461164b57600086600001828154811061160657611606611b32565b906000526020600020015490508087600001848154811061162957611629611b32565b6000918252602080832090910192909255918252600188019052604090208390555b855486908061165c5761165c611b48565b600190038181906000526020600020016000905590558560010160008681526020019081526020016000206000905560019350505050611331565b6000915050611331565b5092915050565b806001600160a01b03163b6000036116de57604051634c9c8ce360e01b81526001600160a01b0382166004820152602401610f22565b600080516020611b5f83398151915280546001600160a01b0319166001600160a01b0392909216919091179055565b6060600080846001600160a01b03168460405161172a9190611b16565b600060405180830381855af49150503d8060008114611765576040519150601f19603f3d011682016040523d82523d6000602084013e61176a565b606091505b509150915061177a8583836117a2565b95945050505050565b341561070e5760405163b398979f60e01b815260040160405180910390fd5b6060826117b7576117b2826117f7565b61140d565b81511580156117ce57506001600160a01b0384163b155b156116a157604051639996b31560e01b81526001600160a01b0385166004820152602401610f22565b8051156118075780518082602001fd5b60405163d6bda27560e01b815260040160405180910390fd5b6001600160a01b0381168114610ef157600080fd5b60006020828403121561184757600080fd5b813561140d81611820565b6000806000806080858703121561186857600080fd5b843561187381611820565b93506020850135925060408501359150606085013561189181611820565b939692955090935050565b634e487b7160e01b600052604160045260246000fd5b600080604083850312156118c557600080fd5b82356118d081611820565b915060208301356001600160401b03808211156118ec57600080fd5b818501915085601f83011261190057600080fd5b8135818111156119125761191261189c565b604051601f8201601f19908116603f0116810190838211818310171561193a5761193a61189c565b8160405282815288602084870101111561195357600080fd5b8260208601602083013760006020848301015280955050505050509250929050565b60006020828403121561198757600080fd5b81356001600160401b038116811461140d57600080fd5b600080604083850312156119b157600080fd5b8235915060208301356119c381611820565b809150509250929050565b60005b838110156119e95781810151838201526020016119d1565b50506000910152565b6020815260008251806020840152611a118160408501602087016119ce565b601f01601f19169190910160400192915050565b600060208284031215611a3757600080fd5b815161140d81611820565b634e487b7160e01b600052601160045260246000fd5b6001600160401b038281168282160390808211156116a1576116a1611a42565b60006001600160401b03808316818103611a9457611a94611a42565b6001019392505050565b8082018082111561133157611331611a42565b808202811582820484141761133157611331611a42565b600082611ae557634e487b7160e01b600052601260045260246000fd5b500490565b8181038181111561133157611331611a42565b600060208284031215611b0f57600080fd5b5051919050565b60008251611b288184602087016119ce565b9190910192915050565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052603160045260246000fdfe360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbca2646970667358221220694efe3646be9c29010375493bb672eb7c87010c525221f0060474617ee8e9de64736f6c63430008160033";

type StakeManagerConstructorParams =
  | [signer?: Signer]
  | ConstructorParameters<typeof ContractFactory>;

const isSuperArgs = (
  xs: StakeManagerConstructorParams
): xs is ConstructorParameters<typeof ContractFactory> => xs.length > 1;

export class StakeManager__factory extends ContractFactory {
  constructor(...args: StakeManagerConstructorParams) {
    if (isSuperArgs(args)) {
      super(...args);
    } else {
      super(_abi, _bytecode, args[0]);
    }
  }

  override getDeployTransaction(
    overrides?: NonPayableOverrides & { from?: string }
  ): Promise<ContractDeployTransaction> {
    return super.getDeployTransaction(overrides || {});
  }
  override deploy(overrides?: NonPayableOverrides & { from?: string }) {
    return super.deploy(overrides || {}) as Promise<
      StakeManager & {
        deploymentTransaction(): ContractTransactionResponse;
      }
    >;
  }
  override connect(runner: ContractRunner | null): StakeManager__factory {
    return super.connect(runner) as StakeManager__factory;
  }

  static readonly bytecode = _bytecode;
  static readonly abi = _abi;
  static createInterface(): StakeManagerInterface {
    return new Interface(_abi) as StakeManagerInterface;
  }
  static connect(
    address: string,
    runner?: ContractRunner | null
  ): StakeManager {
    return new Contract(address, _abi, runner) as unknown as StakeManager;
  }
}
