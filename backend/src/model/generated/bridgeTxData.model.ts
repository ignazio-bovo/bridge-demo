import {Entity as Entity_, Column as Column_, PrimaryColumn as PrimaryColumn_, ManyToOne as ManyToOne_, Index as Index_, StringColumn as StringColumn_, BigIntColumn as BigIntColumn_, BooleanColumn as BooleanColumn_, DateTimeColumn as DateTimeColumn_} from "@subsquid/typeorm-store"
import {Chain} from "./chain.model"
import {Token} from "./token.model"

@Entity_()
export class BridgeTxData {
    constructor(props?: Partial<BridgeTxData>) {
        Object.assign(this, props)
    }

    @PrimaryColumn_()
    id!: string

    @Index_()
    @ManyToOne_(() => Chain, {nullable: true})
    sourceChain!: Chain

    @Index_()
    @ManyToOne_(() => Chain, {nullable: true})
    destinationChain!: Chain

    @StringColumn_({nullable: false})
    sourceAddress!: string

    @StringColumn_({nullable: false})
    destinationAddress!: string

    @BigIntColumn_({nullable: false})
    amount!: bigint

    @BigIntColumn_({nullable: false})
    nonce!: bigint

    @Index_()
    @ManyToOne_(() => Token, {nullable: true})
    token!: Token

    @BooleanColumn_({nullable: false})
    confirmed!: boolean

    @DateTimeColumn_({nullable: false})
    createdAtTimestamp!: Date

    @DateTimeColumn_({nullable: true})
    confirmedAtTimestamp!: Date | undefined | null
}
