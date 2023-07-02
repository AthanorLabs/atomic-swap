<script>
  import { Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell } from 'flowbite-svelte';
  import { Badge, Button, ButtonGroup, Toolbar, ToolbarButton, ToolbarGroup, Heading, Search, Card } from 'flowbite-svelte';
  import TokenIcon from '$lib/TokenIcon.svelte';

  let value = '';

  /*
  interface Pair {
    ticker: string,
    token: string,
    offers: number,
    liquidity_token: number,
    liquidity_xmr: number,
    verified: boolean
  }
  */

  const pairs = [
    { ticker: 'ETH', token: 'eth', offers: 352, liquidity_token: 800, liquidity_xmr: 20, verified: true },
    { ticker: 'USDT', token: '0xdAC17F958D2ee523a2206206994597C13D831ec7', offers: 512, liquidity_token: 252890, liquidity_xmr: 12, verified: true },
    { ticker: 'USDC', token: '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48', offers: 502, liquidity_token: 302472, liquidity_xmr: 10, verified: true },
    { ticker: 'DAI', token: '0xdeadbeefa', offers: 417, liquidity_token: 221254, liquidity_xmr: 8, verified: true },
    { ticker: 'MKR', token: '0xdeadbeefb', offers: 17, liquidity_token: 242, liquidity_xmr: 2, verified: true },
    { ticker: 'PAXG', token: '0xdeadbeefc', offers: 17, liquidity_token: 27, liquidity_xmr: 9, verified: false },
    { ticker: 'SHITCOIN', token: '0xdeadbeefd', offers: 2, liquidity_token: 2930293023, liquidity_xmr: 1, verified: false },
  ]

  $: filteredPairs = pairs.filter(
    (item) => item.ticker.toLowerCase().indexOf(value.toLowerCase()) !== -1
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
        {pairs.length} Pairs
      </p>
    </Card>

    <Card>
      <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Liquidity</h5>
      <p class="font-normal text-gray-700 dark:text-gray-400 leading-tight">
        1812 XMR
      </p>
    </Card>

    <Card>
      <h5 class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Offers</h5>
      <p class="font-normal text-gray-700 dark:text-gray-400 leading-tight">
        1512
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
      {#each filteredPairs as pair (pair.token)}
      <TableBodyRow>
        <TableBodyCell>
          <TokenIcon size="32" ticker={pair.ticker} />
          <div class="ticker">
            <p>{pair.ticker}</p>
            {#if pair.verified}
              <Badge color="green">Verified</Badge>
            {:else}
              <Badge color="dark">Unverified</Badge>
            {/if}
          </div>
        </TableBodyCell>
        <TableBodyCell>
          <ButtonGroup>
            <Button style="border-radius: 5px 0 0 5px;" size="xs">{pair.liquidity_token.toLocaleString()} {pair.ticker}</Button>
            <Button style="border-radius: 0px 5px 5px 0;" size="xs">{pair.liquidity_xmr.toLocaleString()} XMR</Button>
          </ButtonGroup>
        </TableBodyCell>
        <TableBodyCell>{pair.offers}</TableBodyCell>
        <TableBodyCell>
          <ButtonGroup>
            <Button href="/offers/{pair.token}" color="light" size="xs">SEE OFFERS</Button>
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
h2 {
  margin: auto;
  text-align: center;
  font-size: 2em;
}
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
.search p {
  font-size: 1em;
  margin-top: 8px;
}
</style>
