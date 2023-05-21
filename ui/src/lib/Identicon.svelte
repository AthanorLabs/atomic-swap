<script lang="ts">
    import Identicon from 'identicon.js'
  
    export let size: number = 30
    export let peerAddress: string = ''
    var options = {
      //foreground: [0, 0, 0, 255],               // rgba black
      background: [240, 240, 240, 255],         // rgba white
      margin: 0.2,                              // 20% margin
      size,                                // 420px square
      format: 'svg'                             // use SVG instead of PNG
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
  
