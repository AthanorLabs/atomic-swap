<script lang="ts">
	import {
		Button,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import { isLoadingPeers, getPeers, peers } from '../../stores/peerStore';
  import Identicon from '../../lib/Identicon.svelte'
</script>

<div class="peers">
  <p>{$peers.length} Peers</p>
  <br>
	<Table class="border rounded">
		<TableHead>
			<TableHeadCell>Id</TableHeadCell>
			<TableHeadCell>Offers</TableHeadCell>
		</TableHead>
		<TableBody>
			{#each $peers as peerID}
				<TableBodyRow>
					<TableBodyCell>
						<Identicon peerAddress={peerID} />
						{peerID}
					</TableBodyCell>
					<TableBodyCell>?</TableBodyCell>
				</TableBodyRow>
			{/each}
		</TableBody>
	</Table>
</div>

<style>
.peers {
    max-width: 650px;
    margin: auto;
    margin-top: 20px;
    margin-bottom: 30px;
}
:global(.identicon > canvas) {
    border-radius: 50%;
}
</style>