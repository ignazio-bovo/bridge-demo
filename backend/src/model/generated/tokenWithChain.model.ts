import {Entity as Entity_, Column as Column_, PrimaryColumn as PrimaryColumn_, Index as Index_, ManyToOne as ManyToOne_, StringColumn as StringColumn_, BooleanColumn as BooleanColumn_} from "@subsquid/typeorm-store"
import {Token} from "./token.model"
import {Chain} from "./chain.model"

@Index_(["token", "chain"], {unique: false})
@Entity_()
export class TokenWithChain {
    constructor(props?: Partial<TokenWithChain>) {
        Object.assign(this, props)
    }

    @PrimaryColumn_()
    id!: string

    @ManyToOne_(() => Token, {nullable: true})
    token!: Token

    @Index_()
    @ManyToOne_(() => Chain, {nullable: true})
    chain!: Chain

    @StringColumn_({nullable: false})
    address!: string

    @BooleanColumn_({nullable: false})
    native!: boolean
}
