<script lang="ts">
    import OffersTable from '$lib/OffersTable.svelte'
    import TakeDealDialog from '$lib/TakeDealDialog.svelte'
    import type { PageData } from './$types';
    import { offers, selectedOffer, refreshOffers } from '../../../stores/offerStore'

    export let data: PageData
    $: filteredOffers = $offers.filter(off => off.ethAsset === data.token)
</script>

<div>
    {#if offers }
    <OffersTable offers={filteredOffers} />
    <TakeDealDialog />
    {:else}
    <p>loading...</p>
    {/if}
</div>
