<script lang="ts">
    import Identicon from 'identicon.js'
  
    export let size: number = 30
    export let peerAddress: string = ''
    var options = {
      background: [240, 240, 240, 255],
      margin: 0.2,
      size,
      format: 'svg'
    };
    const source = `data:image/svg+xml;base64,${new Identicon(peerAddress, options).toString()}`
</script>

<img width={size} height={size} src={source} alt="peer" />

<style>
img {
    border-radius: 50%;
    display: inline;
}
</style>
