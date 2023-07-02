<script lang="ts">
    import {onMount} from 'svelte';

    import { offers, selectedOffer, refreshOffers } from '../stores/offerStore'
    import { isLoadingPeers, getPeers } from '../stores/peerStore'

    import { Button, Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell } from 'flowbite-svelte';
    import { Toolbar, ToolbarButton, ToolbarGroup } from 'flowbite-svelte';
    import { Heading } from 'flowbite-svelte'

    import Identicon from './Identicon.svelte'

    import xmr from '../assets/coins/xmr.png'
    import eth from '../assets/coins/eth.png'

    $: sortedOffers = $offers
</script>

<div class="offers">
    <Toolbar color="none" style="position:relative;">
        <Heading tag="h5">
            <img width="25" height="25" src={eth} alt="eth" style="display: inline; vertical-align: top;"/>
            <span>ETH / XMR</span>
            <img width="25" height="25" src={xmr} alt="xmr" style="display: inline; vertical-align: top;" />
        </Heading>

        <ToolbarGroup slot="end">
        <ToolbarButton>
            {sortedOffers.length} Offers
        </ToolbarButton>
        <ToolbarButton on:click={getPeers}>
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 12c0-1.232-.046-2.453-.138-3.662a4.006 4.006 0 00-3.7-3.7 48.678 48.678 0 00-7.324 0 4.006 4.006 0 00-3.7 3.7c-.017.22-.032.441-.046.662M19.5 12l3-3m-3 3l-3-3m-12 3c0 1.232.046 2.453.138 3.662a4.006 4.006 0 003.7 3.7 48.656 48.656 0 007.324 0 4.006 4.006 0 003.7-3.7c.017-.22.032-.441.046-.662M4.5 12l3 3m-3-3l-3 3" />
            </svg>              
        </ToolbarButton>
        </ToolbarGroup>
      </Toolbar>
      <br>
    {#if sortedOffers.length > 0}
    <Table class="offers" divClass="relative overflow-x-auto sm:rounded-lg" shadow>
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
            <TableBodyCell>{offer.minAmount}</TableBodyCell>
            <TableBodyCell>{offer.maxAmount}</TableBodyCell>
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
    margin-top: 60px;
    margin-bottom: 40px;
}
:global(.identicon > canvas) {
    border-radius: 50%;
}
</style>
