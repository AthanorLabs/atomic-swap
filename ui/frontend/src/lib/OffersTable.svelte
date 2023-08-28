<script lang="ts">
    import {onMount} from 'svelte';

    import { isLoadingPeers, getPeers } from '../stores/peerStore'
    import { selectedOffer, refreshOffers } from '../stores/offerStore'

    import { Button, Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell } from 'flowbite-svelte';
    import { Toolbar, ToolbarButton, ToolbarGroup } from 'flowbite-svelte';
    import { Heading } from 'flowbite-svelte'

    import Identicon from './Identicon.svelte'
    import TokenIcon from '$lib/TokenIcon.svelte';

    import xmr from '../assets/coins/xmr.png'
    
    import type { Offer, TokenInfo } from '../types'

    export let offers: Offer[]
    $: sortedOffers = offers
    $: count = offers ? offers.length : 0
    export let tokenInfo: TokenInfo

</script>

<div class="offers pb-20">
    <Toolbar color="none" style="position:relative;">
        <Heading tag="h5">
            <TokenIcon size={25} ticker={tokenInfo.symbol} />
            <span>{tokenInfo.symbol} / XMR</span>
            <img width="25" height="25" src={xmr} alt="xmr" style="display: inline; vertical-align: top;" />
        </Heading>

        <ToolbarGroup slot="end">
        <ToolbarButton>{ count } Offers</ToolbarButton>
        <ToolbarButton on:click={getPeers}>
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 12c0-1.232-.046-2.453-.138-3.662a4.006 4.006 0 00-3.7-3.7 48.678 48.678 0 00-7.324 0 4.006 4.006 0 00-3.7 3.7c-.017.22-.032.441-.046.662M19.5 12l3-3m-3 3l-3-3m-12 3c0 1.232.046 2.453.138 3.662a4.006 4.006 0 003.7 3.7 48.656 48.656 0 007.324 0 4.006 4.006 0 003.7-3.7c.017-.22.032-.441.046-.662M4.5 12l3 3m-3-3l-3 3" />
            </svg>              
        </ToolbarButton>
        </ToolbarGroup>
      </Toolbar>
      <br>
    {#if sortedOffers.length > 0}
    <Table class="offers-table border rounded" divClass="relative overflow-x-auto ">
    <TableHead>
        <TableHeadCell>Peer</TableHeadCell>
        <TableHeadCell>Offer Id</TableHeadCell>
        <TableHeadCell>Rate</TableHeadCell>
        <TableHeadCell>Min</TableHeadCell>
        <TableHeadCell>Max</TableHeadCell>
        <TableHeadCell></TableHeadCell>
    </TableHead>
    <TableBody>
        {#each sortedOffers as offer (offer.offerID)}
        <TableBodyRow>
            <TableBodyCell>
                <Identicon peerAddress={offer.peerID}/>
                <span style="display: inline;">{offer.peerID.slice(-8)}</span>
            </TableBodyCell>
            <TableBodyCell>{offer.offerID.slice(0,8)}</TableBodyCell>
            <TableBodyCell>{offer.exchangeRate}</TableBodyCell>
            <TableBodyCell>{Number(offer.minAmount)}</TableBodyCell>
            <TableBodyCell>{Number(offer.maxAmount)}</TableBodyCell>
            <TableBodyCell class="text-right">
                <Button on:click={() => selectedOffer.set(offer)} gradient color="purpleToBlue" size="xs">SWAP</Button>
            </TableBodyCell>
        </TableBodyRow>
        {/each}
    </TableBody>
    </Table>
    {:else}
    <div>
        <p class="text-center">No Offers</p>
    </div>
    {/if}
</div>

<style>
.offers {
    max-width: 650px;
    margin: auto;
    margin-top: 20px;
    margin-bottom: 30px;
}
:global(.identicon > canvas) {
    border-radius: 50%;
}
</style>
