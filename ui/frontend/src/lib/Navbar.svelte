<script>
    import { Navbar, NavBrand, NavLi, NavUl, NavHamburger } from 'flowbite-svelte'
    import { Badge, Indicator } from 'flowbite-svelte'
    
    import { peers, getPeers } from '../stores/peerStore'

    import atomic from '../assets/logo.svg'
</script>


<Navbar let:hidden let:toggle>
    <NavBrand href="/">
        <img width="36" src={atomic} alt="logo" style="margin-right: 10px;" />
        <div>
            <p class="self-center whitespace-nowrap text-xl font-semibold dark:text-white">
                Atomic Swap
            </p>
            <Badge color="blue">Beta</Badge>
        </div>
    </NavBrand>
    <NavUl {hidden}>
        <NavLi href="/">Pairs</NavLi>
        <NavLi href="/">Wallets</NavLi>
        <NavLi href="/">Docs</NavLi>
        <NavLi target="_blank" href="https://github.com/AthanorLabs/atomic-swap">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 16 16">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.012 8.012 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
            </svg>
        </NavLi>
    </NavUl>
    <NavHamburger on:click={toggle} />
    <NavUl {hidden}>
        <!--<Badge rounded color="blue">mainnet</Badge>-->
{#if $peers.length > 0}
        <Badge color="green" rounded class="px-2.5 py-0.5">
            <Indicator color="green" size="xs" class="mr-1"/>{$peers.length.toString()} Peers
        </Badge>
        {:else}
        <Badge color="red" rounded class="px-2.5 py-0.5">
            <Indicator color="red" size="xs" class="mr-1"/>{$peers.length.toString()} Peers
        </Badge>
        {/if}
    </NavUl>
</Navbar>

<style>
</style>
