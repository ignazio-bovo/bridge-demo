import {Entity as Entity_, Column as Column_, PrimaryColumn as PrimaryColumn_, StringColumn as StringColumn_, OneToMany as OneToMany_} from "@subsquid/typeorm-store"
import {TokenWithChain} from "./tokenWithChain.model"
import {BridgeTxData} from "./bridgeTxData.model"

@Entity_()
export class Chain {
    constructor(props?: Partial<Chain>) {
        Object.assign(this, props)
    }

    @PrimaryColumn_()
    id!: string

    @StringColumn_({nullable: false})
    name!: string

    @OneToMany_(() => TokenWithChain, e => e.chain)
    tokens!: TokenWithChain[]

    @OneToMany_(() => BridgeTxData, e => e.destinationChain)
    incomingTransfers!: BridgeTxData[]

    @OneToMany_(() => BridgeTxData, e => e.sourceChain)
    outgoingTransfers!: BridgeTxData[]
}
