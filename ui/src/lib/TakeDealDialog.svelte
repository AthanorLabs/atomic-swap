<script lang="ts">
  import type { CancelResult } from '../types/Cancel'
  import type { NetTakeOfferSyncResult } from '../types/NetTakeOfferSync'
  import { getCorrespondingToken, rpcRequest, getPort } from '../utils'
  import { selectedOffer } from '../stores/offerStore'
  import { getPeers } from '../stores/peerStore'
  //import { currentAccount, sign } from '../stores/metamask'
  import Loader from './Loader.svelte'

  import { Button, Modal } from 'flowbite-svelte'
  import { Badge, Label, Input, Helper, InputAddon, ButtonGroup, Spinner } from 'flowbite-svelte'
  
  import { CheckSolid, XmarkSolid } from 'svelte-awesome-icons';
  
  import eth from '../assets/coins/eth.png'
  
  let popupModal = true;

  const WS_ADDRESS = `ws://127.0.0.1:${getPort()}/ws`

  let amountProvided: number | null = null
  let isSuccess = false
  let isLoadingSwap = false
  let error = ''
  let swapError = ''
  let swapStatus = ''

  $: willReceive =
    amountProvided && amountProvided > 0 && $selectedOffer?.exchangeRate
      ? amountProvided / $selectedOffer.exchangeRate
      : 0

  $: if (
    willReceive !== 0 &&
    $selectedOffer &&
    willReceive < $selectedOffer.minAmount
  ) {
    error = `The amount of ${getCorrespondingToken(
      $selectedOffer.provides
    )} to swap is too low`
  } else if (
    willReceive !== 0 &&
    $selectedOffer &&
    willReceive > $selectedOffer.maxAmount
  ) {
    error = `The amount of ${getCorrespondingToken(
      $selectedOffer.provides
    )} to swap is too high`
  } else {
    error = ''
  }

  const handleSendTakeOffer = () => {
    const offerID = $selectedOffer?.offerID
    const webSocket = new WebSocket(WS_ADDRESS)

    webSocket.onopen = () => {
      console.log('WebSocket opened')
      const req = {
        jsonRPC: '2.0',
        id: 0,
        method: 'net_takeOfferAndSubscribe',
        params: {
          peerID: $selectedOffer?.peerID,
          offerID,
          providesAmount: amountProvided,
        },
      }
      webSocket.send(JSON.stringify(req))
      console.log('takeOfferAndSubscribe sent', req)
    }

    webSocket.onmessage = async (msg) => {
      const { result, error } = JSON.parse(msg.data)
      if (error) {
        console.error(error)
        swapError = error.message
        isSuccess = false
        isLoadingSwap = false
        getPeers()
        return
      }

      const { status } = result
      swapStatus = status
      if (status === "Success") {
        isSuccess = true
        isLoadingSwap = false
        getPeers()
      }
    }

    webSocket.onclose = (event: Event) => {
      console.log('closed:', event)
      swapError = "Swapd websocket closed"
      isLoadingSwap = false
    }

    webSocket.onerror = (event: Event) => {
      console.error(event)
      swapError = event.toString()
      isLoadingSwap = false
    }

    isLoadingSwap = true
  }

  const onReset = (resetOffer = true) => {
    resetOffer && selectedOffer.set(undefined)
    amountProvided = 0
    willReceive = 0
    isSuccess = false
    swapError = ''
    swapStatus = ''
  }
</script>

{#if !!$selectedOffer}
  <Modal bind:open={$selectedOffer} permanent="true" size="xs" class="{isLoadingSwap ? 'modal-hide-footer' : ''}"style="min-width: 300px;"> 
    <section class="container">
      {#if isLoadingSwap}
        <div class="flexBox justify-center align-center text-center">
          <Spinner size={10}/>
          <p class="mt-5 m-auto">Swapping ...</p>
          <p class="mt-1 m-auto">{swapStatus}</p>
        </div>
      {:else if isSuccess}
        <div class="flexBox text-center justify-center">
          <CheckSolid class="m-auto mb-3 w-16 h-16" style="color: #0c4;" />
          <p class="successMessage">
            You received <b>{willReceive} {$selectedOffer.provides}</b>
          </p>
        </div>
      {:else if !!swapError}
        <div class="flexBox text-center justify-center">
          <XmarkSolid class="m-auto mb-3 w-16 h-16" style="color: #e40;" /> 
          <p class="errorMessage">
            {swapError}
          </p>
        </div>
      {:else}
      Offer ID
      <br>
      <Badge border color="blue" large>{$selectedOffer.offerID.slice(0,12)}...</Badge>    
      
      <div class='mt-4 mb-1'>
        <Label 
          for='default-input' class='block mb-2'>
          {getCorrespondingToken($selectedOffer.provides)} amount
          <span>
            (Min {$selectedOffer.exchangeRate * $selectedOffer.minAmount}
            / Max {$selectedOffer.exchangeRate * $selectedOffer.maxAmount})
          </span>
        </Label>
        <Input 
            bind:value={amountProvided}
            invalid={!!error}
            id='large-input'
            size="lg"
            placeholder="Your amount ...">
            <img slot="left" width="32" height="32" src={eth} alt="asset" class="pr-1" />
        </Input>
        <Helper class="mt-2" color="red">{error}</Helper>
      </div>
          
     <p class="text-center pt-4">You will receive<br>{willReceive} XMR</p>
      
      {/if}
    </section>

    <svelte:fragment slot="footer">
      {#if isSuccess}
        <Button on:click={() => onReset(true)} class="w-full" gradient color="purpleToBlue">Done</Button>
      {:else if !!swapError}
        <Button on:click={() => onReset(false)} class="w-full" gradient color="purpleToBlue">Back</Button>
      {:else if !isLoadingSwap}
          <Button on:click={() => selectedOffer.set(null)} color='alternative' class="w-1/2">CANCEL</Button>
          <Button on:click={handleSendTakeOffer} disabled={isLoadingSwap || !!error || !willReceive} class="w-1/2" gradient color="cyanToBlue" s>SWAP</Button>
      {/if}
    </svelte:fragment>
  </Modal>
{/if}

<style>
:global(.modal-hide-footer > div:last-of-type) {
  display: none;
}
</style>
