<script lang="ts">
  import Dialog, { Title, Content, Actions } from '@smui/dialog'
  import Button, { Label } from '@smui/button'
  import type { NetTakeOfferResult } from 'src/types/NetTakeOffer'
  import { getCorrespondingToken, rpcRequest } from 'src/utils'
  import { selectedOffer } from '../stores/offerStore'
  import Textfield from '@smui/textfield'
  import { mdiSwapVertical } from '@mdi/js'
  import { Icon } from '@smui/icon-button'
  import { Svg } from '@smui/common/elements'
  import CircularProgress from '@smui/circular-progress'
  import HelperText from '@smui/textfield/helper-text'

  let amountProvided: number | null = null
  let isSuccess = false
  let receivedAmount = 0
  let isLoadingSwap = false
  let error = ''

  $: console.log(
    'willReceive < $selectedOffer.minAmount',
    willReceive,
    $selectedOffer?.minAmount
  )
  $: willReceive =
    amountProvided && amountProvided > 0 && $selectedOffer?.exchangeRate
      ? amountProvided / $selectedOffer.exchangeRate
      : 0

  $: if (willReceive === 0) {
    error = ''
  }

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

  $: console.log('isSuccess', isSuccess, receivedAmount)
  $: console.log('$selectedOffer.maxAmount', $selectedOffer?.maxAmount)

  const handleSendTakeOffer = async () => {
    isLoadingSwap = true
    await rpcRequest<NetTakeOfferResult | undefined>('net_takeOffer', {
      multiaddr: $selectedOffer?.peer,
      offerID: $selectedOffer?.id,
      providesAmount: Number(amountProvided),
    })
      .then(({ result }) => {
        if (result?.success) {
          receivedAmount = result.receivedAmount
          isSuccess = true
        }
      })
      .catch(console.error)
      .finally(() => (isLoadingSwap = false))
  }

  const onReset = () => {
    selectedOffer.set(undefined)
    amountProvided = 0
    willReceive = 0
    receivedAmount = 0
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
          <div class="loader">
            <CircularProgress
              style="height: 48px; width: 48px;"
              indeterminate
            />
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
    <Actions>
      <Button
        on:click={handleSendTakeOffer}
        disabled={isLoadingSwap || !!error || !willReceive}
      >
        <Label>Swap</Label>
      </Button>
    </Actions>
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

  .loader {
    display: flex;
    justify-content: center;
    align-items: center;
    flex: 1;
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
