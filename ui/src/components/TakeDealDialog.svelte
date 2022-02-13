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

  let amountProvided: number | null = null
  let isSuccess = false
  let isLoadingSwap = false
  let error = ''

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

  $: console.log('isSuccess', isSuccess)

  const handleSendTakeOffer = () => {
    isLoadingSwap = true
    rpcRequest<NetTakeOfferSyncResult | undefined>('net_takeOfferSync', {
      multiaddr: $selectedOffer?.peer,
      offerID: $selectedOffer?.id,
      providesAmount: Number(amountProvided),
    })
      .then(({ result }) => {
        if (result?.status === 'success') {
          isSuccess = true
          getPeers()
        }
      })
      .catch(console.error)
      .finally(() => (isLoadingSwap = false))
  }

  const onReset = () => {
    selectedOffer.set(undefined)
    amountProvided = 0
    willReceive = 0
  }
</script>

{#if !!$selectedOffer}
  <Dialog
    open={true}
    on:SMUIDialog:action={() => console.log('action')}
    on:SMUIDialog:closed={onReset}
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
            <span class="material-icons circleCheck"> check_circle </span>
            <p class="successMessage">
              Yay, you received {willReceive}{getCorrespondingToken(
                $selectedOffer.provides
              )}
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
