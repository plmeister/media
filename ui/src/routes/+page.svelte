<script lang="ts">
  import { onMount } from "svelte";
  import { writable } from "svelte/store";
  import { getState, post } from "$lib/api";

  type QueueItem = {
    title: string;
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

  onMount(() => {
    const init = async () => {
      const s = await getState();
      playerState.set(s);
    };

    init();

    const ws = new WebSocket("ws://localhost:8080/ws");

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

    return () => ws.close();
  });
</script>

<h1>Media Server</h1>

<h2>Now Playing</h2>

<div>
  {$playerState.current?.title ?? "Nothing playing"}
</div>

<div>
  {JSON.stringify($playerState)}
  {$playerState.playing ? "playing" : "not playing"}
</div>

<button onclick={() => post("/control/play")}> Play </button>

<button onclick={() => post("/control/pause")}> Pause </button>

<button onclick={() => post("/control/resume")}> Resume </button>
<button onclick={() => post("/control/next")}> Next </button>

<h2>Queue</h2>

<ul>
  {#each $playerState.queue as item}
    <li>{item.title}</li>
  {/each}
</ul>
