import {Entity as Entity_, Column as Column_, PrimaryColumn as PrimaryColumn_, StringColumn as StringColumn_, IntColumn as IntColumn_, OneToMany as OneToMany_} from "@subsquid/typeorm-store"
import {TokenWithChain} from "./tokenWithChain.model"
import {BridgeTxData} from "./bridgeTxData.model"

@Entity_()
export class Token {
    constructor(props?: Partial<Token>) {
        Object.assign(this, props)
    }

    @PrimaryColumn_()
    id!: string

    @StringColumn_({nullable: true})
    name!: string | undefined | null

    @StringColumn_({nullable: true})
    symbol!: string | undefined | null

    @IntColumn_({nullable: false})
    decimals!: number

    @OneToMany_(() => TokenWithChain, e => e.token)
    chains!: TokenWithChain[]

    @OneToMany_(() => BridgeTxData, e => e.token)
    transfers!: BridgeTxData[]
}
