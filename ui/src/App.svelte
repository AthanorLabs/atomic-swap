<script lang="ts">
  import Spacer from './components/Spacer.svelte'
  import LayoutGrid, { Cell, InnerGrid } from '@smui/layout-grid'
  import Button from '@smui/button'
  // import CircularProgress from '@smui/circular-progress'
  import { peers, getPeers, isLoadingPeers } from './stores/peerStore'
  import { offers } from './stores/offerStore'
  import OffersTable from './components/OffersTable.svelte'
  import StatCard from './components/StatCard.svelte'

  const handleRefreshClick = () => {
    getPeers()
  }
</script>

<main>
  <LayoutGrid>
    <Spacer />
    <Cell spanDevices={{ desktop: 8, tablet: 6, phone: 12 }}>
      <InnerGrid>
        <Cell spanDevices={{ desktop: 2, tablet: 4, phone: 12 }}>
          <StatCard title="Peers" content={$peers.length.toString()} />
        </Cell>
        <Cell spanDevices={{ desktop: 2, tablet: 4, phone: 12 }}>
          <StatCard title="Offers" content={$offers.length.toString()} />
        </Cell>
        <Cell class="refreshButton">
          <Button on:click={handleRefreshClick}>Refresh</Button>
        </Cell>
      </InnerGrid>
      <br />
      <OffersTable />
      <!-- <h2>Peers:</h2>
      {#if $isLoadingPeers}
        <div class="loader">
          <CircularProgress style="height: 32px; width: 32px;" indeterminate />
        </div>
      {:else}
        {#each $peers as peer}
          <pre>{peer}</pre>
        {/each}
      {/if}
      <h2>Offers:</h2>
      <pre>{JSON.stringify($offers, null, 4)}</pre> -->
    </Cell>
  </LayoutGrid>
</main>

<svelte:head>
  <!-- <link rel="stylesheet" href="node_modules/svelte-material-ui/bare.css" /> -->
  <link
    rel="stylesheet"
    href="https://cdn.jsdelivr.net/npm/svelte-material-ui@6.0.0-beta.13/bare.min.css"
  />
  <!-- Material Icons -->
  <link
    rel="stylesheet"
    href="https://fonts.googleapis.com/icon?family=Material+Icons"
  />
  <!-- Roboto -->
  <link
    rel="stylesheet"
    href="https://fonts.googleapis.com/css?family=Roboto:300,400,500,600,700"
  />
  <!-- Roboto Mono -->
  <link
    rel="stylesheet"
    href="https://fonts.googleapis.com/css?family=Roboto+Mono"
  />
</svelte:head>

<style>
  * :global(.refreshButton) {
    display: flex;
    align-items: center;
  }
</style>
