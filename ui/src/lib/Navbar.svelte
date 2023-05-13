<script>
	import { Alert, P } from 'flowbite-svelte'
    import { Navbar, NavBrand, NavLi, NavUl, NavHamburger } from 'flowbite-svelte'
    import { Badge } from 'flowbite-svelte'
    
    import { peers, getPeers } from '../stores/peerStore'

    import atomic from '../assets/logo.svg'
</script>


<Navbar let:hidden let:toggle>
    <NavBrand href="/">
        <img width="36" src={atomic} style="margin-right: 10px;" />
        <div>
            <p class="self-center whitespace-nowrap text-xl font-semibold dark:text-white">
                Atomic Swap
            </p>
            <Badge color="blue">Beta</Badge>
        </div>
    </NavBrand>
    <NavUl {hidden}>
        <NavLi href="/pairs">Pairs</NavLi>
        <NavLi href="/wallets">Wallets</NavLi>
        <NavLi href="/docs">Docs</NavLi>
    </NavUl>
    <NavHamburger on:click={toggle} />
    <NavUl {hidden}>
        <!--<Badge rounded color="blue">mainnet</Badge>-->
        <Badge border color="white"><div class="connected mr-2"></div>{$peers.length.toString()} Peers</Badge>
    </NavUl>
</Navbar>

<style>
    .connected {
        width: 8px; height: 8px;
        background: rgb(82, 246, 49);
        border-radius: 999px;
    }
</style>