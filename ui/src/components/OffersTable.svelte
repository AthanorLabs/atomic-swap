<script lang="ts">
  import DataTable, {
    Head,
    Body,
    Row,
    Cell,
    Label,
    SortValue,
  } from '@smui/data-table'
  import IconButton from '@smui/icon-button'
  import type { Offer } from 'src/types'
  import { offers } from '../stores/offerStore'
  import { isLoadingPeers } from '../stores/peerStore'
  import LinearProgress from '@smui/linear-progress'
  import Identicon from '../components/Identicon.svelte'
  import Tooltip, { Wrapper } from '@smui/tooltip'

  $: sortedOffers = $offers
  $: closed = !$isLoadingPeers
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
      <Cell sortable={false} columnId="id">
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
        <Cell>{offer.id}</Cell>
        <Cell>{offer.exchangeRate}</Cell>
        <Cell>{offer.maxAmount}</Cell>
        <Cell>{offer.minAmount}</Cell>
        <Cell>{offer.provides}</Cell>
      </Row>
    {/each}
  </Body>
  <LinearProgress
    indeterminate
    bind:closed
    aria-label="Data is being loaded..."
    slot="progress"
  />
</DataTable>
