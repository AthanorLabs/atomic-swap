<script lang="ts">
  import unknown from '../assets/coins/unknown.svg'
  
  let url = unknown
  export let size: number = 32
  export let ticker: string = null

  if (ticker) {
    url = new URL(`../assets/coins/${ticker.toLowerCase()}@2x.png`, import.meta.url).href
  }
</script>

<img class="token-icon" width={size} height={size} src={url} alt="token-icon" onerror={ `this.src='${unknown}'`}/>

<style>
.token-icon {
  display: inline-block;
  vertical-align: middle;
}
</style>
