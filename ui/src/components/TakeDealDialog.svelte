<script lang="ts">
  import Dialog, { Title, Content, Actions } from '@smui/dialog'
  import Button, { Label } from '@smui/button'
  import type { NetTakeOfferSyncResult } from 'src/types/NetTakeOfferSync'
  import { getCorrespondingToken, rpcRequest } from 'src/utils'
  import { selectedOffer } from '../stores/offerStore'
  import { getPeers } from '../stores/peerStore'
  import Textfield from '@smui/textfield'
  import { mdiSwapVertical } from '@mdi/js'
  import { Icon } from '@smui/icon-button'
  import { Svg } from '@smui/common/elements'
  import CircularProgress from '@smui/circular-progress'
  import HelperText from '@smui/textfield/helper-text'
  import { currentAccount, sign } from '../stores/metamask'

  const WS_ADDRESS = 'ws://127.0.0.1:8081'

  let amountProvided: number | null = null
  let xmrAddress = ''
  let isSuccess = false
  let isLoadingSwap = false
  let error = ''
  let swapError = ''

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
    const offerID = $selectedOffer?.id
    const webSocket = new WebSocket(WS_ADDRESS)

    webSocket.onopen = () => {
      console.log('opened')
      console.log('sending ws signer msg')
      const req = {
        method: 'signer_subscribe',
        params: {
          jsonRPC: '2.0',
          id: '0',
          offerID,
          ethAddress: $currentAccount,
          xmrAddress,
        },
      }
      webSocket.send(JSON.stringify(req))
      console.log('sent ws signer msg', req)
    }

    webSocket.onmessage = async (msg) => {
      console.log('message to sign:', msg.data)
      const txHash = await sign(msg.data)
      console.log('signed txHash', txHash)
      const out = {
        offerID,
        txHash,
      }
      webSocket.send(JSON.stringify(out))
    }

    webSocket.onclose = (e) => {
      console.log('closed:', e)
    }

    webSocket.onerror = (e) => {
      console.log('error', e)
    }

    isLoadingSwap = true

    rpcRequest<NetTakeOfferSyncResult | undefined>('net_takeOfferSync', {
      multiaddr: $selectedOffer?.peer,
      offerID,
      providesAmount: Number(amountProvided),
    })
      .then(({ result }) => {
        console.log('result NetTakeOfferSyncResult', result)

        if (result?.status === 'Success') {
          isSuccess = true
          getPeers()
        } else if (result?.status === 'Aborted') {
          swapError = 'Something went wrong. Please check your node logs'
        } else if (result?.status === 'Refunded') {
          swapError =
            'Something went wrong. Swap funds refunded, please check the logs for more info'
        }
      })
      .catch((e: Error) => {
        console.error('error when swapping', e)
        swapError = e.message
      })
      .finally(() => {
        webSocket.close()
        isLoadingSwap = false
      })
  }

  const onReset = (resetOffer = true) => {
    resetOffer && selectedOffer.set(undefined)
    amountProvided = 0
    willReceive = 0
    isSuccess = false
    swapError = ''
  }
</script>

{#if !!$selectedOffer}
  <Dialog
    open={true}
    on:SMUIDialog:action={() => console.log('action')}
    on:SMUIDialog:closed={() => onReset(true)}
    aria-labelledby="mandatory-title"
    aria-describedby="mandatory-content"
  >
    <div>
      <Title class="title" id="mandatory-title">
        Swap offer {$selectedOffer.id}
      </Title>
    </div>
    <Content id="mandatory-content">
      <section class="container">
        {#if isLoadingSwap}
          <div class="flexBox">
            <CircularProgress
              style="height: 48px; width: 48px;"
              indeterminate
            />
            <p>Swapping, please be patient...</p>
          </div>
        {:else if isSuccess}
          <div class="flexBox">
            <span class="material-icons circleCheck">check_circle</span>
            <p class="successMessage">
              Yay, you received {willReceive}{$selectedOffer.provides}
            </p>
          </div>
        {:else if !!swapError}
          <div class="flexBox">
            <span class="material-icons circleCross">error_outline</span>
            <p class="errorMessage">
              {swapError}
            </p>
          </div>
        {:else}
          <Textfield
            bind:value={amountProvided}
            variant="outlined"
            label={`${getCorrespondingToken($selectedOffer.provides)} amount`}
            invalid={!!error}
            suffix={getCorrespondingToken($selectedOffer.provides)}
          >
            <HelperText slot="helper">{error}</HelperText>
          </Textfield>
          <Textfield
            bind:value={xmrAddress}
            variant="outlined"
            label={'XMR address'}
          />
          <Icon class="swapIcon" component={Svg} viewBox="0 0 24 24">
            <path fill="currentColor" d={mdiSwapVertical} />
          </Icon>
          <div class="receivingAmount">
            {willReceive}
            {$selectedOffer.provides}
          </div>
        {/if}
      </section>
    </Content>
    {#if isSuccess}
      <Actions>
        <Button>
          <Label>Done</Label>
        </Button>
      </Actions>
    {:else if !!swapError}
      <Button on:click={() => onReset(false)}>
        <Label>Back</Label>
      </Button>
    {:else}
      <Button
        on:click={handleSendTakeOffer}
        disabled={isLoadingSwap || !!error || !willReceive}
      >
        <Label>Swap</Label>
      </Button>
    {/if}
  </Dialog>
{/if}

<style>
  .container {
    margin: 1em;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .receivingAmount {
    font-size: x-large;
  }

  .flexBox {
    display: flex;
    justify-content: center;
    align-items: center;
    flex: 1;
    flex-direction: column;
  }

  .circleCheck {
    font-size: 45px;
    color: darkcyan;
  }

  .circleCross {
    font-size: 45px;
    color: var(--mdc-theme-error, #b00020);
  }

  * :global(.swapIcon) {
    margin-top: 1rem;
    margin-bottom: 1rem;
    height: 3rem;
  }

  * :global(.title) {
    text-overflow: ellipsis;
    width: 100%;
    overflow: hidden;
    word-break: break-all;
    white-space: nowrap;
  }
</style>
