<script lang="ts">
  import Paper, { Title, Subtitle, Content } from '@smui/paper'

  export let title = ''
  export let subtitle = ''
  export let content = ''
</script>

<div>
  <Paper class="root">
    {#if title}
      <Title>{title}</Title>
      <br />
    {/if}
    {#if subtitle}
      <Subtitle>{subtitle}</Subtitle>
      <br />
    {/if}
    {#if content}
      <Content>{content}</Content>
    {/if}
  </Paper>
</div>

<style>
  * :global(.root) {
    display: flex;
    align-items: center;
    flex-direction: column;
  }
</style>
