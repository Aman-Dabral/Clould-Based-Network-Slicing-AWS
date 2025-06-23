<script>
  export let content;
</script>

<div class="card" bind:this={cardElement}>
  {@html content}
</div>

<style>
  .card {
    background: white;
    border-radius: 8px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    padding: 1rem;
    margin: 1rem 0;
    transition: all 0.3s ease;
  }
</style>
