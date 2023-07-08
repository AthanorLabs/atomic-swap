<script>
  import { Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell } from 'flowbite-svelte';
  import { Badge, Button, ButtonGroup, Toolbar, ToolbarButton, ToolbarGroup, Heading, Search, Card } from 'flowbite-svelte';
  import TokenIcon from '$lib/TokenIcon.svelte';

  import { pairs, liquidity, offers, isLoadingPairs, getPairs } from '../stores/pairStore'

  let value = '';

  $: filteredPairs = $pairs.filter(
    (item) => item.asset.toLowerCase().indexOf(value.toLowerCase()) !== -1
  );
</script>

<div class="pairs m-5">
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
    <Card>
      <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Pairs</h5>
      <p class="font-normal text-gray-700 dark:text-gray-400 leading-tight">
        {$pairs.length} Pairs
      </p>
    </Card>

    <Card>
      <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Liquidity</h5>
      <p class="font-normal text-gray-700 dark:text-gray-400 leading-tight">
        {$liquidity} XMR
      </p>
    </Card>

    <Card>
      <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Offers</h5>
      <p class="font-normal text-gray-700 dark:text-gray-400 leading-tight">
        {$offers}
      </p>
    </Card>
  </div>

  {#if filteredPairs.length > 0}
  <Table shadow>
    <TableHead>
      <TableHeadCell>Ticker</TableHeadCell>
      <TableHeadCell>Liquidity</TableHeadCell>
      <TableHeadCell>Offers</TableHeadCell>
      <TableHeadCell></TableHeadCell>
    </TableHead>
    <TableBody class="divide-y">
      {#each filteredPairs as pair (pair.asset)}
      <TableBodyRow>
        <TableBodyCell>
          <TokenIcon size="32" ticker={pair.asset} />
          <div class="ticker">
            <p>{pair.asset}</p>
            {#if pair.verified}
              <Badge color="green">Verified</Badge>
            {:else}
              <Badge color="dark">Unverified</Badge>
            {/if}
          </div>
        </TableBodyCell>
        <TableBodyCell>
          <ButtonGroup>
            <Button style="border-radius: 5px 0 0 5px;" size="xs">{pair.liquidityEth.toLocaleString()} {pair.asset}</Button>
            <Button style="border-radius: 0px 5px 5px 0;" size="xs">{pair.liquidityXmr.toLocaleString()} XMR</Button>
          </ButtonGroup>
        </TableBodyCell>
        <TableBodyCell>{pair.offers}</TableBodyCell>
        <TableBodyCell>
          <ButtonGroup>
            <Button href="/offers/{pair.asset.toLocaleLowerCase()}" color="light" size="xs">SEE OFFERS</Button>
          </ButtonGroup>
        </TableBodyCell>
      </TableBodyRow>
      {/each}
    </TableBody>
  </Table>
  {:else}
  <p class="text-center">No pairs found.</p>
  {/if}
</div>

<style lang="postcss">
.pairs {
  max-width: 750px;
  margin: auto;
  margin-bottom: 50px;
}
.ticker {
  font-size: 1.2em;
  display: inline-block;
  vertical-align: middle;
}
.search {
  max-width: 400px;
  margin: auto;
  margin-top: 60px;
  margin-bottom: 30px;
}
</style>
