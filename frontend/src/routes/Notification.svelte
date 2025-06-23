
<script>
  export let notificationMessage = ''; // Now expected to be a plain string, not a store
  export let notificationType = 'success'; // 'success' or 'error'
  export let close = () => {};

  $: bgColor = notificationType === 'success' ? 'bg-green-500' : 'bg-red-500';
</script>

<div class={`fixed top-28 left-1/2 transform -translate-x-1/2 p-4 rounded ${bgColor} text-white shadow-lg z-50`}>
  <p>{notificationMessage}</p>
  <button
    on:click={close}
    class="absolute top-0 right-0 p-2 text-white"
    aria-label="Close"
  >
    ✕
  </button>
</div>