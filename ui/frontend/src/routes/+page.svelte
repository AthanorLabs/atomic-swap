<script>
  import { Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell } from 'flowbite-svelte';
  import { Badge, Button, ButtonGroup, Toolbar, ToolbarButton, ToolbarGroup, Heading, Search, Card } from 'flowbite-svelte';
  import TokenIcon from '$lib/TokenIcon.svelte';

  import { pairs, liquidity, offers, isLoadingPairs, getPairs } from '../stores/pairStore'

  let value = '';

  $: filteredPairs = $pairs.filter(
    (item) => item.token.symbol.toLowerCase().indexOf(value.toLowerCase()) !== -1
  );
</script>

{#if filteredPairs.length > 0}
<div class="pairs" style="width: 100%; height: 100%;">
  <div class="search">
    <Search size="md" bind:value />
  </div>
  <Toolbar color="none">
    <Heading>
    </Heading>
    <ToolbarGroup slot="end">
    </ToolbarGroup>
  </Toolbar>

  <div class="grid grid-flow-col gap-4 mb-4">
    <Card shadow={false}>
      <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Pairs</h5>
      <p class="font-normal text-gray-700 dark:text-gray-400 leading-tight">
        {$pairs.length} Pairs
      </p>
    </Card>

    <Card shadow={false}>
      <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Reported Liquidity</h5>
      <p class="font-normal text-gray-700 dark:text-gray-400 leading-tight">
        {$liquidity} XMR
      </p>
    </Card>

    <Card shadow={false}>
      <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Offers</h5>
      <p class="font-normal text-gray-700 dark:text-gray-400 leading-tight">
        {$offers}
      </p>
    </Card>
  </div>

  <Table class="border rounded">
    <TableHead>
      <TableHeadCell>Ticker</TableHeadCell>
      <TableHeadCell>Reported Liquidity</TableHeadCell>
      <TableHeadCell>Offers</TableHeadCell>
      <TableHeadCell></TableHeadCell>
    </TableHead>
    <TableBody>
      {#each filteredPairs as pair (pair.token.symbol)}
      <TableBodyRow>
        <TableBodyCell>
          <TokenIcon size="32" ticker={pair.token.symbol} />
          <div class="ticker">
            <p>{pair.token.symbol}</p>
            {#if pair.verified}
              <Badge color="green">Verified</Badge>
            {:else}
              <Badge color="dark">Unverified</Badge>
            {/if}
          </div>
        </TableBodyCell>
        <TableBodyCell>
          {pair.reportedLiquidityXmr.toLocaleString()} XMR
        </TableBodyCell>
        <TableBodyCell>{pair.offers}</TableBodyCell>
        <TableBodyCell>
          <ButtonGroup>
            <Button href="/offers/{pair.ethAsset.toLocaleLowerCase()}" color="light" size="xs">SEE OFFERS</Button>
          </ButtonGroup>
        </TableBodyCell>
      </TableBodyRow>
      {/each}
    </TableBody>
  </Table>
</div>
{:else}
<div class="flex flex-auto justify-center items-center" style="height: 100vh;">
  <p class="text-center">No pairs found.</p>
</div>
{/if}

<style lang="postcss">
.pairs {
  max-width: 750px;
  margin: 0 auto;
  margin-bottom: 50px;
  padding: 25px;
}
.ticker {
  font-size: 1.2em;
  display: inline-block;
  vertical-align: middle;
}
.search {
  max-width: 400px;
  margin: auto;
  margin-top: 0px;
  margin-bottom: 0px;
}
</style>
