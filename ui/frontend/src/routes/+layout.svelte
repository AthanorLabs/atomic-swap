<script>
  import "../app.postcss";
  import Sidebar from "../lib/Sidebar.svelte";
</script>
 
<main>
  <div style="display: flex; height:100vh;">
    <Sidebar></Sidebar>
    <div style="margin-left: 150px; width: 100%;">
      <slot />
    </div>
  </div>
</main>

<style>
</style>
