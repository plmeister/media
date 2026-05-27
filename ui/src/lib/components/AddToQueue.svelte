<script lang="ts">
  import { postContent } from "$lib/api";

  type SourceType = "youtube" | "jellyfin";

  let url = "";
  let type: SourceType = "youtube";
  let loading = false;

  async function add() {
    if (!url.trim()) return;

    loading = true;

    try {
      await postContent("/queue/add", {
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

  <button onclick={add} disabled={loading || !url.trim()}> Add </button>
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
