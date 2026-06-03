<script lang="ts">
  import { postContent } from "$lib/api";
  import type { Device } from "$lib/device";

  type SourceType = "youtube" | "file" | "jellyfin";

  let url = $state<string>("");
  let type = $state<SourceType>("youtube");
  let loading = $state(false);
  let { device }: { device: Device | null } = $props();

  async function add() {
    if (!device) return;
    if (!url.trim()) return;
    loading = true;
    try {
      await postContent(device, "/api/queue/add", {
        type,
        url,
      });
      url = "";
    } finally {
      loading = false;
    }
  }
</script>

<div class="queue-form">
  <select bind:value={type}>
    <option value="youtube">YouTube</option>
    <option value="file">File</option>
    <!--<option value="jellyfin">Jellyfin</option>-->
  </select>

  <input
    placeholder="Paste URL or ID"
    bind:value={url}
    onkeydown={(e) => e.key === "Enter" && add()}
  />

  <button onclick={add} disabled={!device || loading || !url.trim()}>
    {loading ? "Adding.." : "Add"}
  </button>
</div>

<style>
  .queue-form {
    display: flex;
    gap: 0.5rem;
  }

  input {
    flex: 1;
  }
</style>
