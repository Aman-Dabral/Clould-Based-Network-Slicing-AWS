<script>
  import Settings from './Settings.svelte';
  import Card from './Card.svelte';
  import { fade } from 'svelte/transition';

  let showSettings = false;
  let cards = [];
  let locked = false;

  // Methods to expose
  export function appendCard(content) {
    if (!locked) cards = [...cards, { id: Date.now(), content }];
  }

  window.appendCard = appendCard;

  export function InteruptingError(err) {
    locked = true;
    cards = [{ id: Date.now(), content: `<div style="color: red;"><strong>Error:</strong> ${err}</div>` }];
  }

  window.InteruptingError = InteruptingError;

  function toggleSettings() {
    showSettings = !showSettings;
  }
</script>

<nav class="navbar">
  <div class="nav-title">Client Dashboard</div>
  <div class="nav-actions">
    <button class="settings-btn" on:click={toggleSettings}>
      ⚙️
    </button>
  </div>
</nav>

<main class="content p-4">
  
    {#if showSettings}
    <div in:fade>
      <Settings />
    </div>
    {/if}
  <h1>Information Board</h1>
  {#each cards as card (card.id)}
    <Card content={card.content} />
  {/each}
</main>

<style>
  .navbar {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 2rem;
    z-index: 100;
  }

  .nav-title {
    font-weight: bold;
    font-size: 1.2rem;
  }

  .nav-actions {
    position: relative;
  }

  .settings-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
  }

  .content {
    padding-top: 80px;
    max-width: 800px;
    margin: 0 auto;
  }
</style>
