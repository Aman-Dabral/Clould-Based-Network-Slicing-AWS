
<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';
  import Notification from './Notification.svelte';
  import { GetInitialSettings, SubmitSettings } from '../lib/wailsjs/go/main/App';
  import "../app.css";

  // function submitSettingsThroughGo(){
  //   SubmitSettings().then((res) => {
  //     console.log(res);
  //   });
  // }

  function getInitialSettings(){
    GetInitialSettings().then((res) => {
      console.log(res);
      settings.set(JSON.parse(res));
    });
  }

  onMount(() => {
    getInitialSettings();
  });

  const settings = writable({
    GBR: true,
    ConnTo5G: false,
    IoT: false,
    AVRGaming: false,
    Healthcare: false,
    Industry40: false,
    IoTDevices: false,
    PublicSafety: false,
    SmartCityHome: false,
    SmartTransport: false,
    Smartphone: true,
    LTECategory: 1,
    minutesBeforeCloud: 1
  });

  const exclusiveKeys = [
    'IoT', 'AVRGaming', 'Healthcare', 'Industry40',
    'IoTDevices', 'PublicSafety', 'SmartCityHome',
    'SmartTransport', 'Smartphone'
  ];

  const categories = [
    { value: 1, label: 'Category 1: IoT for smart meters, asset tracking, and low-power industrial IoT devices.' },
    { value: 2, label: 'Category 2: Entry-level broadband and IoT for wearable devices and smart appliances.' },
    { value: 3, label: 'Category 3: Consumer broadband and IoT for connected vehicles and home automation.' },
    { value: 4, label: 'Category 4: High-speed broadband and IoT for video streaming and mobile hotspots.' },
    { value: 5, label: 'Category 5: Enhanced broadband and IoT for high-definition video conferencing.' },
    { value: 6, label: 'Category 6: Ultra-broadband and IoT for 4K video streaming and virtual reality.' },
    { value: 7, label: 'Category 7: Advanced broadband and IoT for immersive augmented reality experiences.' },
    { value: 8, label: 'Category 8: Gigabit broadband and IoT for enterprise applications and cloud services.' },
    { value: 9, label: 'Category 9: Ultra-reliable broadband and IoT for critical infrastructure and industrial automation.' },
    { value: 10, label: 'Category 10: High-performance broadband and IoT for autonomous vehicles and smart cities.' },
    { value: 11, label: 'Category 11: Enhanced mobile broadband and IoT for real-time data analytics and edge computing.' },
    { value: 12, label: 'Category 12: Ultra-low latency broadband and IoT for mission-critical communications.' },
    { value: 13, label: 'Category 13: High-capacity broadband and IoT for large-scale industrial IoT deployments.' },
    { value: 14, label: 'Category 14: Advanced broadband and IoT for smart agriculture and environmental monitoring.' },
    { value: 15, label: 'Category 15: Enhanced broadband and IoT for healthcare applications and remote patient monitoring.' },
    { value: 16, label: 'Category 16: Ultra-broadband and IoT for smart transportation and logistics.' },
    { value: 17, label: 'Category 17: High-speed broadband and IoT for entertainment and media services.' },
    { value: 18, label: 'Category 18: Advanced broadband and IoT for education and e-learning platforms.' },
    { value: 19, label: 'Category 19: Enhanced broadband and IoT for financial services and mobile banking.' },
    { value: 20, label: 'Category 20: Ultra-low latency broadband and IoT for gaming and virtual reality.' },
    { value: 21, label: 'Category 21: High-performance broadband and IoT for smart homes and building automation.' },
    { value: 22, label: 'Category 22: Advanced broadband and IoT for public safety and emergency services.' },
  ];

  const showNotification = writable(false);
  const notificationMessage = writable('');
  const notificationType = writable('');
  const isLoading = writable(false);

  function toggleExclusive(key) {
    settings.update(current => {
      const updated = { ...current };
      if (updated[key]) return updated;
      exclusiveKeys.forEach(k => {
        updated[k] = (k === key);
      });
      return updated;
    });
  }

  onMount(() => {
    try {

      // First create GetInitialSettings
      const res = JSON.parse(GetInitialSettings());
      const initialSettings = JSON.parse(res);

      const trueKeys = exclusiveKeys.filter(k => initialSettings[k]);
      if (trueKeys.length !== 1) {
        exclusiveKeys.forEach((k, i) => {
          initialSettings[k] = (i === 0);
        });
      }

      settings.set(initialSettings);
    } catch (error) {
      console.error('Failed to fetch initial settings:', error);
    }
  });

  const submitSettings = () => {
    isLoading.set(true); // Show loader
    try {
      // First create SubmitSettings
      const res = SubmitSettings(JSON.stringify($settings));

      if (res) {
        notificationMessage.set('Settings updated successfully!');
        notificationType.set('success');
      } else {
        notificationMessage.set('Failed to update settings.');
        notificationType.set('error');
      }
    } catch (error) {
      console.error('Error submitting settings:', error);
      notificationMessage.set('An error occurred.');
      notificationType.set('error');
    } finally {
      showNotification.set(true);
      isLoading.set(false); // Hide loader
    }
  };
</script>

<div class="p-4">
  <h1 class="text-2xl font-bold mb-4">Device Settings</h1>

  <form class="space-y-2">
    {#each Object.keys($settings).filter(key => key !== 'LTECategory' && key !== 'minutesBeforeCloud') as key}
      <div class="flex items-center">
        {#if exclusiveKeys.includes(key)}
          <input
            type="checkbox"
            checked={$settings[key]}
            on:change={() => toggleExclusive(key)}
            id={key}
            class="mr-2"
          />
        {:else}
          <input
            type="checkbox"
            bind:checked={$settings[key]}
            id={key}
            class="mr-2"
          />
        {/if}
        <label for={key} class="capitalize">{key}</label>
      </div>
    {/each}

    <div>
      <label for="LTECategory" class="block">LTE/5G Category</label>
      <select id="LTECategory" bind:value={$settings.LTECategory} class="w-full p-2 border">
        {#each categories as { value, label }}
          <option value={value}>{label}</option>
        {/each}
      </select>
    </div>

    <div>
      <label for="minutesBeforeCloud" class="block">Time Between Updates (in minutes)</label>
      <input
        type="number"
        id="minutesBeforeCloud"
        bind:value={$settings.minutesBeforeCloud}
        min="1"
        class="w-full p-2 border"
      />
    </div>

    <button
      type="button"
      on:click={submitSettings}
      class="w-full p-2 bg-blue-500 text-white rounded transition transform active:scale-95 hover:bg-blue-600"
    >
      Save
    </button>
  </form>

  {#if $showNotification}
    <Notification
      notificationMessage={$notificationMessage}
      notificationType={$notificationType}
      close={() => showNotification.set(false)}
    />
  {/if}

  {#if $isLoading}
    <div class="fixed inset-0 bg-black bg-opacity-30 flex justify-center items-center z-50">
      <div class="loader ease-linear rounded-full border-8 border-t-8 border-gray-200 h-16 w-16"></div>
    </div>
  {/if}
</div>

<style>
  .loader {
    border-top-color: #3498db;
    animation: spin 1s infinite linear;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>