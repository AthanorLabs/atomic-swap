<script lang="ts">
    import OffersTable from '$lib/OffersTable.svelte'
    import TakeDealDialog from '$lib/TakeDealDialog.svelte'
    import type { PageData } from './$types';
    import { offers, selectedOffer, refreshOffers } from '../../../stores/offerStore'
    import { EthTokenInfo, getTokenInfo } from '../../../utils'
    import type { TokenInfo } from '../../../types/PersonalTokenInfoResult'

    export let data: PageData
    $: filteredOffers = $offers.filter(off => off.ethAsset === data.token)

    const getToken = async (): Promise<TokenInfo> => {
        if (data.token.toLowerCase() === 'eth') return EthTokenInfo
        return getTokenInfo(data.token) || EthTokenInfo
    }
    let tokenInfoPromise = getToken()
    $: filteredOffers = $offers.filter(off => off.ethAsset.toLowerCase() === data.token.toLowerCase())
</script>

<div class="w-full">
    {#if offers }
        {#await tokenInfoPromise}
            <!-- TODO -->
        {:then token}
            <OffersTable tokenInfo={token} offers={filteredOffers} />
            <TakeDealDialog tokenInfo={token} />
        {:catch error}
            <!-- TODO -->
            <p style="color: red">{error.message}</p>
        {/await}
    {:else}
    <p>loading...</p>
    {/if}
</div>
