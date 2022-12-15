<script lang="ts">
  import DataTable, {
    Head,
    Body,
    Row,
    Cell,
    Label,
    SortValue,
  } from '@smui/data-table'
  import Button from '@smui/button'
  import IconButton from '@smui/icon-button'
  import type { Offer } from 'src/types'
  import { offers, selectedOffer } from '../stores/offerStore'
  import { isLoadingPeers } from '../stores/peerStore'
  import LinearProgress from '@smui/linear-progress'
  import Identicon from '../components/Identicon.svelte'
  import Tooltip, { Wrapper } from '@smui/tooltip'

  $: sortedOffers = $offers
  $: linearProgressClosed = !$isLoadingPeers
  let sort: keyof Offer = 'id'
  let sortDirection: Lowercase<keyof typeof SortValue> = 'ascending'

  function handleSort() {
    sortedOffers = $offers.sort((a, b) => {
      const [aVal, bVal] = [a[sort], b[sort]][
        sortDirection === 'ascending' ? 'slice' : 'reverse'
      ]()
      if (typeof aVal === 'string' && typeof bVal === 'string') {
        return aVal.localeCompare(bVal)
      }
      return Number(aVal) - Number(bVal)
    })
  }
</script>

<!-- 
  curl -X POST http://127.0.0.1:5001 -d '{"jsonrpc":"2.0","id":"0","method":"net_takeOffer", 
  "params":{
    "multiaddr":"/ip4/127.0.0.1/udp/9934/quic-v1/p2p/12D3KooWC547RfLcveQi1vBxACjnT6Uv15V11ortDTuxRWuhubGv",
    "offerID":"cf4bf01a0775a0d13fa41b14516e4b89034300707a1754e0d99b65f6cb6fffb9", 
    "providesAmount": 0.05 
}}' -H 'Content-Type: application/json' 
-->

<!-- {"jsonrpc":"2.0","result":{"success":true,"receivedAmount":2.999999999999},"id":"0"} -->
<div>
  <DataTable
    sortable
    bind:sort
    bind:sortDirection
    on:SMUIDataTable:sorted={handleSort}
    table$aria-label="User list"
    style="width: 100%;"
  >
    <Head>
      <Row>
        <!--
        Note: whatever you supply to "columnId" is
        appended with "-status-label" and used as an ID
        for the hidden label that describes the sort
        status to screen readers.

        You can localize those labels with the
        "sortAscendingAriaLabel" and
        "sortDescendingAriaLabel" props on the DataTable.
      -->
        <Cell columnId="peer">
          <!-- For numeric columns, icon comes first. -->
          <Label>Peer</Label>
          <IconButton class="material-icons">arrow_drop_up</IconButton>
        </Cell>
        <Cell sortable={false} columnId="id" class="idCell">
          <!-- For numeric columns, icon comes first. -->
          <!-- <IconButton class="material-icons">arrow_drop_up</IconButton> -->
          <Label>Offer id</Label>
        </Cell>
        <Cell columnId="exchangeRate">
          <Label>Exchange rate</Label>
          <!-- For non-numeric columns, icon comes second. -->
          <IconButton class="material-icons">arrow_drop_up</IconButton>
        </Cell>
        <Cell columnId="maxAmount">
          <Label>Max amount</Label>
          <IconButton class="material-icons">arrow_drop_up</IconButton>
        </Cell>
        <Cell columnId="minAmount" l>
          <Label>Min amount</Label>
          <IconButton class="material-icons">arrow_drop_up</IconButton>
        </Cell>
        <Cell>
          <Label>Provides</Label>
          <IconButton class="material-icons">arrow_drop_up</IconButton>
        </Cell>
        <!-- Button to take a deal -->
        <Cell />
      </Row>
    </Head>
    <Body>
      {#each sortedOffers as offer (offer.id)}
        <Row>
          <Cell>
            <Wrapper>
              <Identicon peerAddress={offer.peer} />
              <Tooltip xPos="center" style={'min-width: 100px'}>
                {offer.peer}
              </Tooltip>
            </Wrapper>
          </Cell>
          <Cell class="idCell">{offer.id}</Cell>
          <Cell>{offer.exchangeRate}</Cell>
          <Cell>{offer.maxAmount}</Cell>
          <Cell>{offer.minAmount}</Cell>
          <Cell>{offer.provides}</Cell>
          <Cell>
            <Button on:click={() => selectedOffer.set(offer)}>Take</Button>
          </Cell>
        </Row>
      {/each}
    </Body>
    <LinearProgress
      indeterminate
      bind:closed={linearProgressClosed}
      aria-label="Data is being loaded..."
      slot="progress"
    />
  </DataTable>
</div>

<style>
  * :global(.idCell) {
    max-width: 200px;
  }
</style>
