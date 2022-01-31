<script lang="ts">
  import Identicon from 'identicon.js'

  export let size: number = 50
  export let peerAddress: string = ''
  const source = `data:image/png;base64,${new Identicon(peerAddress, {
    background: [255, 255, 255, 255],
    size,
  }).toString()}`
</script>

<img width={size} height={size} src={source} alt="peer" />
