<script lang="ts">
  import { onMount } from "svelte";
  import { writable } from "svelte/store";
  import { getState, post } from "$lib/api";
  import AddToQueue from "$lib/components/AddToQueue.svelte";
  import type { Device } from "$lib/device";

  type QueueItem = {
    title: string;
  };

  type AddItem = {
    type: string;
    source: string;
  };

  type PlayerState = {
    playing: boolean;
    current: QueueItem | null;
    queue: QueueItem[];
  };

  const playerState = writable<PlayerState>({
    playing: false,
    current: null,
    queue: [],
  });

  let ws: WebSocket | null = null;
  let device = $state<Device>({
    id: "",
    baseUrl: "",
    label: "",
  });

  onMount(() => {
    const init = async () => {
      const s = await getState(device);
      playerState.set(s);

      const wsUrl = device.baseUrl
        .replace("http://", "ws://")
        .replace("https://", "wss://");

      ws = new WebSocket(`${wsUrl}/ws`);

      ws.onmessage = (ev) => {
        var m = JSON.parse(ev.data);
        switch (m.type) {
          case "state":
            console.log(m.data);
            playerState.set(m.data);
            break;

          default:
            break;
        }
      };
    };

    init();

    return () => {
      ws?.close();
    };
  });

  function play() {
    if (device) post(device, "/api/control/play");
  }
  function togglePause() {
    if (!$playerState) return;
    if ($playerState.playing) {
      post(device, "/api/control/pause");
    } else {
      post(device, "/api/control/resume");
    }
  }
  function prev() {
    if (device) post(device, "/api/control/prev");
  }
  function next() {
    if (device) post(device, "/api/control/next");
  }
</script>

<div class="device">
  {device?.label ?? "no device"}
</div>

<div class="now-playing">
  <div class="section-label">Now Playing</div>

  <div class="track">
    {$playerState.current?.title ?? "Nothing playing"}
  </div>

  <div class="status">
    {$playerState.playing ? "Playing" : "Paused"}
  </div>

  <div class="controls">
    <button class="btn" onclick={prev}>⏮</button>
    <button class="btn primary" onclick={play}>▶</button>
    <button class="btn" onclick={togglePause}>⏸</button>
    <button class="btn" onclick={next}>⏭</button>
  </div>
</div>

<div class="queue">
  <div class="section-label">Queue</div>

  {#each $playerState.queue as item}
    <div class="queue-item">{item.title}</div>
  {/each}
</div>

<AddToQueue {device} />

<style>
  :global(body) {
    margin: 0;
    background: #0b0f14;
    color: #e6edf3;
    font-family:
      system-ui,
      -apple-system,
      Segoe UI,
      Roboto,
      sans-serif;
  }

  .title {
    padding: 1rem 1.25rem;
    font-size: 1.4rem;
    font-weight: 600;
  }

  .device {
    padding: 0 1.25rem;
    color: #7c8a99;
    font-size: 0.9rem;
    margin-bottom: 1rem;
  }

  .now-playing {
    margin: 1rem;
    padding: 1rem;
    background: #111826;
    border: 1px solid #1f2a3a;
    border-radius: 12px;
  }

  .section-label {
    font-size: 0.75rem;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: #7c8a99;
    margin-bottom: 0.5rem;
  }

  .track {
    font-size: 1.1rem;
    font-weight: 500;
    margin-bottom: 0.3rem;
  }

  .status {
    font-size: 0.85rem;
    color: #7c8a99;
    margin-bottom: 1rem;
  }

  .controls {
    display: flex;
    gap: 0.5rem;
  }

  .btn {
    background: #1a2433;
    border: 1px solid #263447;
    color: #e6edf3;
    padding: 0.5rem 0.7rem;
    border-radius: 10px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .btn:hover {
    background: #223044;
    transform: translateY(-1px);
  }

  .btn:active {
    transform: translateY(0px);
  }

  .btn.primary {
    background: #2d6cdf;
    border-color: #2d6cdf;
  }

  .btn.primary:hover {
    background: #3a7cff;
  }

  .queue {
    margin: 1rem;
    padding: 1rem;
    background: #0f1622;
    border: 1px solid #1f2a3a;
    border-radius: 12px;
  }

  .queue-item {
    padding: 0.5rem;
    border-bottom: 1px solid #1a2433;
    color: #c9d4e2;
  }

  .queue-item:last-child {
    border-bottom: none;
  }
</style>
