<script lang="ts">
  import unknown from '../assets/coins/unknown.svg'
  
  export let size: number = 32
  export let ticker: String = ''
  $: ticker, load()
  $: url = unknown

  async function load() {
    try {
      const path = await import(`../assets/coins/${ticker.toLowerCase()}@2x.png`)
      url = path.default
    } catch (e) {}
  }
</script>

<img class="token-icon" width={size} height={size} src={url} alt="token-icon" onerror={ `this.src='${unknown}'`}/>

<style>
.token-icon {
  display: inline-block;
  vertical-align: middle;
}
</style>
