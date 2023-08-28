<script>
  import { Sidebar, SidebarWrapper, SidebarBrand, SidebarItem, SidebarGroup } from 'flowbite-svelte'
  import { peers, getPeers } from '../stores/peerStore'

  import atomic from '../assets/logo.svg'
  import wallet from '../assets/icons/wallet.svg'
  import bookmark from '../assets/icons/bookmark.svg'
  import transfer from '../assets/icons/transfer.svg'
  import network from '../assets/icons/network.svg'

  import { Badge, Indicator } from 'flowbite-svelte'

  import { page } from '$app/stores';
  $: activeUrl = $page.url.pathname;

  let spanClass = 'flex-1 ml-3 whitespace-nowrap';
</script>

<Sidebar style="width: auto;">
    <SidebarWrapper style="height: 100%;" class="sidebar-wrapper fixed">
      <SidebarGroup>
        <img width="75" src={transfer} alt="logo" class="mx-auto mb-5" style="opacity: 0.75;" />
        <SidebarItem label="Trade" href="/" active={activeUrl === '/' || activeUrl.includes('/offers') }>
          <svelte:fragment slot="icon">
            <img src={transfer} alt="trade" width="24" />
          </svelte:fragment>
        </SidebarItem>
        <SidebarItem label="Swaps" {spanClass}>
          <svelte:fragment slot="icon">
            <img src={bookmark} alt="swaps" width="24" />
          </svelte:fragment>
          <svelte:fragment slot="subtext">
            <Badge color="dark" class="px-2.5">
              0
            </Badge>
          </svelte:fragment>
        </SidebarItem>
        <SidebarItem label="Wallets" {spanClass}>
          <svelte:fragment slot="icon">
            <img src={wallet} alt="wallet" width="24" />
          </svelte:fragment>
        </SidebarItem>
        <SidebarItem label="Peers" href="/peers" active={activeUrl === '/peers'}>
          <svelte:fragment slot="icon">
            <img src={network} alt="network" width="24" />
          </svelte:fragment>
          <svelte:fragment slot="subtext">
            {#if $peers.length > 0}
            <Badge color="green" class="ml-4">
              {$peers.length.toString()}
            </Badge>
            {:else}
            <Badge color="red" class="ml-4">
              {$peers.length.toString()}
            </Badge>
            {/if}
          </svelte:fragment>
        </SidebarItem>
      </SidebarGroup>
    </SidebarWrapper>
  </Sidebar>
  

<style>
  :global(.sidebar-wrapper) {
    /*box-shadow: 0px 0px 15px #0001;*/
    z-index: 9;
    background: #fff;
    border-right: 1px solid #eee;
  }
</style>